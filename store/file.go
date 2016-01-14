package store

import (
	"github.com/disorganizer/brig/util/trie"
	"github.com/jbenet/go-multihash"
)

// File represents a single file in the repository.
// It stores all metadata about it and links to the actual data.
type File struct {
	// Pointer for dynamic loading of bigger data:
	*trie.Node
	s *Store

	Size     FileSize
	Hash     multihash.Multihash
	IpfsHash multihash.Multihash
}

// New returns a file inside a repo.
// Path is relative to the repo root.
func New(path string) (*File, error) {
	// TODO:
	return nil, nil
}

// func (f *File) Open() (Stream, error) {
// 	// Get io.Reader from ipfs cat
// 	// Mask with decompressor
// 	// Mask with decrypter
// 	// return
// 	return nil, nil
// }
