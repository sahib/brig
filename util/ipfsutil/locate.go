package ipfsutil

import (
	"time"

	"github.com/disorganizer/brig/util/security"
	"github.com/ipfs/go-ipfs/blocks/key"
	"github.com/ipfs/go-ipfs/core/commands"
	ma "github.com/jbenet/go-multiaddr"
	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	ipdht "github.com/ipfs/go-ipfs/routing/dht"
	gmh "github.com/jbenet/go-multihash"
)

// PeerInfo holds the addresses and id of a peer.
//
// RANT: This is very much the same as ipfs's PeerInfo.
//       The reason we have this here, because we don't buy in their
//       gx-package-bundling bullshit. It's not possible to convert a
//       jbenet/go-mulithash.Multihash to a gx/$hash/multihash.Multihash or the
//       other way round, effectively making the use of ipfs as library much
//       harder.
type PeerInfo struct {
	ID     string
	Addrs  []ma.Multiaddr
	PubKey security.PubKey
}

// locate finds the object pointed to by `hash`. it will wait
// for max `timeout` duration if it got less than `n` items in that time.
// if `n` is less than 0, all reachable peers that have `hash` will be returned.
// if `n` is 0, locate will return immeditately.
// this operation requires online-mode.
func Locate(node *Node, hash gmh.Multihash, n int, t time.Duration) ([]*PeerInfo, error) {
	if n == 0 {
		return []*PeerInfo{}, nil
	}

	// Note: Do not use Maxint32. That makes ipfs allocate
	//       a whole lot of memory. Just assume that 100 is fine.
	if n < 0 {
		n = 100
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
