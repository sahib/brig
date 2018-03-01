package nodes

import (
	"time"

	h "github.com/sahib/brig/util/hashlib"
	capnp "zombiezen.com/go/capnproto2"
)

const (
	// NodeTypeUnknown should not happen in real programs
	NodeTypeUnknown = iota
	// NodeTypeFile indicates a regular file
	NodeTypeFile
	// NodeTypeDirectory indicates a directory
	NodeTypeDirectory
	// NodeTypeCommit indicates a commit
	NodeTypeCommit
	// NodeTypeGhost indicates a moved node
	NodeTypeGhost
)

// NodeType defines the type of a specific node.
type NodeType uint8

var nodeTypeToString = map[NodeType]string{
	NodeTypeCommit:    "commit",
	NodeTypeGhost:     "ghost",
	NodeTypeFile:      "file",
	NodeTypeDirectory: "directory",
}

func (n NodeType) String() string {
	if name, ok := nodeTypeToString[n]; ok {
		return name
	}

	return "unknown"
}

// Metadatable is a thing that accumulates certain common node attributes.
type Metadatable interface {
	// Name returns the name of the object, i.e. the last part of the path,
	// which is also commonly called 'basename' in unix filesystems.
	Name() string

	// User returns the id of the user that last modified this file.
	// (There is no real ownership)
	User() string

	// Hash returns the hash value of the node.
	//
	// It is an error to modify the hash value.
	// If you need to modify it, you have to make an own copy via .Clone().
	Hash() h.Hash

	// Size returns the size of the node in bytes.
	Size() uint64

	// ModTime returns the time when the last modification to the node happened.
	ModTime() time.Time

	// Path of this node.
	Path() string

	// GetType returns the type of the node.
	Type() NodeType

	// INode shall return a unique identifier for this node that does
	// not change, even when the content of the node changes.
	Inode() uint64

	// Content will return the hash of the content of this file/
	// It is valid to return nil if the file is empty.
	Content() h.Hash
}

// Serializable is a thing that can be converted to a capnproto message.
type Serializable interface {
	ToCapnp() (*capnp.Message, error)
	FromCapnp(*capnp.Message) error
}

// HierarchyEntry represents a thing that is placed in
// a file hierarchy and may have other children beneath it.
type HierarchyEntry interface {
	// NChildren returns the total number of children to a node.
	NChildren(lkr Linker) int

	// Child returns a named child.
	Child(lkr Linker, name string) (Node, error)

	// Parent returns the parent node or nil if there is none.
	Parent(lkr Linker) (Node, error)

	// SetParent sets the parent new.  Care must be taken to remove old
	// references to the node to avoid loops.
	SetParent(lkr Linker, nd Node) error
}

// Streamable represents a thing that can be streamed,
// given a cryptographic key.
type Streamable interface {
	Key() []byte
}

// Node is a single node in brig's MDAG.
// It is currently either a Commit, a File or a Directory.
type Node interface {
	Metadatable
	Serializable
	HierarchyEntry
}

// ModNode is a node that supports modification of
// it's core attributes. File and Directory are settable,
// but a commit is not.
type ModNode interface {
	Node

	SetSize(size uint64)
	SetModTime(modTime time.Time)
	SetName(name string)
	SetUser(user string)

	// TODO: write some assumptions about this.
	NotifyMove(lkr Linker, newPath string) error

	// TODO: Should this be part of this interface?
	Copy(inode uint64) ModNode
}
