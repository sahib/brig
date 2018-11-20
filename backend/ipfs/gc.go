package ipfs

import (
	cid "gx/ipfs/QmPSQnBKM9g7BaUcZCvswUJVscQ1ipjmwxN5PXCjkp9EQ7/go-cid"

	corerepo "github.com/ipfs/go-ipfs/core/corerepo"
	h "github.com/sahib/brig/util/hashlib"
)

// GC triggers IPFS' garbage collector.
// It returns the hashes of the collected nodes.
func (nd *Node) GC() ([]h.Hash, error) {
	gcOutChan := corerepo.GarbageCollectAsync(nd.ipfsNode, nd.ctx)
	killed := []h.Hash{}

	// CollectResult blocks until garbarge collection is finished:
	err := corerepo.CollectResult(nd.ctx, gcOutChan, func(k cid.Cid) {
		killed = append(killed, h.Hash(k.Hash()))
	})

	if err != nil {
		return nil, err
	}

	return killed, nil
}
