package net

import (
	"bytes"
	"context"
	"io"
	"net"

	"github.com/disorganizer/brig/net/capnp"
	"zombiezen.com/go/capnproto2/rpc"
)

type Client struct {
	bk Backend

	ctx     context.Context
	conn    *rpc.Conn
	rawConn net.Conn
	api     capnp.API
}

func Dial(who string, ctx context.Context, bk Backend) (*Client, error) {
	rawConn, err := bk.Dial(who, "caprpc")
	if err != nil {
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
