package ipfsutil

import (
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"

	core "github.com/ipfs/go-ipfs/core"
	fsrepo "github.com/ipfs/go-ipfs/repo/fsrepo"

	"golang.org/x/net/context"
)

var (
	// ErrTimeout is returned when ipfs takes longer than the supplied duration.
	ErrTimeout = errors.New("IPFS operation timed out")
	// ErrIsOffline is returned when an online operation was done offline.
	ErrIsOffline = errors.New("Node is offline")
)

func createNode(nd *Node, online bool, ctx context.Context) (*core.IpfsNode, error) {
	// `nd` only contains the prepopulated fields as in New().
	rp, err := fsrepo.Open(nd.Path)
	if err != nil {
		log.Errorf("Unable to open repo `%s`: %v", nd.Path, err)
		return nil, err
	}

	swarmAddrs := []string{
		fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", nd.SwarmPort),
		fmt.Sprintf("/ip6/::/tcp/%d", nd.SwarmPort),
	}

	if err := rp.SetConfigKey("Addresses.Swarm", swarmAddrs); err != nil {
		return nil, err
	}

	cfg := &core.BuildCfg{
		Repo:   rp,
		Online: online,
	}

	ipfsNode, err := core.NewNode(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return ipfsNode, nil
}

// New creates a new ipfs node manager.
// No daemon is started yet.
func New(ipfsPath string) *Node {
	return NewWithPort(ipfsPath, 4001)
}

func NewWithPort(ipfsPath string, swarmPort int) *Node {
	return &Node{
		Path:      ipfsPath,
		SwarmPort: swarmPort,
		ipfsNode:  nil,
		Context:   nil,
		Cancel:    nil,
	}
}

func (n *Node) IsOnline() bool {
	return n.ipfsNode != nil && n.ipfsNode.OnlineMode()
}

func (n *Node) Online() error {
	if n.IsOnline() {
		return nil
	}

	// close potential offline node:
	if n.ipfsNode != nil {
		if err := n.Close(); err != nil {
			return err
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	nd, err := createNode(n, true, ctx)

	if err != nil {
		return err
	}

	n.ipfsNode = nd
	n.Cancel = cancel
	n.Context = ctx
	return nil
}

func (n *Node) Offline() error {
	if !n.IsOnline() {
		return nil
	}

	if err := n.Close(); err != nil {
		return err
	}

	// Offline daemon will be started on next proc() call.
	n.ipfsNode = nil
	n.Cancel = nil
	n.Context = nil
	return nil
}

func (n *Node) proc() (*core.IpfsNode, error) {
	if n.IsOnline() {
		return n.ipfsNode, nil
	}

	if n.ipfsNode == nil {
		ctx, cancel := context.WithCancel(context.Background())
		nd, err := createNode(n, false, ctx)

		if err != nil {
			return nil, err
		}

		n.ipfsNode = nd
		n.Cancel = cancel
		n.Context = ctx
	}

	return n.ipfsNode, nil
}

// Close shuts down the ipfs node.
// It may not be used afterwards.
func (n *Node) Close() error {
	nd := n.ipfsNode
	if nd != nil {
		n.ipfsNode = nil
		return nd.Close()
	}

	if n.Cancel != nil {
		n.Cancel()
		n.Cancel = nil
	}
	return nil
}

// Identity returns the base58 encoded id of the own ipfs node.
func (n *Node) Identity() (string, error) {
	nd, err := n.proc()
	if err != nil {
		log.Warningf("ipfs identity: %v", err)
		return "", err
	}

	return nd.Identity.Pretty(), nil
}
