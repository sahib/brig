package store

import (
	"fmt"
	"testing"
)

func TestFSInsertRoot(t *testing.T) {
	withDummyKv(t, func(kv KV) {
		fs := NewFilesystem(kv)
		root, err := emptyDirectory(fs, nil, "/")
		if err != nil {
			t.Errorf("Creating empty dir failed: %v", err)
			return
		}

		if err := fs.StageNode(root); err != nil {
			t.Errorf("Staging root failed: %v", err)
			return
		}

		sameRoot, err := fs.ResolveDirectory("/")
		if err != nil {
			t.Errorf("Resolving root failed: %v", err)
			return
		}

		if sameRoot == nil {
			t.Errorf("Resolving root  failed (is nil)")
			return
		}

		ptrRoot, err := fs.ResolveDirectory("/")

		fmt.Println(nodePath(sameRoot))
		fmt.Printf("%p %p %p\n", ptrRoot, sameRoot, root)
	})
}
