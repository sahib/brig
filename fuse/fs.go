package fuse

import (
	"bazil.org/fuse/fs"
	"github.com/disorganizer/brig/store"
)

type FS struct {
	Store *store.Store
}

func (sys *FS) Root() (fs.Node, error) {
	return &Dir{File: sys.Store.Root, fs: sys}, nil
}
