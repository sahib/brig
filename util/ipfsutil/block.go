package ipfsutil

import (
	"errors"
	"math"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ipfs/go-ipfs/blocks"
	"github.com/ipfs/go-ipfs/blocks/key"
	"github.com/ipfs/go-ipfs/core/commands"
	ipdht "github.com/ipfs/go-ipfs/routing/dht"
	ma "github.com/jbenet/go-multiaddr"
	gmh "github.com/jbenet/go-multihash"
	"golang.org/x/net/context"
)

var (
	// ErrTimeout is returned when ipfs takes longer than the supplied duration.
	ErrTimeout = errors.New("IPFS operation timed out")
	// ErrIsOffline is returned when an online operation was done offline.
	ErrIsOffline = errors.New("Node is offline")
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
// It will timeout with ErrTimeout if the operation takes too long,
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
		return nil, ErrTimeout
	}

	if err != nil {
		return nil, err
	}

	return block.Data, nil
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

// PeerInfo holds the addresses and id of a peer.
//
// RANT: This is very much the same as ipfs's PeerInfo.
//       The reason we have this here, because we don't buy in their
//       gx-package-bundling bullshit. It's not possible to convert a
//       jbenet/go-mulithash.Multihash to a gx/$hash/multihash.Multihash or the
//       other way round, effectively making the use of ipfs as library much
//       harder.
type PeerInfo struct {
	ID    string
	Addrs []ma.Multiaddr
}

// Locate finds the object pointed to by `hash`. It will wait
// for max `timeout` duration if it got less than `n` items in that time.
// If `n` is less than 0, all reachable peers that have `hash` will be returned.
// If `n` is 0, Locate will return immeditately.
// This operation requires online-mode.
func Locate(node *Node, hash gmh.Multihash, n int, t time.Duration) ([]*PeerInfo, error) {
	if n == 0 {
		return []*PeerInfo{}, nil
	}

	if n < 0 {
		n = math.MaxInt32
	}

	if !node.IsOnline() {
		return nil, ErrIsOffline
	}

	nd, err := node.proc()
	if err != nil {
		log.Warningf("ipfs dht: %v", err)
		return nil, err
	}

	dht, ok := nd.Routing.(*ipdht.IpfsDHT)
	if !ok {
		return nil, commands.ErrNotDHT
	}

	ctx, cancel := context.WithTimeout(node.Context, t)
	defer cancel()

	k := key.B58KeyDecode(hash.B58String())
	peers := dht.FindProvidersAsync(ctx, k, n)
	infos := []*PeerInfo{}

	for info := range peers {
		// Converting equal struct into each other is my favourite thing.
		peerInfo := &PeerInfo{
			ID: info.ID.Pretty(),
		}

		for _, addr := range info.Addrs {
			peerInfo.Addrs = append(peerInfo.Addrs, ma.Cast(addr.Bytes()))
		}

		infos = append(infos, peerInfo)
	}

	return infos, nil
}

// TODO: needed?
// func LocateNTimes(node *Node, hash gmh.Multihash, n int, t time.Duration, times int) (*PeerInfo, error) {
// 	for i := 0; i < times; i++ {
// 		peers, err := Locate(node, hash, n, t)
// 		if err != nil && err != ErrTimeout {
// 			return nil, err
// 		}
//
// 		if len(peers) > 0 && len(peers[0].Addrs) > 0 {
// 			return peers[0], nil
// 		}
//
// 		time.Sleep(500 * time.Millisecond)
// 	}
//
// 	return nil, nil
//
// }
