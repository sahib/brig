package blockstore

import (
	"context"

	mh "gx/ipfs/QmPnFwZ2JXKnXgMw8CdBPxn7FWh6LLdjUjxV1fKHuJnkr8/go-multihash"
	blocks "gx/ipfs/QmTRCUvZLiir12Qr6MV3HKfKMHX8Nf1Vddn6t2g5nsQSb9/go-block-format"
	cid "gx/ipfs/QmapdYm1b22Frv3k17fqrBYTFRxwiaVJkB299Mfn33edeB/go-cid"
)

// idstore wraps a BlockStore to add support for identity hashes
type idstore struct {
	bs Blockstore
}

func NewIdStore(bs Blockstore) Blockstore {
	return &idstore{bs}
}

func extractContents(k *cid.Cid) (bool, []byte) {
	dmh, err := mh.Decode(k.Hash())
	if err != nil || dmh.Code != mh.ID {
		return false, nil
	}
	return true, dmh.Digest
}

func (b *idstore) DeleteBlock(k *cid.Cid) error {
	isId, _ := extractContents(k)
	if isId {
		return nil
	}
	return b.bs.DeleteBlock(k)
}

func (b *idstore) Has(k *cid.Cid) (bool, error) {
	isId, _ := extractContents(k)
	if isId {
		return true, nil
	}
	return b.bs.Has(k)
}

func (b *idstore) Get(k *cid.Cid) (blocks.Block, error) {
	isId, bdata := extractContents(k)
	if isId {
		return blocks.NewBlockWithCid(bdata, k)
	}
	return b.bs.Get(k)
}

func (b *idstore) Put(bl blocks.Block) error {
	isId, _ := extractContents(bl.Cid())
	if isId {
		return nil
	}
	return b.bs.Put(bl)
}

func (b *idstore) PutMany(bs []blocks.Block) error {
	toPut := make([]blocks.Block, 0, len(bs))
	for _, bl := range bs {
		isId, _ := extractContents(bl.Cid())
		if isId {
			continue
		}
		toPut = append(toPut, bl)
	}
	return b.bs.PutMany(toPut)
}

func (b *idstore) HashOnRead(enabled bool) {
	b.bs.HashOnRead(enabled)
}

func (b *idstore) AllKeysChan(ctx context.Context) (<-chan *cid.Cid, error) {
	return b.bs.AllKeysChan(ctx)
}
