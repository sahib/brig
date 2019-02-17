// +build !windows

package fuse

import (
	"bazil.org/fuse/fs"
	log "github.com/Sirupsen/logrus"
)

const (
	enableDebugLogs = false
)

func debugLog(format string, args ...interface{}) {
	if enableDebugLogs {
		log.Debugf(format, args...)
	}
}

// Filesystem is the entry point to the fuse filesystem
type Filesystem struct {
	root string
	m    *Mount
}

// Root returns the topmost directory node.
// This depends on what the user choose to select,
// but usually it's "/".
func (fs *Filesystem) Root() (fs.Node, error) {
	return &Directory{path: fs.root, m: fs.m}, nil
}
