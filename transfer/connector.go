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
)

var (
	// ErrListenerWasClosed is returned by Accept() when connector quit
	ErrListenerWasClosed = errors.New("Listener was closed")
)

// Connector is the connection layer of brig.
// It manages a pool of connections to all registered remotes
// and dials new connections to new remotes once added.
//
// It offers the caller to dial to a remote brig daemon
// via a so called APIClient. The caller can then call the supported methods
// to query information from the remote daemon.
//
// It's also supported to broadcast messages to all connected remotes.
// Broadcasted messages are not guaranteed to be delivered and should
// be therefore only used to quickly share updates of e.g. files on our side.
// Incoming broadcast messages and APIClient requests will be handled by
// the connector and the methods implemented in serverops.go.
type Connector struct {
	// Underlying protocol layer.
	layer Layer

	// Open repo. required for answering requests.
	// (might be nil for tests if no handlers are tested)
	rp *repo.Repository

	// Conversation pool handling.
	cp *conversationPool
}

// conversationPool holds all open conversations and tries to keep them
// up-to-date. It also occasionally pings other remotes in order to
// check if they're still alive.
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

	// changeCh will be triggered once RemoteStore was changed.
	changeCh chan *repo.RemoteChange

	// updateTicker updates the connection pool when triggered.
	updateTicker *time.Ticker
}

func newConversationPool(rp *repo.Repository, layer Layer) *conversationPool {
	cp := &conversationPool{
		open:         make(map[id.ID]Conversation),
		heartbeat:    make(map[id.ID]*ipfsutil.Pinger),
		rp:           rp,
		layer:        layer,
		changeCh:     make(chan *repo.RemoteChange, 10),
		updateTicker: time.NewTicker(120 * time.Second),
	}

	rp.Remotes.Register(func(change *repo.RemoteChange) {
		go func() {
			// Wait a bit of time in the case that both parties
			// just added them as remotes respecitvely.
			time.Sleep(500 * time.Millisecond)

			// Better check if the channel wasn't closed yet:
			cp.mu.Lock()
			if cp.changeCh != nil {
				cp.changeCh <- change
			}
			cp.mu.Unlock()
		}()
	})

	// Make sure we immediately dis/connect to other peers
	// when the remote store changes externally.
	go func() {
		for change := range cp.changeCh {
			// Check how we need to handle that change:
			doRemove, doUpdate := false, false
			switch change.ChangeType {
			case repo.RemoteChangeAdded:
				//doUpdate = true
			case repo.RemoteChangeModified:
				doUpdate, doRemove = true, true
			case repo.RemoteChangeRemoved:
				doRemove = true
			default:
				log.Warningf("Invalid remote change type: %d", change.ChangeType)
				return
			}

			if doRemove {
				if err := cp.Forget(change.OldRemote); err != nil {
					log.Warningf(
						"Cannot forget `%s` from connection pool: %v",
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

	// Keep connection pool updates in a certain time interval:
	go func() {
		for range cp.updateTicker.C {
			cp.UpdateConnections()
		}
	}()

	return cp
}

// UpdateConnections checks which remotes currently do not have an
// connection and attempts to dial them.
// Note: There can be only active UpdateConnections at a time.
func (cp *conversationPool) UpdateConnections() {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	for _, remote := range cp.rp.Remotes.List() {
		// We already have an open connection:
		_, ok := cp.open[remote.ID()]
		if ok {
			continue
		}

		// Ask layer to dial to the remote:
		cnv, err := cp.layer.Dial(remote)
		if err != nil {
			log.Warningf("Could not connect to `%s`: %v", remote.ID(), err)
			continue
		}

		if err := cp.rememberUnlocked(remote, cnv); err != nil {
			log.Warningf("Cannot create pinger: %v", err)
		}
	}
}

func (cp *conversationPool) WaitForUpdate() {
	// UpdateConnections locks the mutex
	cp.mu.Lock()
	cp.mu.Unlock()
}

// Forget removes a peer from the pool and cleans up after it.
func (cp *conversationPool) Forget(peer id.Peer) error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	peerID := peer.ID()
	cnv, ok := cp.open[peerID]
	if !ok {
		return repo.ErrNoSuchRemote(peerID)
	}

	if pinger, ok := cp.heartbeat[peerID]; ok {
		pinger.Close()
	}

	delete(cp.open, peerID)
	delete(cp.heartbeat, peerID)
	return cnv.Close()
}

// Remember adds a conversation to a specific peer to the pool.
func (cp *conversationPool) Remember(peer id.Peer, cnv Conversation) error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	return cp.rememberUnlocked(peer, cnv)
}

func (cp *conversationPool) rememberUnlocked(peer id.Peer, cnv Conversation) error {
	cp.open[peer.ID()] = cnv

	// Create a new pinger if not already done:
	if _, ok := cp.heartbeat[peer.ID()]; !ok {
		pinger, err := cp.rp.IPFS.Ping(peer.Hash())
		if err != nil {
			return err
		}
		cp.heartbeat[peer.ID()] = pinger
	}

	return nil
}

// Iter iterates over all conversations in the pool
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

// LastSeen returns the time when we've last seen `peer`
func (cp *conversationPool) LastSeen(peer id.Peer) time.Time {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	if pinger := cp.heartbeat[peer.ID()]; pinger != nil {
		return pinger.LastSeen()
	}

	return time.Unix(0, 0)
}

// Close the complete conversation pool and free resources.
func (cp *conversationPool) Close() error {
	var errs util.Errors

	// Make sure we kill down the go routines
	// in newConversationPool to prevent leaks.
	cp.updateTicker.Stop()

	cp.mu.Lock()
	defer cp.mu.Unlock()

	// Set changeCh to nil, so potential remote
	// events notice that the pool was closed.
	close(cp.changeCh)
	cp.changeCh = nil

	// Close all conversations. Does not need to be done by Layer.
	for _, cnv := range cp.open {
		if err := cnv.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	// Reset conversations maps:
	cp.open = make(map[id.ID]Conversation)
	cp.heartbeat = make(map[id.ID]*ipfsutil.Pinger)
	return errs.ToErr()
}

// ipfsDialer uses ipfs to create a net.Conn to another node,
// which is referenced by a it's peer hash.
type ipfsDialer struct {
	layer Layer
	node  *ipfsutil.Node
}

func (id *ipfsDialer) Dial(peer id.Peer) (net.Conn, error) {
	log.Debugf("IPFS dialing to %v", peer.Hash())
	return id.node.Dial(peer.Hash(), id.layer.ProtocolID())
}

// listenerFilter only Accept()s connections that are listed
// in the remote store of brig. Other connection attemtps
// are logged and Accept() will continue to wait for connections.
type listenerFilter struct {
	ls   net.Listener
	rms  repo.RemoteStore
	quit chan bool
}

// newListenerFilter constructs a new filter with the remotes in `rms` over `ls`.
func newListenerFilter(ls net.Listener, rms repo.RemoteStore) *listenerFilter {
	return &listenerFilter{
		ls:   ls,
		rms:  rms,
		quit: make(chan bool, 1),
	}
}

// Accept blocks until a valid, authenticated connection arrives.
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

// Close wakes up all Accept() calls.
func (lf *listenerFilter) Close() error {
	// quit is buffered, this will return immediately.
	// The Accept() loop might have errored out before,
	// so we don't want this to block if it won't be read.
	lf.quit <- true
	return lf.ls.Close()
}

// Addr returns the Addr() of the underlying listener.
func (lf *listenerFilter) Addr() net.Addr {
	return lf.Addr()
}

// NewConnector returns an unconnected Connector.
// Connect() should be called in order to be fully working.
func NewConnector(layer Layer, rp *repo.Repository) *Connector {
	cnc := &Connector{
		rp:    rp,
		layer: layer,
		cp:    newConversationPool(rp, layer),
	}

	// handlerMap maps the request type to the respective method
	// value of this connector. This way the methods complies
	// "HandlerFunc" but still gets passed a Connector.
	handlerMap := map[wire.RequestType]HandlerFunc{
		wire.RequestType_FETCH:         cnc.handleFetch,
		wire.RequestType_UPDATE_FILE:   cnc.handleUpdateFile,
		wire.RequestType_STORE_VERSION: cnc.handleStoreVersion,
	}

	// Register them so layer knows which one to call:
	for typ, handler := range handlerMap {
		layer.RegisterHandler(typ, handler)
	}

	return cnc
}

// DialID looks up the hash of `ident` from the remotes.
// Otherwise it's the same as the regular Dial()
func (cn *Connector) DialID(ident id.ID) (*APIClient, error) {
	remote, err := cn.rp.Remotes.Get(ident)
	if err != nil {
		return nil, err
	}

	return cn.Dial(remote)
}

// Dial returns an APIClient that is connected to `peer`.
// If there's already a conversation to that peer no new
// connection will be created.
//
// It will return ErrOffline if the connector is not connected.
//
// Note: All connections will be multiplexed over ipfs' swarm.
func (cn *Connector) Dial(peer id.Peer) (*APIClient, error) {
	if !cn.IsInOnlineMode() {
		return nil, ErrOffline
	}

	// Check if we're allowed to connect to that node:
	if _, err := cn.rp.Remotes.Get(peer.ID()); err != nil {
		return nil, err
	}

	cnv, err := cn.layer.Dial(peer)
	if err != nil {
		return nil, err
	}

	if err := cn.cp.Remember(peer, cnv); err != nil {
		return nil, err
	}

	return newAPIClient(cnv, cn.rp.IPFS)
}

// Repo returns the repo used to build this Connector.
func (cn *Connector) Repo() *repo.Repository {
	return cn.rp
}

// IsOnline checks if we believe that `peer` is still online.
// It does not actually ping the peer, but checks if the last
// ping was not too long ago.
func (cn *Connector) IsOnline(peer id.Peer) bool {
	if !cn.IsInOnlineMode() {
		return false
	}

	if time.Since(cn.cp.LastSeen(peer)) < 15*time.Second {
		return true
	}

	return false
}

// Broadcaster returns a Broadcaster struct that can be used
// to broadcast messages to all connected remotes.
//
// NOTE: Peers that are offline or are just about to connect
//       might not retrieve the message.
//
// It's only factored out of Connector for cosmetic reasons.
func (cn *Connector) Broadcaster() *Broadcaster {
	return &Broadcaster{cn}
}

// broadcast implements the actual network broadcasting.
// It just sends the request over all conversations in the pool.
func (cn *Connector) broadcast(req *wire.Request) error {
	var errs util.Errors

	req.ID = 0

	for cnv := range cn.cp.Iter() {
		if err := cnv.SendAsync(req, nil); err != nil {
			errs = append(errs, err)
		}
	}

	return errs.ToErr()
}

// Connect establishes the connection pool and makes Dial() possible.
// Broadcasting will work shortly after the Connect()
//
// If you need to wait for the pool to be ready for broadcasting, you
// can use WaitForPool(), if you need to wait for a single peer,
// just wait for a successful Dial().
func (cn *Connector) Connect() error {
	ls, err := cn.rp.IPFS.Listen(cn.layer.ProtocolID())
	if err != nil {
		return err
	}

	// Make sure we filter unauthorized incoming connections:
	filter := newListenerFilter(ls, cn.rp.Remotes)
	dialer := &ipfsDialer{cn.layer, cn.rp.IPFS}
	if err := cn.layer.Connect(filter, dialer); err != nil {
		return err
	}

	go cn.cp.UpdateConnections()
	return nil
}

// WaitForPool blocks until the connection pool tried it's best to
// connect to all remotes or until RemoteStore changed and the pool
// adapted to reflect the change.
//
// Or in short words: It waits until Connector portrays reality.
// If the connector is offline, WaitForPool returns ErrOffline immediately.
func (cn *Connector) WaitForPool() error {
	if !cn.IsInOnlineMode() {
		return ErrOffline
	}

	// If WaitForPool is called directly after Connect,
	// it might lock the mutex before UpdateConnections does.
	time.Sleep(10 * time.Millisecond)

	// Wait until UpdateConnections ran through:
	cn.cp.WaitForUpdate()
	return nil
}

// Disconnect closes the connection pool and
// resets the connector to the state it had after NewConnector().
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

// Close is the same as Disconnect()
func (cn *Connector) Close() error {
	return cn.Disconnect()
}

// IsInOnlineMode returns true after a successful call to Connect()
func (cn *Connector) IsInOnlineMode() bool {
	return cn.layer.IsInOnlineMode()
}
