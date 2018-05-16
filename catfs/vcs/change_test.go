package vcs

import (
	"fmt"
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
			Mask:        ChangeTypeMove | ChangeTypeRemove,
			Head:        head,
			Next:        next,
			Curr:        curr,
			ReferToPath: "/something",
		}

		msg, err := change.ToCapnp()
		require.Nil(t, err)

		newChange := &Change{}
		require.Nil(t, newChange.FromCapnp(msg))

		require.Equal(t, newChange.ReferToPath, "/something")
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
		require.Equal(t, change.ReferToPath, "/x")
		require.Equal(
			t,
			change.Mask,
			ChangeTypeRemove|ChangeTypeMove|ChangeTypeModify|ChangeTypeAdd,
		)
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
				fmt.Println(lkrDst.LookupModNode("/x"))
				_, err := lkrDst.LookupGhost("/x")
				require.Nil(t, err)

				dstY, err := lkrDst.LookupFile("/y")
				require.Nil(t, err)
				fmt.Println("dst y", dstY)
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

				fmt.Println("Hist", srcChanges)

				ch := CombineChanges(srcChanges)
				require.Nil(t, ch.Replay(lkrDst))

				tc.check(t, lkrSrc, lkrDst, srcNd)
			})
		})
	}
}
