package ipfsutil

import (
	"io"

	"github.com/disorganizer/brig/interfaces"
	h "github.com/disorganizer/brig/util/hashlib"
)

type IpfsBackend struct {
	*Node
}

func (ib *IpfsBackend) Add(r io.Reader) (*h.Hash, error) {
	mh, err := Add(ib.Node, r)
	if err != nil {
		return nil, err
	}

	return &h.Hash{mh}, nil
}

func (ib *IpfsBackend) Cat(hash *h.Hash) (interfaces.OutStream, error) {
	return Cat(ib.Node, hash.Multihash)
}

func (ib *IpfsBackend) Pin(hash *h.Hash) error {
	return ib.Node.Pin(hash.Multihash)
}

func (ib *IpfsBackend) Unpin(hash *h.Hash) error {
	return ib.Node.Unpin(hash.Multihash)
}

func (ib *IpfsBackend) IsPinned(hash *h.Hash) (bool, error) {
	return ib.Node.IsPinned(hash.Multihash)
}
