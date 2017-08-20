package cafs

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"unsafe"

	"github.com/disorganizer/brig/cafs/db"
	n "github.com/disorganizer/brig/cafs/nodes"
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

// Basic test to see if the root node can be inserted and stored.
// A new staging commit should be also created in the background.
// On the second run, the root node should be already cached.
func TestLinkerInsertRoot(t *testing.T) {
	withDummyKv(t, func(kv db.Database) {
		lkr := NewLinker(kv)
		root, err := n.NewEmptyDirectory(lkr, nil, "/", 2)
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

		fmt.Println(status.Hash())
	})
}

func TestLinkerRefs(t *testing.T) {
	author := n.AuthorOfStage()
	withDummyKv(t, func(kv db.Database) {
		lkr := NewLinker(kv)
		root, err := lkr.Root()
		if err != nil {
			t.Fatalf("Failed to create root: %v", err)
		}

		newFile, err := n.NewEmptyFile(root, "cat.png", 2)
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

		if _, err := lkr.Head(); !IsErrNoSuchRef(err) {
			t.Fatalf("There is a HEAD from start?!")
		}

		cmt, err := lkr.Status()
		if err != nil || cmt == nil {
			t.Fatalf("Failed to retrieve status: %v", err)
		}

		if err := lkr.MakeCommit(author, "First commit"); err != nil {
			t.Fatalf("Making commit failed: %v", err)
		}

		// TODO: Check that stage/{tree,objects} is empty.

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

		fmt.Println("is it here?")
		if err := lkr.MakeCommit(author, "No."); err != ErrNoChange {
			t.Fatalf("Committing without change led to a new commit: %v", err)
		}
	})
}

func TestLinkerNested(t *testing.T) {
	withDummyKv(t, func(kv db.Database) {
		lkr := NewLinker(kv)
		root, err := lkr.Root()
		if err != nil {
			t.Fatalf("Fetching initial root failed: %v", err)
			return
		}

		sub, err := n.NewEmptyDirectory(lkr, root, "sub", 3)
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

		subpub, err := n.NewEmptyDirectory(lkr, sameSubDir, "pub", 4)
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
			t.Fatalf("New / and old / have different hashes, despite being same instance")
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

		if err := lkr.MakeCommit(n.AuthorOfStage(), "first message"); err != nil {
			t.Fatalf("Making first commit failed: %v", err)
		}
	})
}

func modifyFile(t *testing.T, lkr *Linker, file *n.File, seed int) {
	root, err := lkr.Root()
	if err != nil {
		t.Fatalf("Failed to get root: %v", err)
	}

	if err := root.RemoveChild(lkr, file); err != nil && !n.IsNoSuchFileError(err) {
		t.Fatalf("Unable to remove %s from /: %v", file.Path(), err)
	}

	file.SetSize(uint64(seed))
	file.SetContent(lkr, h.TestDummy(t, byte(seed)))

	if err := root.Add(lkr, file); err != nil {
		t.Fatalf("Unable to add %s to /: %v", file.Path(), err)
	}

	if err := lkr.StageNode(file); err != nil {
		t.Fatalf("Failed to stage %s for second: %v", file.Path(), err)
	}
}

func TestCheckoutFile(t *testing.T) {
	withDummyKv(t, func(kv db.Database) {
		lkr := NewLinker(kv)
		if err := lkr.MakeCommit(n.AuthorOfStage(), "initial commit"); err != nil {
			t.Fatalf("Initial commit failed: %v", err)
		}

		root, err := lkr.Root()
		if err != nil {
			t.Fatalf("Getting root failed: %v", err)
		}

		file, err := n.NewEmptyFile(root, "cat.png", 3)
		if err != nil {
			t.Fatalf("Failed to create cat.png: %v", err)
		}

		modifyFile(t, lkr, file, 1)
		oldFileHash := file.Hash().Clone()

		if err := lkr.MakeCommit(n.AuthorOfStage(), "second commit"); err != nil {
			t.Fatalf("Failed to make second commit: %v", err)
		}

		modifyFile(t, lkr, file, 2)

		if err := lkr.MakeCommit(n.AuthorOfStage(), "third commit"); err != nil {
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

		if err := lkr.CheckoutFile(lastCommit, file); err != nil {
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
	if err := lkr.MakeCommit(n.AuthorOfStage(), "initial commit"); err != nil {
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
	withDummyKv(t, func(kv db.Database) {
		lkr := NewLinker(kv)
		root, err := lkr.Root()
		if err != nil {
			t.Fatalf("Failed to retrieve root: %v", err)
		}

		sub, err := n.NewEmptyDirectory(lkr, root, "sub", 3)
		if err != nil {
			t.Fatalf("Creating empty sub dir failed: %v", err)
			return
		}

		if err := lkr.StageNode(sub); err != nil {
			t.Fatalf("Staging /sub failed: %v", err)
		}

		file1, err := n.NewEmptyFile(sub, "a.png", 4)
		if err != nil {
			t.Fatalf("Failed to create empty file1: %v", err)
		}

		file2, err := n.NewEmptyFile(root, "a.png", 5)
		if err != nil {
			t.Fatalf("Failed to create empty file2: %v", err)
		}

		file3, err := n.NewEmptyFile(root, "b.png", 6)
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

		fmt.Println(file1.Hash())
		fmt.Println(file2.Hash())
		fmt.Println(file3.Hash())
		if file1.Hash().Equal(file2.Hash()) {
			t.Fatalf("file1 and file2 hash is equal: %v", file1.Hash())
		}
		if file2.Hash().Equal(file3.Hash()) {
			t.Fatalf("file2 and file3 hash is equal: %v", file2.Hash())
		}

		// Make sure we load the actual hases from disk:
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
