package vcs

import (
	"testing"

	c "github.com/sahib/brig/catfs/core"
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
