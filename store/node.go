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

	// ID returns the numeric identifier of this node.
	// It stays the same, even when the node was modified
	// or the path was moved.
	ID() uint64
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

func nodeDepth(nd Node) int {
	var depth int
	var curr Node = nd
	var err error

	for curr != nil {
		curr, err = curr.Parent()
		if err != nil {
			return -1
		}

		depth++
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

type metaRecord struct {
	hash    *Hash
	name    string
	size    uint64
	modTime time.Time
}

func (mr *metaRecord) Hash() *Hash        { return mr.hash }
func (mr *metaRecord) Name() string       { return mr.name }
func (mr *metaRecord) Size() uint64       { return mr.size }
func (mr *metaRecord) ModTime() time.Time { return mr.modTime }

func Metadata(nd Node) Metadatable {
	return &metaRecord{
		hash:    nd.Hash(),
		name:    nd.Name(),
		size:    nd.Size(),
		modTime: nd.ModTime(),
	}
}
