package transfer

import (
	"net"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/id"
	"github.com/disorganizer/brig/repo"
	"github.com/disorganizer/brig/transfer/wire"
	"github.com/disorganizer/brig/util"
	"github.com/disorganizer/brig/util/ipfsutil"
)

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

func (cn *Connector) Dial(peer id.Peer) (*APIClient, error) {
	if !cn.IsInOnlineMode() {
		return nil, ErrOffline
	}

	cnv, err := cn.layer.Dial(peer)
	if err != nil {
		return nil, err
	}

	cn.mu.Lock()
	cn.open[peer.ID()] = cnv
	cn.mu.Unlock()

	return newAPIClient(cnv, cn.rp.IPFS)
}

func (cn *Connector) IsOnline(peer id.Peer) bool {
	if !cn.IsInOnlineMode() {
		return false
	}

	// cn.mu.Lock()
	// defer cn.mu.Unlock()
	return false
}

func (cn *Connector) Broadcast(req *wire.Request) error {
	var errs util.Errors

	for _, cnv := range cn.open {
		if err := cnv.SendAsync(req, nil); err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func (c *Connector) Layer() Layer {
	return c.layer
}

func (cn *Connector) Connect() error {
	ls, err := cn.rp.IPFS.Listen(cn.layer.ProtocolID())
	if err != nil {
		return err
	}

	go func() {
		for remote := range cn.rp.Remotes.Iter() {
			cnv, err := cn.layer.Dial(remote)
			if err != nil {
				log.Warningf("Could not connect to `%s`: %v", remote.ID(), err)
				continue
			}

			cn.mu.Lock()
			cn.open[remote.ID()] = cnv
			cn.mu.Unlock()
		}
	}()

	return cn.layer.Connect(ls, &dialer{cn.layer, cn.rp.IPFS})
}

func (cn *Connector) Disconnect() error {
	cn.mu.Lock()
	defer cn.mu.Unlock()

	for _, cnv := range cn.open {
		cnv.Close()
	}

	return cn.layer.Disconnect()
}

func (cn *Connector) IsInOnlineMode() bool {
	cn.mu.Lock()
	defer cn.mu.Unlock()

	return cn.layer.IsInOnlineMode()
}

// func (c *Connector) Auth(ID id.ID, peerHash string) error {
// 	c.mu.Lock()
// 	defer c.mu.Unlock()
//
// 	if c.client == nil {
// 		return ErrOffline
// 	}
//
// 	return c.client.Auth(ID, finger)
// }
