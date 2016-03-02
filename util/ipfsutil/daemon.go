package ipfsutil

import (
	log "github.com/Sirupsen/logrus"

	core "github.com/ipfs/go-ipfs/core"
	fsrepo "github.com/ipfs/go-ipfs/repo/fsrepo"

	"golang.org/x/net/context"
)

func createNode(ipfsPath string, online bool, ctx context.Context) (*core.IpfsNode, error) {
	// Basic ipfsnode setup
	r, err := fsrepo.Open(ipfsPath)
	if err != nil {
		log.Errorf("Unable to open repo `%s`: %v", ipfsPath, err)
		return nil, err
	}

	cfg := &core.BuildCfg{
		Repo:   r,
		Online: online,
	}

	nd, err := core.NewNode(ctx, cfg)
	nd.OnlineMode()
	if err != nil {
		return nil, err
	}

	return nd, nil
}

// New creates a new ipfs node manager.
// No daemon is started yet.
func New(ipfsPath string) *Node {
	return &Node{
		Path:     ipfsPath,
		ipfsNode: nil,
		Context:  nil,
		Cancel:   nil,
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
	nd, err := createNode(n.Path, true, ctx)

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
		nd, err := createNode(n.Path, false, ctx)

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
	if n.Cancel != nil {
		n.Cancel()
	}

	return n.ipfsNode.Close()
}
