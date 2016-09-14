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

func TestInsertMultiple(t *testing.T) {
	testPaths := []struct {
		path  string
		isDir bool
	}{
		{"/home/sahib", true},
		// {"/home/sahib", true},
		// {"/home/sahib/music.txt", false},
	}

	withDummyKv(t, func(kv KV) {
		fs := NewFilesystem(kv)

		for _, elem := range testPaths {
			fmt.Println("============", elem.path)

			dir, err := mkdir(fs, elem.path, true)
			if err != nil {
				t.Errorf("Failed to mkdir parents of %s: %v", elem.path, err)
				return
			}

			dirPath := NodePath(dir)
			if dirPath != elem.path {
				t.Errorf("`%s` was inserted at `%s` :(", elem.path, dirPath)
				return
			}

		}
	})
}
