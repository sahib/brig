package ipfs

import (
	"fmt"
	cid "gx/ipfs/QmPSQnBKM9g7BaUcZCvswUJVscQ1ipjmwxN5PXCjkp9EQ7/go-cid"
	mh "gx/ipfs/QmPnFwZ2JXKnXgMw8CdBPxn7FWh6LLdjUjxV1fKHuJnkr8/go-multihash"

	coreiface "github.com/ipfs/go-ipfs/core/coreapi/interface"

	h "github.com/sahib/brig/util/hashlib"
)

// Pin does the same as `ipfs pin add <hash>`
// XXX: Issue: brig assumes it is the only instance of pin/unpin something
//      of their nodes. Otherwise doable.
func (nd *Node) Pin(hash h.Hash) error {
	// Lock the store:
	defer nd.ipfsNode.Blockstore.PinLock().Unlock()

	p, err := coreiface.ParsePath(hash.B58String())
	if err != nil {
		return err
	}

	dagnode, err := nd.api.ResolveNode(nd.ctx, p)
	if err != nil {
		return fmt.Errorf("pin: %s", err)
	}

	err = nd.ipfsNode.Pinning.Pin(nd.ctx, dagnode, true)
	if err != nil {
		return fmt.Errorf("pin: %s", err)
	}

	return nd.ipfsNode.Pinning.Flush()
}

// Unpin does the same as `ipfs pin rm <hash>`
func (nd *Node) Unpin(hash h.Hash) error {
	// Lock the store:
	defer nd.ipfsNode.Blockstore.PinLock().Unlock()

	cid := cid.NewCidV0(mh.Multihash(hash))
	if err := nd.ipfsNode.Pinning.Unpin(nd.ctx, cid, true); err != nil {
		return err
	}

	return nd.ipfsNode.Pinning.Flush()
}

// IsPinned returns true if `hash` is pinned.
func (nd *Node) IsPinned(hash h.Hash) (bool, error) {
	cid := cid.NewCidV0(mh.Multihash(hash))
	mode, _, err := nd.ipfsNode.Pinning.IsPinned(cid)
	if err != nil {
		return false, err
	}

	switch mode {
	case "direct", "internal":
		return true, nil
	case "indirect", "recursive":
		return true, nil
	default:
		return false, nil
	}
}
