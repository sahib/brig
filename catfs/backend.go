package catfs

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/disorganizer/brig/catfs/mio"
	"github.com/disorganizer/brig/catfs/mio/chunkbuf"
	h "github.com/disorganizer/brig/util/hashlib"
)

type ErrNoSuchHash struct {
	what h.Hash
}

func (eh ErrNoSuchHash) Error() string {
	return fmt.Sprintf("No such hash: %s", eh.what.B58String())
}

type FsBackend interface {
	Cat(hash h.Hash) (mio.Stream, error)
	Add(r io.Reader) (h.Hash, error)

	Pin(hash h.Hash) error
	Unpin(hash h.Hash) error
	IsPinned(hash h.Hash) (bool, error)
}

type MemFsBackend struct {
	data map[string][]byte
	pins map[string]bool
}

func NewMemFsBackend() *MemFsBackend {
	return &MemFsBackend{
		data: make(map[string][]byte),
		pins: make(map[string]bool),
	}
}

func (mb *MemFsBackend) Cat(hash h.Hash) (mio.Stream, error) {
	data, ok := mb.data[hash.B58String()]
	if !ok {
		return nil, ErrNoSuchHash{hash}
	}

	return chunkbuf.NewChunkBuffer(data), nil
}

func (mb *MemFsBackend) Add(r io.Reader) (h.Hash, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	hash := h.Sum(data)
	mb.data[hash.B58String()] = data
	return hash, nil
}

func (mb *MemFsBackend) Pin(hash h.Hash) error {
	mb.pins[hash.B58String()] = true
	return nil
}

func (mb *MemFsBackend) Unpin(hash h.Hash) error {
	mb.pins[hash.B58String()] = false
	return nil
}

func (mb *MemFsBackend) IsPinned(hash h.Hash) (bool, error) {
	isPinned, ok := mb.pins[hash.B58String()]
	return isPinned && ok, nil
}

var _ FsBackend = &MemFsBackend{}
