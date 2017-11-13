package ipfs

import (
	"context"
	"time"

	blocks "gx/ipfs/QmSn9Td7xgxm9EV7iEjTckpUWmWApggzPxu7eFGWkkpwin/go-block-format"

	cid "gx/ipfs/QmNp85zy9RLrQ5oQD4hPyS39ezrrXpcaa7R4Y9kxdWQLLQ/go-cid"

	"github.com/disorganizer/brig/net/peer"
	"github.com/disorganizer/brig/util"
	h "github.com/disorganizer/brig/util/hashlib"
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

func (nd *Node) PublishName(name peer.Name) error {
	// Build all names under we can find this node:
	fullName := "brig:" + string(name)
	userName := "brig:" + string(name.WithoutResource())
	domainId := "brig:" + string(name.Domain())

	if err := nd.addBlock([]byte(fullName)); err != nil {
		return err
	}

	if fullName != userName {
		if err := nd.addBlock([]byte(userName)); err != nil {
			return err
		}
	}

	if domainId != "" {
		if err := nd.addBlock([]byte(name)); err != nil {
			return err
		}
	}

	return nil
}

func (nd *Node) ResolveName(name peer.Name) ([]peer.Info, error) {
	return nil, nil
}

// Identity returns the base58 encoded id of the own ipfs node.
func (nd *Node) Identity() (string, error) {
	return nd.ipfsNode.Identity.Pretty(), nil
}

// Locate finds the object pointed to by `hash`. it will wait
// for max `timeout` duration if it got less than `n` items in that time.
// if `n` is less than 0, all reachable peers that have `hash` will be returned.
// if `n` is 0, locate will return immeditately.
// this operation requires online-mode.
func (nd *Node) ResolveName(name peer.Name) ([]peer.Info, error) {
	if n == 0 {
		return []*PeerInfo{}, nil
	}

	// Note: Do not use Maxint32. That makes ipfs allocate
	//       a whole lot of memory. Just assume that 100 is fine.
	if n < 0 {
		n = 100
	}

	if !nd.IsOnline() {
		return nil, ErrIsOffline
	}

	ctx, cancel := context.WithTimeout(nd.ctx, 30*time.Second)
	defer cancel()

	cid.Decode()

	k := key.B58KeyDecode(hash.B58String())
	peers := nd.ipfsNode.Routing.FindProvidersAsync(ctx, k, 5)
	infos := []*PeerInfo{}

	for info := range peers {
		// Converting equal struct into each other is my favourite thing.
		peerInfo := &PeerInfo{
			ID:     info.ID.Pretty(),
			PubKey: node.ipfsNode.Peerstore.PubKey(info.ID),
		}

		for _, addr := range info.Addrs {
			peerInfo.Addrs = append(peerInfo.Addrs, ma.Cast(addr.Bytes()))
		}

		infos = append(infos, peerInfo)
	}

	return infos, nil
}
