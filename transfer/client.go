package transfer

import (
	"errors"
	"io"
)
import "github.com/disorganizer/brig/transfer/proto"

type Client struct {
	im   io.ReadWriteCloser
	ptcl *ClientProtocol
}

func NewClient(im io.ReadWriteCloser) *Client {
	return &Client{
		im:   im,
		ptcl: NewClientProtocol(im),
	}
}

func (c *Client) Send(req *proto.Request) (*proto.Response, error) {
	if err := c.ptcl.Encode(req); err != nil {
		return nil, err
	}

	resp, err := c.ptcl.Decode()
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) Close() error {
	return c.im.Close()
}

func (c *Client) unpack(req *proto.Request) ([]byte, error) {
	resp, err := c.Send(req)
	if err != nil {
		return nil, err
	}

	if resp.GetError() != "" {
		return nil, errors.New(resp.GetError())
	}

	return resp.GetData(), nil
}

func (c *Client) DoFetch() ([]byte, error) {
	return c.unpack(&proto.Request{
		Type: proto.RequestType_FETCH.Enum(),
	})
}
