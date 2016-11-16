package store

import (
	"strings"
	"time"

	"github.com/disorganizer/brig/store/wire"
	"github.com/disorganizer/brig/util/ipfsutil"
)

const (
	NodeTypeUnknown = iota
	NodeTypeFile
	NodeTypeDirectory
	NodeTypeCommit
)

type NodeType uint8

func (nt NodeType) String() string {
	switch nt {
	case NodeTypeFile:
		return "file"
	case NodeTypeDirectory:
		return "directory"
	case NodeTypeCommit:
		return "commit"
	}

	return "unknown"
}

// Metadatable is a thing that accumulates certain common node attributes.
type Metadatable interface {
	// Name returns the name of the object, i.e. the last part of the path,
	// which is also commonly called 'basename' in unix filesystems.
	Name() string

	// Hash returns the hash value of the node.
	//
	// It is an error to modify the hash value.
	// If you need to modify it, you have to make an own copy via .Clone().
	Hash() *Hash

	// Size returns the size of the node in bytes.
	Size() uint64

	// ModTime returns the time when the last modification to the node happened.
	ModTime() time.Time

	// Path of this node.
	Path() string
}

// Serializable is a thing that can be converted to a wire.Node protobuf message.
type Serializable interface {
	ToProto() (*wire.Node, error)
	FromProto(*wire.Node) error
}

// HierarchyEntry represents a thing that is placed in
// a file hierarchy and may have other children beneath it.
type HierarchyEntry interface {
	// NChildren returns the total number of children to a node.
	NChildren() int

	// Child returns a named child.
	Child(name string) (Node, error)

	// Parent returns the parent node or nil if there is none.
	Parent() (Node, error)

	// SetParent sets the parent new.  Care must be taken to remove old
	// references to the node to avoid loops.
	SetParent(nd Node) error

	// GetType returns the type of the node.
	GetType() NodeType
}

// Streamable represents a thing that can be streamed,
// given a cryptographic key.
type Streamable interface {
	Key() []byte
	Stream() (ipfsutil.Reader, error)
}

// Node is a single node in brig's MDAG.
// It is currently either a Commit, a File or a Directory.
type Node interface {
	Metadatable
	Serializable
	HierarchyEntry

	// ID returns the numeric identifier of this node.
	// It stays the same, even when the node was modified
	// or the path was moved.
	ID() uint64
}

// SettableNode is a node that supports modification of
// it's core attributes. File and Directory are settable,
// but a commit is not.
type SettableNode interface {
	Node

	SetSize(size uint64)
	SetModTime(modTime time.Time)
	SetName(name string)
	SetHash(hash *Hash)
}

//////////////// UTILITY FUNCTIONS ////////////////

func prefixSlash(s string) string {
	if !strings.HasPrefix(s, "/") {
		return "/" + s
	}

	return s
}

// NodeDepth returns the depth of the node.
// It does this by looking at the path separators.
// The depth of "/" is defined as 0.
func NodeDepth(nd Node) int {
	path := nd.Path()
	if path == "/" {
		return 0
	}

	depth := 0
	for _, rn := range path {
		if rn == '/' {
			depth++
		}
	}

	return depth
}

func nodeRemove(nd Node) error {
	parDir, err := nodeParentDir(nd)
	if err != nil {
		return err
	}

	// Cannot remove root:
	if parDir == nil {
		return nil
	}

	return parDir.RemoveChild(nd)
}

func nodeParentDir(nd Node) (*Directory, error) {
	par, err := nd.Parent()
	if err != nil {
		return nil, err
	}

	if par == nil {
		return nil, nil
	}

	parDir, ok := par.(*Directory)
	if !ok {
		return nil, ErrBadNode
	}

	return parDir, nil
}
