package transfer

import (
	"crypto/tls"
	"path/filepath"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/im"
	"github.com/disorganizer/brig/repo"
	"github.com/tsuibin/goxmpp2/xmpp"
)

// Connector is a pool of xmpp connections.
// It listens for new connections and establishes a ServerProtocol on them.
// Clients can talk to other parties with the Talk() function, which
// esatblishes a ready to use ClientProtocol.
//
// Connector tries to hold the connections open as long as possible,
// since building the conversations is expensive due to OTR.
//
// It is okay to call Connector from more than one goroutine.
type Connector struct {
	// The "own" client. Created on Connect()
	xmpp *im.Client

	// Open repo. required for answering requests.
	// (might be nil for tests if no handlers are tested)
	rp *repo.Repository

	// Map of open conversations
	open map[xmpp.JID]*im.Conversation

	// lock for `open`
	mu sync.Mutex
}

// NewConnector returns an unconnected Connector.
func NewConnector(rp *repo.Repository) *Connector {
	return &Connector{
		rp:   rp,
		open: make(map[xmpp.JID]*im.Conversation),
	}
}

// Mainloop for incoming conversations.
func (c *Connector) loop() {
	for {
		cnv := c.xmpp.Listen()
		if cnv == nil {
			log.Debugf("connector: server: quitting loop...")
			break
		}

		// Establish a ServerProtocol on the conversation:
		go func(cnv *im.Conversation) {
			server := NewServer(cnv, c.rp)
			if err := server.Serve(); err != nil {
				log.Warningf("connector: server: %v", err)
			}

			if err := cnv.Close(); err != nil {
				log.Warningf("connector: server: Could not terminate conv: %v", err)
			}

			c.mu.Lock()
			delete(c.open, cnv.Jid)
			c.mu.Unlock()
		}(cnv)
	}
}

func (c *Connector) Talk(jid xmpp.JID) (*Client, error) {
	c.mu.Lock()
	if cnv, ok := c.open[jid]; ok {
		c.mu.Unlock()
		return NewClient(cnv), nil
	}
	c.mu.Unlock()

	cnv, err := c.xmpp.Dial(jid)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.open[jid] = cnv
	c.mu.Unlock()

	return NewClient(cnv), nil
}

func (c *Connector) Connect(jid xmpp.JID, password string) error {
	// Already connected?
	c.mu.Lock()
	if c.xmpp != nil {
		return nil
	}
	c.mu.Unlock()

	serverName := jid.Domain()
	cfg := &im.Config{
		Jid:             jid,
		Password:        password,
		TLSConfig:       tls.Config{ServerName: serverName},
		KeyPath:         filepath.Join(c.rp.InternalFolder, "otr.key"),
		FingerprintPath: filepath.Join(c.rp.InternalFolder, "otr.buddies"),
	}

	xmpp, err := im.NewClient(cfg)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.xmpp = xmpp
	c.mu.Unlock()

	go c.loop()
	return nil
}

func (c *Connector) Disconnect() error {
	// Already disconnected?
	if c.xmpp == nil {
		return nil
	}

	c.mu.Lock()
	cl := c.xmpp
	c.xmpp = nil
	c.mu.Unlock()

	return cl.Close()
}

func (c *Connector) IsOnline() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.xmpp != nil
}
