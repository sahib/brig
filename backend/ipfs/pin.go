package ipfs

import (
	mh "gx/ipfs/QmZyZDi491cCNTLfAhwcaDii2Kg4pwKRkhqQzURGDvY6ua/go-multihash"
	cid "gx/ipfs/QmcZfnkapfECQGcLZaf9B79NRg7cRa9EnZh4LSbkCzwNvY/go-cid"

	h "github.com/sahib/brig/util/hashlib"

	ipfspath "github.com/ipfs/go-ipfs/path"

	"github.com/ipfs/go-ipfs/core/corerepo"
	"github.com/ipfs/go-ipfs/pin"
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
	path, err := ipfspath.ParsePath(hash.B58String())
	if err != nil {
		return err
	}

	paths := []string{path.String()}
	if _, err := corerepo.Unpin(nd.ipfsNode, nd.ctx, paths, true); err != nil {
		if err != pin.ErrNotPinned {
			return err
		}

		return nil
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
