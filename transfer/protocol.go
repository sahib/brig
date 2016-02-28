package transfer

import (
	"io"

	"github.com/disorganizer/brig/transfer/proto"
	"github.com/disorganizer/brig/util/protocol"
)

type ClientProtocol struct {
	p *protocol.Protocol
}

func NewClientProtocol(rw io.ReadWriter) *ClientProtocol {
	return &ClientProtocol{protocol.NewProtocol(rw, true)}
}

func (c *ClientProtocol) Encode(req *proto.Request) error {
	return c.p.Send(req)
}

func (c *ClientProtocol) Decode() (*proto.Response, error) {
	resp := &proto.Response{}
	if err := c.p.Recv(resp); err != nil {
		return nil, err
	}

	return resp, nil
}

type ServerProtocol struct {
	p *protocol.Protocol
}

func NewServerProtocol(rw io.ReadWriter) *ServerProtocol {
	return &ServerProtocol{protocol.NewProtocol(rw, true)}
}

func (c *ServerProtocol) Encode(resp *proto.Response) error {
	return c.p.Send(resp)
}

func (c *ServerProtocol) Decode() (*proto.Request, error) {
	req := &proto.Request{}
	if err := c.p.Recv(req); err != nil {
		return nil, err
	}

	return req, nil
}
