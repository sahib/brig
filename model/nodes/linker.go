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

	NextUID() uint64
}

////////////////////////////
// MOCKING IMPLEMENTATION //
////////////////////////////

// MockLinker is supposed to be used for testing.
type MockLinker struct {
	id_count uint64
	root     *Directory
	paths    map[string]Node
	hashes   map[string]Node
}

func NewMockLinker() *MockLinker {
	return &MockLinker{
		paths:  make(map[string]Node),
		hashes: make(map[string]Node),
	}
}

func (ml *MockLinker) Root() (*Directory, error) {
	if ml.root != nil {
		return ml.root, nil
	}

	root, err := NewEmptyDirectory(ml, nil, "")
	if err != nil {
		return nil, err
	}

	ml.root = root
	return root, nil
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

func (ml *MockLinker) NextUID() uint64 {
	ml.id_count++
	return ml.id_count
}

func (ml *MockLinker) MemSetRoot(root *Directory) {
	ml.root = root
}

func (ml *MockLinker) MemIndexSwap(nd Node, oldHash h.Hash) {
	delete(ml.hashes, oldHash.B58String())
	ml.AddNode(nd)
}

func (ml *MockLinker) AddNode(nd Node) {
	ml.hashes[nd.Hash().B58String()] = nd
	ml.paths[nd.Path()] = nd
}
