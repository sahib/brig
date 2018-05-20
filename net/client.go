package net

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"

	log "github.com/Sirupsen/logrus"
	e "github.com/pkg/errors"
	netBackend "github.com/sahib/brig/net/backend"
	"github.com/sahib/brig/net/capnp"
	"github.com/sahib/brig/net/peer"
	"github.com/sahib/brig/repo"
	"zombiezen.com/go/capnproto2/rpc"
)

type Client struct {
	bk netBackend.Backend

	ctx      context.Context
	conn     *rpc.Conn
	rawConn  net.Conn
	authConn *AuthReadWriter
	api      capnp.API
}

func Dial(name string, rp *repo.Repository, bk netBackend.Backend, ctx context.Context) (*Client, error) {
	remote, err := rp.Remotes.Remote(name)
	if err != nil {
		return nil, err
	}

	addr := remote.Fingerprint.Addr()
	ctl, err := DialByAddr(addr, remote.Fingerprint, rp, bk, ctx)
	if err != nil {
		return nil, e.Wrapf(err, "by-addr")
	}

	return ctl, nil
}

func DialByAddr(
	addr string,
	fingerprint peer.Fingerprint,
	rp *repo.Repository,
	bk netBackend.Backend,
	ctx context.Context,
) (*Client, error) {
	kr := rp.Keyring()
	ownPubKey, err := kr.OwnPubKey()
	if err != nil {
		return nil, err
	}

	// Low level by addr, not by brig's remote name:
	log.Debugf("raw dial to %s", addr)
	rawConn, err := bk.Dial(addr, "brig/caprpc")
	if err != nil {
		return nil, e.Wrapf(err, "raw")
	}

	ownName := rp.Owner
	if fingerprint == "" {
		return nil, fmt.Errorf("Rejecting own, empty fingerprint... bug?")
	}

	authConn := NewAuthReadWriter(rawConn, kr, ownPubKey, ownName, func(pubKey []byte) error {
		if !fingerprint.PubKeyMatches(pubKey) {
			return fmt.Errorf("remote pubkey does not match fingerprint")
		}

		return nil
	})

	// Trigger the authentication:
	// (otherwise it would be triggered on the first read/write)
	if err := authConn.Trigger(); err != nil {
		return nil, e.Wrapf(err, "auth")
	}

	// Setup capnp-rpc:
	transport := rpc.StreamTransport(rawConn)
	clientConn := rpc.NewConn(transport)
	api := capnp.API{Client: clientConn.Bootstrap(ctx)}

	return &Client{
		ctx:      ctx,
		authConn: authConn,
		conn:     clientConn,
		rawConn:  rawConn,
		api:      api,
	}, nil
}

func PeekRemotePubkey(
	addr string,
	rp *repo.Repository,
	bk netBackend.Backend,
	ctx context.Context,
) ([]byte, string, error) {
	kr := rp.Keyring()
	ownPubKey, err := kr.OwnPubKey()
	if err != nil {
		return nil, "", err
	}

	log.Debugf("peek to %s", addr)
	rawConn, err := bk.Dial(addr, "brig/caprpc")
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

func (cl *Client) Ping() error {
	call := cl.api.Ping(cl.ctx, func(p capnp.Meta_ping_Params) error {
		return nil
	})

	_, err := call.Struct()
	return err
}

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

func (cl *Client) FetchPatch(fromIndex int64) ([]byte, error) {
	call := cl.api.FetchPatch(cl.ctx, func(p capnp.Sync_fetchPatch_Params) error {
		p.SetFromIndex(fromIndex)
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

	return data, nil
}

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
