package nodes

import (
	"fmt"

	h "github.com/disorganizer/brig/util/hashlib"
)

// Linker will tell a node how it relates to other nodes
// and gives it the ability to resolve other nodes by hash.
// Apart from that it gives the underlying linker implementation
// the possibility to be notified when a hash changes.
type Linker interface {
	Root() (*Directory, error)
	LookupNode(path string) (Node, error)
	NodeByHash(hash h.Hash) (Node, error)

	// MemIndexSwap should be called when
	// the hash of a node changes.
	MemIndexSwap(nd Node, oldHash h.Hash)

	// MemSetRoot should be called when the current root directory changed.
	MemSetRoot(root *Directory)

	NextUID() (uint64, error)
}

////////////////////////////
// MOCKING IMPLEMENTATION //
////////////////////////////

// MockLinker is supposed to be used for testing.
type MockLinker struct {
	id_count uint64
	paths    map[string]Node
	hashes   map[string]Node
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

func (ml *MockLinker) NextID() uint64 {
	return ml.id_count++
}

func (ml *MockLinker) MemSetRoot(nd Node, oldHash h.Hash)   { /* No-Op */ }
func (ml *MockLinker) MemIndexSwap(nd Node, oldHash h.Hash) { /* No-Op */ }
