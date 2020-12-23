package net

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"

	e "github.com/pkg/errors"
	netBackend "github.com/sahib/brig/net/backend"
	"github.com/sahib/brig/net/capnp"
	"github.com/sahib/brig/net/peer"
	"github.com/sahib/brig/repo"
	log "github.com/sirupsen/logrus"
	"zombiezen.com/go/capnproto2/rpc"
)

// Client is a client for inter-remote communication.
// It implements convenient methods to talk to other brig instances.
type Client struct {
	bk netBackend.Backend

	ctx      context.Context
	conn     *rpc.Conn
	rawConn  net.Conn
	authConn *AuthReadWriter
	api      capnp.API
}

// Dial creates a new Client connected to `name`.
func Dial(ctx context.Context, name string, rp *repo.Repository, bk netBackend.Backend, pingMap *PingMap) (*Client, error) {
	remote, err := rp.Remotes.Remote(name)
	if err != nil {
		return nil, err
	}

	addr := remote.Fingerprint.Addr()
	ctl, err := DialByAddr(ctx, addr, remote.Fingerprint, rp, bk, pingMap)
	if err != nil {
		return nil, e.Wrapf(err, "by-addr")
	}

	return ctl, nil
}

// DialByAddr is like Dial but does not get its info from the remote list.
func DialByAddr(
	ctx context.Context,
	addr string,
	fingerprint peer.Fingerprint,
	rp *repo.Repository,
	bk netBackend.Backend,
	pingMap *PingMap,
) (*Client, error) {
	kr := rp.Keyring()
	ownPubKey, err := kr.OwnPubKey()
	if err != nil {
		return nil, err
	}

	// Low level by addr, not by brig's remote name:
	log.Debugf("raw dial to %s:%s", addr, fingerprint.PubKeyID())
	rawConn, err := bk.Dial(addr, fingerprint.PubKeyID(), "brig/caprpc")
	if err != nil {
		pingMap.hintNetAttempt(addr, false)
		return nil, e.Wrapf(err, "raw")
	}

	ownName := rp.Owner
	if fingerprint == "" {
		return nil, fmt.Errorf("rejecting own, empty fingerprint... bug?")
	}

	authConn := NewAuthReadWriter(rawConn, kr, ownPubKey, ownName, func(pubKey []byte) error {
		if !fingerprint.PubKeyMatches(pubKey) {
			pingMap.hintNetAttempt(addr, false)
			return fmt.Errorf("remote pubkey does not match fingerprint")
		}

		return nil
	})

	// Trigger the authentication:
	// (otherwise it would be triggered on the first read/write)
	if err := authConn.Trigger(); err != nil {
		pingMap.hintNetAttempt(addr, false)
		return nil, e.Wrapf(err, "auth")
	}

	pingMap.hintNetAttempt(addr, true)

	// Setup capnp-rpc:
	transport := rpc.StreamTransport(rawConn)
	clientConn := rpc.NewConn(transport, rpc.ConnLog(nil))
	api := capnp.API{Client: clientConn.Bootstrap(ctx)}

	return &Client{
		ctx:      ctx,
		authConn: authConn,
		conn:     clientConn,
		rawConn:  rawConn,
		api:      api,
	}, nil
}

// PeekRemotePubkey connects to `addr` and tries to read the public key they claim.
func PeekRemotePubkey(
	ctx context.Context,
	addr string,
	rp *repo.Repository,
	bk netBackend.Backend,
) ([]byte, string, error) {
	kr := rp.Keyring()
	ownPubKey, err := kr.OwnPubKey()
	if err != nil {
		return nil, "", err
	}

	log.Debugf("peek to %s", addr)
	rawConn, err := bk.Dial(addr, "", "brig/caprpc")
	if err != nil {
		return nil, "", e.Wrapf(err, "raw")
	}

	owner := rp.Owner
	authConn := NewAuthReadWriter(rawConn, kr, ownPubKey, owner, func(_ []byte) error {
		return nil
	})

	// io.EOF is expected, since other side will close not auth'd connection
	// after it failed to check our public key.
	if err := authConn.Trigger(); err != nil && err != io.EOF {
		log.Warningf("peek: %v", err)
	}

	return authConn.RemotePubKey(), authConn.RemoteName(), nil
}

// Close will close the connection from the client side
func (cl *Client) Close() error {
	return cl.conn.Close()
}

/////////////////////
// ACTUAL COMMANDS //
/////////////////////

// Ping will contact the remote. This will only work if both remotes are authenticated.
// This in contrast to the backend ping, which will work when there is a network connection.
func (cl *Client) Ping() error {
	call := cl.api.Ping(cl.ctx, func(p capnp.Meta_ping_Params) error {
		return nil
	})

	_, err := call.Struct()
	return err
}

// FetchStore tries to fetch all store data from the remote.
// This will only work when the other store allowed us to access all folders.
// (See IsCompleteFetchAllowed)
func (cl *Client) FetchStore() (*bytes.Buffer, error) {
	call := cl.api.FetchStore(cl.ctx, func(p capnp.Sync_fetchStore_Params) error {
		return nil
	})

	result, err := call.Struct()
	if err != nil {
		return nil, err
	}

	data, err := result.Data()
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(data), nil
}

// FetchPatch tries to get a set of changes since `fromIndex`.
// The serialized patch is returned as byte slice.
func (cl *Client) FetchPatch(fromIndex int64) ([]byte, error) {
	call := cl.api.FetchPatch(cl.ctx, func(p capnp.Sync_fetchPatch_Params) error {
		p.SetFromIndex(fromIndex)
		return nil
	})

	result, err := call.Struct()
	if err != nil {
		return nil, err
	}

	return result.Data()
}

// FetchPatches tries to get a set of changes since `fromIndex`, packages as
// individual changes.  The serialized patch is returned as byte slice.
func (cl *Client) FetchPatches(fromIndex int64) ([]byte, error) {
	call := cl.api.FetchPatches(cl.ctx, func(p capnp.Sync_fetchPatches_Params) error {
		p.SetFromIndex(fromIndex)
		return nil
	})

	result, err := call.Struct()
	if err != nil {
		return nil, err
	}

	return result.Data()
}

// IsCompleteFetchAllowed asks the remote if we can use FetchStore.
func (cl *Client) IsCompleteFetchAllowed() (bool, error) {
	call := cl.api.IsCompleteFetchAllowed(cl.ctx, func(p capnp.Sync_isCompleteFetchAllowed_Params) error {
		return nil
	})

	result, err := call.Struct()
	if err != nil {
		return false, err
	}

	return result.IsAllowed(), nil
}

// IsPushAllowed asks the remote if we may push to them.
func (cl *Client) IsPushAllowed() (bool, error) {
	call := cl.api.IsPushAllowed(cl.ctx, func(p capnp.Sync_isPushAllowed_Params) error {
		return nil
	})

	result, err := call.Struct()
	if err != nil {
		return false, err
	}

	return result.IsAllowed(), nil
}

// Push asks the remote to do a "brig sync" with us.
func (cl *Client) Push() error {
	call := cl.api.Push(cl.ctx, func(p capnp.Sync_push_Params) error {
		return nil
	})

	_, err := call.Struct()
	return err
}
