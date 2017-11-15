package ipfs

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	p2pnet "gx/ipfs/QmNa31VPzC561NWwRsJLE7nGYZYuuD2QfpK2b1q9BK54J1/go-libp2p-net"
	ipfspeer "gx/ipfs/QmXYjuNuxVzXKJCfWasQk1RqkhVLDM9jtUKhqc2WPQmFSB/go-libp2p-peer"

	pstore "gx/ipfs/QmPgDWmTmuzvP7QE5zwo1TmjbJme9pmZHNujB2453jkCTr/go-libp2p-peerstore"
	peer "gx/ipfs/QmXYjuNuxVzXKJCfWasQk1RqkhVLDM9jtUKhqc2WPQmFSB/go-libp2p-peer"
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

	peerID, err := peer.IDB58Decode(peerHash)
	if err != nil {
		return nil, err
	}

	peerInfo := pstore.PeerInfo{ID: peerID}
	fmt.Println("Connect")
	if err := nd.ipfsNode.PeerHost.Connect(nd.ctx, peerInfo); err != nil {
		return nil, err
	}

	protoId := pro.ID(protocol)
	fmt.Println("New stream")
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

func (lst *Listener) Close() error {
	lst.cancel()
	return nil
}

type Pinger struct {
	lastSeen  time.Time
	roundtrip time.Duration
	cancel    func()
	mu        sync.Mutex
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

func (p *Pinger) Close() error {
	p.cancel()
	return nil
}

// Ping returns a new Pinger. It can be used to
// query the time the remote was last seen. It will be
// constantly updated until close is called on it.
func (nd *Node) Ping(addr string) (*Pinger, error) {
	if !nd.IsOnline() {
		return nil, ErrIsOffline
	}

	peerID, err := ipfspeer.IDB58Decode(addr)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(nd.ctx)
	pingCh, err := nd.ipfsNode.Ping.Ping(ctx, peerID)
	if err != nil {
		// If peer cannot be rached, we will bail out here.
		return nil, err
	}

	pinger := &Pinger{
		lastSeen: time.Now(),
		cancel:   cancel,
	}

	go func() {
		for roundtrip := range pingCh {
			pinger.mu.Lock()
			pinger.roundtrip = roundtrip
			pinger.lastSeen = time.Now()
			pinger.mu.Unlock()
		}
	}()

	return pinger, nil
}
