package ipfsutil

import (
	log "github.com/Sirupsen/logrus"

	core "github.com/ipfs/go-ipfs/core"
	fsrepo "github.com/ipfs/go-ipfs/repo/fsrepo"

	"golang.org/x/net/context"
)

// TODO: Close, Offline, Online functions for *Node

// StartNode starts an ipfs node on the repo.
func StartNode(ipfsPath string) (*Node, error) {
	// Basic ipfsnode setup
	r, err := fsrepo.Open(ipfsPath)
	if err != nil {
		log.Errorf("Unable to open repo `%s`: %v", ipfsPath, err)
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	cfg := &core.BuildCfg{
		Repo:   r,
		Online: false,
	}

	nd, err := core.NewNode(ctx, cfg)
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

func (n *Node) Close() error {
	return n.IpfsNode.Close()
}
