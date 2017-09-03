package catfs

import (
	"io"

	"github.com/disorganizer/brig/catfs/db"
	"github.com/disorganizer/brig/interfaces"
	h "github.com/disorganizer/brig/util/hashlib"
)

type StorageBackend interface {
	Cat(hash *h.Hash) (interfaces.OutStream, error)
	Add(r io.Reader) (*h.Hash, error)
	Pin(hash *h.Hash) error
	Unpin(hash *h.Hash) error
	IsPinned(hash *h.Hash) (bool, error)
}

type Model struct {
	Storage  StorageBackend
	Database db.Database
}

func NewModel(path string, kv db.Database, store StorageBackend) (*Model, error) {
	return &Model{
		Storage:  store,
		Database: kv,
	}, nil
}

func (m *Model) Import(r io.Reader) error {
	return nil
}

func (m *Model) Export(w io.Writer) error {
	return nil
}
