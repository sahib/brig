package transfer

import (
	"encoding/json"
	"io"
)

type Client struct {
	im      io.ReadWriter
	encoder *json.Encoder
	decoder *json.Decoder
}

func NewClient(im io.ReadWriter) *Client {
	return &Client{
		im:      im,
		encoder: json.NewEncoder(im),
		decoder: json.NewDecoder(im),
	}
}

func (c *Client) Send(cmd *Command) (*Response, error) {
	if err := c.encoder.Encode(&cmd); err != nil {
		return nil, err
	}

	resp := &Response{}
	if err := c.decoder.Decode(resp); err != nil {
		return nil, err
	}

	return resp, nil
}
