package net

import (
	"context"
	"fmt"
	"net"

	"zombiezen.com/go/capnproto2/rpc"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/backend"
	"github.com/disorganizer/brig/net/capnp"
	"github.com/disorganizer/brig/net/peer"
	"github.com/disorganizer/brig/repo"
	"github.com/disorganizer/brig/util/server"
)

type Server struct {
	bk         backend.Backend
	baseServer *server.Server
	hdl        *handler
}

func (sv *Server) Serve() error {
	return sv.baseServer.Serve()
}

func (sv *Server) Close() error {
	return sv.baseServer.Close()
}

func NewServer(rp *repo.Repository, bk backend.Backend) (*Server, error) {
	hdl := &handler{
		rp: rp,
		bk: bk,
	}

	lst, err := bk.Listen("brig/caprpc")
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	baseServer, err := server.NewServer(lst, hdl, ctx)
	if err != nil {
		return nil, err
	}

	return &Server{
		baseServer: baseServer,
		bk:         bk,
		hdl:        hdl,
	}, nil
}

func (sv *Server) Locate(who peer.Name) ([]peer.Info, error) {
	return sv.bk.ResolveName(who)
}

func (sv *Server) Identity() (peer.Info, error) {
	return sv.bk.Identity()
}

/////////////////////////////////////
// INTERNAL HANDLER IMPLEMENTATION //
/////////////////////////////////////

type handler struct {
	bk backend.Backend
	rp *repo.Repository
}

func (hdl *handler) Handle(ctx context.Context, conn net.Conn) {
	keyring := hdl.rp.Keyring()
	ownPubKey, err := keyring.OwnPubKey()
	if err != nil {
		log.Warnf("Failed to retrieve own pubkey: %v", err)
		return
	}

	authConn := NewAuthReadWriter(conn, keyring, ownPubKey, func(pubKey []byte) error {
		remotes, err := hdl.rp.Remotes.ListRemotes()
		if err != nil {
			return err
		}

		// Create a temporary fingerprint to get a hashed version of pubkey.
		remoteFp := peer.BuildFingerprint("", pubKey)

		// Linear scan over all remotes.
		// If this proves to be a performance problem, we can fix it later.
		for _, remote := range remotes {
			if remote.Fingerprint.PubKeyID() == remoteFp.PubKeyID() {
				log.Infof("Starting connection with %s", remote.Fingerprint.Addr())
				return nil
			}
		}

		return fmt.Errorf("Remote uses no public key known to us")
	})

	if err := authConn.Trigger(); err != nil {
		log.Warnf("Failed to authenticate connection: %v", err)
		return
	}

	transport := rpc.StreamTransport(conn)
	srv := capnp.API_ServerToClient(hdl)
	rpcConn := rpc.NewConn(transport, rpc.MainInterface(srv.Client))

	if err := rpcConn.Wait(); err != nil {
		log.Warnf("Serving rpc failed: %v", err)
	}

	if err := rpcConn.Close(); err != nil {
		log.Warnf("Failed to close rpc conn: %v", err)
	}
}

// Quit is being called by the base server implementation
func (hdl *handler) Quit() error {
	return nil
}
