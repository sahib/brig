package catfs

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/sahib/brig/catfs/mio"
	"github.com/sahib/brig/catfs/mio/chunkbuf"
	h "github.com/sahib/brig/util/hashlib"
)

// ErrNoSuchHash should be returned whenever the backend is unable
// to find an object referenced to by this hash.
type ErrNoSuchHash struct {
	what h.Hash
}

func (eh ErrNoSuchHash) Error() string {
	return fmt.Sprintf("No such hash: %s", eh.what.B58String())
}

// FsBackend is the interface that needs to be implemented by the data
// management layer.
type FsBackend interface {
	// Cat should find the object referenced to by `hash` and
	// make its data available as mio.Stream.
	Cat(hash h.Hash) (mio.Stream, error)

	// Add should read all data in `r` and return the hash under
	// which it can be accessed on later.
	Add(r io.Reader) (h.Hash, error)

	// Pin gives the object at `hash` a "pin".
	// (i.e. it marks the file to be stored indefinitely in local storage)
	// When pinning an explicit pin with an implicit pin, the explicit pin
	// will stay. Upgrading from implicit to explicit is possible though.
	Pin(hash h.Hash, explicit bool) error

	// Unpin removes a previously added pin.
	// If an object is already unpinned this is a no op.
	Unpin(hash h.Hash, explicit bool) error

	// IsPinned return two boolean values:
	// - If the first value is true, the file is pinned.
	// - If the second value is true, it was explicitly pinned by the user.
	IsPinned(hash h.Hash) (bool, bool, error)
}

// MemFsBackend is a mock structure that implements FsBackend.
type MemFsBackend struct {
	data map[string][]byte
	pins map[string]*memPinInfo
}

// NewMemFsBackend returns a MemFsBackend (useful for writing tests)
func NewMemFsBackend() *MemFsBackend {
	return &MemFsBackend{
		data: make(map[string][]byte),
		pins: make(map[string]*memPinInfo),
	}
}

// Cat implements FsBackend.Cat by querying memory.
func (mb *MemFsBackend) Cat(hash h.Hash) (mio.Stream, error) {
	data, ok := mb.data[hash.B58String()]
	if !ok {
		return nil, ErrNoSuchHash{hash}
	}

	return chunkbuf.NewChunkBuffer(data), nil
}

// Add implements FsBackend.Add by storing the data in memory.
func (mb *MemFsBackend) Add(r io.Reader) (h.Hash, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	hash := h.SumSHA256(data)
	mb.data[hash.B58String()] = data
	return hash, nil
}

type memPinInfo struct {
	isPinned, isExplicit bool
}

// Pin implements FsBackend.Pin by storing a marker in memory.
func (mb *MemFsBackend) Pin(hash h.Hash, explicit bool) error {
	_, isExplicit, err := mb.IsPinned(hash)
	if err != nil {
		return err
	}

	if !explicit && isExplicit {
		// should not overwrite.
		return nil
	}

	mb.pins[hash.B58String()] = &memPinInfo{true, explicit}
	return nil
}

// Unpin implements FsBackend.Unpin by removing a marker in memory.
func (mb *MemFsBackend) Unpin(hash h.Hash, explicit bool) error {
	isPinned, isExplicit, err := mb.IsPinned(hash)
	if err != nil {
		return err
	}

	if !isPinned {
		return nil
	}

	if !explicit && isExplicit {
		return nil
	}

	mb.pins[hash.B58String()] = &memPinInfo{false, false}
	return nil
}

// IsPinned implements FsBackend.IsPinned by querying a marker in memory.
func (mb *MemFsBackend) IsPinned(hash h.Hash) (bool, bool, error) {
	info, ok := mb.pins[hash.B58String()]
	if !ok {
		return false, false, nil
	}

	return info.isPinned, info.isExplicit, nil
}
