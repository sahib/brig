package ipfs

import (
	"io"

	log "github.com/Sirupsen/logrus"
	coreiface "github.com/ipfs/go-ipfs/core/coreapi/interface"
	coreunix "github.com/ipfs/go-ipfs/core/coreunix"
	"github.com/sahib/brig/catfs/mio"
	h "github.com/sahib/brig/util/hashlib"
)

type ipfsFile struct {
	coreiface.UnixfsFile
}

func (ipf *ipfsFile) WriteTo(w io.Writer) (int64, error) {
	return io.Copy(w, ipf)
}

// Cat returns an io.Reader that reads from ipfs.
func (nd *Node) Cat(hash h.Hash) (mio.Stream, error) {
	fpath, err := coreiface.ParsePath(hash.B58String())
	if err != nil {
		return nil, err
	}

	file, err := nd.api.Unixfs().Get(nd.ctx, fpath)
	if err != nil {
		return nil, err
	}

	return &ipfsFile{UnixfsFile: file}, nil
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
