package vcs

import (
	"testing"

	c "github.com/sahib/brig/catfs/core"
	n "github.com/sahib/brig/catfs/nodes"
	"github.com/stretchr/testify/require"
)

func TestChangeMarshalling(t *testing.T) {
	c.WithDummyLinker(t, func(lkr *c.Linker) {
		head, err := lkr.Head()
		require.Nil(t, err)

		curr := c.MustTouch(t, lkr, "/x", 1)
		next := c.MustCommit(t, lkr, "hello")

		change := &Change{
			Mask:    ChangeTypeMove | ChangeTypeRemove,
			Head:    head,
			Next:    next,
			Curr:    curr,
			MovedTo: "/something",
		}

		msg, err := change.ToCapnp()
		require.Nil(t, err)

		newChange := &Change{}
		require.Nil(t, newChange.FromCapnp(msg))

		require.Equal(t, newChange.MovedTo, "/something")
		require.Equal(t, newChange.Mask, ChangeTypeMove|ChangeTypeRemove)
		require.Equal(t, newChange.Curr, curr)
		require.Equal(t, newChange.Head, head)
		require.Equal(t, newChange.Next, next)

		// This check helps failing when adding new fields:
		require.Equal(t, change, newChange)
	})
}

func TestChangeCombine(t *testing.T) {
	c.WithDummyLinker(t, func(lkr *c.Linker) {
		x := c.MustTouch(t, lkr, "/x", 1)
		c.MustCommit(t, lkr, "1")
		c.MustModify(t, lkr, x, 2)
		c.MustCommit(t, lkr, "2")
		y := c.MustMove(t, lkr, x, "/y")
		c.MustCommit(t, lkr, "move")
		c.MustRemove(t, lkr, y)
		ghost, err := lkr.LookupGhost("/y")
		require.Nil(t, err)

		status, err := lkr.Status()
		require.Nil(t, err)

		changes, err := History(lkr, ghost, status, nil)
		require.Nil(t, err)
		require.Len(t, changes, 4)
		require.Equal(t, changes[0].Mask, ChangeTypeRemove)
		require.Equal(t, changes[1].Mask, ChangeTypeMove)
		require.Equal(t, changes[2].Mask, ChangeTypeModify)
		require.Equal(t, changes[3].Mask, ChangeTypeAdd)

		change := CombineChanges(changes)
		require.Equal(t, change.MovedTo, "")
		require.Equal(t, change.WasPreviouslyAt, "/x")
		require.Equal(
			t,
			change.Mask,
			ChangeTypeRemove|ChangeTypeMove|ChangeTypeModify|ChangeTypeAdd,
		)
	})
}

func TestChangeCombineMoveBackAndForth(t *testing.T) {
	c.WithDummyLinker(t, func(lkr *c.Linker) {
		x := c.MustTouch(t, lkr, "/x", 1)
		c.MustCommit(t, lkr, "1")
		y := c.MustMove(t, lkr, x, "/y")
		c.MustCommit(t, lkr, "2")
		xx := c.MustMove(t, lkr, y, "/x")
		c.MustCommit(t, lkr, "3")

		status, err := lkr.Status()
		require.Nil(t, err)

		changes, err := History(lkr, xx, status, nil)
		require.Nil(t, err)
		require.Len(t, changes, 4)
		require.Equal(t, changes[0].Mask, ChangeTypeNone)
		require.Equal(t, changes[1].Mask, ChangeTypeMove)
		require.Equal(t, changes[2].Mask, ChangeTypeMove)
		require.Equal(t, changes[3].Mask, ChangeTypeAdd)

		change := CombineChanges(changes)
		require.Equal(t, "/x", change.WasPreviouslyAt)
		require.Equal(t, "", change.MovedTo)
		require.Equal(t, ChangeTypeAdd, change.Mask)
	})
}

func TestChangeRemoveAndReadd(t *testing.T) {
	c.WithDummyLinker(t, func(lkr *c.Linker) {
		x := c.MustTouch(t, lkr, "/x", 1)
		c.MustCommit(t, lkr, "1")
		c.MustRemove(t, lkr, x)
		c.MustCommit(t, lkr, "2")
		xx := c.MustTouch(t, lkr, "/x", 2)
		c.MustCommit(t, lkr, "3")

		status, err := lkr.Status()
		require.Nil(t, err)

		changes, err := History(lkr, xx, status, nil)
		require.Nil(t, err)
		require.Len(t, changes, 4)
		require.Equal(t, changes[0].Mask, ChangeTypeNone)
		require.Equal(t, changes[1].Mask, ChangeTypeAdd|ChangeTypeModify)
		require.Equal(t, changes[2].Mask, ChangeTypeRemove)
		require.Equal(t, changes[3].Mask, ChangeTypeAdd)

		change := CombineChanges(changes)
		require.Equal(t, "", change.MovedTo)
		require.Equal(t, ChangeTypeAdd|ChangeTypeModify, change.Mask)
	})
}

func TestChangeReplay(t *testing.T) {
	tcs := []struct {
		name  string
		setup func(t *testing.T, lkrSrc, lkrDst *c.Linker) n.ModNode
		check func(t *testing.T, lkrSrc, lkrDst *c.Linker, srcNd n.ModNode)
	}{
		{
			name: "basic-add",
			setup: func(t *testing.T, lkrSrc, lkrDst *c.Linker) n.ModNode {
				return c.MustTouch(t, lkrSrc, "/x", 1)
			},
			check: func(t *testing.T, lkrSrc, lkrDst *c.Linker, srcNd n.ModNode) {
				dstX, err := lkrDst.LookupFile("/x")
				require.Nil(t, err)

				require.Equal(t, dstX.Size(), srcNd.Size())
				require.Equal(t, dstX.TreeHash(), srcNd.TreeHash())
				require.Equal(t, dstX.BackendHash(), srcNd.BackendHash())
				require.Equal(t, dstX.ContentHash(), srcNd.ContentHash())
			},
		}, {
			name: "basic-modify",
			setup: func(t *testing.T, lkrSrc, lkrDst *c.Linker) n.ModNode {
				c.MustTouch(t, lkrDst, "/x", 0)
				return c.MustTouch(t, lkrSrc, "/x", 1)
			},
			check: func(t *testing.T, lkrSrc, lkrDst *c.Linker, srcNd n.ModNode) {
				dstX, err := lkrDst.LookupFile("/x")
				require.Nil(t, err)

				require.Equal(t, dstX.Size(), srcNd.Size())
				require.Equal(t, dstX.TreeHash(), srcNd.TreeHash())
				require.Equal(t, dstX.BackendHash(), srcNd.BackendHash())
				require.Equal(t, dstX.ContentHash(), srcNd.ContentHash())
			},
		}, {
			name: "basic-remove",
			setup: func(t *testing.T, lkrSrc, lkrDst *c.Linker) n.ModNode {
				c.MustTouch(t, lkrDst, "/x", 1)
				srcX := c.MustTouch(t, lkrSrc, "/x", 1)
				return c.MustRemove(t, lkrSrc, srcX)
			},
			check: func(t *testing.T, lkrSrc, lkrDst *c.Linker, srcNd n.ModNode) {
				// it's enough to assert that it's a ghost now:
				_, err := lkrDst.LookupGhost("/x")
				require.Nil(t, err)
			},
		}, {
			name: "basic-move",
			setup: func(t *testing.T, lkrSrc, lkrDst *c.Linker) n.ModNode {
				c.MustTouch(t, lkrDst, "/x", 1)
				srcX := c.MustTouch(t, lkrSrc, "/x", 1)
				c.MustCommit(t, lkrSrc, "move")
				return c.MustMove(t, lkrSrc, srcX, "/y")
			},
			check: func(t *testing.T, lkrSrc, lkrDst *c.Linker, srcNd n.ModNode) {
				// it's enough to assert that it's a ghost now:
				_, err := lkrDst.LookupGhost("/x")
				require.Nil(t, err)

				_, err = lkrDst.LookupFile("/y")
				require.Nil(t, err)
			},
		}, {
			name: "basic-all",
			setup: func(t *testing.T, lkrSrc, lkrDst *c.Linker) n.ModNode {
				c.MustTouch(t, lkrDst, "/x", 1)

				srcX := c.MustTouch(t, lkrSrc, "/x", 1)
				c.MustCommit(t, lkrSrc, "touch")
				srcY := c.MustMove(t, lkrSrc, srcX, "/y").(*n.File)
				c.MustCommit(t, lkrSrc, "move")
				return c.MustRemove(t, lkrSrc, srcY)
			},
			check: func(t *testing.T, lkrSrc, lkrDst *c.Linker, srcNd n.ModNode) {
				// it's enough to assert that it's a ghost now:
				_, err := lkrDst.LookupGhost("/x")
				require.Nil(t, err)

				_, err = lkrDst.LookupGhost("/y")
				require.Nil(t, err)
			},
		}, {
			name: "basic-mkdir",
			setup: func(t *testing.T, lkrSrc, lkrDst *c.Linker) n.ModNode {
				return c.MustMkdir(t, lkrSrc, "/sub")
			},
			check: func(t *testing.T, lkrSrc, lkrDst *c.Linker, srcNd n.ModNode) {
				dir, err := lkrDst.LookupDirectory("/sub")
				require.Nil(t, err)
				require.Equal(t, dir.Path(), "/sub")
			},
		}, {
			name: "edge-conflicting-types",
			setup: func(t *testing.T, lkrSrc, lkrDst *c.Linker) n.ModNode {
				// Directory and file:
				c.MustMkdir(t, lkrDst, "/sub")
				return c.MustTouch(t, lkrSrc, "/sub", 1)
			},
			check: func(t *testing.T, lkrSrc, lkrDst *c.Linker, srcNd n.ModNode) {
				// The directory was purged and the file should appear:
				// The policy here is "trust the remote, it's his metadata"
				_, err := lkrDst.LookupFile("/sub")
				require.Nil(t, err)
			},
		}, {
			name: "edge-modified-ghost",
			setup: func(t *testing.T, lkrSrc, lkrDst *c.Linker) n.ModNode {
				srcX := c.MustTouch(t, lkrSrc, "/x", 1)
				c.MustCommit(t, lkrSrc, "1")
				c.MustModify(t, lkrSrc, srcX, 2)
				c.MustCommit(t, lkrSrc, "2")
				return c.MustRemove(t, lkrSrc, srcX)
			},
			check: func(t *testing.T, lkrSrc, lkrDst *c.Linker, srcNd n.ModNode) {
				_, err := lkrDst.LookupGhost("/x")
				require.Nil(t, err)
			},
		}, {
			name: "edge-mkdir-existing",
			setup: func(t *testing.T, lkrSrc, lkrDst *c.Linker) n.ModNode {
				c.MustMkdir(t, lkrDst, "/sub")
				return c.MustMkdir(t, lkrSrc, "/sub")
			},
			check: func(t *testing.T, lkrSrc, lkrDst *c.Linker, srcNd n.ModNode) {
				dir, err := lkrDst.LookupDirectory("/sub")
				require.Nil(t, err)
				require.Equal(t, dir.Path(), "/sub")
			},
		}, {
			name: "edge-mkdir-existing-non-empty",
			setup: func(t *testing.T, lkrSrc, lkrDst *c.Linker) n.ModNode {
				c.MustMkdir(t, lkrDst, "/sub")
				c.MustTouch(t, lkrDst, "/sub/x", 1)
				return c.MustMkdir(t, lkrSrc, "/sub")
			},
			check: func(t *testing.T, lkrSrc, lkrDst *c.Linker, srcNd n.ModNode) {
				dir, err := lkrDst.LookupDirectory("/sub")
				require.Nil(t, err)
				require.Equal(t, dir.Path(), "/sub")

				dstX, err := lkrDst.LookupFile("/sub/x")
				require.Nil(t, err)
				require.Equal(t, dstX.Path(), "/sub/x")
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			c.WithLinkerPair(t, func(lkrSrc, lkrDst *c.Linker) {
				srcNd := tc.setup(t, lkrSrc, lkrDst)
				srcHead := c.MustCommit(t, lkrSrc, "post setup")

				srcChanges, err := History(lkrSrc, srcNd, srcHead, nil)
				require.Nil(t, err)

				ch := CombineChanges(srcChanges)
				require.Nil(t, ch.Replay(lkrDst))

				tc.check(t, lkrSrc, lkrDst, srcNd)
			})
		})
	}
}
