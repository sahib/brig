package vcs

import (
	"fmt"
	"testing"

	c "github.com/sahib/brig/catfs/core"
	"github.com/sahib/brig/catfs/db"
	ie "github.com/sahib/brig/catfs/errors"
	n "github.com/sahib/brig/catfs/nodes"
	h "github.com/sahib/brig/util/hashlib"
	"github.com/stretchr/testify/require"
)

func TestResetFile(t *testing.T) {
	c.WithDummyKv(t, func(kv db.Database) {
		lkr := c.NewLinker(kv)
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

		c.MustModify(t, lkr, file, 1)
		oldFileHash := file.TreeHash().Clone()

		if err := lkr.MakeCommit(n.AuthorOfStage, "second commit"); err != nil {
			t.Fatalf("Failed to make second commit: %v", err)
		}

		c.MustModify(t, lkr, file, 2)
		headFileHash := file.TreeHash().Clone()

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

		if err := ResetNode(lkr, lastCommit, "/cat.png"); err != nil {
			t.Fatalf("Failed to checkout file before commit: %v", err)
		}

		lastVersion, err := lkr.LookupFile("/cat.png")
		if err != nil {
			t.Fatalf("Failed to lookup /cat.png post checkout")
		}

		if !lastVersion.TreeHash().Equal(oldFileHash) {
			t.Fatalf("Hash of checkout'd file is not from second commit")
		}

		if lastVersion.Size() != 1 {
			t.Fatalf("Size of checkout'd file is not from second commit")
		}

		if err := ResetNode(lkr, initCmt, "/cat.png"); err != nil {
			t.Fatalf("Failed to checkout file at init: %v", err)
		}

		_, err = lkr.LookupFile("/cat.png")
		if !ie.IsNoSuchFileError(err) {
			t.Fatalf("Different error: %v", err)
		}

		if err := ResetNode(lkr, head, "/cat.png"); err != nil {
			t.Fatalf("Failed to checkout file at head: %v", err)
		}

		headVersion, err := lkr.LookupFile("/cat.png")
		if err != nil {
			t.Fatalf("Failed to lookup /cat.png post checkout")
		}

		if !headVersion.TreeHash().Equal(headFileHash) {
			t.Fatalf(
				"Hash differs between new and head reset: %v != %v",
				headVersion.TreeHash(),
				headFileHash,
			)
		}
	})
}

func TestFindPathAt(t *testing.T) {
	c.WithDummyLinker(t, func(lkr *c.Linker) {
		nd := c.MustTouch(t, lkr, "/x", 1)
		c1 := c.MustCommit(t, lkr, "1")
		c.MustMove(t, lkr, nd, "/y")
		c.MustCommit(t, lkr, "2")

		oldPath, err := findPathAt(lkr, c1, "/y")
		require.Nil(t, err)
		fmt.Println("OLD PATH", oldPath)
	})
}

// Reset a file that was moved in earlier incarnations.
func TestResetMovedFile(t *testing.T) {
	c.WithDummyLinker(t, func(lkr *c.Linker) {
		sub := c.MustMkdir(t, lkr, "/sub")
		nd := c.MustTouch(t, lkr, "/sub/x", 1)
		c1 := c.MustCommit(t, lkr, "1")
		c.MustMove(t, lkr, nd, "/y")
		c.MustModify(t, lkr, nd, 2)
		c.MustCommit(t, lkr, "2")

		// This should reset /y to content=1.
		err := ResetNode(lkr, c1, "/y")
		require.Nil(t, err)

		root, err := lkr.Root()
		require.Nil(t, err)

		children, err := root.ChildrenSorted(lkr)
		require.Nil(t, err)
		require.Len(t, children, 2)
		require.Equal(t, children[0].Type(), n.NodeType(n.NodeTypeDirectory))
		require.Equal(t, children[0].Path(), "/sub")

		require.Equal(t, children[1].Type(), n.NodeType(n.NodeTypeFile))
		require.Equal(t, children[1].Path(), "/y")
		require.Equal(t, children[1].BackendHash(), h.TestDummy(t, 1))

		subChildren, err := sub.ChildrenSorted(lkr)
		require.Nil(t, err)
		require.Len(t, subChildren, 1)
		require.Equal(t, subChildren[0].Type(), n.NodeType(n.NodeTypeGhost))
		require.Equal(t, subChildren[0].Path(), "/sub/x")
	})
}
