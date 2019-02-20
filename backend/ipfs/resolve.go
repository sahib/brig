package ipfs

import (
	"context"

	blocks "gx/ipfs/QmRcHuYzAyswytBuMF78rj3LTChYszomRFXNg4685ZN1WM/go-block-format"

	cid "gx/ipfs/QmPSQnBKM9g7BaUcZCvswUJVscQ1ipjmwxN5PXCjkp9EQ7/go-cid"

	log "github.com/Sirupsen/logrus"
	"github.com/sahib/brig/net/peer"
	h "github.com/sahib/brig/util/hashlib"
)

// addBlock creates a new block with `data`.
// The hash of the data is returned.
// It is no error if the block already exists.
func (nd *Node) addBlock(data []byte) (h.Hash, error) {
	block := blocks.NewBlock(data)
	if err := nd.ipfsNode.Blocks.AddBlock(block); err != nil {
		return nil, err
	}

	return h.Hash(block.Cid().Hash()), nil
}

// PublishName makes the string `name` known in the network.
// XXX: doable.
func (nd *Node) PublishName(name string) error {
	// Build all names under we can find this node:
	fullName := "brig:" + string(name)
	hash, err := nd.addBlock([]byte(fullName))
	if err != nil {
		return err
	}

	log.Debugf("Publishing name `%s` as `%s`", name, hash.B58String())
	return nil
}

// Identity returns the base58 encoded id of the own ipfs node.
func (nd *Node) Identity() (peer.Info, error) {
	return peer.Info{
		Name: "ipfs", // The name is not used currently.
		Addr: nd.ipfsNode.Identity.Pretty(),
	}, nil
}

// ResolveName finds the object pointed to by `hash`. it will wait
// for max `timeout` duration if it got less than `n` items in that time.
// if `n` is less than 0, all reachable peers that have `hash` will be returned.
// if `n` is 0, locate will return immeditately.
// this operation requires online-mode.
func (nd *Node) ResolveName(ctx context.Context, name string) ([]peer.Info, error) {
	if !nd.IsOnline() {
		return nil, ErrIsOffline
	}

	name = "brig:" + name
	hash := h.Hash(blocks.NewBlock([]byte(name)).Multihash())
	log.Debugf("Trying to locate %v (hash: %v)", name, hash.B58String())

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
