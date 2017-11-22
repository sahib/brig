package net

import (
	"bytes"
	"context"
	"fmt"
	"net"

	log "github.com/Sirupsen/logrus"
	"zombiezen.com/go/capnproto2/rpc"

	"github.com/disorganizer/brig/backend"
	"github.com/disorganizer/brig/net/capnp"
	"github.com/disorganizer/brig/net/peer"
	"github.com/disorganizer/brig/repo"
)

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

func (hdl *handler) GetStore(call capnp.Sync_getStore) error {
	fs, err := hdl.rp.OwnFS(hdl.bk)
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	if err := fs.Export(buf); err != nil {
		return err
	}

	return call.Results.SetData(buf.Bytes())
}

func (hdl *handler) Ping(call capnp.Meta_ping) error {
	return call.Results.SetReply("ALIVE")
}

func (hdl *handler) Version(call capnp.API_version) error {
	call.Results.SetVersion(1)
	return nil
}
