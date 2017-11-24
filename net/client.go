package net

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"

	netBackend "github.com/disorganizer/brig/net/backend"
	"github.com/disorganizer/brig/net/capnp"
	"github.com/disorganizer/brig/net/peer"
	"github.com/disorganizer/brig/repo"
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
	ctl, err := DialByAddr(addr, remote.Fingerprint, rp.Keyring(), bk, ctx)
	if err != nil {
		return nil, err
	}

	// Save the remote's public key for later.
	// Might be used e.g. in locate()
	remotePubKey, err := ctl.RemotePubKey()
	if err != nil {
		return nil, err
	}

	if err := rp.Keyring().SavePubKey(name, remotePubKey); err != nil {
		return nil, err
	}

	return ctl, nil
}

func DialByAddr(
	addr string,
	fingerprint peer.Fingerprint,
	kr *repo.Keyring,
	bk netBackend.Backend,
	ctx context.Context,
) (*Client, error) {
	ownPubKey, err := kr.OwnPubKey()
	if err != nil {
		return nil, err
	}

	// Low level by addr, not by brig's remote name:
	rawConn, err := bk.Dial(addr, "brig/caprpc")
	if err != nil {
		return nil, err
	}

	authConn := NewAuthReadWriter(rawConn, kr, ownPubKey, func(pubKey []byte) error {
		// Skip authentication if no fingerprint was supplied:
		if string(fingerprint) == "" {
			return nil
		}

		if !fingerprint.PubKeyMatches(pubKey) {
			return fmt.Errorf("remote pubkey does not match fingerprint")
		}

		return nil
	})

	// Trigger the authentication:
	// (otherwise it would be triggered on the first read/write)
	if err := authConn.Trigger(); err != nil {
		return nil, err
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

// Close will close the connection from the client side
func (cl *Client) Close() error {
	return cl.conn.Close()
}

func (ctl *Client) RemotePubKey() ([]byte, error) {
	return ctl.authConn.RemotePubKey()
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

func (cl *Client) GetStore() (io.Reader, error) {
	call := cl.api.GetStore(cl.ctx, func(p capnp.Sync_getStore_Params) error {
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

	return bytes.NewReader(data), nil
}
