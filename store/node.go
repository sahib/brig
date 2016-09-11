package store

import (
	"path"
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
}

//////////////// UTILITY FUNCTIONS ////////////////

func prefixSlash(s string) string {
	if !strings.HasPrefix(s, "/") {
		return "/" + s
	}

	return s
}

func nodePath(nd Node) string {
	var err error
	elems := []string{}

	for nd != nil {
		elems = append(elems, nd.Name())

		nd, err = nd.Parent()
		if err != nil || nd == nil {
			break
		}
	}

	for i := 0; i < len(elems)/2; i++ {
		elems[i], elems[len(elems)-i-1] = elems[len(elems)-i-1], elems[i]
	}

	return prefixSlash(path.Join(elems...))
}
