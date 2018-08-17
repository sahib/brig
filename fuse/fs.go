package fuse

import (
	"bazil.org/fuse/fs"
	"github.com/sahib/brig/catfs"
)

// Filesystem is the entry point to the fuse filesystem
type Filesystem struct {
	root string
	cfs  *catfs.FS
}

// Root returns the topmost directory node.
// This depends on what the user choose to select,
// but usually it's "/".
func (fs *Filesystem) Root() (fs.Node, error) {
	return &Directory{path: fs.root, cfs: fs.cfs}, nil
}
