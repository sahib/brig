package transfer

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/id"
	"github.com/disorganizer/brig/repo"
	"github.com/disorganizer/brig/transfer/wire"
	"github.com/disorganizer/brig/util"
	"github.com/disorganizer/brig/util/ipfsutil"
	"github.com/gogo/protobuf/proto"
)

var (
	ErrListenerWasClosed = errors.New("Listener was closed")
)

type Connector struct {
	// Underlying protocol layer.
	layer Layer

	// Open repo. required for answering requests.
	// (might be nil for tests if no handlers are tested)
	rp *repo.Repository

	// Conversation pool handling.
	cp *conversationPool
}

type conversationPool struct {
	// Map of open conversations
	open map[id.ID]Conversation

	// Map from hash id to last seen timestamp
	heartbeat map[id.ID]*ipfsutil.Pinger

	// lock for `open`
	mu sync.Mutex

	// repo passed from Connector
	rp *repo.Repository

	// layer passed from Connector
	layer Layer

	changeCh     chan *repo.RemoteChange
	updateTicker *time.Ticker
}

func newConversationPool(rp *repo.Repository, layer Layer) *conversationPool {
	cp := &conversationPool{
		open:         make(map[id.ID]Conversation),
		heartbeat:    make(map[id.ID]*ipfsutil.Pinger),
		rp:           rp,
		layer:        layer,
		changeCh:     make(chan *repo.RemoteChange),
		updateTicker: time.NewTicker(120 * time.Second),
	}

	rp.Remotes.Register(func(change *repo.RemoteChange) {
		cp.changeCh <- change
	})

	// Make sure we immediately dis/connect to other peers
	// when the remote store changes externally.
	go func() {
		for change := range cp.changeCh {
			doRemove, doUpdate := false, false
			switch change.ChangeType {
			case repo.RemoteChangeAdded:
				doUpdate = true
			case repo.RemoteChangeModified:
				doUpdate, doRemove = true, true
			case repo.RemoteChangeRemoved:
				doRemove = true
			default:
				log.Warningf("Invalid remote change type: %d", change.ChangeType)
				return
			}

			if doRemove {
				if err := cp.Remove(change.OldRemote); err != nil {
					log.Warningf(
						"Cannot remove `%s` from connection pool: %v",
						change.OldRemote.ID(),
						err,
					)
				}
			}

			if doUpdate {
				cp.UpdateConnections()
			}
		}
	}()

	// Keep connection pool updates in a certain interval:
	go func() {
		for range cp.updateTicker.C {
			cp.UpdateConnections()
		}
	}()

	return cp
}

func (cp *conversationPool) UpdateConnections() {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	for _, remote := range cp.rp.Remotes.List() {
		_, ok := cp.open[remote.ID()]

		// We already have an open connection:
		if ok {
			continue
		}

		cnv, err := cp.layer.Dial(remote)

		if err != nil {
			log.Warningf("Could not connect to `%s`: %v", remote.ID(), err)
			continue
		}

		if err := cp.setUnlocked(remote, cnv); err != nil {
			log.Warningf("Cannot create pinger: %v", err)
		}
	}
}

func (cp *conversationPool) Remove(peer id.Peer) error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	peerID := peer.ID()
	cnv, ok := cp.open[peerID]
	if !ok {
		return repo.ErrNoSuchRemote(peerID)
	}

	pinger, ok := cp.heartbeat[peerID]
	if !ok {
		return fmt.Errorf("No pinger for `%s`?", peerID)
	}

	delete(cp.open, peerID)
	delete(cp.heartbeat, peerID)

	pinger.Close()
	return cnv.Close()
}

// Set add a conversation for a specific to the pool.
func (cp *conversationPool) Set(peer id.Peer, cnv Conversation) error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	return cp.setUnlocked(peer, cnv)
}

func (cp *conversationPool) setUnlocked(peer id.Peer, cnv Conversation) error {
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
func (cp *conversationPool) Iter() <-chan Conversation {
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
func (cp *conversationPool) LastSeen(peer id.Peer) time.Time {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	if pinger := cp.heartbeat[peer.ID()]; pinger != nil {
		return pinger.LastSeen()
	}

	return time.Unix(0, 0)
}

// Close the complete conversation pool and free ressources.
func (cp *conversationPool) Close() error {
	var errs util.Errors

	// Make sure we kill down the go routines
	// in newConversationPool to prevent leaks.
	close(cp.changeCh)
	cp.updateTicker.Stop()

	cp.mu.Lock()
	defer cp.mu.Unlock()

	for _, cnv := range cp.open {
		if err := cnv.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	cp.open = make(map[id.ID]Conversation)
	cp.heartbeat = make(map[id.ID]*ipfsutil.Pinger)
	return errs.ToErr()
}

// dialer uses ipfs to create a net.Conn to another node.
type dialer struct {
	layer Layer
	node  *ipfsutil.Node
}

func (d *dialer) Dial(peer id.Peer) (net.Conn, error) {
	log.Debugf("IPFS dialing to %v", peer.Hash())
	return d.node.Dial(peer.Hash(), d.layer.ProtocolID())
}

type listenerFilter struct {
	ls   net.Listener
	rms  repo.RemoteStore
	quit chan bool
}

func newListenerFilter(ls net.Listener, rms repo.RemoteStore) *listenerFilter {
	return &listenerFilter{
		ls:   ls,
		rms:  rms,
		quit: make(chan bool, 1),
	}
}

func (lf *listenerFilter) Accept() (net.Conn, error) {
	for {
		conn, err := lf.ls.Accept()
		if err != nil {
			return nil, err
		}

		select {
		case <-lf.quit:
			return nil, ErrListenerWasClosed
		default:
			break
		}

		streamConn, ok := conn.(*ipfsutil.StreamConn)
		if !ok {
			return nil, fmt.Errorf("Not used with ipfs listener?")
		}

		hash := streamConn.PeerHash()

		// Check if we know of this hash:
		for _, remote := range lf.rms.List() {
			if remote.Hash() == hash {
				return streamConn, nil
			}
		}

		log.Warningf("Denying incoming connection from `%s`", hash)
	}
}

func (lf *listenerFilter) Close() error {
	// quit is buffered, this will return immediately.
	// The Accept() loop might have errored out before,
	// so we don't want this to block if it won't be read.
	lf.quit <- true
	return lf.ls.Close()
}

func (lf *listenerFilter) Addr() net.Addr {
	return lf.Addr()
}

// NewConnector returns an unconnected Connector.
func NewConnector(layer Layer, rp *repo.Repository) *Connector {
	cnc := &Connector{
		rp:    rp,
		layer: layer,
		cp:    newConversationPool(rp, layer),
	}

	handlerMap := map[wire.RequestType]HandlerFunc{
		wire.RequestType_FETCH:         cnc.handleFetch,
		wire.RequestType_UPDATE_FILE:   cnc.handleUpdateFile,
		wire.RequestType_STORE_VERSION: cnc.handleStoreVersion,
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

func (c *Connector) Repo() *repo.Repository {
	return c.rp
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

func (cn *Connector) Broadcaster() *Broadcaster {
	return &Broadcaster{cn}
}

func (cn *Connector) broadcast(req *wire.Request) error {
	var errs util.Errors

	req.ID = proto.Int64(0)

	for cnv := range cn.cp.Iter() {
		if err := cnv.SendAsync(req, nil); err != nil {
			errs = append(errs, err)
		}
	}

	return errs.ToErr()
}

func (cn *Connector) Connect() error {
	ls, err := cn.rp.IPFS.Listen(cn.layer.ProtocolID())
	if err != nil {
		return err
	}

	// Make sure we filter unauthorized incoming connections:
	filter := newListenerFilter(ls, cn.rp.Remotes)
	dialer := &dialer{cn.layer, cn.rp.IPFS}

	if err := cn.layer.Connect(filter, dialer); err != nil {
		return err
	}

	go cn.cp.UpdateConnections()
	return nil
}

func (cn *Connector) Disconnect() error {
	errs := util.Errors{}
	if err := cn.cp.Close(); err != nil {
		errs = append(errs, err)
	}

	if err := cn.layer.Disconnect(); err != nil {
		errs = append(errs, err)
	}

	return errs.ToErr()
}

func (cn *Connector) Close() error {
	return cn.Disconnect()
}

func (cn *Connector) IsInOnlineMode() bool {
	return cn.layer.IsInOnlineMode()
}
