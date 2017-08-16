package cafs

import (
	"fmt"
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
	dirname := path.Dir(touchPath)
	parent, err := lkr.LookupDirectory(dirname)
	if err != nil {
		t.Fatalf("touch: Failed to lookup: %s", dirname)
	}

	file, err := n.NewEmptyFile(parent, path.Base(touchPath), lkr.NextInode())
	if err != nil {
		t.Fatalf("touch: Creating dummy file failed: %v", err)
	}

	file.SetHash(lkr, h.TestDummy(t, seed))

	if err := parent.Add(lkr, file); err != nil {
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

func TestRemove(t *testing.T) {
	withDummyKv(t, func(kv db.Database) {
		lkr := NewLinker(kv)
		dir, err := mkdir(lkr, "/some/nested/directory", true)
		if err != nil {
			t.Fatalf("Failed to mkdir a nested directory: %v", err)
		}

		assertDir(t, lkr, "/some/nested/directory", true)

		path := "/some/nested/directory/cat.png"
		touchFile(t, lkr, path, 1)

		// Check file removal with ghost creation:

		file, err := lkr.LookupFile(path)
		if err != nil {
			t.Fatalf("Failed to lookup nested file: %v", err)
		}

		parentDir, err := remove(lkr, file, true)
		if err != nil {
			t.Fatalf("Remove failed: %v", err)
		}

		if !parentDir.Hash().Equal(dir.Hash()) {
			t.Fatalf("Hash differs on %s and %s", dir.Path(), parentDir.Hash())
		}

		// Check that a ghost was created for the removed file:

		ghost, err := lkr.LookupGhost(path)
		if err != nil {
			t.Fatalf("Looking up ghost failed: %v", err)
		}

		oldFile, err := ghost.OldFile()
		if err != nil {
			t.Fatalf("Failed to retrieve old file from ghost: %v", err)
		}

		if !oldFile.Hash().Equal(file.Hash()) {
			t.Fatal("Old file and original file hashes differ!")
		}

		// Check directory removal:

		nestedDir, err := lkr.LookupDirectory("/some/nested")
		if err != nil {
			t.Fatalf("Lookup on /some/nested failed: %v", err)
		}

		nestedParentDir, err := nestedDir.Parent(lkr)
		if err != nil {
			t.Fatalf("Getting parent of /some/nested failed: %v", err)
		}

		parentDir, err = remove(lkr, nestedDir, true)
		if err != nil {
			t.Fatalf("Directory removal failed: %v", err)
		}

		if !parentDir.Hash().Equal(nestedParentDir.Hash()) {
			t.Fatalf("Hash differs on %s and %s", nestedParentDir.Path(), parentDir.Hash())
		}

		root, err := lkr.Root()
		n.Walk(lkr, root, true, func(nd n.Node) error {
			fmt.Println(nd.Path())
			return nil
		})

	})
}
