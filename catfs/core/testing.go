package core

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"
	"testing"

	"github.com/sahib/brig/catfs/db"
	ie "github.com/sahib/brig/catfs/errors"
	n "github.com/sahib/brig/catfs/nodes"
	h "github.com/sahib/brig/util/hashlib"
	"github.com/stretchr/testify/require"
)

// WithDummyKv creates a testing key value store and passes it to `fn`.
func WithDummyKv(t *testing.T, fn func(kv db.Database)) {
	dbPath, err := ioutil.TempDir("", "brig-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	defer os.RemoveAll(dbPath)

	kv, err := db.NewDiskDatabase(dbPath)
	if err != nil {
		t.Fatalf("Could not create dummy kv for tests: %v", err)
	}

	fn(kv)

	if err := kv.Close(); err != nil {
		t.Fatalf("Closing the dummy kv failed: %v", err)
	}
}

// WithDummyLinker creates a testing linker and passes it to `fn`.
func WithDummyLinker(t *testing.T, fn func(lkr *Linker)) {
	WithDummyKv(t, func(kv db.Database) {
		lkr := NewLinker(kv)
		require.Nil(t, lkr.SetOwner("alice"))
		MustCommit(t, lkr, "init")

		fn(lkr)
	})
}

// WithReloadingLinker creates a testing linker and passes it to `fn1`.
// It then closes the linker and lets it load a second time and passes it to `fn2`.
// This is useful to test persistency issues.
func WithReloadingLinker(t *testing.T, fn1 func(lkr *Linker), fn2 func(lkr *Linker)) {
	WithDummyKv(t, func(kv db.Database) {
		lkr1 := NewLinker(kv)
		require.Nil(t, lkr1.SetOwner("alice"))
		MustCommit(t, lkr1, "init")

		fn1(lkr1)

		lkr2 := NewLinker(kv)
		fn2(lkr2)
	})
}

// WithLinkerPair creates two linkers, useful for testing syncing.
func WithLinkerPair(t *testing.T, fn func(lkrSrc, lkrDst *Linker)) {
	WithDummyLinker(t, func(lkrSrc *Linker) {
		WithDummyLinker(t, func(lkrDst *Linker) {
			require.Nil(t, lkrSrc.SetOwner("src"))
			require.Nil(t, lkrDst.SetOwner("dst"))
			fn(lkrSrc, lkrDst)
		})
	})
}

// AssertDir asserts the existence of a directory.
func AssertDir(t *testing.T, lkr *Linker, path string, shouldExist bool) {
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

// MustMkdir creates a directory or fails on `t`.
func MustMkdir(t *testing.T, lkr *Linker, repoPath string) *n.Directory {
	dir, err := Mkdir(lkr, repoPath, true)
	if err != nil {
		t.Fatalf("Failed to create directories %s: %v", repoPath, err)
	}

	return dir
}

// MustTouch creates a new node at `touchPath` and sets its content hash
// to a hash derived from `seed`.
func MustTouch(t *testing.T, lkr *Linker, touchPath string, seed byte) *n.File {
	dirname := path.Dir(touchPath)
	parent, err := lkr.LookupDirectory(dirname)
	if err != nil {
		t.Fatalf("touch: Failed to lookup: %s", dirname)
	}

	basePath := path.Base(touchPath)
	file := n.NewEmptyFile(parent, basePath, lkr.owner, lkr.NextInode())

	file.SetBackend(lkr, h.TestDummy(t, seed))
	file.SetContent(lkr, h.TestDummy(t, seed))
	file.SetKey(make([]byte, 32))

	child, err := parent.Child(lkr, basePath)
	if err != nil {
		t.Fatalf("touch: Failed to lookup child: %v %v", touchPath, err)
	}

	if child != nil {
		if err := parent.RemoveChild(lkr, child); err != nil {
			t.Fatalf("touch: failed to remove previous node: %v", err)
		}
	}

	if err := parent.Add(lkr, file); err != nil {
		t.Fatalf("touch: Adding %s to root failed: %v", touchPath, err)
	}

	if err := lkr.StageNode(file); err != nil {
		t.Fatalf("touch: Staging %s failed: %v", touchPath, err)
	}

	return file
}

// MustMove moves the node `nd` to `destPath` or fails `t`.
func MustMove(t *testing.T, lkr *Linker, nd n.ModNode, destPath string) n.ModNode {
	if err := Move(lkr, nd, destPath); err != nil {
		t.Fatalf("move of %s to %s failed: %v", nd.Path(), destPath, err)
	}

	newNd, err := lkr.LookupModNode(destPath)
	if err != nil {
		t.Fatalf("Failed to lookup dest path `%s` of new node: %v", destPath, err)
	}

	return newNd
}

// MustRemove removes the node `nd` or fails.
func MustRemove(t *testing.T, lkr *Linker, nd n.ModNode) n.ModNode {
	if _, _, err := Remove(lkr, nd, true, false); err != nil {
		t.Fatalf("Failed to remove %s: %v", nd.Path(), err)
	}

	newNd, err := lkr.LookupModNode(nd.Path())
	if err != nil {
		t.Fatalf("Failed to lookup dest path `%s` of deleted node: %v", nd.Path(), err)
	}

	return newNd
}

// MustCommit commits the current state with `msg`.
func MustCommit(t *testing.T, lkr *Linker, msg string) *n.Commit {
	if err := lkr.MakeCommit(n.AuthorOfStage, msg); err != nil {
		t.Fatalf("Failed to make commit with msg %s: %v", msg, err)
	}

	head, err := lkr.Head()
	if err != nil {
		t.Fatalf("Failed to retrieve head after commit: %v", err)
	}

	return head
}

// MustCommitIfPossible with is like MustCommit, but allows empty changesets.
func MustCommitIfPossible(t *testing.T, lkr *Linker, msg string) *n.Commit {
	haveChanges, err := lkr.HaveStagedChanges()
	if err != nil {
		t.Fatalf("Failed to check for changes: %v", err)
	}

	if haveChanges {
		return MustCommit(t, lkr, msg)
	}

	return nil
}

// MustTouchAndCommit is a combined MustTouch and MustCommit.
func MustTouchAndCommit(t *testing.T, lkr *Linker, path string, seed byte) (*n.File, *n.Commit) {
	file, err := Stage(lkr, path, h.TestDummy(t, seed), h.TestDummy(t, seed), uint64(seed), nil, time.Now())
	if err != nil {
		t.Fatalf("Failed to stage %s at %d: %v", path, seed, err)
	}

	return file, MustCommit(t, lkr, fmt.Sprintf("cmt %d", seed))
}

// MustModify changes the content of an existing node.
func MustModify(t *testing.T, lkr *Linker, file *n.File, seed int) {
	parent, err := lkr.LookupDirectory(path.Dir(file.Path()))
	// root, err := lkr.Root()
	if err != nil {
		t.Fatalf("Failed to get root: %v", err)
	}

	if err := parent.RemoveChild(lkr, file); err != nil && !ie.IsNoSuchFileError(err) {
		t.Fatalf("Unable to remove %s from /: %v", file.Path(), err)
	}

	file.SetSize(uint64(seed))
	file.SetBackend(lkr, h.TestDummy(t, byte(seed)))
	file.SetContent(lkr, h.TestDummy(t, byte(seed)))

	if err := parent.Add(lkr, file); err != nil {
		t.Fatalf("Unable to add %s to /: %v", file.Path(), err)
	}

	if err := lkr.StageNode(file); err != nil {
		t.Fatalf("Failed to stage %s for second: %v", file.Path(), err)
	}
}

// MustLookupDirectory loads an existing dir or fails.
func MustLookupDirectory(t *testing.T, lkr *Linker, path string) *n.Directory {
	dir, err := lkr.LookupDirectory(path)
	if err != nil {
		t.Fatalf("Failed to lookup directory %v: %v", path, err)
	}

	return dir
}
