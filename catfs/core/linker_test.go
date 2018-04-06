package core

import (
	"io/ioutil"
	"os"
	"sort"
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
	t.Parallel()

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

		if !status.Root().Equal(root.Hash()) {
			t.Fatalf("status.root and root differ: %v <-> %v", status.Root(), root.Hash())
		}
	})
}

func TestLinkerRefs(t *testing.T) {
	t.Parallel()

	author := n.AuthorOfStage
	WithDummyKv(t, func(kv db.Database) {
		lkr := NewLinker(kv)
		root, err := lkr.Root()
		if err != nil {
			t.Fatalf("Failed to create root: %v", err)
		}

		newFile, err := n.NewEmptyFile(root, "cat.png", "u", 2)
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

		// TODO: Check that stage/{tree,objects,moves} is empty.

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
	t.Parallel()

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

		if !newRootDir.Hash().Equal(root.Hash()) {
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

func TestCheckoutFile(t *testing.T) {
	t.Parallel()

	WithDummyKv(t, func(kv db.Database) {
		lkr := NewLinker(kv)
		if err := lkr.MakeCommit(n.AuthorOfStage, "initial commit"); err != nil {
			t.Fatalf("Initial commit failed: %v", err)
		}

		initCmt, err := lkr.Head()
		if err != nil {
			t.Fatalf("Failed to get initial head")
		}

		root, err := lkr.Root()
		if err != nil {
			t.Fatalf("Getting root failed: %v", err)
		}

		file, err := n.NewEmptyFile(root, "cat.png", "u", 3)
		if err != nil {
			t.Fatalf("Failed to create cat.png: %v", err)
		}

		MustModify(t, lkr, file, 1)
		oldFileHash := file.Hash().Clone()

		if err := lkr.MakeCommit(n.AuthorOfStage, "second commit"); err != nil {
			t.Fatalf("Failed to make second commit: %v", err)
		}

		MustModify(t, lkr, file, 2)
		headFileHash := file.Hash().Clone()

		if err := lkr.MakeCommit(n.AuthorOfStage, "third commit"); err != nil {
			t.Fatalf("Failed to make third commit: %v", err)
		}

		head, err := lkr.Head()
		if err != nil {
			t.Fatalf("Failed to get HEAD: %v", err)
		}

		lastCommitNd, err := head.Parent(lkr)
		if err != nil {
			t.Fatalf("Failed to get second commit: %v", err)
		}

		lastCommit := lastCommitNd.(*n.Commit)

		if err := lkr.CheckoutFile(lastCommit, "/cat.png"); err != nil {
			t.Fatalf("Failed to checkout file before commit: %v", err)
		}

		lastVersion, err := lkr.LookupFile("/cat.png")
		if err != nil {
			t.Fatalf("Failed to lookup /cat.png post checkout")
		}

		if !lastVersion.Hash().Equal(oldFileHash) {
			t.Fatalf("Hash of checkout'd file is not from second commit")
		}

		if lastVersion.Size() != 1 {
			t.Fatalf("Size of checkout'd file is not from second commit")
		}

		if err := lkr.CheckoutFile(initCmt, "/cat.png"); err != nil {
			t.Fatalf("Failed to checkout file at init: %v", err)
		}

		_, err = lkr.LookupFile("/cat.png")
		if !ie.IsNoSuchFileError(err) {
			t.Fatalf("Different error: %v", err)
		}

		if err := lkr.CheckoutFile(head, "/cat.png"); err != nil {
			t.Fatalf("Failed to checkout file at head: %v", err)
		}

		headVersion, err := lkr.LookupFile("/cat.png")
		if err != nil {
			t.Fatalf("Failed to lookup /cat.png post checkout")
		}

		if !headVersion.Hash().Equal(headFileHash) {
			t.Fatalf(
				"Hash differs between new and head reset: %v != %v",
				headVersion.Hash(),
				headFileHash,
			)
		}
	})
}

// Test if Linker can load objects after closing/re-opening the kv.
func TestLinkerPersistence(t *testing.T) {
	t.Parallel()

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

	oldHeadHash := head.Hash().Clone()

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

	if !oldHeadHash.Equal(head.Hash()) {
		t.Fatalf("HEAD hash differs before and after reload: %v <-> %v", oldHeadHash, head.Hash())
	}

	if err := kv.Close(); err != nil {
		t.Fatalf("Closing the second kv failed: %v", err)
	}
}

func TestCollideSameObjectHash(t *testing.T) {
	t.Parallel()

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

		file1, err := n.NewEmptyFile(sub, "a.png", "u", 4)
		if err != nil {
			t.Fatalf("Failed to create empty file1: %v", err)
		}

		file2, err := n.NewEmptyFile(root, "a.png", "u", 5)
		if err != nil {
			t.Fatalf("Failed to create empty file2: %v", err)
		}

		file3, err := n.NewEmptyFile(root, "b.png", "u", 6)
		if err != nil {
			t.Fatalf("Failed to create empty file3: %v", err)
		}

		file1.SetContent(lkr, h.TestDummy(t, 1))
		file2.SetContent(lkr, h.TestDummy(t, 1))
		file3.SetContent(lkr, h.TestDummy(t, 1))

		// TODO: Shouldn't NewEmptyFile call this? It gets the parent...
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

		if file1.Hash().Equal(file2.Hash()) {
			t.Fatalf("file1 and file2 hash is equal: %v", file1.Hash())
		}
		if file2.Hash().Equal(file3.Hash()) {
			t.Fatalf("file2 and file3 hash is equal: %v", file2.Hash())
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

		if file1Reset.Hash().Equal(file2Reset.Hash()) {
			t.Fatalf("file1Reset and file2Reset hash is equal: %v", file1.Hash())
		}
		if file2Reset.Hash().Equal(file3Reset.Hash()) {
			t.Fatalf("file2Reset and file3Reset hash is equal: %v", file2.Hash())
		}
	})
}

func TestHaveStagedChanges(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	WithDummyLinker(t, func(lkr *Linker) {
		file := MustTouch(t, lkr, "/x.png", 1)

		contents := []h.Hash{file.Content()}
		result, err := lkr.FilesByContents(contents)

		require.Nil(t, err)

		resultFile, ok := result[file.Content().B58String()]
		require.True(t, ok)
		require.Equal(t, file, resultFile)
		require.Len(t, result, 1)
	})
}

func TestResolveRef(t *testing.T) {
	t.Parallel()

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
			refname := "HEAD"
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

		_, err = lkr.ResolveRef("HE^^AD")
		require.Equal(t, err, ie.ErrNoSuchRef("he^^ad"))
	})
}

type iterResult struct {
	path, commit string
}

func TestIterAll(t *testing.T) {
	t.Parallel()

	WithDummyLinker(t, func(lkr *Linker) {
		init, err := lkr.Head()
		require.Nil(t, err)
		c0 := init.Hash().B58String()

		x := MustTouch(t, lkr, "/x", 1)
		MustTouch(t, lkr, "/y", 1)
		c1 := MustCommit(t, lkr, "first").Hash().B58String()
		MustModify(t, lkr, x, 2)

		status, err := lkr.Status()
		require.Nil(t, err)
		c2 := status.Hash().B58String()

		results := []iterResult{}
		require.Nil(t, lkr.IterAll(nil, nil, func(nd n.ModNode, cmt *n.Commit) error {
			results = append(results, iterResult{nd.Path(), cmt.Hash().B58String()})
			return nil
		}))

		expected := []iterResult{
			{"/", c2},
			{"/x", c2},
			{"/y", c2},
			{"/", c1},
			{"/x", c1},
			{"/", c0},
		}

		sort.Slice(results, func(i, j int) bool {
			// Do not change orderings between commits:
			if results[i].commit != results[j].commit {
				return false
			}

			return results[i].path < results[j].path
		})

		for idx, result := range results {
			require.Equal(t, result.path, expected[idx].path)
			require.Equal(t, result.commit, expected[idx].commit)
		}
	})
}
