package net

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"zombiezen.com/go/capnproto2/rpc"

	log "github.com/Sirupsen/logrus"
	"github.com/sahib/brig/backend"
	"github.com/sahib/brig/net/capnp"
	"github.com/sahib/brig/net/peer"
	"github.com/sahib/brig/repo"
	"github.com/sahib/brig/util/server"
)

type Server struct {
	bk         backend.Backend
	baseServer *server.Server
	hdl        *handler
	pingMap    *PingMap
}

func (sv *Server) Serve() error {
	return sv.baseServer.Serve()
}

func (sv *Server) Close() error {
	return sv.baseServer.Close()
}

func (sv *Server) Quit() {
	sv.baseServer.Quit()
}

func publishSelf(bk backend.Backend, owner string) error {
	// Example: alice@wonderland.org/resource
	name := peer.Name(owner)

	// Publish the full name.
	if err := bk.PublishName(owner); err != nil {
		return err
	}

	// Also publish alice@wonderland.org
	if noRes := name.WithoutResource(); noRes != string(name) {
		if err := bk.PublishName(noRes); err != nil {
			return err
		}
	}

	// Publish wonderland.org
	if domain := name.Domain(); domain != "" {
		if err := bk.PublishName(domain); err != nil {
			return err
		}
	}

	if user := name.User(); user != string(name) {
		if err := bk.PublishName(user); err != nil {
			return err
		}
	}

	return nil
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

	if err := publishSelf(bk, rp.Owner); err != nil {
		log.Warningf("Failed to publish `%v` to the network: %v", rp.Owner, err)
		log.Warningf("You will not be visible to other users.")
	}

	return &Server{
		baseServer: baseServer,
		bk:         bk,
		hdl:        hdl,
		pingMap:    NewPingMap(bk),
	}, nil
}

const (
	LocateNone  = 0
	LocateExact = 1 << iota
	LocateDomain
	LocateUser
	LocateEmail
	LocateAll = LocateExact | LocateDomain | LocateUser | LocateEmail
)

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
		default:
			return mask, fmt.Errorf("Invalid locate mask name `%s`", part)
		}
	}

	return mask, nil
}

func (sv *Server) Locate(who peer.Name, timeoutSec int, mask LocateMask) (map[LocateMask][]peer.Info, error) {
	uniqueNames := make(map[string]LocateMask)

	// Example: donald@whitehouse.gov/ovaloffice
	uniqueNames[string(who)] = mask & LocateExact

	// Example: whitehouse.gov
	uniqueNames[who.Domain()] = mask & LocateDomain

	// Example: donald
	uniqueNames[who.User()] = mask & LocateUser

	// Example: donald@whitehouse.gov
	uniqueNames[who.WithoutResource()] = mask & LocateEmail

	resultMu := &sync.Mutex{}
	results := make(map[LocateMask][]peer.Info)
	errors := make(map[LocateMask]error)

	wg := &sync.WaitGroup{}
	for name, mask := range uniqueNames {
		// It's not enabled:
		if mask == 0 {
			continue
		}

		wg.Add(1)

		go func(name string, mask LocateMask) {
			defer wg.Done()

			peers, err := sv.bk.ResolveName(name, timeoutSec)
			if err != nil {
				resultMu.Lock()
				errors[mask] = err
				resultMu.Unlock()
				return
			}

			// Collect results:
			resultMu.Lock()
			for _, peer := range peers {
				results[mask] = append(results[mask], peer)
			}
			resultMu.Unlock()
		}(name, mask)
	}

	wg.Wait()

	if len(results) == 0 && len(errors) == 0 {
		// Nothing found.
		return nil, nil
	}

	if len(results) != 0 && len(errors) == 0 {
		// Found something.
		return results, nil
	}

	if len(results) == 0 && len(errors) != 0 {
		// Several errors happened.
		return nil, fmt.Errorf("Several errors: %v", errors)
	}

	// Silence errors if we have some results:
	log.Debugf("locate had errors, but still got results. Errors: %v", errors)
	return results, nil
}

// PeekFingerprint fetches the fingerprint of a peer without authenticating
// ourselves or them.
func (sv *Server) PeekFingerprint(ctx context.Context, addr string) (peer.Fingerprint, error) {
	// Query the remotes pubkey and use it to build the remotes' fingerprint.
	// If not available we just send an empty string back to the client.
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	// Dial peer without authentication:
	emptyFp := peer.Fingerprint("")
	ctl, err := DialByAddr(addr, emptyFp, sv.hdl.rp.Keyring(), sv.bk, ctx)
	if err != nil {
		log.Warningf(
			"locate: failed to dial to `%s` (%s): %v",
			addr, addr, err,
		)
		return peer.Fingerprint(""), nil
	}

	// Quickly check if the other side is online:
	if err := ctl.Ping(); err != nil {
		return peer.Fingerprint(""), err
	}

	// Fetch their remote pub key to build the fingerprint:
	remotePubKey, err := ctl.RemotePubKey()
	if err != nil {
		return peer.Fingerprint(""), err
	}

	return peer.BuildFingerprint(addr, remotePubKey), nil
}

func (sv *Server) Identity() (peer.Info, error) {
	return sv.bk.Identity()
}

func (sv *Server) PingMap() *PingMap {
	return sv.pingMap
}

func (sv *Server) IsOnline() bool {
	return sv.bk.IsOnline()
}

func (sv *Server) Connect() error {
	return sv.bk.Connect()
}

func (sv *Server) Disconnect() error {
	return sv.bk.Disconnect()
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

	// Take the raw connection we get and add an authentication layer on top of it.
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

	// Trigger the authentication.
	// (would trigger with the first read/writer elsewhise)
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
		// Close seems to be complaining that the conn was
		// already closed, but be safe and expect this.
		if err != rpc.ErrConnClosed {
			log.Warnf("Failed to close rpc conn: %v", err)
		}
	}
}

// Quit is being called by the base server implementation
func (hdl *handler) Quit() error {
	return nil
}
