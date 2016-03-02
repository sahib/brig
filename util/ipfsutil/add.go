package ipfsutil

import (
	"io"

	coreunix "github.com/ipfs/go-ipfs/core/coreunix"
	"github.com/jbenet/go-multihash"

	log "github.com/Sirupsen/logrus"
)

// Add reads `r` and adds it to ipfs.
// The resulting content hash is returned.
func Add(node *Node, r io.Reader) (multihash.Multihash, error) {
	nd, err := node.proc()
	if err != nil {
		return nil, err
	}

	hash, err := coreunix.Add(nd, r)
	if err != nil {
		log.Warningf("ipfs add: %v", err)
		return nil, err
	}

	return multihash.FromB58String(hash)
}
