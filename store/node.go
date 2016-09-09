package store

import (
	"path"
	"strings"
	"sync"
	"time"

	"github.com/disorganizer/brig/store/wire"
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

// TODO: Document api
// TODO: Split in sub-interfaces
type Node interface {
	sync.Locker

	// Unmarshalling
	ToProto() (*wire.Node, error)
	FromProto(*wire.Node) error

	// Metadata
	Name() string
	Hash() *Hash
	Size() uint64
	ModTime() time.Time

	// Hierarchy
	NChildren() int
	Child(name string) (Node, error)
	Parent() (Node, error)
	SetParent(nd Node) error
	GetType() NodeType
}

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
