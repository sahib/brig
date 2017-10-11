package ipfsutil

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/util"
	"github.com/ipfs/go-ipfs/blocks"
	"github.com/ipfs/go-ipfs/blocks/key"
	gmh "github.com/jbenet/go-multihash"
	"golang.org/x/net/context"
)

// AddBlock creates a new block with `data`.
// The hash of the data is returned.
// It is no error if the block already exists.
func AddBlock(node *Node, data []byte) (gmh.Multihash, error) {
	nd, err := node.proc()
	if err != nil {
		log.Warningf("ipfs block-add: %v", err)
		return nil, err
	}

	block := blocks.NewBlock(data)
	k, err := nd.Blocks.AddBlock(block)

	if err != nil {
		return nil, err
	}

	mh, err := gmh.FromB58String(k.B58String())
	if err != nil {
		return nil, err
	}

	return mh, nil
}

// CatBlock retuns the data stored in the block pointed to by `hash`.
// It will timeout with util.ErrTimeout if the operation takes too long,
// this includes querying for an non-existing hash.
//
// This operation works offline and online, but if the block is stored
// elsewhere on the net, node must be online to find the block.
func CatBlock(node *Node, hash gmh.Multihash, timeout time.Duration) ([]byte, error) {
	nd, err := node.proc()
	if err != nil {
		log.Warningf("ipfs block-cat: %v", err)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(node.Context, timeout)
	defer cancel()

	k := key.B58KeyDecode(hash.B58String())
	block, err := nd.Blocks.GetBlock(ctx, k)
	if err == context.DeadlineExceeded {
		return nil, util.ErrTimeout
	}

	if err != nil {
		return nil, err
	}

	return block.Data(), nil
}

// DelBlock deletes the block pointed to by `hash`.
func DelBlock(node *Node, hash gmh.Multihash) error {
	nd, err := node.proc()
	if err != nil {
		log.Warningf("ipfs block-del: %v", err)
		return err
	}

	k := key.B58KeyDecode(hash.B58String())
	return nd.Blocks.DeleteBlock(k)
}
