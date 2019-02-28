package net

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"

	"zombiezen.com/go/capnproto2/rpc"

	e "github.com/pkg/errors"
	"github.com/sahib/brig/backend"
	"github.com/sahib/brig/net/capnp"
	"github.com/sahib/brig/net/peer"
	"github.com/sahib/brig/repo"
	"github.com/sahib/brig/util/server"
	log "github.com/sirupsen/logrus"
)

// Server implements the server for inter-remote communication.
type Server struct {
	bk         backend.Backend
	baseServer *server.Server
	hdl        *connHandler
	pingMap    *PingMap
}

// Serve blocks and serves request until quit was called.
func (sv *Server) Serve() error {
	return e.Wrapf(sv.baseServer.Serve(), "serve")
}

// Close will clean up resources.
func (sv *Server) Close() error {
	return sv.baseServer.Close()
}

// Quit will shut down the server and unblock Serve()
func (sv *Server) Quit() {
	sv.baseServer.Quit()
}

func publishSelf(bk backend.Backend, owner string) error {
	// Example: alice@wonderland.org/resource
	name := peer.Name(owner)

	// Publish the full name »alice@wonderland.org/resource«
	if err := bk.PublishName(owner); err != nil {
		return err
	}

	// Also publish »alice@wonderland.org«
	if noRes := name.WithoutResource(); noRes != string(name) {
		if err := bk.PublishName(noRes); err != nil {
			return err
		}
	}

	// Publish »wonderland.org«
	if domain := name.Domain(); domain != "" {
		if err := bk.PublishName(domain); err != nil {
			return err
		}
	}

	// Publish »alice«
	if user := name.User(); user != string(name) {
		if err := bk.PublishName(user); err != nil {
			return err
		}
	}

	return nil
}

// NewServer returns a new inter-remote server.
func NewServer(rp *repo.Repository, bk backend.Backend) (*Server, error) {
	hdl := &connHandler{
		rp: rp,
		bk: bk,
	}

	lst, err := bk.Listen("brig/caprpc")
	if err != nil {
		return nil, e.Wrapf(err, "listen")
	}

	ctx := context.Background()
	baseServer, err := server.NewServer(ctx, lst, hdl)
	if err != nil {
		return nil, e.Wrapf(err, "new-server")
	}

	log.Debugf("publishing own identity to network: %s", rp.Owner)
	if err := publishSelf(bk, rp.Owner); err != nil {
		log.Warningf("failed to publish `%v` to the network: %v", rp.Owner, err)
		log.Warningf("you will not be visible to other users.")
	}

	return &Server{
		baseServer: baseServer,
		bk:         bk,
		hdl:        hdl,
		pingMap:    NewPingMap(bk),
	}, nil
}

const (
	// LocateNone is used when no part of the name should be searched.
	LocateNone = 0
	// LocateExact means that we should search for the name exactly.
	LocateExact = 1 << iota
	// LocateDomain means that we only search for the domain name only.
	LocateDomain
	// LocateUser means that we only search for the user name only.
	LocateUser
	// LocateEmail means that we only search for the user@domain part only.
	LocateEmail
	// LocateAll means that we search for everything.
	LocateAll = LocateExact | LocateDomain | LocateUser | LocateEmail
)

// LocateMask is a combination of the individual LocateXXX settings
// and tells Locate() what parts of the name to search for.
type LocateMask int

func (lm LocateMask) String() string {
	if lm == LocateNone {
		return ""
	}

	parts := []string{}
	if lm&LocateExact != 0 {
		parts = append(parts, "exact")
	}
	if lm&LocateDomain != 0 {
		parts = append(parts, "domain")
	}
	if lm&LocateUser != 0 {
		parts = append(parts, "user")
	}
	if lm&LocateEmail != 0 {
		parts = append(parts, "email")
	}

	return strings.Join(parts, ",")
}

// LocateMaskFromString builds a LocateMask from a comma separated string.
// This is the inverse of mask.String().
func LocateMaskFromString(s string) (LocateMask, error) {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return LocateNone, nil
	}

	mask := LocateMask(LocateNone)
	parts := strings.Split(s, ",")
	for _, part := range parts {
		switch part {
		case "exact":
			mask |= LocateExact
		case "domain":
			mask |= LocateDomain
		case "user":
			mask |= LocateUser
		case "email":
			mask |= LocateEmail
		case "all":
			mask |= LocateAll
		default:
			return mask, fmt.Errorf("Invalid locate mask name `%s`", part)
		}
	}

	return mask, nil
}

// LocateResult is one result returned by Locate's result channel.
type LocateResult struct {
	Peers []peer.Info
	Mask  LocateMask
	Name  string
	Err   error
}

// Locate tries to find other remotes named `who`.
// It also tries to find different variations/parts of `who`, defined by `mask`.
// It does not block, but returns a channel where the results are being pushed to.
// This is a very slow operation.
func (sv *Server) Locate(ctx context.Context, who peer.Name, mask LocateMask) chan LocateResult {
	uniqueNames := make(map[string]LocateMask)

	// Example: donald@whitehouse.gov/ovaloffice
	uniqueNames[string(who)] = mask & LocateExact

	// Example: whitehouse.gov
	uniqueNames[who.Domain()] = mask & LocateDomain

	// Example: donald
	uniqueNames[who.User()] = mask & LocateUser

	// Example: donald@whitehouse.gov
	uniqueNames[who.WithoutResource()] = mask & LocateEmail

	resultCh := make(chan LocateResult)

	wg := &sync.WaitGroup{}
	for name, mask := range uniqueNames {
		if name == "" {
			continue
		}

		// It's not enabled:
		if mask == 0 {
			continue
		}

		wg.Add(1)

		go func(name string, mask LocateMask) {
			defer wg.Done()

			peers, err := sv.bk.ResolveName(ctx, name)
			log.Debugf("Found peers: %v", peers)
			resultCh <- LocateResult{
				Peers: peers,
				Err:   err,
				Name:  name,
				Mask:  mask,
			}
		}(name, mask)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	return resultCh
}

// PeekFingerprint fetches the fingerprint of a peer without authenticating
// ourselves or them.
func (sv *Server) PeekFingerprint(ctx context.Context, addr string) (peer.Fingerprint, string, error) {
	// Query the remotes pubkey and use it to build the remotes' fingerprint.
	// If not available we just send an empty string back to the client.
	pubKey, remoteName, err := PeekRemotePubkey(ctx, addr, sv.hdl.rp, sv.bk)
	if err != nil {
		log.Warningf(
			"locate: failed to dial to `%s` (%s): %v",
			addr, addr, err,
		)
		return peer.Fingerprint(""), "", nil
	}

	return peer.BuildFingerprint(addr, pubKey), remoteName, nil
}

// Identity returns the backend's Identity (i.e. addr)
func (sv *Server) Identity() (peer.Info, error) {
	return sv.bk.Identity()
}

// PingMap returns the ping map associated with this server.
func (sv *Server) PingMap() *PingMap {
	return sv.pingMap
}

// IsOnline returns true if we are online.
func (sv *Server) IsOnline() bool {
	return sv.bk.IsOnline()
}

// Connect will connect to the network (this is the default already)
func (sv *Server) Connect() error {
	return sv.bk.Connect()
}

// Disconnect will stop network operations immediately.
func (sv *Server) Disconnect() error {
	return sv.bk.Disconnect()
}

/////////////////////////////////////
// INTERNAL HANDLER IMPLEMENTATION //
/////////////////////////////////////

type connHandler struct {
	bk backend.Backend
	rp *repo.Repository
}

// Handle is called whenever we receive a new connection from another brig peer.
func (hdl *connHandler) Handle(ctx context.Context, conn net.Conn) {
	// We are currently not allowing more than one parallel connection.
	// This is not a technical problem, but more due to the fact that it makes
	// it easier to pass the current remote to the active handler.
	// Make sure to reset the current remote:
	keyring := hdl.rp.Keyring()
	ownPubKey, err := keyring.OwnPubKey()
	if err != nil {
		log.Warnf("Failed to retrieve own pubkey: %v", err)
		return
	}

	ownFingerprint := peer.BuildFingerprint("", ownPubKey)

	// The respective handler should get its own context it can listen to.
	reqCtx, reqCancel := context.WithCancel(ctx)
	reqHdl := &requestHandler{
		bk:  hdl.bk,
		rp:  hdl.rp,
		ctx: reqCtx,
	}

	// This func will be called during the authentication process.
	// It checks if the pub key the other side send us can be
	// related to one of the allowed remotes. If not, the connection
	// will be dropped.
	authChecker := func(pubKey []byte) error {
		remotes, err := hdl.rp.Remotes.ListRemotes()
		if err != nil {
			return err
		}

		// Create a temporary fingerprint to get a hashed version of pubkey.
		remoteFp := peer.BuildFingerprint("", pubKey)
		if remoteFp == ownFingerprint {
			return fmt.Errorf("cannot dial self")
		}

		// Linear scan over all remotes.
		// If this proves to be a performance problem, we can fix it later.
		for _, remote := range remotes {
			if remote.Fingerprint.PubKeyID() == remoteFp.PubKeyID() {
				log.Infof("starting connection with addr `%s`", remote.Fingerprint.Addr())
				reqHdl.currRemoteName = remote.Name
				return nil
			}
		}

		return fmt.Errorf("remote uses no public key known to us")
	}

	// Take the raw connection we get and add an authentication layer on top of it.
	authConn := NewAuthReadWriter(conn, keyring, ownPubKey, hdl.rp.Owner, authChecker)

	// Trigger the authentication. This is not strictly necessary and would
	// happen anyways on the first read/write on the connection. But doing it
	// here catches errors early.
	if err := authConn.Trigger(); err != nil {
		log.Warnf("failed to authenticate connection: %v", err)
		reqCancel()
		return
	}

	// The connection is considered authenticated at this point.
	// Initialize the capnp rpc protocol over it.
	transport := rpc.StreamTransport(conn)
	srv := capnp.API_ServerToClient(reqHdl)
	rpcConn := rpc.NewConn(
		transport,
		rpc.MainInterface(srv.Client),
		rpc.ConnLog(nil),
	)

	// Wait until either side quits the connection in the background.
	// The number of open connections is limited by the base server.
	go func() {
		defer reqCancel()

		if err := rpcConn.Wait(); err != nil {
			log.Warnf("serving rpc failed: %v", err)
		}

		if err := rpcConn.Close(); err != nil {
			// Close seems to be complaining that the conn was
			// already closed, but be safe and expect this.
			if err != rpc.ErrConnClosed {
				log.Warnf("failed to close rpc conn: %v", err)
			}
		}
	}()
}

// Quit is being called by the base server implementation
func (hdl *connHandler) Quit() error {
	return nil
}
