package client

import (
	"context"
	"fmt"
	"net"
	"net/url"

	"github.com/sahib/brig/server/capnp"
	"zombiezen.com/go/capnproto2/rpc"
)

// Client is a helper API that implements the rpc interface to brig and makes
// all data easily accessible from Go.  Note that this layer is needed, so we
// could later support other languages.
type Client struct {
	ctx     context.Context
	conn    *rpc.Conn
	rawConn net.Conn
	api     capnp.API
}

func connFromURL(s string) (net.Conn, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case "tcp":
		return net.Dial(u.Scheme, u.Host)
	case "unix":
		return net.Dial(u.Scheme, u.Path)
	default:
		return nil, fmt.Errorf("unsupported protocol: %v", u.Scheme)
	}
}

// Dial will attempt to connect to brigd under the specified port
func Dial(ctx context.Context, daemonURL string) (*Client, error) {
	rawConn, err := connFromURL(daemonURL)
	if err != nil {
		return nil, err
	}

	transport := rpc.StreamTransport(rawConn)
	conn := rpc.NewConn(transport, rpc.ConnLog(nil))
	api := capnp.API{Client: conn.Bootstrap(ctx)}

	return &Client{
		ctx:     ctx,
		rawConn: rawConn,
		conn:    conn,
		api:     api,
	}, nil
}

// LocalAddr return info about the local addr
func (cl *Client) LocalAddr() net.Addr {
	return cl.rawConn.LocalAddr()
}

// RemoteAddr return info about the remote addr
func (cl *Client) RemoteAddr() net.Addr {
	return cl.rawConn.RemoteAddr()
}

// Close will close the connection from the client side
func (cl *Client) Close() error {
	return cl.conn.Close()
}
