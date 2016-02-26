package store

import (
	"crypto/tls"
	"path/filepath"

	"github.com/disorganizer/brig/im"
	"github.com/tsuibin/goxmpp2/xmpp"
)

type Connector struct {
	repoPath string
	client   *im.Client
}

func NewConnector(repoPath string) *Connector {
	return &Connector{
		repoPath: repoPath,
	}
}

func (c *Connector) Connect(jid xmpp.JID, password string) error {
	// Already connected?
	if c.client != nil {
		return nil
	}

	serverName := jid.Domain()
	cfg := &im.Config{
		Jid:             jid,
		Password:        password,
		TLSConfig:       tls.Config{ServerName: serverName},
		KeyPath:         filepath.Join(c.repoPath, "otr.key"),
		FingerprintPath: filepath.Join(c.repoPath, "otr.buddies"),
	}

	client, err := im.NewClient(cfg)
	if err != nil {
		return err
	}

	c.client = client
	return nil
}

func (c *Connector) Disconnect() error {
	// Already disconnected?
	if c.client == nil {
		return nil
	}

	cl := c.client
	c.client = nil
	return cl.Close()
}

func (c *Connector) IsOnline() bool {
	return c.client != nil
}
