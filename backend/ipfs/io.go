package ipfs

import (
	"io"

	log "github.com/Sirupsen/logrus"
	coreunix "github.com/ipfs/go-ipfs/core/coreunix"
	"github.com/sahib/brig/catfs/mio"
	h "github.com/sahib/brig/util/hashlib"
)

// Cat returns an io.Reader that reads from ipfs.
func (nd *Node) Cat(hash h.Hash) (mio.Stream, error) {
	reader, err := coreunix.Cat(nd.ctx, nd.ipfsNode, hash.B58String())
	if err != nil {
		log.Warningf("ipfs cat: %v", err)
		return nil, err
	}

	return reader, nil
}

// Add reads `r` and adds it to ipfs.
// The resulting content hash is returned.
func (nd *Node) Add(r io.Reader) (h.Hash, error) {
	hash, err := coreunix.Add(nd.ipfsNode, r)
	if err != nil {
		log.Warningf("ipfs add: %v", err)
		return nil, err
	}

	return h.FromB58String(hash)
}
