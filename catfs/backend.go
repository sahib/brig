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
	Pin(hash h.Hash) error

	// Unpin removes a previously added pin.
	// If an object is already unpinned this is a no op.
	Unpin(hash h.Hash) error

	// IsPinned checks if the file is pinned.
	IsPinned(hash h.Hash) (bool, error)

	// IsCached checks if the file contents are available locally.
	IsCached(hash h.Hash) (bool, error)

	// CachedSize returns the backend size for a given hash
	// return value of MaxUint64 := uint64(1<<64 - 1) indicates that cachedSize is unknown
	CachedSize(hash h.Hash) (uint64, error)
}

// MemFsBackend is a mock structure that implements FsBackend.
type MemFsBackend struct {
	data map[string][]byte
	pins map[string]bool
}

// NewMemFsBackend returns a MemFsBackend (useful for writing tests)
func NewMemFsBackend() *MemFsBackend {
	return &MemFsBackend{
		data: make(map[string][]byte),
		pins: make(map[string]bool),
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

	hash := h.SumWithBackendHash(data)
	mb.data[hash.B58String()] = data
	return hash, nil
}

// Pin implements FsBackend.Pin by storing a marker in memory.
func (mb *MemFsBackend) Pin(hash h.Hash) error {
	mb.pins[hash.B58String()] = true
	return nil
}

// Unpin implements FsBackend.Unpin by removing a marker in memory.
func (mb *MemFsBackend) Unpin(hash h.Hash) error {
	mb.pins[hash.B58String()] = false
	return nil
}

// IsPinned implements FsBackend.IsPinned by querying a marker in memory.
func (mb *MemFsBackend) IsPinned(hash h.Hash) (bool, error) {
	isPinned, ok := mb.pins[hash.B58String()]
	if !ok {
		return false, nil
	}

	return isPinned, nil
}

// IsCached implements FsBackend.IsCached by checking if the file exists.
// If yes, the file is cached always.
func (mb *MemFsBackend) IsCached(hash h.Hash) (bool, error) {
	_, ok := mb.data[hash.B58String()]
	return ok, nil
}

// CachedSize implements FsBackend.CachedSize by returnig data size
// If yes, the file is cached always.
func (mb *MemFsBackend) CachedSize(hash h.Hash) (uint64, error) {
	data, ok := mb.data[hash.B58String()]
	// MaxUint64 indicates that cachedSize is unknown
	MaxUint64 := uint64(1<<64 - 1)
	if !ok {
		return MaxUint64, nil
	}
	return uint64(len(data)), nil
}

