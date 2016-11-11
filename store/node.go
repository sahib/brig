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

// NOTE: The name sounds funny.
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

type Serializable interface {
	ToProto() (*wire.Node, error)
	FromProto(*wire.Node) error
}

type HierarchyEntry interface {
	NChildren() int
	Child(name string) (Node, error)
	Parent() (Node, error)
	SetParent(nd Node) error
	GetType() NodeType
}

type Streamable interface {
	Key() []byte
	Stream() (ipfsutil.Reader, error)
}

// TODO: Document api
// TODO: StreambleNode?
type Node interface {
	// TODO: sync.Locker needed?
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

func NodePath(nd Node) string {
	// TODO: Remove; not needed anymore.
	return nd.Path()
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
