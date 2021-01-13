package nodes

import (
	"time"

	capnp_model "github.com/sahib/brig/catfs/nodes/capnp"
	h "github.com/sahib/brig/util/hashlib"
	capnp "zombiezen.com/go/capnproto2"
)

// NodeType defines the type of a specific node.
type NodeType uint8

const (
	// NodeTypeUnknown should not happen in real programs
	NodeTypeUnknown = NodeType(iota)
	// NodeTypeFile indicates a regular file
	NodeTypeFile
	// NodeTypeDirectory indicates a directory
	NodeTypeDirectory
	// NodeTypeCommit indicates a commit
	NodeTypeCommit
	// NodeTypeGhost indicates a moved node
	NodeTypeGhost
)

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

	// Size returns the size of the node in bytes.
	Size() uint64

	// CachedSize returns the size of the node at the backend in bytes.
	CachedSize() uint64

	// ModTime returns the time when the last modification to the node happened.
	ModTime() time.Time

	// Path of this node.
	Path() string

	// GetType returns the type of the node.
	Type() NodeType

	// INode shall return a unique identifier for this node that does
	// not change, even when the content of the node changes.
	Inode() uint64

	// TreeHash returns the hash value of the node.
	//
	// It is an error to modify the hash value.
	// If you need to modify it, you have to make an own copy via .Clone().
	TreeHash() h.Hash

	// ContentHash is the actual plain text hash of the node.
	// This is used for comparing file and directory equality.
	ContentHash() h.Hash

	// BackendHash returns the hash under which the stored content
	// can be read from the backend.
	// It is valid to return nil if the file is empty.
	BackendHash() h.Hash
}

// Serializable is a thing that can be converted to a capnproto message.
type Serializable interface {
	ToCapnp() (*capnp.Message, error)
	FromCapnp(*capnp.Message) error

	ToCapnpNode(seg *capnp.Segment, capNd capnp_model.Node) error
	FromCapnpNode(capNd capnp_model.Node) error
}

// HierarchyEntry represents a thing that is placed in
// a file hierarchy and may have other children beneath it.
type HierarchyEntry interface {
	// NChildren returns the total number of children to a node.
	NChildren() int

	// Child returns a named child.
	Child(lkr Linker, name string) (Node, error)

	// Parent returns the parent node or nil if there is none.
	Parent(lkr Linker) (Node, error)

	// SetParent sets the parent new. Care must be taken to remove old
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

	// SetSize sets the size of the node in bytes
	SetSize(size uint64)

	// SetModTime updates the modtime timestamp
	SetModTime(modTime time.Time)

	// SetName sets the user that last modified the file
	SetName(name string)

	// SetUser sets the user that last modified the file
	SetUser(user string)

	// NotifyMove tells the node that it was moved.
	// It should be called whenever the path of the node changed.
	// (i.e. not only the name, but parts of the parent path)
	NotifyMove(lkr Linker, parent *Directory, newPath string) error

	// Copy creates a copy of this node with the inode `inode`.
	Copy(inode uint64) ModNode
}
