package cafs

import (
	"fmt"
	"io/ioutil"
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

	//defer os.RemoveAll(dbPath)

	kv, err := db.NewDiskvDatabase(dbPath)
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
		root, err := n.NewEmptyDirectory(lkr, nil, "/")
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
		fs := NewLinker(kv)
		root, err := fs.Root()
		if err != nil {
			t.Fatalf("Failed to create root: %v", err)
		}

		newFile, err := n.NewEmptyFile(fs, root, "cat.png")
		if err != nil {
			t.Fatalf("Failed to create empty file: %v", err)
		}

		newFile.SetSize(10)
		newFile.SetHash(fs, h.TestDummy(t, 1))

		if err := root.Add(fs, newFile); err != nil {
			t.Fatalf("Adding empty file failed: %v", err)
		}

		if err := fs.StageNode(newFile); err != nil {
			t.Fatalf("Staging new file failed: %v", err)
		}

		if _, err := fs.Head(); !IsErrNoSuchRef(err) {
			t.Fatalf("There is a HEAD from start?!")
		}

		cmt, err := fs.Status()
		if err != nil || cmt == nil {
			t.Fatalf("Failed to retrieve status: %v", err)
		}

		fmt.Println("----> MAKE COMMIT")
		if err := fs.MakeCommit(author, "First commit"); err != nil {
			t.Fatalf("Making commit failed: %v", err)
		}

		head, err := fs.Head()
		if err != nil {
			t.Fatalf("Obtaining HEAD failed: %v", err)
		}

		status, err := fs.Status()
		if err != nil {
			t.Fatalf("Failed to obtain the status: %v", err)
		}

		if !head.Root().Equal(status.Root()) {
			t.Fatalf("HEAD and CURR are not equal after first commit.")
		}

		if err := fs.MakeCommit(author, "No."); err != ErrNoChange {
			t.Fatalf("Committing without change led to a new commit.")
		}
	})
}

func TestFSInsertTwoLevelDir(t *testing.T) {
	withDummyKv(t, func(kv db.Database) {
		fs := NewLinker(kv)
		root, err := fs.Root()
		if err != nil {
			t.Fatalf("Fetching initial root failed: %v", err)
			return
		}

		sub, err := n.NewEmptyDirectory(fs, root, "sub")
		if err != nil {
			t.Fatalf("Creating empty sub dir failed: %v", err)
			return
		}

		par, err := sub.Parent(fs)
		if err != nil {
			t.Fatalf("Failed to get parent of /sub")
		}

		if par.Path() != "/" {
			t.Fatalf("Parent path of /sub is not /")
		}

		if topPar, err := par.Parent(fs); topPar != nil || err != nil {
			t.Fatalf("Parent of / is not nil: %v (%v)", topPar, err)
		}

		if err := fs.StageNode(sub); err != nil {
			t.Fatalf("Staging /sub failed: %v", err)
		}

		sameSubDir, err := fs.ResolveDirectory("/sub")
		if err != nil {
			t.Fatalf("Resolving /sub failed: %v", err)
		}

		_, err = fs.NodeByInode(sameSubDir.Inode())
		if err != nil {
			t.Fatalf("Resolving /sub by ID (%d) failed: %v", sameSubDir.Inode(), err)
		}

		subpub, err := n.NewEmptyDirectory(fs, sameSubDir, "pub")
		if err != nil {
			t.Fatalf("Creating of deep sub failed")
		}

		if err := fs.StageNode(subpub); err != nil {
			t.Fatalf("Staging /sub/pub failed: %v", err)
		}

		newRootDir, err := fs.ResolveDirectory("/")
		if err != nil {
			t.Fatalf("Failed to resolve new root dir")
		}

		if !newRootDir.Hash().Equal(root.Hash()) {
			t.Fatalf("New / and old / have different hashes, despite being same instance")
		}

		count := 0
		if err := n.Walk(fs, root, true, func(c n.Node) error { count++; return nil }); err != nil {
			t.Fatalf("Failed to walk the tree: %v", err)
		}

		if count != 3 {
			t.Fatalf("There are more or less than 3 elems in the tree: %d", count)
		}

		// Index shall only contain the nodes with their most current hash values.
		if len(fs.index) != 3 {
			t.Fatalf("Index does not contain the expected 3 elements.")
		}
	})
}

//
// func withEmptyRoot(t *testing.T, f func(fs *FS, root *Directory)) {
// 	withDummyKv(t, func(kv KV) {
// 		fs := NewFilesystem(kv)
// 		root, err := newEmptyDirectory(fs, nil, "/")
// 		if err != nil {
// 			t.Fatalf("Creating empty dir failed: %v", err)
// 			return
// 		}
//
// 		if err := fs.StageNode(root); err != nil {
// 			t.Fatalf("Failed to stage root: %v", err)
// 			return
// 		}
//
// 		if err := fs.MakeCommit(StageAuthor(), "initial commit"); err != nil {
// 			t.Fatalf("Failed to create initial commit")
// 			return
// 		}
//
// 		f(fs, root)
// 	})
// }
//
// func modFile(t *testing.T, fs *FS, file *File, seed int) {
// 	root, err := fs.Root()
// 	if err != nil {
// 		t.Fatalf("Failed to get root: %v", err)
// 		return
// 	}
//
// 	if err := root.RemoveChild(file); err != nil && !IsNoSuchFileError(err) {
// 		t.Fatalf("Unable to remove %s from /: %v", file.Path(), err)
// 		return
// 	}
//
// 	file.SetSize(uint64(seed))
// 	file.SetHash(dummyHash(t, byte(seed)))
//
// 	if err := root.Add(file); err != nil {
// 		t.Fatalf("Unable to add %s to /: %v", file.Path(), err)
// 		return
// 	}
//
// 	if err := fs.StageNode(file); err != nil {
// 		t.Fatalf("Failed to stage %s for second: %v", file.Path(), err)
// 		return
// 	}
// }
//
// func TestCheckoutFile(t *testing.T) {
// 	withEmptyRoot(t, func(fs *FS, root *Directory) {
// 		file, err := newEmptyFile(fs, root, "cat.png")
// 		if err != nil {
// 			t.Fatalf("Failed to create cat.png: %v", err)
// 			return
// 		}
//
// 		modFile(t, fs, file, 1)
//
// 		if err := fs.MakeCommit(StageAuthor(), "second commit"); err != nil {
// 			t.Fatalf("Failed to make second commit: %v", err)
// 			return
// 		}
//
// 		modFile(t, fs, file, 2)
//
// 		if err := fs.MakeCommit(StageAuthor(), "third commit"); err != nil {
// 			t.Fatalf("Failed to make third commit: %v", err)
// 			return
// 		}
//
// 		head, err := fs.Head()
// 		if err != nil {
// 			t.Fatalf("Failed to get HEAD: %v", err)
// 			return
// 		}
//
// 		lastCommitNd, err := head.Parent()
// 		if err != nil {
// 			t.Fatalf("Failed to get second commit: %v", err)
// 			return
// 		}
//
// 		lastCommit := lastCommitNd.(*Commit)
//
// 		if err := fs.CheckoutFile(lastCommit, file); err != nil {
// 			t.Fatalf("Failed to checkout file before commit: %v", err)
// 			return
// 		}
//
// 		lastVersion, err := fs.LookupFile("/cat.png")
// 		if err != nil {
// 			t.Fatalf("Failed to lookup /cat.png post checkout")
// 			return
// 		}
//
// 		if !lastVersion.Hash().Equal(dummyHash(t, 1)) {
// 			t.Fatalf("Hash of checkout'd file is not from second commit")
// 			return
// 		}
//
// 		if lastVersion.Size() != 1 {
// 			t.Fatalf("Size of checkout'd file is not from second commit")
// 			return
// 		}
// 	})
// }
