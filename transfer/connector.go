package transfer

import (
	"crypto/tls"
	"path/filepath"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/repo"
	"github.com/disorganizer/brig/transfer/xmpp"
	goxmpp "github.com/tsuibin/goxmpp2/xmpp"
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
	// Client is the underlying otr authenticated xmpp client, created on Connect()
	client *xmpp.Client

	// Open repo. required for answering requests.
	// (might be nil for tests if no handlers are tested)
	rp *repo.Repository

	// Map of open conversations
	open map[goxmpp.JID]*xmpp.Conversation

	// lock for `open`
	mu sync.Mutex
}

// NewConnector returns an unconnected Connector.
func NewConnector(rp *repo.Repository) *Connector {
	return &Connector{
		rp:   rp,
		open: make(map[goxmpp.JID]*xmpp.Conversation),
	}
}

// Mainloop for incoming conversations.
func (c *Connector) loop() {
	for {
		cnv := c.client.Listen()
		if cnv == nil {
			log.Debugf("connector: server: quitting loop...")
			break
		}

		// Establish a ServerProtocol on the conversation:
		go func(cnv *xmpp.Conversation) {
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

func (c *Connector) Talk(jid goxmpp.JID) (*Client, error) {
	c.mu.Lock()
	if cnv, ok := c.open[jid]; ok {
		c.mu.Unlock()
		return NewClient(cnv), nil
	}
	c.mu.Unlock()

	cnv, err := c.client.Dial(jid)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.open[jid] = cnv
	c.mu.Unlock()

	return NewClient(cnv), nil
}

func (c *Connector) Connect(jid goxmpp.JID, password string) error {
	// Already connected?
	c.mu.Lock()
	if c.client != nil {
		return nil
	}
	c.mu.Unlock()

	serverName := jid.Domain()
	cfg := &xmpp.Config{
		Jid:             jid,
		Password:        password,
		TLSConfig:       tls.Config{ServerName: serverName},
		KeyPath:         filepath.Join(c.rp.InternalFolder, "otr.key"),
		FingerprintPath: filepath.Join(c.rp.InternalFolder, "otr.buddies"),
	}

	xmpp, err := xmpp.NewClient(cfg)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.client = xmpp
	c.mu.Unlock()

	go c.loop()
	return nil
}

func (c *Connector) Disconnect() error {
	// Already disconnected?
	if c.client == nil {
		return nil
	}

	c.mu.Lock()
	cl := c.client
	c.client = nil
	c.mu.Unlock()

	return cl.Close()
}

func (c *Connector) IsOnline() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.client != nil
}

func (c *Connector) Auth(jid goxmpp.JID, finger string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client == nil {
		return ErrOffline
	}

	return c.client.Auth(jid, finger)
}

func (c *Connector) Fingerprint() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client == nil {
		return "", ErrOffline
	}

	return c.client.Fingerprint(), nil
}
