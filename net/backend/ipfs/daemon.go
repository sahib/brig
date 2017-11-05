package main

import (
	"errors"
	"fmt"
	p2pnet "gx/ipfs/QmNa31VPzC561NWwRsJLE7nGYZYuuD2QfpK2b1q9BK54J1/go-libp2p-net"
	pstore "gx/ipfs/QmPgDWmTmuzvP7QE5zwo1TmjbJme9pmZHNujB2453jkCTr/go-libp2p-peerstore"
	peer "gx/ipfs/QmXYjuNuxVzXKJCfWasQk1RqkhVLDM9jtUKhqc2WPQmFSB/go-libp2p-peer"
	pro "gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	"net"

	log "github.com/Sirupsen/logrus"

	core "github.com/ipfs/go-ipfs/core"
	fsrepo "github.com/ipfs/go-ipfs/repo/fsrepo"

	"golang.org/x/net/context"
)

var (
	// ErrIsOffline is returned when an online operation was done offline.
	ErrIsOffline = errors.New("Node is offline")
)

// Node remembers the settings needed for accessing the ipfs daemon.
type Node struct {
	Path      string
	SwarmPort int

	ipfsNode *core.IpfsNode
	ctx      context.Context
	cancel   context.CancelFunc
}

func createNode(path string, swarmPort int, ctx context.Context) (*core.IpfsNode, error) {
	rp, err := fsrepo.Open(path)
	if err != nil {
		log.Errorf("Unable to open repo `%s`: %v", path, err)
		return nil, err
	}

	swarmAddrs := []string{
		fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", swarmPort),
		fmt.Sprintf("/ip6/::/tcp/%d", swarmPort),
	}

	if err := rp.SetConfigKey("Addresses.Swarm", swarmAddrs); err != nil {
		return nil, err
	}

	cfg := &core.BuildCfg{
		Repo:   rp,
		Online: true,
	}

	ipfsNode, err := core.NewNode(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return ipfsNode, nil
}

// New creates a new ipfs node manager.
// No daemon is started yet.
func New(ipfsPath string) (*Node, error) {
	return NewWithPort(ipfsPath, 4001)
}

func NewWithPort(ipfsPath string, swarmPort int) (*Node, error) {
	ctx, cancel := context.WithCancel(context.Background())
	ipfsNode, err := createNode(ipfsPath, swarmPort, ctx)
	if err != nil {
		return nil, err
	}

	return &Node{
		Path:      ipfsPath,
		SwarmPort: swarmPort,
		ipfsNode:  ipfsNode,
		ctx:       ctx,
		cancel:    cancel,
	}, nil
}

func (nd *Node) IsOnline() bool {
	return nd.ipfsNode.OnlineMode()
}

// Close shuts down the ipfs node.
// It may not be used afterwards.
func (nd *Node) Close() error {
	nd.cancel()
	return nd.ipfsNode.Close()
}

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
	conCh  chan p2pnet.Stream
	ctx    context.Context
	cancel func()
}

func (nd *Node) Listen(protocol string) (*Listener, error) {
	ctx, cancel := context.WithCancel(nd.ctx)
	lst := &Listener{
		conCh:  make(chan p2pnet.Stream),
		ctx:    ctx,
		cancel: cancel,
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

func (lst *Listener) Close() error {
	lst.cancel()
	return nil
}
