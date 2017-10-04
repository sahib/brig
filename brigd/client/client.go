package client

import (
	"context"
	"fmt"
	"net"

	"github.com/disorganizer/brig/brigd/capnp"
	"zombiezen.com/go/capnproto2/rpc"
)

type Client struct {
	ctx     context.Context
	conn    *rpc.Conn
	tcpConn net.Conn

	api capnp.API
}

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
		ctx:     context.Background(),
		conn:    clientConn,
		tcpConn: tcpConn,
		api:     api,
	}, nil
}

func (cl *Client) LocalAddr() net.Addr {
	return cl.tcpConn.LocalAddr()
}

func (cl *Client) RemoteAddr() net.Addr {
	return cl.tcpConn.RemoteAddr()
}

func (cl *Client) Close() error {
	return cl.conn.Close()
}
