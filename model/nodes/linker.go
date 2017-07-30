package nodes

import (
	"fmt"

	h "github.com/disorganizer/brig/util/hashlib"
)

// Linker will tell a node how it relates to other nodes
// and gives it the ability to resolve other nodes by hash.
type Linker interface {
	Root() (*Directory, error)
	LookupNode(path string) (Node, error)
	NodeByHash(hash h.Hash) (Node, error)

	MemIndexSwap(nd Node, oldHash h.Hash)
	MemSetRoot(root *Directory)
}

////////////////////////////
// MOCKING IMPLEMENTATION //
////////////////////////////

// MockLinker is supposed to be used for testing.
type MockLinker struct {
	paths  map[string]Node
	hashes map[string]Node
}

func NewMockLinker() *MockLinker {
	return &MockLinker{
		paths:  make(map[string]Node),
		hashes: make(map[string]Node),
	}
}

func (ml *MockLinker) LookupNode(path string) (Node, error) {
	if node, ok := ml.paths[path]; ok {
		return node, nil
	}

	return nil, NoSuchFile(path)
}

func (ml *MockLinker) NodeByHash(hash h.Hash) (Node, error) {
	if node, ok := ml.hashes[hash.B58String()]; ok {
		return node, nil
	}

	return nil, fmt.Errorf("No such hash")
}

func (ml *MockLinker) MemSetRoot(nd Node, oldHash h.Hash)   { /* No-Op */ }
func (ml *MockLinker) MemIndexSwap(nd Node, oldHash h.Hash) { /* No-Op */ }
