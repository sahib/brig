package fuse

import (
	"bazil.org/fuse/fs"
	"github.com/disorganizer/brig/util/trie"
)

type FS struct {
	Trie trie.Trie
}

func (f *FS) Root() (fs.Node, error) {
	return &Dir{Node: f.Trie.Root(), fs: f}, nil
}
