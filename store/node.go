package store

import (
	"path"
	"sync"
	"time"
)

type Node interface {
	sync.Locker

	Marshal() ([]byte, error)
	Unmarshal(data []byte) error

	Name() string
	Hash() *Hash
	Size() uint64
	ModTime() time.Time

	NChildren() int
	Child(name string) (Node, error)

	Parent() (Node, error)
	SetParent(nd Node) error
}

func nodePath(nd Node) string {
	var err error
	elems := []string{}

	for nd != nil {
		elems = append(elems, nd.Name())

		nd, err = nd.Parent()
		if err != nil {
			break
		}
	}

	for i := 0; i < len(elems)/2; i++ {
		elems[i], elems[len(elems)-i-1] = elems[len(elems)-i-1], elems[i]
	}

	return "/" + path.Join(elems...)
}
