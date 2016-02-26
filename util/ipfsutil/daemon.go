package ipfsutil

import (
	log "github.com/Sirupsen/logrus"

	core "github.com/ipfs/go-ipfs/core"
	fsrepo "github.com/ipfs/go-ipfs/repo/fsrepo"

	"golang.org/x/net/context"
)

// StartNode starts an ipfs node on the repo.
// If `online` is true, the node will try to connect to the global ipfs net.
func StartNode(ipfsPath string, online bool) (*Node, error) {
	// Basic ipfsnode setup
	r, err := fsrepo.Open(ipfsPath)
	if err != nil {
		log.Errorf("Unable to open repo `%s`: %v", ipfsPath, err)
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	cfg := &core.BuildCfg{
		Repo:   r,
		Online: online,
	}

	nd, err := core.NewNode(ctx, cfg)
	nd.OnlineMode()
	if err != nil {
		return nil, err
	}

	return &Node{
		IpfsNode: nd,
		Path:     ipfsPath,
		Context:  ctx,
		Cancel:   cancel,
	}, nil
}

func (n *Node) IsOnline() bool {
	return n.IpfsNode.OnlineMode()
}

// Close shuts down the ipfs node.
// It may not be used afterwards.
func (n *Node) Close() error {
	return n.IpfsNode.Close()
}
