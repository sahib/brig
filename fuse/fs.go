package fuse

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/store"
)

// FS represents a Filesystem.
type FS struct {
	Store *store.Store
}

// Root returns the topmost directory node.
// It will have the path "/".
func (fs *FS) Root() (fs.Node, error) {
	root, err := fs.Store.Root()
	if err != nil {
		log.Warningf("fs: failed to retrieve root: %v", err)
		return nil, fuse.EIO
	}

	return &Dir{Directory: root, fs: fs}, nil
}
