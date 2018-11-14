package ipfs

import (
	"context"
	"net"
	"sync"
	"time"

	netBackend "github.com/sahib/brig/net/backend"

	ipfspeer "gx/ipfs/QmTRhk7cgjUf2gfQ3p2M9KPECNZEW9XUrmHcFCgog4cPgB/go-libp2p-peer"
	p2pnet "gx/ipfs/QmXuRkCR7BNQa9uqfpTiFWsTQLzmTWYg91Ja1w95gnqb6u/go-libp2p-net"

	pstore "gx/ipfs/QmTTJcDL3gsnGDALjh2fDGg1onGRUdVgNL2hU2WEZcVrMX/go-libp2p-peerstore"
	ping "gx/ipfs/QmUDTcnDp2WssbmiDLC6aYurUeyt7QeRakHUQMxA2mZ5iB/go-libp2p/p2p/protocol/ping"
	pro "gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
)

type streamAddr struct {
	protocol string
	peer     string
}

func (sa *streamAddr) Network() string {
	return sa.protocol
}

func (sa *streamAddr) String() string {
	return sa.peer
}

type stdStream struct {
	p2pnet.Stream
}

func (st *stdStream) LocalAddr() net.Addr {
	return &streamAddr{
		protocol: string(st.Protocol()),
		peer:     st.Stream.Conn().LocalPeer().Pretty(),
	}
}

func (st *stdStream) RemoteAddr() net.Addr {
	return &streamAddr{
		protocol: string(st.Protocol()),
		peer:     st.Stream.Conn().RemotePeer().Pretty(),
	}
}

func (nd *Node) Dial(peerHash, protocol string) (net.Conn, error) {
	if !nd.IsOnline() {
		return nil, ErrIsOffline
	}

	peerID, err := ipfspeer.IDB58Decode(peerHash)
	if err != nil {
		return nil, err
	}

	peerInfo := pstore.PeerInfo{ID: peerID}
	if err := nd.ipfsNode.PeerHost.Connect(nd.ctx, peerInfo); err != nil {
		return nil, err
	}

	protoId := pro.ID(protocol)
	stream, err := nd.ipfsNode.PeerHost.NewStream(nd.ctx, peerID, protoId)
	if err != nil {
		return nil, err
	}

	return &stdStream{Stream: stream}, nil
}

/////////////////////////////
// LISTENER IMPLEMENTATION //
/////////////////////////////

type Listener struct {
	self     string
	protocol string

	conCh  chan p2pnet.Stream
	ctx    context.Context
	cancel func()
}

func (nd *Node) Listen(protocol string) (net.Listener, error) {
	if !nd.IsOnline() {
		return nil, ErrIsOffline
	}

	ctx, cancel := context.WithCancel(nd.ctx)
	lst := &Listener{
		protocol: protocol,
		self:     nd.ipfsNode.Identity.String(),
		conCh:    make(chan p2pnet.Stream),
		ctx:      ctx,
		cancel:   cancel,
	}

	protoId := pro.ID(protocol)
	nd.ipfsNode.PeerHost.SetStreamHandler(protoId, func(stream p2pnet.Stream) {
		select {
		case lst.conCh <- stream:
		case <-ctx.Done():
			stream.Close()
		}
	})

	return lst, nil
}

func (lst *Listener) Accept() (net.Conn, error) {
	select {
	case <-lst.ctx.Done():
		return nil, nil
	case stream := <-lst.conCh:
		return &stdStream{Stream: stream}, nil
	}
}

func (lst *Listener) Addr() net.Addr {
	return &streamAddr{
		protocol: lst.protocol,
		peer:     lst.self,
	}
}

func (lst *Listener) SetDeadline(t time.Time) error {
	// NOTE: Implement, if we need a stoppable quit.
	return nil
}

func (lst *Listener) Close() error {
	lst.cancel()
	return nil
}

type Pinger struct {
	lastSeen  time.Time
	roundtrip time.Duration
	cancel    func()
	mu        sync.Mutex
	isClosed  bool
}

func (p *Pinger) LastSeen() time.Time {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.lastSeen
}

func (p *Pinger) Roundtrip() time.Duration {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.roundtrip
}

func (p *Pinger) Err() error {
	return nil
}

func (p *Pinger) Close() error {
	p.cancel()

	p.mu.Lock()
	defer p.mu.Unlock()

	p.isClosed = true
	return nil
}

// Ping returns a new Pinger. It can be used to
// query the time the remote was last seen. It will be
// constantly updated until close is called on it.
func (nd *Node) Ping(addr string) (netBackend.Pinger, error) {
	if !nd.IsOnline() {
		return nil, ErrIsOffline
	}

	peerID, err := ipfspeer.IDB58Decode(addr)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(nd.ctx)
	pingCh, err := ping.Ping(ctx, nd.ipfsNode.PeerHost, peerID)
	if err != nil {
		// If peer cannot be rached, we will bail out here.
		cancel()
		return nil, err
	}

	pinger := &Pinger{
		lastSeen: time.Now(),
		cancel:   cancel,
	}

	// pingCh will also be closed by ipfs's Ping().
	// This will happen once cancel() is called.
	go func() {
		for roundtrip := range pingCh {
			pinger.mu.Lock()
			pinger.roundtrip = roundtrip
			pinger.lastSeen = time.Now()

			isClosed := pinger.isClosed
			pinger.mu.Unlock()

			if isClosed {
				break
			}

			time.Sleep(5 * time.Second)
		}
	}()

	return pinger, nil
}
