package main

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/sahib/brig/catfs/mio"
	h "github.com/sahib/brig/util/hashlib"
	log "github.com/sirupsen/logrus"
)

// TmpFsBackend is a mock structure that implements FsBackend.
type TmpFsBackend struct {
	root string
}

type streamWrapper struct {
	*os.File
}

func (sw streamWrapper) WriteTo(w io.Writer) (n int64, err error) {
	return io.Copy(w, sw.File)
}

// NewTmpFsBackend returns a TmpFsBackend (useful for writing tests)
func NewTmpFsBackend(root string) (*TmpFsBackend, error) {
	if err := os.MkdirAll(root, 0700); err != nil {
		return nil, err
	}
	return &TmpFsBackend{
		root: root,
	}, nil
}

// Cat implements FsBackend.Cat by querying memory.
func (tb *TmpFsBackend) Cat(hash h.Hash) (mio.Stream, error) {
	path := filepath.Join(tb.root, hash.B58String())

	/* #nosec */
	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return streamWrapper{fd}, nil
}

// Add implements FsBackend.Add by storing the data in memory.
func (tb *TmpFsBackend) Add(r io.Reader) (h.Hash, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	hash := h.SumWithBackendHash(data)
	path := filepath.Join(tb.root, hash.B58String())
	if err := ioutil.WriteFile(path, data, 0600); err != nil {
		return nil, err
	}

	return hash, nil
}

// Pin implements FsBackend.Pin by storing a marker in memory.
func (tb *TmpFsBackend) Pin(hash h.Hash) error {
	path := filepath.Join(tb.root, hash.B58String()+"-pin")
	return ioutil.WriteFile(path, []byte{}, 0600)
}

// Unpin implements FsBackend.Unpin by removing a marker in memory.
func (tb *TmpFsBackend) Unpin(hash h.Hash) error {
	path := filepath.Join(tb.root, hash.B58String()+"-pin")
	if err := os.Remove(path); err != nil {
		log.Debugf("unpin failed: %v", err)
	}

	return nil
}

// IsPinned implements FsBackend.IsPinned by querying a marker in memory.
func (tb *TmpFsBackend) IsPinned(hash h.Hash) (bool, error) {
	path := filepath.Join(tb.root, hash.B58String()+"-pin")
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// IsCached implements FsBackend.IsCached by always returning true.
func (tb *TmpFsBackend) IsCached(hash h.Hash) (bool, error) {
	return true, nil
}

// CachedSize implements FsBackend.CachedSize by returning file size
func (tb *TmpFsBackend) CachedSize(hash h.Hash) (uint64, error) {
	path := filepath.Join(tb.root, hash.B58String())

	fi, err := os.Stat(path)
	if err != nil {
		return uint64(1<<64 - 1), err // MaxUint64 indicates unknown
	}
	return uint64(fi.Size()), nil
}

