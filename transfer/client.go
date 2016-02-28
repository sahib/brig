package transfer

import "io"
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

func (c *Client) SendClone() ([]byte, error) {
	// TODO
	return nil, nil
}
