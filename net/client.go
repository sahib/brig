package net

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"

	netBackend "github.com/disorganizer/brig/net/backend"
	"github.com/disorganizer/brig/net/capnp"
	"github.com/disorganizer/brig/repo"
	"zombiezen.com/go/capnproto2/rpc"
)

type Client struct {
	bk netBackend.Backend

	ctx     context.Context
	conn    *rpc.Conn
	rawConn net.Conn
	api     capnp.API
}

// func Dial(addr string, ctx context.Context, bk netBackend.Backend) (*Client, error) {
func Dial(name string, rp *repo.Repository, bk netBackend.Backend, ctx context.Context) (*Client, error) {
	remote, err := rp.Remotes.Remote(name)
	if err != nil {
		return nil, err
	}

	addr := remote.Fingerprint.Addr()
	keyring := rp.Keyring()
	ownPubKey, err := keyring.OwnPubKey()
	if err != nil {
		return nil, err
	}

	// Low level by addr, not by brig's remote name:
	rawConn, err := bk.Dial(addr, "brig/caprpc")
	if err != nil {
		return nil, err
	}

	authConn := NewAuthReadWriter(rawConn, keyring, ownPubKey, func(pubKey []byte) error {
		if !remote.Fingerprint.PubKeyMatches(pubKey) {
			return fmt.Errorf("remote pubkey does not match fingerprint")
		}

		return nil
	})

	// Trigger the authentication:
	// (otherwise it would be triggered on the first read/write)
	if err := authConn.Trigger(); err != nil {
		return nil, err
	}

	transport := rpc.StreamTransport(rawConn)
	clientConn := rpc.NewConn(transport)
	api := capnp.API{Client: clientConn.Bootstrap(ctx)}

	return &Client{
		ctx:     ctx,
		conn:    clientConn,
		rawConn: rawConn,
		api:     api,
	}, nil
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

func (cl *Client) PubKeyData() ([]byte, error) {
	call := cl.api.PubKey(cl.ctx, func(p capnp.Meta_pubKey_Params) error {
		return nil
	})

	result, err := call.Struct()
	if err != nil {
		return nil, err
	}

	return result.Key()
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
