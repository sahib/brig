package client

import (
	"context"
	"fmt"
	"net"

	"github.com/disorganizer/brig/server/capnp"
	"zombiezen.com/go/capnproto2/rpc"
)

// Client is a helper API that implements the rpc interface to brig and makes
// all data easily accessible from Go.  Note that this layer is needed, so we
// could later support other languages.
type Client struct {
	ctx     context.Context
	conn    *rpc.Conn
	tcpConn net.Conn

	api capnp.API
}

// Dial will attempt to connect to brigd under the specified port
func Dial(ctx context.Context, port int) (*Client, error) {
	addr := fmt.Sprintf("localhost:%d", port)
	tcpConn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	transport := rpc.StreamTransport(tcpConn)
	clientConn := rpc.NewConn(transport)
	api := capnp.API{Client: clientConn.Bootstrap(ctx)}

	return &Client{
		ctx:     ctx,
		conn:    clientConn,
		tcpConn: tcpConn,
		api:     api,
	}, nil
}

// Return info about the local addr
func (cl *Client) LocalAddr() net.Addr {
	return cl.tcpConn.LocalAddr()
}

// Return info about the remote addr
func (cl *Client) RemoteAddr() net.Addr {
	return cl.tcpConn.RemoteAddr()
}

// Close will close the connection from the client side
func (cl *Client) Close() error {
	return cl.conn.Close()
}
