package core

import (
	"path"
	"sort"
	"strings"
	"time"
	"testing"

	ie "github.com/sahib/brig/catfs/errors"
	n "github.com/sahib/brig/catfs/nodes"
	h "github.com/sahib/brig/util/hashlib"
	"github.com/stretchr/testify/require"
)

func TestMkdir(t *testing.T) {
	WithDummyLinker(t, func(lkr *Linker) {
		// Test nested creation without -p like flag:
		dir, err := Mkdir(lkr, "/deep/nested", false)
		if err == nil || dir != nil {
			t.Fatalf("Nested mkdir without -p should have failed: %v", err)
		}

		AssertDir(t, lkr, "/", true)
		AssertDir(t, lkr, "/deep", false)
		AssertDir(t, lkr, "/deep/nested", false)

		// Test mkdir -p like creating of nested dirs:
		dir, err = Mkdir(lkr, "/deep/nested", true)
		if err != nil {
			t.Fatalf("mkdir -p failed: %v", err)
		}

		AssertDir(t, lkr, "/", true)
		AssertDir(t, lkr, "/deep", true)
		AssertDir(t, lkr, "/deep/nested", true)

		// Attempt to mkdir the same directory once more:
		dir, err = Mkdir(lkr, "/deep/nested", true)
		if err != nil {
			t.Fatalf("second mkdir -p failed: %v", err)
		}

		// Also without -p, it should just return the respective dir.
		// (i.e. work like LookupDirectory)
		// Note: This is a difference to the traditional mkdir.
		dir, err = Mkdir(lkr, "/deep/nested", false)
		if err != nil {
			t.Fatalf("second mkdir without -p failed: %v", err)
		}

		// See if an attempt at creating the root failed,
		// should not and just work like lkr.LookupDirectory("/")
		dir, err = Mkdir(lkr, "/", false)
		if err != nil {
			t.Fatalf("mkdir root failed (without -p): %v", err)
		}

		root, err := lkr.Root()
		if err != nil {
			t.Fatalf("Failed to retrieve root: %v", err)
		}

		if !dir.TreeHash().Equal(root.TreeHash()) {
			t.Fatal("Root and mkdir('/') differ!")
		}

		// Try to mkdir over a regular file:
		MustTouch(t, lkr, "/cat.png", 1)

		// This should fail, since we cannot create it.
		dir, err = Mkdir(lkr, "/cat.png", false)
		if err == nil {
			t.Fatal("Creating directory on file should have failed!")
		}

		// Same even for -p
		dir, err = Mkdir(lkr, "/cat.png", true)
		if err == nil {
			t.Fatal("Creating directory on file should have failed!")
		}
	})
}

func TestRemove(t *testing.T) {
	WithDummyLinker(t, func(lkr *Linker) {
		dir, err := Mkdir(lkr, "/some/nested/directory", true)
		if err != nil {
			t.Fatalf("Failed to mkdir a nested directory: %v", err)
		}

		AssertDir(t, lkr, "/some/nested/directory", true)

		path := "/some/nested/directory/cat.png"
		MustTouch(t, lkr, path, 1)

		// Check file removal with ghost creation:
		file, err := lkr.LookupFile(path)
		if err != nil {
			t.Fatalf("Failed to lookup nested file: %v", err)
		}

		// Fill in a dummy file hash, so we get a ghost instance
		parentDir, _, err := Remove(lkr, file, true, false)
		if err != nil {
			t.Fatalf("Remove failed: %v", err)
		}

		if !parentDir.TreeHash().Equal(dir.TreeHash()) {
			t.Fatalf("Hash differs on %s and %s", dir.Path(), parentDir.TreeHash())
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

		if !oldFile.TreeHash().Equal(file.TreeHash()) {
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
		parentDir, ghost, err = Remove(lkr, nestedDir, true, false)
		if err != nil {
			t.Fatalf("Directory removal failed: %v", err)
		}

		if ghost == nil || ghost.Type() != n.NodeTypeGhost {
			t.Fatalf("Ghost node does not look like a ghost: %v", ghost)
		}

		if !parentDir.TreeHash().Equal(nestedParentDir.TreeHash()) {
			t.Fatalf("Hash differs on %s and %s", nestedParentDir.Path(), parentDir.TreeHash())
		}
	})
}

func TestRemoveGhost(t *testing.T) {
	WithDummyLinker(t, func(lkr *Linker) {
		file := MustTouch(t, lkr, "/x", 1)
		par, err := n.ParentDirectory(lkr, file)
		if err != nil {
			t.Fatalf("Failed to get get parent directory of /x: %v", err)
		}

		if err := par.RemoveChild(lkr, file); err != nil {
			t.Fatalf("Removing child /x failed: %v", err)
		}

		ghost, err := n.MakeGhost(file, 42)
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
		if _, _, err := Remove(lkr, ghost, true, false); err != ErrIsGhost {
			t.Fatalf("Removing ghost failed other than expected: %v", err)
		}
	})
}

func TestRemoveExistingGhost(t *testing.T) {
	WithDummyLinker(t, func(lkr *Linker) {
		nd := MustTouch(t, lkr, "/x", 1)
		_, ghost, err := Remove(lkr, nd, true, true)
		require.Nil(t, err)

		_, _, err = Remove(lkr, ghost, false, true)
		require.Nil(t, err)

		_, _, err = Remove(lkr, ghost, true, true)
		require.NotNil(t, err)
	})
}

func moveValidCheck(t *testing.T, lkr *Linker, srcPath, dstPath string) {
	nd, err := lkr.LookupNode(srcPath)

	if err == nil {
		if nd.Type() != n.NodeTypeGhost {
			t.Fatalf("Source node still exists! (%v): %v", srcPath, nd.Type())
		}
	} else if !ie.IsNoSuchFileError(err) {
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

var moveAndCopyTestCases = []struct {
	name        string
	isErrorCase bool
	setup       func(t *testing.T, lkr *Linker) (n.ModNode, string)
}{
	{
		name:        "basic",
		isErrorCase: false,
		setup: func(t *testing.T, lkr *Linker) (n.ModNode, string) {
			MustMkdir(t, lkr, "/a/b/c")
			return MustTouch(t, lkr, "/a/b/c/x", 1), "/a/b/y"
		},
	}, {
		name:        "basic-directory",
		isErrorCase: false,
		setup: func(t *testing.T, lkr *Linker) (n.ModNode, string) {
			return MustMkdir(t, lkr, "/a/b/short"), "/a/b/looooong"
		},
	}, {
		name:        "basic-same-level",
		isErrorCase: false,
		setup: func(t *testing.T, lkr *Linker) (n.ModNode, string) {
			return MustTouch(t, lkr, "/a", 1), "/b"
		},
	}, {
		name:        "basic-root-to-sub",
		isErrorCase: false,
		setup: func(t *testing.T, lkr *Linker) (n.ModNode, string) {
			MustTouch(t, lkr, "/README.md", 1)
			MustMkdir(t, lkr, "/sub")
			return MustTouch(t, lkr, "/x", 1), "/sub"
		},
	}, {
		name:        "into-directory",
		isErrorCase: false,
		setup: func(t *testing.T, lkr *Linker) (n.ModNode, string) {
			MustMkdir(t, lkr, "/a/b/c")
			MustMkdir(t, lkr, "/a/b/d")
			return MustTouch(t, lkr, "/a/b/c/x", 1), "/a/b/d"
		},
	}, {
		name:        "into-nonempty-directory",
		isErrorCase: false,
		setup: func(t *testing.T, lkr *Linker) (n.ModNode, string) {
			MustMkdir(t, lkr, "/a/b/c")
			MustMkdir(t, lkr, "/a/b/d")
			MustTouch(t, lkr, "/a/b/d/y", 1)
			return MustTouch(t, lkr, "/a/b/c/x", 1), "/a/b/d"
		},
	}, {
		name:        "error-to-directory-contains-file",
		isErrorCase: true,
		setup: func(t *testing.T, lkr *Linker) (n.ModNode, string) {
			MustMkdir(t, lkr, "/src")
			MustMkdir(t, lkr, "/dst")
			MustTouch(t, lkr, "/dst/x", 1)
			return MustTouch(t, lkr, "/src/x", 1), "/dst"
		},
	}, {
		name:        "error-file-over-existing",
		isErrorCase: false,
		setup: func(t *testing.T, lkr *Linker) (n.ModNode, string) {
			MustMkdir(t, lkr, "/src")
			MustMkdir(t, lkr, "/dst")
			MustTouch(t, lkr, "/dst/x", 1)
			return MustTouch(t, lkr, "/src/x", 1), "/dst/x"
		},
	}, {
		name:        "error-file-over-ghost",
		isErrorCase: false,
		setup: func(t *testing.T, lkr *Linker) (n.ModNode, string) {
			MustMkdir(t, lkr, "/src")
			MustMkdir(t, lkr, "/dst")
			destFile := MustTouch(t, lkr, "/dst/x", 1)
			MustRemove(t, lkr, destFile)
			return MustTouch(t, lkr, "/src/x", 1), "/dst/x"
		},
	}, {
		name:        "error-src-equal-dst",
		isErrorCase: true,
		setup: func(t *testing.T, lkr *Linker) (n.ModNode, string) {
			return MustTouch(t, lkr, "/x", 1), "/x"
		},
	}, {
		name:        "error-into-own-subdir",
		isErrorCase: true,
		setup: func(t *testing.T, lkr *Linker) (n.ModNode, string) {
			// We should not be able to move "/dir" into itself.
			dir := MustMkdir(t, lkr, "/dir")
			MustTouch(t, lkr, "/dir/x", 1)
			return dir, "/dir/own"
		},
	},
}

func TestMoveSingle(t *testing.T) {
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

	for _, tc := range moveAndCopyTestCases {
		t.Run(tc.name, func(t *testing.T) {
			WithDummyLinker(t, func(lkr *Linker) {
				// Setup src and dest dir with a file in it named like src.
				srcNd, dstPath := tc.setup(t, lkr)
				srcPath := srcNd.Path()

				if err := Move(lkr, srcNd, dstPath); err != nil {
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

func TestMoveDirectoryWithChild(t *testing.T) {
	WithDummyLinker(t, func(lkr *Linker) {
		MustMkdir(t, lkr, "/src")
		oldFile := MustTouch(t, lkr, "/src/x", 1)
		oldFile = oldFile.Copy(oldFile.Inode()).(*n.File)

		MustCommit(t, lkr, "before move")

		dir, err := lkr.LookupDirectory("/src")
		require.Nil(t, err)

		MustMove(t, lkr, dir, "/dst")
		MustCommit(t, lkr, "after move")

		file, err := lkr.LookupFile("/dst/x")
		require.Nil(t, err)
		require.Equal(t, h.TestDummy(t, 1), file.BackendHash())

		_, err = lkr.LookupGhost("/src")
		require.Nil(t, err)

		// This will resolve to the old file:
		_, err = lkr.LookupFile("/src/x")
		require.NotNil(t, err)
	})
}

func TestMoveDirectory(t *testing.T) {
	WithDummyLinker(t, func(lkr *Linker) {
		srcDir := MustMkdir(t, lkr, "/src")
		MustMkdir(t, lkr, "/src/sub")
		MustTouch(t, lkr, "/src/sub/x", 23)
		MustTouch(t, lkr, "/src/y", 23)

		dstDir := MustMove(t, lkr, srcDir, "/dst")

		expect := []string{
			"/dst/sub/x",
			"/dst/sub",
			"/dst/y",
			"/dst",
		}

		got := []string{}
		require.Nil(t, n.Walk(lkr, dstDir, true, func(child n.Node) error {
			got = append(got, child.Path())
			return nil
		}))

		sort.Strings(expect)
		sort.Strings(got)

		require.Equal(t, len(expect), len(got))
		for idx := range expect {
			if got[idx] != expect[idx] {
				t.Errorf(
					"Moved node child `%s` does not match `%s`",
					got[idx],
					expect[idx],
				)
			}
		}
	})
}

func TestMoveDirectoryWithGhosts(t *testing.T) {
	WithDummyLinker(t, func(lkr *Linker) {
		srcDir := MustMkdir(t, lkr, "/src")
		MustMkdir(t, lkr, "/src/sub")
		xFile := MustTouch(t, lkr, "/src/sub/x", 23)
		MustTouch(t, lkr, "/src/y", 23)
		MustMove(t, lkr, xFile, "/src/z")

		dstDir := MustMove(t, lkr, srcDir, "/dst")

		expect := []string{
			"/dst",
			"/dst/sub",
			"/dst/sub/x",
			"/dst/y",
			"/dst/z",
		}

		// Be evil and clear the mem cache in order to check if all changes
		// were checked into the staging area.
		lkr.MemIndexClear()

		got := []string{}
		require.Nil(t, n.Walk(lkr, dstDir, true, func(child n.Node) error {
			got = append(got, child.Path())
			return nil
		}))

		// Check if the moved directory contains the right paths:
		sort.Strings(got)
		for idx, expectPath := range expect {
			if expectPath != got[idx] {
				t.Fatalf("%d: %s != %s", idx, expectPath, got[idx])
			}
		}

		ghost, err := lkr.LookupNode(got[2])
		require.Nil(t, err)

		status, err := lkr.Status()
		require.Nil(t, err)
		require.Equal(t, "/src/sub/x", ghost.(*n.Ghost).OldNode().Path())

		twin, _, err := lkr.MoveMapping(status, ghost)
		require.Nil(t, err)
		require.Equal(t, "/dst/z", twin.Path())
	})
}

func TestStage(t *testing.T) {
	WithDummyLinker(t, func(lkr *Linker) {
		// Initial stage of the file:
		key := make([]byte, 32)

		contentHash1 := h.TestDummy(t, 1)
		backendHash1 := h.TestDummy(t, 1)
		file, err := Stage(lkr, "/photos/moose.png", contentHash1, backendHash1, 2, key, time.Now())
		if err != nil {
			t.Fatalf("Adding of /photos/moose.png failed: %v", err)
		}

		contentHash2 := h.TestDummy(t, 2)
		backendHash2 := h.TestDummy(t, 2)
		file, err = Stage(lkr, "/photos/moose.png", contentHash2, backendHash2, 3, key, time.Now())
		if err != nil {
			t.Fatalf("Adding of /photos/moose.png failed: %v", err)
		}

		if !file.BackendHash().Equal(h.TestDummy(t, 2)) {
			t.Fatalf(
				"File content after update is not what's advertised: %v",
				file.TreeHash(),
			)
		}
	})
}

func TestStageDirOverGhost(t *testing.T) {
	WithDummyLinker(t, func(lkr *Linker) {
		empty := MustMkdir(t, lkr, "/empty")
		MustMove(t, lkr, empty, "/moved_empty")
		MustMkdir(t, lkr, "/empty")
		dir, err := lkr.LookupDirectory("/empty")
		require.Nil(t, err)

		require.Equal(t, dir.Path(), "/empty")
		if dir.Type() != n.NodeTypeDirectory {
			t.Fatalf("/empty is not a directory")
		}
	})
}

func TestCopy(t *testing.T) {
	for _, tc := range moveAndCopyTestCases {
		t.Run(tc.name, func(t *testing.T) {

			WithDummyLinker(t, func(lkr *Linker) {
				// Setup src and dest dir with a file in it named like src.
				srcNd, dstPath := tc.setup(t, lkr)
				srcPath := srcNd.Path()

				newNd, err := Copy(lkr, srcNd, dstPath)
				if newNd != nil {
					if !strings.HasPrefix(newNd.Path(), dstPath) {
						t.Fatalf(
							"Node was copied to wrong path: %v (want %v)",
							newNd.Path(),
							dstPath,
						)
					}

					// Make sure the new copy is reachable from parent:
					par, err := lkr.LookupDirectory(path.Dir(newNd.Path()))
					if err != nil {
						t.Fatalf("Failed to lookup parent: %v", err)
					}

					newChildNd, err := par.Child(lkr, newNd.Name())
					if err != nil {
						t.Fatalf("Failed to get base path: %v", err)
					}

					newNd = newChildNd.(n.ModNode)
				}

				if oldNd, lookErr := lkr.LookupNode(srcPath); oldNd == nil || lookErr != nil {
					t.Fatalf("source node does not exist or is not accesible: %v %v", err, tc.isErrorCase)
				}

				if err != nil {
					if tc.isErrorCase {
						node, err := lkr.LookupNode(srcPath)
						if err != nil {
							t.Fatalf("Source node vanished during errorneous copy: %v", err)
						}

						if node.Type() == n.NodeTypeGhost {
							t.Fatalf("Source node was converted to a ghost: %v", node.Path())
						}
					} else {
						t.Fatalf("Copy failed unexpectly: %v", err)
					}

					// No need to test more.
					return
				}

				if tc.isErrorCase {
					t.Fatalf("test should have failed")
				}

				if newNd == nil {
					t.Fatalf("Dest node does not exist after copy: %v", err)
				}

				if !newNd.BackendHash().Equal(srcNd.BackendHash()) {
					t.Logf("Content of src and dst differ after copy")
					t.Logf("WANT: %v", srcNd.BackendHash())
					t.Logf("GOT : %v", newNd.BackendHash())
					t.Fatalf("Check Copy()")
				}

				if newNd.Inode() < srcNd.Inode() {
					t.Fatalf("New inode has <= inode of src")
				}

				// Sanity check: do not rely on Copy() to return a valid, staged node.
				// Check if we can look it up after the Copy too.
				nd, err := lkr.LookupNode(newNd.Path())
				require.Nil(t, err)
				require.NotNil(t, nd)
			})
		})
	}
}
