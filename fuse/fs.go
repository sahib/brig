package fuse

import (
	"bazil.org/fuse/fs"
	"github.com/disorganizer/brig/util/trie"
)

type FS struct {
	trie trie.Trie
}

func (f *FS) Root() (fs.Node, error) {
	// TODO
	return nil, nil
}
