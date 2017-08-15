package cafs

import (
	"path"
	"testing"

	"github.com/disorganizer/brig/cafs/db"
	n "github.com/disorganizer/brig/cafs/nodes"
	h "github.com/disorganizer/brig/util/hashlib"
)

func assertDir(t *testing.T, lkr *Linker, path string, shouldExist bool) {
	dir, err := lkr.LookupDirectory(path)
	if shouldExist {
		if err != nil {
			t.Fatalf("exist-check: Directory lookup failed for %s: %v", path, err)
		}

		if dir == nil || dir.Path() != path {
			t.Fatalf("exist-check: directory does not exist:  %s -> %v", path, dir)
		}
	} else {
		if dir != nil {
			t.Fatalf("exist-check: Dir exists, but should not: %v", path)
		}
	}
}

func touchFile(t *testing.T, lkr *Linker, touchPath string, seed byte) {
	root, err := lkr.Root()
	if err != nil {
		t.Fatalf("Failed to retrieve root: %v", err)
	}

	file, err := n.NewEmptyFile(root, path.Base(touchPath), lkr.NextInode())
	if err != nil {
		t.Fatalf("touch: Creating dummy file failed: %v", err)
	}

	file.SetHash(lkr, h.TestDummy(t, seed))

	if err := root.Add(lkr, file); err != nil {
		t.Fatalf("touch: Adding %s to root failed: %v", touchPath, err)
	}

	if err := lkr.StageNode(file); err != nil {
		t.Fatalf("touch: Staging %s failed: %v", touchPath, err)
	}

}

func TestMkdir(t *testing.T) {
	withDummyKv(t, func(kv db.Database) {
		lkr := NewLinker(kv)

		// Test nested creation without -p like flag:
		dir, err := mkdir(lkr, "/deep/nested", false)
		if err == nil || dir != nil {
			t.Fatalf("Nested mkdir without -p should have failed: %v", err)
		}

		assertDir(t, lkr, "/", true)
		assertDir(t, lkr, "/deep", false)
		assertDir(t, lkr, "/deep/nested", false)

		// Test mkdir -p like creating of nested dirs:
		dir, err = mkdir(lkr, "/deep/nested", true)
		if err != nil {
			t.Fatalf("mkdir -p failed: %v", err)
		}

		assertDir(t, lkr, "/", true)
		assertDir(t, lkr, "/deep", true)
		assertDir(t, lkr, "/deep/nested", true)

		// Attempt to mkdir the same directory once more:
		dir, err = mkdir(lkr, "/deep/nested", true)
		if err != nil {
			t.Fatalf("second mkdir -p failed: %v", err)
		}

		// Also without -p, it should just return the respective dir.
		// (i.e. work like LookupDirectory)
		// Note: This is a difference to the traditional mkdir.
		dir, err = mkdir(lkr, "/deep/nested", false)
		if err != nil {
			t.Fatalf("second mkdir without -p failed: %v", err)
		}

		// See if an attempt at creating the root failed,
		// should not and just work like lkr.LookupDirectory("/")
		dir, err = mkdir(lkr, "/", false)
		if err != nil {
			t.Fatalf("mkdir root failed (without -p): %v", err)
		}

		root, err := lkr.Root()
		if err != nil {
			t.Fatalf("Failed to retrieve root: %v", err)
		}

		if !dir.Hash().Equal(root.Hash()) {
			t.Fatal("Root and mkdir('/') differ!")
		}

		// Try to mkdir over a regular file:
		touchFile(t, lkr, "/cat.png", 1)

		// This should fail, since we cannot create it.
		dir, err = mkdir(lkr, "/cat.png", false)
		if err == nil {
			t.Fatal("Creating directory on file should have failed!")
		}

		// Same even for -p
		dir, err = mkdir(lkr, "/cat.png", true)
		if err == nil {
			t.Fatal("Creating directory on file should have failed!")
		}
	})
}
