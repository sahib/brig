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
	"github.com/disorganizer/brig/util/security"
)

var (
	ErrListenerWasClosed = errors.New("Listener was closed")
)

type Connector struct {
	layer Layer

	// Open repo. required for answering requests.
	// (might be nil for tests if no handlers are tested)
	rp *repo.Repository

	// Conversation pool handling.
	cp *conversationPool
}

type authTunnel struct {
	priv security.PrivKey
	pub  security.PubKey
}

func (at *authTunnel) Encrypt(data []byte) ([]byte, error) {
	// TODO: use keys.
	// return at.pub.Encrypt(data)
	return data, nil
}

func (at *authTunnel) Decrypt(data []byte) ([]byte, error) {
	// TODO: use keys.
	// return at.priv.Decrypt(data)
	return data, nil
}

type authManager struct {
	node *ipfsutil.Node
}

func (am *authManager) TunnelFor(hash string) (security.Tunnel, error) {
	pub, err := am.node.PublicKeyFor(hash)
	if err != nil {
		return nil, err
	}

	priv, err := am.node.PrivateKey()
	if err != nil {
		return nil, err
	}

	return &authTunnel{priv, pub}, nil
}

type conversationPool struct {
	// Map of open conversations
	open map[id.ID]Conversation

	// Map from hash id to last seen timestamp
	heartbeat map[id.ID]*ipfsutil.Pinger

	// lock for `open`
	mu sync.Mutex

	rp *repo.Repository
}

func newConversationPool(rp *repo.Repository) *conversationPool {
	return &conversationPool{
		open:      make(map[id.ID]Conversation),
		heartbeat: make(map[id.ID]*ipfsutil.Pinger),
		rp:        rp,
	}
}

// Set add a conversation for a specific to the pool.
func (cp *conversationPool) Set(peer id.Peer, cnv Conversation) error {
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
func (cp *conversationPool) Iter() chan Conversation {
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
		for remote := range lf.rms.Iter() {
			if remote.Hash() == hash {
				return streamConn, nil
			}
		}

		log.Warningf("Denying incoming connection from `%s`", hash)
	}

	return nil, ErrListenerWasClosed
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
		cp:    newConversationPool(rp),
	}

	layer.SetAuthManager(&authManager{rp.IPFS})

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

	// Make sure we filter unauthorized incoming connections:
	filter := newListenerFilter(ls, cn.rp.Remotes)

	if err := cn.layer.Connect(filter, &dialer{cn.layer, cn.rp.IPFS}); err != nil {
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
