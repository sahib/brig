package catfs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/disorganizer/brig/catfs/db"
	n "github.com/disorganizer/brig/catfs/nodes"
	h "github.com/disorganizer/brig/util/hashlib"
)

func withDummyKv(t *testing.T, fn func(kv db.Database)) {
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

func withDummyLinker(t *testing.T, fn func(lkr *Linker)) {
	withDummyKv(t, func(kv db.Database) {
		lkr := NewLinker(kv)
		mustCommit(t, lkr, "init")

		fn(lkr)
	})
}

func withLinkerPair(t *testing.T, fn func(lkrSrc, lkrDst *Linker)) {
	withDummyLinker(t, func(lkrSrc *Linker) {
		withDummyLinker(t, func(lkrDst *Linker) {
			lkrSrc.SetOwner(n.NewPerson("src", h.TestDummy(t, 23)))
			lkrDst.SetOwner(n.NewPerson("dst", h.TestDummy(t, 42)))

			fn(lkrSrc, lkrDst)
		})
	})
}

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

func mustMkdir(t *testing.T, lkr *Linker, repoPath string) *n.Directory {
	dir, err := mkdir(lkr, repoPath, true)
	if err != nil {
		t.Fatalf("Failed to create directories %s: %v", repoPath, err)
	}

	return dir
}

func mustTouch(t *testing.T, lkr *Linker, touchPath string, seed byte) *n.File {
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

func mustMove(t *testing.T, lkr *Linker, nd n.ModNode, destPath string) n.ModNode {
	if err := move(lkr, nd, destPath); err != nil {
		t.Fatalf("move of %s to %s failed: %v", nd.Path(), destPath, err)
	}

	newNd, err := lkr.LookupModNode(destPath)
	if err != nil {
		t.Fatalf("Failed to lookup dest path `%s` of new node: %v", destPath, err)
	}

	return newNd
}

func mustRemove(t *testing.T, lkr *Linker, nd n.ModNode) n.ModNode {
	if _, _, err := remove(lkr, nd, true, false); err != nil {
		t.Fatalf("Failed to remove %s: %v", nd.Path(), err)
	}

	newNd, err := lkr.LookupModNode(nd.Path())
	if err != nil {
		t.Fatalf("Failed to lookup dest path `%s` of deleted node: %v", nd.Path(), err)
	}

	return newNd
}

func mustCommit(t *testing.T, lkr *Linker, msg string) *n.Commit {
	if err := lkr.MakeCommit(n.AuthorOfStage(), msg); err != nil {
		t.Fatalf("Failed to make commit with msg %s: %v", msg, err)
	}

	head, err := lkr.Head()
	if err != nil {
		t.Fatalf("Failed to retrieve head after commit: %v", err)
	}

	return head
}

func mustTouchAndCommit(t *testing.T, lkr *Linker, path string, seed byte) (*n.File, *n.Commit) {

	info := &NodeUpdate{
		Hash:   h.TestDummy(t, seed),
		Size:   uint64(seed),
		Author: "",
		Key:    nil,
	}

	file, err := stage(lkr, path, info)
	if err != nil {
		t.Fatalf("Failed to stage %s at %d: %v", path, seed, err)
	}

	return file, mustCommit(t, lkr, fmt.Sprintf("cmt %d", seed))
}
