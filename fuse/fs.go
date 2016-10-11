package fuse

import (
	"bazil.org/fuse/fs"
	"github.com/disorganizer/brig/store"
)

// FS represents a Filesystem.
type Filesystem struct {
	Store *store.Store
}

// Root returns the topmost directory node.
// It will have the path "/".
func (filesystem *Filesystem) Root() (fs.Node, error) {
	return &Dir{path: "/", fsys: filesystem}, nil
}
