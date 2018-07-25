package vcs

import (
	"testing"

	c "github.com/sahib/brig/catfs/core"
	n "github.com/sahib/brig/catfs/nodes"
	h "github.com/sahib/brig/util/hashlib"
	"github.com/stretchr/testify/require"
)

func TestPatchMarshalling(t *testing.T) {
	c.WithDummyLinker(t, func(lkr *c.Linker) {
		head, err := lkr.Head()
		require.Nil(t, err)

		curr := c.MustTouch(t, lkr, "/x", 1)
		next := c.MustCommit(t, lkr, "hello")

		change1 := &Change{
			Mask:    ChangeTypeMove | ChangeTypeRemove,
			Head:    head,
			Next:    next,
			Curr:    curr,
			MovedTo: "/something1",
		}

		c.MustModify(t, lkr, curr, 2)
		nextNext := c.MustCommit(t, lkr, "hello")

		change2 := &Change{
			Mask:    ChangeTypeAdd | ChangeTypeModify,
			Head:    next,
			Next:    nextNext,
			Curr:    curr,
			MovedTo: "/something2",
		}

		patch := &Patch{
			FromIndex: head.Index(),
			Changes:   []*Change{change2, change1},
		}

		msg, err := patch.ToCapnp()
		require.Nil(t, err)

		newPatch := &Patch{}
		require.Nil(t, newPatch.FromCapnp(msg))

		require.Equal(t, patch, newPatch)
	})
}

func TestPrefixTrie(t *testing.T) {
	prefixes := []string{
		"/a",
		"/b",
		"/c/d",
	}

	root := buildPrefixTrie(prefixes)
	require.True(t, hasValidPrefix(root, "/a"))
	require.True(t, hasValidPrefix(root, "/a/x/y/z"))
	require.True(t, hasValidPrefix(root, "/b/c"))
	require.True(t, hasValidPrefix(root, "/c/d/e"))

	require.False(t, hasValidPrefix(root, "/c/e/d"))
	require.False(t, hasValidPrefix(root, "/c/a/b"))
	require.False(t, hasValidPrefix(root, "/"))
	require.False(t, hasValidPrefix(root, "/d"))
}

func TestMakePatch(t *testing.T) {
	c.WithLinkerPair(t, func(lkrSrc, lkrDst *c.Linker) {
		init, err := lkrSrc.Head()
		require.Nil(t, err)

		srcX := c.MustTouch(t, lkrSrc, "/x", 1)
		srcY := c.MustTouch(t, lkrSrc, "/y", 2)
		c.MustMkdir(t, lkrSrc, "/sub")
		c.MustMkdir(t, lkrSrc, "/empty")
		srcZ := c.MustTouch(t, lkrSrc, "/sub/z", 3)
		c.MustCommit(t, lkrSrc, "3 files")

		patch, err := MakePatch(lkrSrc, init, []string{"/"})
		require.Nil(t, err)

		require.Nil(t, ApplyPatch(lkrDst, patch))
		dstX, err := lkrDst.LookupFile("/x")
		require.Nil(t, err)
		require.Equal(t, dstX.ContentHash(), h.TestDummy(t, 1))

		dstY, err := lkrDst.LookupFile("/y")
		require.Nil(t, err)
		require.Equal(t, dstY.ContentHash(), h.TestDummy(t, 2))

		dstZ, err := lkrDst.LookupFile("/sub/z")
		require.Nil(t, err)
		require.Equal(t, dstZ.ContentHash(), h.TestDummy(t, 3))

		_, err = lkrDst.LookupDirectory("/empty")
		require.Nil(t, err)

		///////////////////

		c.MustModify(t, lkrSrc, srcX, 4)
		c.MustMove(t, lkrSrc, srcY, "/y_moved")
		c.MustRemove(t, lkrSrc, srcZ)
		c.MustTouch(t, lkrSrc, "/empty/not_empty_anymore", 42)

		// TODO: Also fix ordering in Replay/ApplyPatch?
		patch, err = MakePatch(lkrSrc, init, []string{"/"})
		require.Nil(t, err)
		require.Nil(t, ApplyPatch(lkrDst, patch))

		dstYMoved, err := lkrDst.LookupFile("/y_moved")
		require.Nil(t, err)
		require.Equal(t, dstYMoved.Path(), "/y_moved")

		dstYGhost, err := lkrDst.LookupGhost("/y")
		require.Nil(t, err)
		require.Equal(t, dstYGhost.Path(), "/y")

		dstZGhost, err := lkrDst.LookupGhost("/sub/z")
		require.Nil(t, err)
		require.Equal(t, dstZGhost.Path(), "/sub/z")

		dstNotEmptyFile, err := lkrDst.LookupFile("/empty/not_empty_anymore")
		require.Nil(t, err)
		require.Equal(t, dstNotEmptyFile.Path(), "/empty/not_empty_anymore")
	})
}

func TestMakePatchWithOrderConflict(t *testing.T) {
	c.WithLinkerPair(t, func(lkrSrc, lkrDst *c.Linker) {
		init, err := lkrSrc.Head()
		require.Nil(t, err)

		srcX := c.MustTouch(t, lkrSrc, "/x", 1)
		srcY := c.MustTouch(t, lkrSrc, "/y", 2)
		c.MustCommit(t, lkrSrc, "pre-move")

		c.MustMove(t, lkrSrc, srcX, "/z")
		c.MustMove(t, lkrSrc, srcY, "/z")
		c.MustCommit(t, lkrSrc, "post-move")

		patch, err := MakePatch(lkrSrc, init, []string{"/"})
		require.Nil(t, err)

		// All files should be mentioned in the patch.
		// x and y are ghosts, z is the only real file.
		// Since y was moved last it has a move marker, x not.
		require.Equal(t, "/x", patch.Changes[1].Curr.Path())
		require.Equal(t, "", patch.Changes[1].MovedTo)
		require.Equal(t, "", patch.Changes[1].WasPreviouslyAt)
		require.Equal(t, n.NodeTypeGhost, patch.Changes[1].Curr.Type())

		require.Equal(t, "/y", patch.Changes[0].Curr.Path())
		require.Equal(t, "/z", patch.Changes[0].MovedTo)
		require.Equal(t, n.NodeTypeGhost, patch.Changes[0].Curr.Type())

		require.Equal(t, "/z", patch.Changes[2].Curr.Path())
		require.Equal(t, "", patch.Changes[2].MovedTo)
		require.Equal(t, "/y", patch.Changes[2].WasPreviouslyAt)
		require.Equal(t, n.NodeTypeFile, patch.Changes[2].Curr.Type())

		require.Nil(t, ApplyPatch(lkrDst, patch))
		dstZ, err := lkrDst.LookupFile("/z")
		require.Nil(t, err)
		require.Equal(t, dstZ.ContentHash(), h.TestDummy(t, 2))

		dstX, err := lkrDst.LookupGhost("/x")
		require.Nil(t, err)
		require.Equal(t, n.NodeTypeGhost, dstX.Type())

		dstY, err := lkrDst.LookupGhost("/y")
		require.Nil(t, err)
		require.Equal(t, n.NodeTypeGhost, dstY.Type())
	})
}

func TestMakePatchDirMoveAllChildren(t *testing.T) {
	c.WithLinkerPair(t, func(lkrSrc, lkrDst *c.Linker) {
		init, err := lkrSrc.Head()
		require.Nil(t, err)

		c.MustMkdir(t, lkrSrc, "/src")
		subX := c.MustTouch(t, lkrSrc, "/src/x", 1)
		subY := c.MustTouch(t, lkrSrc, "/src/y", 2)
		preMove := c.MustCommit(t, lkrSrc, "base")

		patch1, err := MakePatch(lkrSrc, init, []string{"/"})
		require.Nil(t, err)

		require.Nil(t, ApplyPatch(lkrDst, patch1))
		srcDir, err := lkrDst.LookupDirectory("/src")
		require.Nil(t, err)
		require.Equal(t, 2, srcDir.NChildren())

		///////////

		c.MustMkdir(t, lkrSrc, "/dst")
		c.MustMove(t, lkrSrc, subX, "/dst/x")
		c.MustMove(t, lkrSrc, subY, "/dst/y")
		c.MustCommit(t, lkrSrc, "post-move")

		patch2, err := MakePatch(lkrSrc, preMove, []string{"/"})
		require.Nil(t, err)

		require.Nil(t, ApplyPatch(lkrDst, patch2))

		srcDir, err = lkrDst.LookupDirectory("/src")
		require.Nil(t, err)
		require.Equal(t, 2, srcDir.NChildren())

		_, err = lkrDst.LookupGhost("/src/x")
		require.Nil(t, err)
		_, err = lkrDst.LookupGhost("/src/x")
		require.Nil(t, err)

		_, err = lkrDst.LookupFile("/dst/x")
		require.Nil(t, err)
		_, err = lkrDst.LookupFile("/dst/x")
		require.Nil(t, err)
	})
}

func TestMakePatchDirMoveCompletely(t *testing.T) {
	c.WithLinkerPair(t, func(lkrSrc, lkrDst *c.Linker) {
		init, err := lkrSrc.Head()
		require.Nil(t, err)

		realSrcDir := c.MustMkdir(t, lkrSrc, "/src")
		c.MustTouch(t, lkrSrc, "/src/x", 1)
		c.MustTouch(t, lkrSrc, "/src/y", 2)
		preMove := c.MustCommit(t, lkrSrc, "base")

		patch1, err := MakePatch(lkrSrc, init, []string{"/"})
		require.Nil(t, err)

		require.Nil(t, ApplyPatch(lkrDst, patch1))
		srcDir, err := lkrDst.LookupDirectory("/src")
		require.Nil(t, err)
		require.Equal(t, 2, srcDir.NChildren())

		///////////

		c.MustMove(t, lkrSrc, realSrcDir, "/dst")
		c.MustCommit(t, lkrSrc, "post-move")

		patch2, err := MakePatch(lkrSrc, preMove, []string{"/"})
		require.Nil(t, err)

		require.Nil(t, ApplyPatch(lkrDst, patch2))

		srcDirGhost, err := lkrDst.LookupGhost("/src")
		require.Nil(t, err)
		require.Equal(t, 2, srcDirGhost.NChildren())

		_, err = lkrDst.LookupNode("/src/x")
		require.NotNil(t, err)
		_, err = lkrDst.LookupNode("/src/x")
		require.NotNil(t, err)

		_, err = lkrDst.LookupFile("/dst/x")
		require.Nil(t, err)
		_, err = lkrDst.LookupFile("/dst/x")
		require.Nil(t, err)
	})
}
