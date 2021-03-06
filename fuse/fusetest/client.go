package fusetest

import (
	"context"
	"net"
	"net/http"

	"github.com/sahib/brig/util"
)

// Client allows controlling the
type Client struct {
	httpClient *http.Client
}

// Dial returns a client suitable for controlling a fusetest server.
func Dial(url string) (*Client, error) {
	scheme, addr, err := util.URLToSchemeAndAddr(url)
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial(scheme, addr)
			},
		},
	}
	return &Client{
		httpClient: httpClient,
	}, nil
}

// QuitServer sends a command that tells the server to quit.
// The request will block until the quit was carried out.
func (ctl *Client) QuitServer() error {
	req, err := http.NewRequest("GET", "/quit", nil)
	if err != nil {
		return err
	}

	_, err = ctl.httpClient.Do(req)
	return err
}
