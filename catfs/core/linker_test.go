package core

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"testing"
	"unsafe"

	"github.com/sahib/brig/catfs/db"
	ie "github.com/sahib/brig/catfs/errors"
	n "github.com/sahib/brig/catfs/nodes"
	h "github.com/sahib/brig/util/hashlib"
	"github.com/stretchr/testify/require"
)

// Basic test to see if the root node can be inserted and stored.
// A new staging commit should be also created in the background.
// On the second run, the root node should be already cached.
func TestLinkerInsertRoot(t *testing.T) {
	WithDummyKv(t, func(kv db.Database) {
		lkr := NewLinker(kv)
		root, err := n.NewEmptyDirectory(lkr, nil, "/", "u", 2)
		if err != nil {
			t.Fatalf("Creating empty root dir failed: %v", err)
		}

		if err := lkr.StageNode(root); err != nil {
			t.Fatalf("Staging root failed: %v", err)
		}

		sameRoot, err := lkr.ResolveDirectory("/")
		if err != nil {
			t.Fatalf("Resolving root failed: %v", err)
		}

		if sameRoot == nil {
			t.Fatal("Resolving root  failed (is nil)")
		}

		if path := sameRoot.Path(); path != "/" {
			t.Fatalf("Path of root is not /: %s", path)
		}

		ptrRoot, err := lkr.ResolveDirectory("/")
		if err != nil {
			t.Fatalf("Second lookup of root failed: %v", err)
		}

		if unsafe.Pointer(ptrRoot) != unsafe.Pointer(sameRoot) {
			t.Fatal("Second root did not come from the cache")
		}

		status, err := lkr.Status()
		if err != nil {
			t.Fatalf("Failed to retrieve status: %v", err)
		}

		if !status.Root().Equal(root.TreeHash()) {
			t.Fatalf("status.root and root differ: %v <-> %v", status.Root(), root.TreeHash())
		}
	})
}

func TestLinkerRefs(t *testing.T) {
	author := n.AuthorOfStage
	WithDummyKv(t, func(kv db.Database) {
		lkr := NewLinker(kv)
		root, err := lkr.Root()
		if err != nil {
			t.Fatalf("Failed to create root: %v", err)
		}

		newFile := n.NewEmptyFile(root, "cat.png", "u", 2)
		if err != nil {
			t.Fatalf("Failed to create empty file: %v", err)
		}

		newFile.SetSize(10)
		newFile.SetContent(lkr, h.TestDummy(t, 1))

		if err := root.Add(lkr, newFile); err != nil {
			t.Fatalf("Adding empty file failed: %v", err)
		}

		if err := lkr.StageNode(newFile); err != nil {
			t.Fatalf("Staging new file failed: %v", err)
		}

		if _, err := lkr.Head(); !ie.IsErrNoSuchRef(err) {
			t.Fatalf("There is a HEAD from start?!")
		}

		cmt, err := lkr.Status()
		if err != nil || cmt == nil {
			t.Fatalf("Failed to retrieve status: %v", err)
		}

		if err := lkr.MakeCommit(author, "First commit"); err != nil {
			t.Fatalf("Making commit failed: %v", err)
		}

		// Assert that staging is empy (except the "/stage/STATUS" part)
		foundKeys := []string{}
		keys, err := kv.Keys("stage")
		require.Nil(t, err)

		for _, key := range keys {
			foundKeys = append(foundKeys, strings.Join(key, "/"))
		}

		require.Equal(t, []string{"stage/STATUS"}, foundKeys)

		head, err := lkr.Head()
		if err != nil {
			t.Fatalf("Obtaining HEAD failed: %v", err)
		}

		status, err := lkr.Status()
		if err != nil {
			t.Fatalf("Failed to obtain the status: %v", err)
		}

		if !head.Root().Equal(status.Root()) {
			t.Fatalf("HEAD and CURR are not equal after first commit.")
		}

		if err := lkr.MakeCommit(author, "No."); err != ie.ErrNoChange {
			t.Fatalf("Committing without change led to a new commit: %v", err)
		}
	})
}

func TestLinkerNested(t *testing.T) {
	WithDummyKv(t, func(kv db.Database) {
		lkr := NewLinker(kv)
		root, err := lkr.Root()
		if err != nil {
			t.Fatalf("Fetching initial root failed: %v", err)
			return
		}

		sub, err := n.NewEmptyDirectory(lkr, root, "sub", "u", 3)
		if err != nil {
			t.Fatalf("Creating empty sub dir failed: %v", err)
			return
		}

		par, err := sub.Parent(lkr)
		if err != nil {
			t.Fatalf("Failed to get parent of /sub")
		}

		if par.Path() != "/" {
			t.Fatalf("Parent path of /sub is not /")
		}

		if topPar, err := par.Parent(lkr); topPar != nil || err != nil {
			t.Fatalf("Parent of / is not nil: %v (%v)", topPar, err)
		}

		if err := lkr.StageNode(sub); err != nil {
			t.Fatalf("Staging /sub failed: %v", err)
		}

		sameSubDir, err := lkr.ResolveDirectory("/sub")
		if err != nil {
			t.Fatalf("Resolving /sub failed: %v", err)
		}

		_, err = lkr.NodeByInode(sameSubDir.Inode())
		if err != nil {
			t.Fatalf("Resolving /sub by ID (%d) failed: %v", sameSubDir.Inode(), err)
		}

		subpub, err := n.NewEmptyDirectory(lkr, sameSubDir, "pub", "u", 4)
		if err != nil {
			t.Fatalf("Creating of deep sub failed")
		}

		if err := lkr.StageNode(subpub); err != nil {
			t.Fatalf("Staging /sub/pub failed: %v", err)
		}

		newRootDir, err := lkr.ResolveDirectory("/")
		if err != nil {
			t.Fatalf("Failed to resolve new root dir")
		}

		if !newRootDir.TreeHash().Equal(root.TreeHash()) {
			t.Fatalf("New / and old / have different hashes, despite being same instance %p %p", newRootDir, root)
		}

		count := 0
		if err := n.Walk(lkr, root, true, func(c n.Node) error { count++; return nil }); err != nil {
			t.Fatalf("Failed to walk the tree: %v", err)
		}

		if count != 3 {
			t.Fatalf("There are more or less than 3 elems in the tree: %d", count)
		}

		// Index shall only contain the nodes with their most current hash values.
		if len(lkr.index) != 3 {
			t.Fatalf("Index does not contain the expected 3 elements.")
		}

		gc := NewGarbageCollector(lkr, kv, nil)
		if err := gc.Run(true); err != nil {
			t.Fatalf("Garbage collector failed to run: %v", err)
		}

		if err := lkr.MakeCommit(n.AuthorOfStage, "first message"); err != nil {
			t.Fatalf("Making first commit failed: %v", err)
		}
	})
}

// Test if Linker can load objects after closing/re-opening the kv.
func TestLinkerPersistence(t *testing.T) {
	dbPath, err := ioutil.TempDir("", "brig-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	defer os.RemoveAll(dbPath)

	kv, err := db.NewDiskDatabase(dbPath)
	if err != nil {
		t.Fatalf("Could not create dummy kv for tests: %v", err)
	}

	lkr := NewLinker(kv)
	if err := lkr.MakeCommit(n.AuthorOfStage, "initial commit"); err != nil {
		t.Fatalf("Failed to create initial commit out of nothing: %v", err)
	}

	head, err := lkr.Head()
	if err != nil {
		t.Fatalf("Failed to retrieve Head after initial commit: %v", err)
	}

	oldHeadHash := head.TreeHash().Clone()

	if err := kv.Close(); err != nil {
		t.Fatalf("Closing the dummy kv failed: %v", err)
	}

	kv, err = db.NewDiskDatabase(dbPath)
	if err != nil {
		t.Fatalf("Could not create second dummy kv: %v", err)
	}

	lkr = NewLinker(kv)
	head, err = lkr.Head()
	if err != nil {
		t.Fatalf("Failed to retrieve head after kv reload: %v", err)
	}

	if !oldHeadHash.Equal(head.TreeHash()) {
		t.Fatalf("HEAD hash differs before and after reload: %v <-> %v", oldHeadHash, head.TreeHash())
	}

	if err := kv.Close(); err != nil {
		t.Fatalf("Closing the second kv failed: %v", err)
	}
}

func TestCollideSameObjectHash(t *testing.T) {
	WithDummyKv(t, func(kv db.Database) {
		lkr := NewLinker(kv)
		root, err := lkr.Root()
		if err != nil {
			t.Fatalf("Failed to retrieve root: %v", err)
		}

		sub, err := n.NewEmptyDirectory(lkr, root, "sub", "u", 3)
		if err != nil {
			t.Fatalf("Creating empty sub dir failed: %v", err)
			return
		}

		if err := lkr.StageNode(sub); err != nil {
			t.Fatalf("Staging /sub failed: %v", err)
		}

		file1 := n.NewEmptyFile(sub, "a.png", "u", 4)
		if err != nil {
			t.Fatalf("Failed to create empty file1: %v", err)
		}

		file2 := n.NewEmptyFile(root, "a.png", "u", 5)
		if err != nil {
			t.Fatalf("Failed to create empty file2: %v", err)
		}

		file3 := n.NewEmptyFile(root, "b.png", "u", 6)
		if err != nil {
			t.Fatalf("Failed to create empty file3: %v", err)
		}

		file1.SetContent(lkr, h.TestDummy(t, 1))
		file2.SetContent(lkr, h.TestDummy(t, 1))
		file3.SetContent(lkr, h.TestDummy(t, 1))

		if err := sub.Add(lkr, file1); err != nil {
			t.Fatalf("Failed to add file1: %v", err)
		}
		if err := root.Add(lkr, file2); err != nil {
			t.Fatalf("Failed to add file2: %v", err)
		}
		if err := root.Add(lkr, file3); err != nil {
			t.Fatalf("Failed to add file3: %v", err)
		}

		if err := lkr.StageNode(file1); err != nil {
			t.Fatalf("Failed to stage file1: %v", err)
		}
		if err := lkr.StageNode(file2); err != nil {
			t.Fatalf("Failed to stage file2: %v", err)
		}
		if err := lkr.StageNode(file3); err != nil {
			t.Fatalf("Failed to stage file3: %v", err)
		}

		if file1.TreeHash().Equal(file2.TreeHash()) {
			t.Fatalf("file1 and file2 hash is equal: %v", file1.TreeHash())
		}
		if file2.TreeHash().Equal(file3.TreeHash()) {
			t.Fatalf("file2 and file3 hash is equal: %v", file2.TreeHash())
		}

		// Make sure we load the actual hashes from disk:
		lkr.MemIndexClear()
		file1Reset, err := lkr.LookupFile("/sub/a.png")
		if err != nil {
			t.Fatalf("Re-Lookup of file1 failed: %v", err)
		}
		file2Reset, err := lkr.LookupFile("/a.png")
		if err != nil {
			t.Fatalf("Re-Lookup of file2 failed: %v", err)
		}
		file3Reset, err := lkr.LookupFile("/b.png")
		if err != nil {
			t.Fatalf("Re-Lookup of file3 failed: %v", err)
		}

		if file1Reset.TreeHash().Equal(file2Reset.TreeHash()) {
			t.Fatalf("file1Reset and file2Reset hash is equal: %v", file1.TreeHash())
		}
		if file2Reset.TreeHash().Equal(file3Reset.TreeHash()) {
			t.Fatalf("file2Reset and file3Reset hash is equal: %v", file2.TreeHash())
		}
	})
}

func TestHaveStagedChanges(t *testing.T) {
	WithDummyLinker(t, func(lkr *Linker) {
		hasChanges, err := lkr.HaveStagedChanges()
		if err != nil {
			t.Fatalf("have staged changes failed before touch: %v", err)
		}
		if hasChanges {
			t.Fatalf("HaveStagedChanges has changes before something happened")
		}

		MustTouch(t, lkr, "/x.png", 1)

		hasChanges, err = lkr.HaveStagedChanges()
		if err != nil {
			t.Fatalf("have staged changes failed after touch: %v", err)
		}
		if !hasChanges {
			t.Fatalf("HaveStagedChanges has no changes after something happened")
		}

		MustCommit(t, lkr, "second")

		hasChanges, err = lkr.HaveStagedChanges()
		if err != nil {
			t.Fatalf("have staged changes failed after commit: %v", err)
		}
		if hasChanges {
			t.Fatalf("HaveStagedChanges has changes after commit")
		}
	})
}

func TestFilesByContent(t *testing.T) {
	WithDummyLinker(t, func(lkr *Linker) {
		file := MustTouch(t, lkr, "/x.png", 1)

		contents := []h.Hash{file.BackendHash()}
		result, err := lkr.FilesByContents(contents)

		require.Nil(t, err)

		resultFile, ok := result[file.BackendHash().B58String()]
		require.True(t, ok)
		require.Len(t, result, 1)
		require.Equal(t, file, resultFile)
	})
}

func TestResolveRef(t *testing.T) {
	WithDummyLinker(t, func(lkr *Linker) {
		initCmt, err := lkr.Head()
		require.Nil(t, err)

		cmts := []*n.Commit{initCmt}
		for idx := 0; idx < 10; idx++ {
			_, cmt := MustTouchAndCommit(t, lkr, "/x", byte(idx))
			cmts = append([]*n.Commit{cmt}, cmts...)
		}

		// Insert the init cmt a few times as fodder:
		cmts = append(cmts, initCmt)
		cmts = append(cmts, initCmt)
		cmts = append(cmts, initCmt)

		for nUp := 0; nUp < len(cmts)+3; nUp++ {
			refname := "head"
			for idx := 0; idx < nUp; idx++ {
				refname += "^"
			}

			expect := initCmt
			if nUp < len(cmts) {
				expect = cmts[nUp]
			}

			ref, err := lkr.ResolveRef(refname)
			require.Nil(t, err)
			require.Equal(t, expect, ref)
		}

		_, err = lkr.ResolveRef("he^^ad")
		require.Equal(t, err, ie.ErrNoSuchRef("he^^ad"))
	})
}

type iterResult struct {
	path, commit string
}

func TestIterAll(t *testing.T) {
	WithDummyLinker(t, func(lkr *Linker) {
		init, err := lkr.Head()
		require.Nil(t, err)
		c0 := init.TreeHash().B58String()

		x := MustTouch(t, lkr, "/x", 1)
		MustTouch(t, lkr, "/y", 1)
		first := MustCommit(t, lkr, "first")
		c1 := first.TreeHash().B58String()
		MustModify(t, lkr, x, 2)

		status, err := lkr.Status()
		require.Nil(t, err)
		c2 := status.TreeHash().B58String()

		results := []iterResult{}
		require.Nil(t, lkr.IterAll(nil, nil, func(nd n.ModNode, cmt *n.Commit) error {
			results = append(results, iterResult{nd.Path(), cmt.TreeHash().B58String()})
			return nil
		}))

		sort.Slice(results, func(i, j int) bool {
			// Do not change orderings between commits:
			if results[i].commit != results[j].commit {
				return false
			}

			return results[i].path < results[j].path
		})

		expected := []iterResult{
			{"/", c2},
			{"/x", c2},
			{"/y", c2},
			{"/", c1},
			{"/x", c1},
			{"/", c0},
		}

		for idx, result := range results {
			require.Equal(t, result.path, expected[idx].path)
			require.Equal(t, result.commit, expected[idx].commit)
		}
	})
}

func TestAtomic(t *testing.T) {
	WithDummyLinker(t, func(lkr *Linker) {
		err := lkr.Atomic(func() (bool, error) {
			MustTouch(t, lkr, "/x", 1)
			return false, nil
		})

		require.Nil(t, err)

		err = lkr.Atomic(func() (bool, error) {
			MustTouch(t, lkr, "/y", 1)
			return true, errors.New("artificial error")
		})

		require.NotNil(t, err)

		err = lkr.Atomic(func() (bool, error) {
			MustTouch(t, lkr, "/z", 1)
			panic("woah")
		})

		require.NotNil(t, err)

		x, err := lkr.LookupFile("/x")
		require.Nil(t, err)
		require.Equal(t, x.Path(), "/x")

		_, err = lkr.LookupFile("/y")
		require.NotNil(t, err)
		require.True(t, ie.IsNoSuchFileError(err))

		_, err = lkr.LookupFile("/z")
		require.NotNil(t, err)
		require.True(t, ie.IsNoSuchFileError(err))
	})

}

func TestCommitByIndex(t *testing.T) {
	// Note: WithReloadingLinker creates an init commit.
	WithDummyLinker(t, func(lkr *Linker) {
		head, err := lkr.Head()
		require.Nil(t, err)
		require.Equal(t, head.Index(), int64(0))

		status, err := lkr.Status()
		require.Nil(t, err)
		require.Equal(t, int64(1), status.Index())

		// Must modify something to commit:
		MustTouch(t, lkr, "/x", 1)

		require.Nil(t, lkr.MakeCommit("me", "is mario"))
		newHead, err := lkr.Head()
		require.Nil(t, err)
		require.Equal(t, int64(1), newHead.Index())

		status, err = lkr.Status()
		require.Nil(t, err)
		require.Equal(t, int64(2), status.Index())

		// Lookup the just created commits:

		// Pre-existing init commit:
		c1, err := lkr.CommitByIndex(0)
		require.Nil(t, err)
		require.Equal(t, "init", c1.Message())

		// Our commit:
		c2, err := lkr.CommitByIndex(1)
		require.Nil(t, err)
		require.Equal(t, "is mario", c2.Message())

		// Same as the status commit:
		c3, err := lkr.CommitByIndex(2)
		require.Nil(t, err)
		require.NotNil(t, c3)
		require.Equal(t, status.TreeHash(), c3.TreeHash())

		// Not existing:
		c4, err := lkr.CommitByIndex(3)
		require.True(t, ie.IsErrNoSuchCommitIndex(err))
		require.Nil(t, c4)
	})
}

func TestLookupNodeAt(t *testing.T) {
	WithDummyLinker(t, func(lkr *Linker) {
		fmt.Println("start")
		for idx := byte(0); idx < 10; idx++ {
			MustTouchAndCommit(t, lkr, "/x", idx)
		}
		fmt.Println("done")

		for idx := 0; idx < 10; idx++ {
			// commit index of 0 is init, so + 1
			cmt, err := lkr.CommitByIndex(int64(idx + 1))
			require.Nil(t, err)

			nd, err := lkr.LookupNodeAt(cmt, "/x")
			require.Nil(t, err)
			require.Equal(t, nd.ContentHash(), h.TestDummy(t, byte(idx)))
		}

		// Init should not exist:
		init, err := lkr.CommitByIndex(0)
		require.Nil(t, err)

		nd, err := lkr.LookupNodeAt(init, "/x")
		require.Nil(t, nd)
		require.True(t, ie.IsNoSuchFileError(err))

		// Stage should have the last change:
		stage, err := lkr.CommitByIndex(11)
		require.Nil(t, err)

		stageNd, err := lkr.LookupNodeAt(stage, "/x")
		require.Nil(t, err)
		require.Equal(t, stageNd.ContentHash(), h.TestDummy(t, 9))

		// quick check to see if the next commit is really empty
		// (tests only the test setup)
		last, err := lkr.CommitByIndex(12)
		require.True(t, ie.IsErrNoSuchCommitIndex(err))
		require.Nil(t, last)
	})
}
