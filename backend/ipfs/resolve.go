package ipfs

import (
	"context"
	"time"

	blocks "gx/ipfs/QmSn9Td7xgxm9EV7iEjTckpUWmWApggzPxu7eFGWkkpwin/go-block-format"

	cid "gx/ipfs/QmNp85zy9RLrQ5oQD4hPyS39ezrrXpcaa7R4Y9kxdWQLLQ/go-cid"
	u "gx/ipfs/QmSU6eubNdhXjFBJBSksTp8kv8YRub8mGAPv8tVJHmL2EU/go-ipfs-util"

	"github.com/sahib/brig/net/peer"
	"github.com/sahib/brig/util"
	h "github.com/sahib/brig/util/hashlib"
)

// addBlock creates a new block with `data`.
// The hash of the data is returned.
// It is no error if the block already exists.
func (nd *Node) addBlock(data []byte) (h.Hash, error) {
	block := blocks.NewBlock(data)
	k, err := nd.ipfsNode.Blocks.AddBlock(block)
	if err != nil {
		return nil, err
	}

	return h.Hash(k.Hash()), nil
}

// catBlock retuns the data stored in the block pointed to by `hash`.
// It will timeout with util.ErrTimeout if the operation takes too long,
// this includes querying for an non-existing hash.
//
// This operation works offline and online, but if the block is stored
// elsewhere on the net, node must be online to find the block.
func (nd *Node) catBlock(hash h.Hash, timeout time.Duration) ([]byte, error) {
	ctx, cancel := context.WithTimeout(nd.ctx, timeout)
	defer cancel()

	k, err := cid.Decode(hash.B58String())
	if err != nil {
		return nil, err
	}

	block, err := nd.ipfsNode.Blocks.GetBlock(ctx, k)
	if err == context.DeadlineExceeded {
		return nil, util.ErrTimeout
	}

	if err != nil {
		return nil, err
	}

	return block.RawData(), nil
}

func (nd *Node) PublishName(name string) error {
	// Build all names under we can find this node:
	fullName := "brig:" + string(name)
	_, err := nd.addBlock([]byte(fullName))
	return err
}

// Identity returns the base58 encoded id of the own ipfs node.
func (nd *Node) Identity() (peer.Info, error) {
	return peer.Info{
		Name: "", // TODO: Should we return something meaningful here?
		Addr: nd.ipfsNode.Identity.Pretty(),
	}, nil
}

// Locate finds the object pointed to by `hash`. it will wait
// for max `timeout` duration if it got less than `n` items in that time.
// if `n` is less than 0, all reachable peers that have `hash` will be returned.
// if `n` is 0, locate will return immeditately.
// this operation requires online-mode.
func (nd *Node) ResolveName(name string) ([]peer.Info, error) {
	if !nd.IsOnline() {
		return nil, ErrIsOffline
	}

	// In theory, we could do this also for domains.
	hash := u.Hash([]byte(name))

	ctx, cancel := context.WithTimeout(nd.ctx, 30*time.Second)
	defer cancel()

	k, err := cid.Decode(hash.B58String())
	if err != nil {
		return nil, err
	}

	peers := nd.ipfsNode.Routing.FindProvidersAsync(ctx, k, 10)
	infos := []peer.Info{}

	for info := range peers {
		// Converting equal struct into each other is my favourite thing.
		peerInfo := peer.Info{
			Addr: info.ID.Pretty(),
			Name: peer.Name(name),
		}

		infos = append(infos, peerInfo)
	}

	return infos, nil
}
