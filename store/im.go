package store

import (
	"crypto/tls"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/im"
	"github.com/disorganizer/brig/transfer"
	"github.com/tsuibin/goxmpp2/xmpp"
)

type Connector struct {
	repoPath string
	xmpp     *im.Client
}

func NewConnector(repoPath string) *Connector {
	return &Connector{
		repoPath: repoPath,
	}
}

func (c *Connector) loop() {
	// Listen to incoming connections:
	for {
		cnv := c.xmpp.Listen()
		if cnv == nil {
			log.Debugf("connector: server: quitting loop...")
			break
		}

		go func(cnv *im.Conversation) {
			server := transfer.NewServer(cnv)
			if err := server.Serve(); err != nil {
				log.Warningf("connector: server: %v", err)
			}

			if err := cnv.Close(); err != nil {
				log.Warningf("connector: server: Could not terminate conv: %v", err)
			}
		}(cnv)
	}
}

func (c *Connector) Talk(jid xmpp.JID) (*transfer.Client, error) {
	cnv, err := c.xmpp.Dial(jid)
	if err != nil {
		return nil, err
	}

	return transfer.NewClient(cnv), nil
}

func (c *Connector) Connect(jid xmpp.JID, password string) error {
	// Already connected?
	if c.xmpp != nil {
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

	xmpp, err := im.NewClient(cfg)
	if err != nil {
		return err
	}

	c.xmpp = xmpp
	go c.loop()

	return nil
}

func (c *Connector) Disconnect() error {
	// Already disconnected?
	if c.xmpp == nil {
		return nil
	}

	cl := c.xmpp
	c.xmpp = nil
	return cl.Close()
}

func (c *Connector) IsOnline() bool {
	return c.xmpp != nil
}
