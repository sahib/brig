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

func touchFile(t *testing.T, lkr *Linker, touchPath string, seed byte) *n.File {
	dirname := path.Dir(touchPath)
	parent, err := lkr.LookupDirectory(dirname)
	if err != nil {
		t.Fatalf("touch: Failed to lookup: %s", dirname)
	}

	file, err := n.NewEmptyFile(parent, path.Base(touchPath), lkr.NextInode())
	if err != nil {
		t.Fatalf("touch: Creating dummy file failed: %v", err)
	}

	file.SetContent(lkr, h.TestDummy(t, seed))

	if err := parent.Add(lkr, file); err != nil {
		t.Fatalf("touch: Adding %s to root failed: %v", touchPath, err)
	}

	if err := lkr.StageNode(file); err != nil {
		t.Fatalf("touch: Staging %s failed: %v", touchPath, err)
	}

	return file
}

func mustMkdir(t *testing.T, lkr *Linker, repoPath string) {
	if _, err := mkdir(lkr, repoPath, true); err != nil {
		t.Fatalf("Failed to create directories %s: %v", repoPath, err)
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

		// Fill in a dummy file hash, so we get a ghost instance
		parentDir, _, err := remove(lkr, file, true)
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

		// Just fill in a dummy moved to ref, to get a ghost.
		parentDir, ghost, err = remove(lkr, nestedDir, true)
		if err != nil {
			t.Fatalf("Directory removal failed: %v", err)
		}

		if ghost == nil || ghost.Type() != n.NodeTypeGhost {
			t.Fatalf("Ghost node does not look like a ghost: %v", ghost)
		}

		if !parentDir.Hash().Equal(nestedParentDir.Hash()) {
			t.Fatalf("Hash differs on %s and %s", nestedParentDir.Path(), parentDir.Hash())
		}
	})
}

func TestRemoveGhost(t *testing.T) {
	withDummyKv(t, func(kv db.Database) {
		lkr := NewLinker(kv)
		file := touchFile(t, lkr, "/x", 1)
		par, err := n.ParentDirectory(lkr, file)
		if err != nil {
			t.Fatalf("Failed to get get parent directory of /x: %v", err)
		}

		if err := par.RemoveChild(lkr, file); err != nil {
			t.Fatalf("Removing child /x failed: %v", err)
		}

		ghost, err := n.MakeGhost(file, nil, 42)
		if err != nil {
			t.Fatalf("Failed to summon ghost: %v", err)
		}

		if err := par.Add(lkr, ghost); err != nil {
			t.Fatalf("Re-adding ghost failed: %v", err)
		}

		if err := lkr.StageNode(ghost); err != nil {
			t.Fatalf("Staging ghost failed: %v", err)
		}

		// Try to remove a ghost:
		if _, _, err := remove(lkr, ghost, true); err != ErrIsGhost {
			t.Fatalf("Removing ghost failed or succeeded: %v", err)
		}
	})
}

func moveValidCheck(t *testing.T, lkr *Linker, srcPath, dstPath string) {
	nd, err := lkr.LookupNode(srcPath)

	if err == nil {
		if nd.Type() != n.NodeTypeGhost {
			t.Fatalf("Source node still exists! (%v): %v", srcPath, nd.Type())
		}
	} else if !n.IsNoSuchFileError(err) {
		t.Fatalf("Looking up source node failed: %v", err)
	}

	lkDestNode, err := lkr.LookupNode(dstPath)
	if err != nil {
		t.Fatalf("Looking up dest path failed: %v", err)
	}

	if lkDestNode.Path() != dstPath {
		t.Fatalf("Dest nod and dest path differ: %v <-> %v", lkDestNode.Path(), dstPath)
	}
}

func moveInvalidCheck(t *testing.T, lkr *Linker, srcPath, dstPath string) {
	node, err := lkr.LookupNode(srcPath)
	if err != nil {
		t.Fatalf("Source node vanished during errorneous move: %v", err)
	}

	if node.Type() == n.NodeTypeGhost {
		t.Fatalf("Source node was converted to a ghost: %v", node.Path())
	}
}

func TestMove(t *testing.T) {
	// Cases to cover for move():
	// 1.        Dest exists:
	// 1.1.      Is a directory.
	// 1.1.1  E  This directory contains basename(src) and it is a file.
	// 1.1.2  E  This directory contains basename(src) and it is a non-empty dir.
	// 1.1.3  V  This directory contains basename(src) and it is a empty dir.
	// 2.        Dest does not exist.
	// 2.1    V  dirname(dest) exists and is a directory.
	// 2.2    E  dirname(dest) does not exists.
	// 2.2    E  dirname(dest) exists and is not a directory.
	// 3.     E  Overlap of src and dest paths (src in dest)

	// Checks for valid cases (V):
	// 1) src is gone.
	// 2) dest is the same node as before.
	// 3) dest has the correct path.

	// Checks for invalid cases (E):
	// 1) src is not gone.

	var tcs = []struct {
		name        string
		isErrorCase bool
		setup       func(t *testing.T, lkr *Linker) (n.ModNode, string)
	}{
		{
			name:        "basic",
			isErrorCase: false,
			setup: func(t *testing.T, lkr *Linker) (n.ModNode, string) {
				mustMkdir(t, lkr, "/a/b/c")
				return touchFile(t, lkr, "/a/b/c/x", 1), "/a/b/y"
			},
		}, {
			name:        "move-into-directory",
			isErrorCase: false,
			setup: func(t *testing.T, lkr *Linker) (n.ModNode, string) {
				mustMkdir(t, lkr, "/a/b/c")
				mustMkdir(t, lkr, "/a/b/d")
				return touchFile(t, lkr, "/a/b/c/x", 1), "/a/b/d"
			},
		}, {
			name:        "error-move-to-directory-contains-file",
			isErrorCase: true,
			setup: func(t *testing.T, lkr *Linker) (n.ModNode, string) {
				mustMkdir(t, lkr, "/src")
				mustMkdir(t, lkr, "/dst")
				touchFile(t, lkr, "/dst/x", 1)
				return touchFile(t, lkr, "/src/x", 1), "/dst"
			},
		}, {
			name:        "error-move-file-over-existing",
			isErrorCase: false,
			setup: func(t *testing.T, lkr *Linker) (n.ModNode, string) {
				mustMkdir(t, lkr, "/src")
				mustMkdir(t, lkr, "/dst")
				touchFile(t, lkr, "/dst/x", 1)
				return touchFile(t, lkr, "/src/x", 1), "/dst/x"
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			withDummyKv(t, func(kv db.Database) {
				lkr := NewLinker(kv)

				// Setup src and dest dir with a file in it named like src.
				srcNd, dstPath := tc.setup(t, lkr)
				srcPath := srcNd.Path()

				if err := move(lkr, srcNd, dstPath); err != nil {
					if tc.isErrorCase {
						moveInvalidCheck(t, lkr, srcPath, dstPath)
					} else {
						t.Fatalf("Move failed unexpectly: %v", err)
					}
				} else {
					moveValidCheck(t, lkr, srcPath, dstPath)
				}
			})
		})
	}
}

func TestStage(t *testing.T) {
	withDummyKv(t, func(kv db.Database) {
		lkr := NewLinker(kv)

		update := &NodeUpdate{
			Hash:   h.TestDummy(t, 1),
			Size:   3,
			Author: "me",
			Key:    make([]byte, 32),
		}

		// Initial stage of the file:
		file, err := stage(lkr, "/photos/moose.png", update)
		if err != nil {
			t.Fatalf("Adding of /photos/moose.png failed: %v", err)
		}

		if !file.Content().Equal(h.TestDummy(t, 1)) {
			t.Fatalf("File content after stage is not what's advertised: %v", file.Content())
		}

		update = &NodeUpdate{
			Hash:   h.TestDummy(t, 2),
			Size:   3,
			Author: "me",
			Key:    make([]byte, 32),
		}

		file, err = stage(lkr, "/photos/moose.png", update)
		if err != nil {
			t.Fatalf("Adding of /photos/moose.png failed: %v", err)
		}

		if !file.Content().Equal(h.TestDummy(t, 2)) {
			t.Fatalf("File content after update is not what's advertised: %v", file.Content())
		}
	})
}
