package ipfsutil

import (
	"github.com/ipfs/go-ipfs/blocks/key"
	"github.com/ipfs/go-ipfs/core/corerepo"
	path "github.com/ipfs/go-ipfs/path"
	"github.com/jbenet/go-multihash"
)

func (node *Node) Pin(hash multihash.Multihash) error {
	nd, err := node.proc()
	if err != nil {
		return err
	}

	defer nd.Blockstore.PinLock().Unlock()

	paths := []string{path.FromKey(key.Key(hash)).String()}
	if _, err := corerepo.Pin(nd, node.Context, paths, true); err != nil {
		return err
	}

	return nil
}

func (node *Node) Unpin(hash multihash.Multihash) error {
	nd, err := node.proc()
	if err != nil {
		return err
	}

	paths := []string{path.FromKey(key.Key(hash)).String()}
	if _, err := corerepo.Unpin(nd, node.Context, paths, true); err != nil {
		return err
	}

	return nil
}

func (node *Node) IsPinned(hash multihash.Multihash) (bool, error) {
	nd, err := node.proc()
	if err != nil {
		return false, err
	}

	_, isPinned, err := nd.Pinning.IsPinned(key.Key(hash))
	if err != nil {
		return false, err
	}

	return isPinned, nil
}
