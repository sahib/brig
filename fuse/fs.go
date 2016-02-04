package fuse

import (
	"bazil.org/fuse/fs"
	"github.com/disorganizer/brig/store"
)

// FS represents a Filesystem.
type FS struct {
	Store *store.Store
}

// Root returns the topmost directory node.
// It will have the path "/".
func (sys *FS) Root() (fs.Node, error) {
	return &Dir{File: sys.Store.Root, fs: sys}, nil
}
