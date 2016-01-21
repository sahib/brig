package fuse

import (
	"bazil.org/fuse/fs"
	"github.com/disorganizer/brig/store"
)

type FS struct {
	Store *store.Store
}

func (f *FS) Root() (fs.Node, error) {
	return &Dir{Node: f.Store.Trie.Root(), fs: f}, nil
}
