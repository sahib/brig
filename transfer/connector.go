package transfer

import (
	"net"
	"sync"
	"time"

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

	// Conversation pool handling.
	cp *ConversationPool
}

type ConversationPool struct {
	// Map of open conversations
	open map[id.ID]Conversation

	// Map from hash id to last seen timestamp
	heartbeat map[id.ID]*ipfsutil.Pinger

	// lock for `open`
	mu sync.Mutex

	rp *repo.Repository
}

func newConversationPool(rp *repo.Repository) *ConversationPool {
	return &ConversationPool{
		open:      make(map[id.ID]Conversation),
		heartbeat: make(map[id.ID]*ipfsutil.Pinger),
		rp:        rp,
	}
}

// Set add a conversation for a specific to the pool.
func (cp *ConversationPool) Set(peer id.Peer, cnv Conversation) error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	cp.open[peer.ID()] = cnv

	if _, ok := cp.heartbeat[peer.ID()]; !ok {
		pinger, err := cp.rp.IPFS.Ping(peer.Hash())
		if err != nil {
			return err
		}
		cp.heartbeat[peer.ID()] = pinger
	}

	return nil
}

// Iter iterates over the conversation pool.
func (cp *ConversationPool) Iter() chan Conversation {
	cnvs := make(chan Conversation)
	go func() {
		cp.mu.Lock()
		defer cp.mu.Unlock()

		for _, cnv := range cp.open {
			cnvs <- cnv
		}
		close(cnvs)
	}()
	return cnvs
}

// LastSeen timestamp of a specific peer.
func (cp *ConversationPool) LastSeen(peer id.Peer) time.Time {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	if pinger := cp.heartbeat[peer.ID()]; pinger != nil {
		return pinger.LastSeen()
	}

	return time.Unix(0, 0)
}

// Close the complete conversation pool and free ressources.
func (cp *ConversationPool) Close() error {
	var errs util.Errors

	cp.mu.Lock()
	defer cp.mu.Unlock()

	for _, cnv := range cp.open {
		if err := cnv.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	cp.open = make(map[id.ID]Conversation)
	cp.heartbeat = make(map[id.ID]*ipfsutil.Pinger)

	return errs
}

// dialer uses ipfs to create a net.Conn to another node.
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
	cnc := &Connector{
		rp:    rp,
		layer: layer,
		cp:    newConversationPool(rp),
	}

	handlerMap := map[wire.RequestType]HandlerFunc{
		wire.RequestType_FETCH:       cnc.handleFetch,
		wire.RequestType_UPDATE_FILE: cnc.handleUpdateFile,
	}

	for typ, handler := range handlerMap {
		layer.RegisterHandler(typ, handler)
	}

	return cnc
}

func (cn *Connector) Dial(peer id.Peer) (*APIClient, error) {
	if !cn.IsInOnlineMode() {
		return nil, ErrOffline
	}

	// TODO: use the remote here somehow :)
	_, err := cn.rp.Remotes.Get(peer.ID())
	if err != nil {
		return nil, err
	}

	cnv, err := cn.layer.Dial(peer)
	if err != nil {
		return nil, err
	}

	if err := cn.cp.Set(peer, cnv); err != nil {
		return nil, err
	}

	return newAPIClient(cnv, cn.rp.IPFS)
}

func (cn *Connector) IsOnline(peer id.Peer) bool {
	if !cn.IsInOnlineMode() {
		return false
	}

	if time.Since(cn.cp.LastSeen(peer)) < 15*time.Second {
		return true
	}

	return false
}

func (cn *Connector) Broadcast(req *wire.Request) error {
	var errs util.Errors

	for cnv := range cn.cp.Iter() {
		if err := cnv.SendAsync(req, nil); err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func (cn *Connector) Layer() Layer {
	return cn.layer
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

			if err := cn.cp.Set(remote, cnv); err != nil {
				log.Warningf("Cannot create pinger: %v", err)
			}
		}
	}()

	return cn.layer.Connect(ls, &dialer{cn.layer, cn.rp.IPFS})
}

func (cn *Connector) Disconnect() error {
	errs := util.Errors{}
	if err := cn.cp.Close(); err != nil {
		errs = append(errs, err)
	}
	if err := cn.layer.Disconnect(); err != nil {
		errs = append(errs, err)
	}
	return errs
}

func (cn *Connector) Close() error {
	return cn.Disconnect()
}

func (cn *Connector) IsInOnlineMode() bool {
	return cn.layer.IsInOnlineMode()
}
