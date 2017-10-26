package fuse

import (
	"bazil.org/fuse/fs"
	"github.com/disorganizer/brig/catfs"
)

// Filesystem is the entry point to the fuse filesystem
type Filesystem struct {
	cfs *catfs.FS
}

// Root returns the topmost directory node.
// It will have the path "/".
func (fs *Filesystem) Root() (fs.Node, error) {
	return &Directory{path: "/", cfs: fs.cfs}, nil
}
