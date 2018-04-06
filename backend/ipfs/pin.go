package ipfs

import (
	"fmt"
	mh "gx/ipfs/QmZyZDi491cCNTLfAhwcaDii2Kg4pwKRkhqQzURGDvY6ua/go-multihash"
	cid "gx/ipfs/QmcZfnkapfECQGcLZaf9B79NRg7cRa9EnZh4LSbkCzwNvY/go-cid"

	h "github.com/sahib/brig/util/hashlib"

	"github.com/ipfs/go-ipfs/pin"
)

func (nd *Node) Pin(hash h.Hash, explicit bool) error {
	// Lock the store:
	defer nd.ipfsNode.Blockstore.PinLock().Unlock()

	// This is a hack. In order to store implicit pins we use
	// "recursive" as type, while we use "Direct" for explicit pins.
	// Since we always use single files and blocks as hashes, there
	// (should?) is no difference between them.
	pinMode := pin.Recursive
	if explicit {
		pinMode = pin.Direct
	}

	cid := cid.NewCidV0(mh.Multihash(hash))
	mode, _, err := nd.ipfsNode.Pinning.IsPinned(cid)
	if err != nil {
		return err
	}

	fmt.Println("PINNING", hash, "with", pinMode, "was", mode)

	switch mode {
	case "direct", "internal":
		// It's already explicit.
		return nil
	case "indirect", "recursive":
		if !explicit {
			// I's pinned implicitly already.
			return nil
		}

		// Explicit pin requested, unpin previous implicit.
		if err := nd.ipfsNode.Pinning.Unpin(nd.ctx, cid, true); err != nil {
			return err
		}
	}

	nd.ipfsNode.Pinning.PinWithMode(cid, pinMode)
	return nd.ipfsNode.Pinning.Flush()
}

func (nd *Node) Unpin(hash h.Hash, explicit bool) error {
	// Lock the store:
	defer nd.ipfsNode.Blockstore.PinLock().Unlock()

	cid := cid.NewCidV0(mh.Multihash(hash))
	mode, _, err := nd.ipfsNode.Pinning.IsPinned(cid)
	if err != nil {
		return err
	}

	switch mode {
	case "direct", "internal":
		if !explicit {
			// Pinned explicit, but no explicit unpin. Keep it.
			return nil
		}
	}

	// Explicit pin requested, unpin previous implicit.
	if err := nd.ipfsNode.Pinning.Unpin(nd.ctx, cid, true); err != nil {
		return err
	}

	return nd.ipfsNode.Pinning.Flush()
}

func (nd *Node) IsPinned(hash h.Hash) (bool, bool, error) {
	cid := cid.NewCidV0(mh.Multihash(hash))
	mode, _, err := nd.ipfsNode.Pinning.IsPinned(cid)
	if err != nil {
		return false, false, err
	}

	switch mode {
	case "direct", "internal":
		return true, true, nil
	case "indirect", "recursive":
		return true, false, nil
	default:
		return false, false, nil
	}
}

func (nd *Node) ExplicitPins() ([]h.Hash, error) {
	hashes := []h.Hash{}
	for _, cid := range nd.ipfsNode.Pinning.DirectKeys() {
		hashes = append(hashes, h.Hash(cid.Hash()))
	}

	return hashes, nil
}
