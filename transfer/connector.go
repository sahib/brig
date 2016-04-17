package transfer

import (
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

type Connector struct {
	layer Layer

	// Open repo. required for answering requests.
	// (might be nil for tests if no handlers are tested)
	rp *repo.Repository

	// Map of open conversations
	open map[id.ID]Conversation

	// Map from hash id to last seen timestamp
	heartbeat map[id.ID]*ipfsutil.Pinger

	// lock for `open`
	mu sync.Mutex
}

// dialer uses ipfs to create a net.Conn to another node.
type dialer struct {
	layer Layer
	node  *ipfsutil.Node
}

func (d *dialer) Dial(peer id.Peer) (net.Conn, error) {
	return d.node.Dial(peer.Hash(), d.layer.ProtocolID())
}

// listenerFilter filters
type listenerFilter struct {
	ls   net.Listener
	rms  repo.RemoteStore
	quit bool
}

func newListenerFilter(ls net.Listener, rms repo.RemoteStore) *listenerFilter {
	return &listenerFilter{
		ls:  ls,
		rms: rms,
	}
}

func (lf *listenerFilter) Accept() (net.Conn, error) {
	for !lf.quit {
		conn, err := lf.ls.Accept()
		if err != nil {
			return nil, err
		}

		streamConn, ok := conn.(*ipfsutil.StreamConn)
		if !ok {
			// TODO
			return nil, fmt.Errorf("Not used with ipfs listener?")
		}

		hash := streamConn.PeerHash()

		// Check if we know of this hash:
		for remote := range lf.rms.Iter() {
			if remote.Hash() == hash {
				return streamConn, nil
			}
		}

		log.Warningf("Denying incoming connection from `%s`", hash)
	}

	return nil, fmt.Errorf("Listener was closed")
}

func (lf *listenerFilter) Close() error {
	lf.quit = true
	return lf.ls.Close()
}

func (lf *listenerFilter) Addr() net.Addr {
	return lf.Addr()
}

// NewConnector returns an unconnected Connector.
func NewConnector(layer Layer, rp *repo.Repository) *Connector {
	// TODO: pass authMgr.
	// authMgr := MockAuthSuccess
	cnc := &Connector{
		rp:        rp,
		layer:     layer,
		open:      make(map[id.ID]Conversation),
		heartbeat: make(map[id.ID]*ipfsutil.Pinger),
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

func partnerIsAllowed(rms repo.RemoteStore, ID id.ID) error {
	_, err := rms.Get(ID)
	if err != nil {
		return err
	}

	return nil
}

func (cn *Connector) Dial(peer id.Peer) (*APIClient, error) {
	if !cn.IsInOnlineMode() {
		return nil, ErrOffline
	}

	if err := partnerIsAllowed(cn.rp.Remotes, peer.ID()); err != nil {
		return nil, err
	}

	// Lookup if a conversation was already established:
	cn.mu.Lock()
	cnv, ok := cn.open[peer.ID()]
	cn.mu.Unlock()

	if ok {
		return newAPIClient(cnv, cn.rp.IPFS)
	}

	cnv, err := cn.layer.Dial(peer)
	if err != nil {
		return nil, err
	}

	// Remember conversation:
	cn.mu.Lock()
	cn.open[peer.ID()] = cnv
	cn.mu.Unlock()

	return newAPIClient(cnv, cn.rp.IPFS)
}

func (cn *Connector) IsOnline(peer id.Peer) bool {
	if !cn.IsInOnlineMode() {
		return false
	}

	cn.mu.Lock()
	defer cn.mu.Unlock()

	pinger, ok := cn.heartbeat[peer.ID()]
	if !ok {
		var err error

		pinger, err = cn.rp.IPFS.Ping(peer.Hash())
		if err != nil {
			return false
		}

		cn.heartbeat[peer.ID()] = pinger
	}

	if time.Since(pinger.LastSeen()) < 5*time.Second {
		return true
	}

	// If creating the pinger worked, remote should be online.
	return true
}

func (cn *Connector) Broadcast(req *wire.Request) error {
	var errs util.Errors

	cn.mu.Lock()
	defer cn.mu.Unlock()

	for _, cnv := range cn.open {
		if err := cnv.SendAsync(req, nil); err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func (cn *Connector) Layer() Layer {
	cn.mu.Lock()
	defer cn.mu.Unlock()

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

			cn.mu.Lock()
			cn.open[remote.ID()] = cnv
			cn.mu.Unlock()
		}
	}()

	return cn.layer.Connect(ls, &dialer{cn.layer, cn.rp.IPFS})
}

func (cn *Connector) Disconnect() error {
	var errs util.Errors

	cn.mu.Lock()
	defer cn.mu.Unlock()

	for _, cnv := range cn.open {
		peer := cnv.Peer()

		delete(cn.open, peer.ID())
		delete(cn.heartbeat, peer.ID())

		if err := cnv.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	return cn.layer.Disconnect()
}

func (cn *Connector) Close() error {
	return cn.Disconnect()
}

func (cn *Connector) IsInOnlineMode() bool {
	cn.mu.Lock()
	defer cn.mu.Unlock()

	return cn.layer.IsInOnlineMode()
}
