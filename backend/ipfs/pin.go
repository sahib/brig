package ipfs

import (
	cid "gx/ipfs/QmNp85zy9RLrQ5oQD4hPyS39ezrrXpcaa7R4Y9kxdWQLLQ/go-cid"
	mh "gx/ipfs/QmU9a9NV9RdPNwZQDYd5uKsm6N6LJLSvLbywDDYFbaaC6P/go-multihash"

	h "github.com/disorganizer/brig/util/hashlib"

	ipfspath "github.com/ipfs/go-ipfs/path"

	"github.com/ipfs/go-ipfs/core/corerepo"
)

func (nd *Node) Pin(hash h.Hash) error {
	defer nd.ipfsNode.Blockstore.PinLock().Unlock()

	path, err := ipfspath.ParsePath(hash.B58String())
	if err != nil {
		return err
	}

	paths := []string{path.String()}
	if _, err := corerepo.Pin(nd.ipfsNode, nd.ctx, paths, true); err != nil {
		return err
	}

	return nil
}

func (nd *Node) Unpin(hash h.Hash) error {
	// TODO: Check if this works.
	path, err := ipfspath.ParsePath(hash.B58String())
	if err != nil {
		return err
	}

	paths := []string{path.String()}
	if _, err := corerepo.Unpin(nd.ipfsNode, nd.ctx, paths, true); err != nil {
		return err
	}

	return nil
}

func (nd *Node) IsPinned(hash h.Hash) (bool, error) {
	cid := cid.NewCidV0(mh.Multihash(hash))
	_, isPinned, err := nd.ipfsNode.Pinning.IsPinned(cid)
	if err != nil {
		return false, err
	}

	return isPinned, nil
}
