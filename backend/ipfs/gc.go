package ipfs

import (
	cid "gx/ipfs/QmcZfnkapfECQGcLZaf9B79NRg7cRa9EnZh4LSbkCzwNvY/go-cid"

	corerepo "github.com/ipfs/go-ipfs/core/corerepo"
	h "github.com/sahib/brig/util/hashlib"
)

func (nd *Node) GC() ([]h.Hash, error) {
	gcOutChan := corerepo.GarbageCollectAsync(nd.ipfsNode, nd.ctx)
	killed := []h.Hash{}

	// CollectResult blocks until garbarge collection is finished:
	err := corerepo.CollectResult(nd.ctx, gcOutChan, func(k *cid.Cid) {
		killed = append(killed, h.Hash(k.Hash()))
	})

	if err != nil {
		return nil, err
	}

	return killed, nil
}
