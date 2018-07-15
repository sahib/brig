package ipfs

import (
	"fmt"
	mh "gx/ipfs/QmPnFwZ2JXKnXgMw8CdBPxn7FWh6LLdjUjxV1fKHuJnkr8/go-multihash"
	cid "gx/ipfs/QmapdYm1b22Frv3k17fqrBYTFRxwiaVJkB299Mfn33edeB/go-cid"

	core "github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/path"
	"github.com/ipfs/go-ipfs/path/resolver"
	uio "github.com/ipfs/go-ipfs/unixfs/io"
	h "github.com/sahib/brig/util/hashlib"
)

func (nd *Node) Pin(hash h.Hash) error {
	// Lock the store:
	defer nd.ipfsNode.Blockstore.PinLock().Unlock()

	rslv := &resolver.Resolver{
		DAG:         nd.ipfsNode.DAG,
		ResolveOnce: uio.ResolveUnixfsOnce,
	}

	p, err := path.ParsePath(hash.B58String())
	if err != nil {
		return err
	}

	dagnode, err := core.Resolve(nd.ctx, nd.ipfsNode.Namesys, rslv, p)
	if err != nil {
		return fmt.Errorf("pin: %s", err)
	}

	err = nd.ipfsNode.Pinning.Pin(nd.ctx, dagnode, true)
	if err != nil {
		return fmt.Errorf("pin: %s", err)
	}

	return nd.ipfsNode.Pinning.Flush()
}

func (nd *Node) Unpin(hash h.Hash) error {
	// Lock the store:
	defer nd.ipfsNode.Blockstore.PinLock().Unlock()

	cid := cid.NewCidV0(mh.Multihash(hash))
	if err := nd.ipfsNode.Pinning.Unpin(nd.ctx, cid, true); err != nil {
		return err
	}

	return nd.ipfsNode.Pinning.Flush()
}

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
