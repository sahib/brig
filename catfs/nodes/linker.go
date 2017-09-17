package nodes

import (
	"fmt"

	ie "github.com/disorganizer/brig/catfs/errors"
	h "github.com/disorganizer/brig/util/hashlib"
)

// Linker will tell a node how it relates to other nodes
// and gives it the ability to resolve other nodes by hash.
// Apart from that it gives the underlying linker implementation
// the possibility to be notified when a hash changes.
type Linker interface {
	// Root should return the current root directory.
	Root() (*Directory, error)

	// LookupNode should resolve `path` starting from the root directory.
	// TODO: Is LookupNode("/") the same as Root()?
	// If the path does not exist an error is returned and can be checked
	// with IsNoSuchFileError()
	LookupNode(path string) (Node, error)

	// NodeByHash resolves the hash to a specific node.
	// If the node does not exist, nil is returned.
	NodeByHash(hash h.Hash) (Node, error)

	// MemIndexSwap should be called when
	// the hash of a node changes.
	MemIndexSwap(nd Node, oldHash h.Hash)

	// MemSetRoot should be called when the current root directory changed.
	MemSetRoot(root *Directory)
}

////////////////////////////
// MOCKING IMPLEMENTATION //
////////////////////////////

// MockLinker is supposed to be used for testing.
// It simply holds all nodes in memory. New nodes should be added via AddNode.
type MockLinker struct {
	root   *Directory
	paths  map[string]Node
	hashes map[string]Node
}

// NewMockLinker returns a Linker that can be easily used for testing.
func NewMockLinker() *MockLinker {
	return &MockLinker{
		paths:  make(map[string]Node),
		hashes: make(map[string]Node),
	}
}

// Root returns the currently set root.
// If none was created yet, an empty directory is returned.
func (ml *MockLinker) Root() (*Directory, error) {
	if ml.root != nil {
		return ml.root, nil
	}

	root, err := NewEmptyDirectory(ml, nil, "", 0)
	if err != nil {
		return nil, err
	}

	ml.root = root
	return root, nil
}

// LookupNode tries to lookup if there is already a node with this path.
func (ml *MockLinker) LookupNode(path string) (Node, error) {
	if node, ok := ml.paths[path]; ok {
		return node, nil
	}

	return nil, ie.NoSuchFile(path)
}

// NodeByHash will return a previosuly added node (via AddNode) by it's hash.
func (ml *MockLinker) NodeByHash(hash h.Hash) (Node, error) {
	if node, ok := ml.hashes[hash.B58String()]; ok {
		return node, nil
	}

	return nil, fmt.Errorf("No such hash")
}

// MemSetRoot sets the current root to be `root`.
func (ml *MockLinker) MemSetRoot(root *Directory) {
	ml.root = root
}

// MemIndexSwap will replace a node (referenced by `oldHash`) with `nd`.
// The path does not change.
func (ml *MockLinker) MemIndexSwap(nd Node, oldHash h.Hash) {
	delete(ml.hashes, oldHash.B58String())
	ml.AddNode(nd)
}

// AddNode will add a node to the memory index.
// This is not part of the linker interface.
func (ml *MockLinker) AddNode(nd Node) {
	ml.hashes[nd.Hash().B58String()] = nd
	ml.paths[nd.Path()] = nd
}
