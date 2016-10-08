package store

import (
	"fmt"
	"testing"
	"unsafe"
)

func TestFSInsertRoot(t *testing.T) {
	withDummyKv(t, func(kv KV) {
		fs := NewFilesystem(kv)
		root, err := newEmptyDirectory(fs, nil, "/")
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

		if path := NodePath(sameRoot); path != "/" {
			t.Errorf("Path of root is not /: %s", path)
			return
		}

		ptrRoot, err := fs.ResolveDirectory("/")
		if err != nil {
			t.Errorf("Second lookup of root failed?")
			return
		}

		if unsafe.Pointer(ptrRoot) != unsafe.Pointer(sameRoot) {
			t.Errorf("Second root did not come from the cache")
			return
		}
	})
}

func TestFSInsertTwoLevelDir(t *testing.T) {
	withDummyKv(t, func(kv KV) {
		fs := NewFilesystem(kv)

		root, err := fs.Root()
		if err != nil {
			t.Errorf("Creating empty dir failed: %v", err)
			return
		}

		sub, err := newEmptyDirectory(fs, root, "sub")
		if err != nil {
			t.Errorf("Creating empty sub dir failed: %v", err)
			return
		}

		par, err := sub.Parent()
		if err != nil {
			t.Errorf("Failed to get parent of /sub")
		}

		if par.Path() != "/" {
			t.Errorf("Parent path of /sub is not /")
			return
		}

		if topPar, err := par.Parent(); topPar != nil || err != nil {
			t.Errorf("Parent of / is not nil: %v (%v)", topPar, err)
		}

		fmt.Println("staging sub")

		if err := fs.StageNode(sub); err != nil {
			t.Errorf("Staging /sub failed: %v", err)
			return
		}

		sameSubDir, err := fs.ResolveDirectory("/sub")
		if err != nil {
			t.Errorf("Resolving /sub failed: %v", err)
			return
		}

		subpub, err := newEmptyDirectory(fs, sameSubDir, "pub")
		if err != nil {
			t.Errorf("Creating of deep sub failed")
			return
		}

		if err := fs.StageNode(subpub); err != nil {
			t.Errorf("Staging /sub/pub failed: %v", err)
			return
		}

		newRootDir, err := fs.ResolveDirectory("/")
		if err != nil {
			t.Errorf("Failed to resolve new root dir")
			return
		}

		fmt.Println(newRootDir, newRootDir)
		if !newRootDir.Hash().Equal(root.Hash()) {
			t.Errorf("New / and old / have different hashes, despite being same instance")
			return
		}

		count := 0
		if err := Walk(root, true, func(c Node) error { count++; return nil }); err != nil {
			t.Errorf("Failed to walk the tree: %v", err)
			return
		}

		if count != 3 {
			t.Errorf("There are more or less than 3 elems in the tree: %d", count)
			return
		}

		// Index shall only contain the nodes with their most current hash values.
		if len(fs.index) != 3 {
			t.Errorf("Index does not contain the expected 3 elements.")
			return
		}
	})
}
