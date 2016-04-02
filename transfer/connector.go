package transfer

import (
	"net"
	"sync"

	"github.com/disorganizer/brig/id"
	"github.com/disorganizer/brig/repo"
	"github.com/disorganizer/brig/util/ipfsutil"
)

// Connector is a pool of metadata connections.
// It listens for new connections and establishes a ServerProtocol on them.
// Clients can talk to other parties with the Talk() function, which
// esatblishes a ready to use ClientProtocol.
//
// Connector tries to hold the connections open as long as possible,
// since building the conversations is expensive due to OTR.
//
// It is okay to call Connector from more than one goroutine.
type Connector struct {
	layer Layer

	// Open repo. required for answering requests.
	// (might be nil for tests if no handlers are tested)
	rp *repo.Repository

	// Map of open conversations
	open map[id.ID]Conversation

	// lock for `open`
	mu sync.Mutex
}

type dialer struct {
	layer Layer
	node  *ipfsutil.Node
}

func (d *dialer) Dial(peer id.Peer) (net.Conn, error) {
	return d.node.Dial(peer.Hash(), d.layer.ProtocolID())
}

// NewConnector returns an unconnected Connector.
func NewConnector(layer Layer, rp *repo.Repository) *Connector {
	// TODO: pass authMgr.
	// authMgr := MockAuthSuccess

	// TODO: register handlers
	// layer.RegisterHandler(...)

	return &Connector{
		rp:    rp,
		layer: layer,
		open:  make(map[id.ID]Conversation),
	}
}

// Mainloop for incoming conversations.
// func (c *Connector) loop() {
// 	for {
// 		cnv := c.client.Listen()
// 		if cnv == nil {
// 			log.Debugf("connector: server: quitting loop...")
// 			break
// 		}
//
// 		// Establish a ServerProtocol on the conversation:
// 		go func(cnv *xmpp.Conversation) {
// 			server := NewServer(cnv, c.rp)
// 			if err := server.Serve(); err != nil {
// 				log.Warningf("connector: server: %v", err)
// 			}
//
// 			if err := cnv.Close(); err != nil {
// 				log.Warningf("connector: server: Could not terminate conv: %v", err)
// 			}
//
// 			c.mu.Lock()
// 			delete(c.open, cnv.Jid)
// 			c.mu.Unlock()
// 		}(cnv)
// 	}
// }
//
// func (c *Connector) Talk(ID goxmpp.JID) (*Client, error) {
// 	c.mu.Lock()
// 	if cnv, ok := c.open[ID]; ok {
// 		c.mu.Unlock()
// 		return NewClient(cnv), nil
// 	}
// 	c.mu.Unlock()
//
// 	cnv, err := c.client.Dial(ID)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	c.mu.Lock()
// 	c.open[ID] = cnv
// 	c.mu.Unlock()
//
// 	return NewClient(cnv), nil
// }

func (c *Connector) Layer() Layer {
	return c.layer
}

func (c *Connector) Connect() error {
	ls, err := c.rp.IPFS.Listen(c.layer.ProtocolID())
	if err != nil {
		return err
	}

	return c.layer.Connect(ls, &dialer{c.layer, c.rp.IPFS})
}

func (c *Connector) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.layer.Disconnect()
}

func (c *Connector) IsOnlineMode() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.layer.IsOnlineMode()
}

func (c *Connector) Auth(ID id.ID, peerHash string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client == nil {
		return ErrOffline
	}

	return c.client.Auth(ID, finger)
}
