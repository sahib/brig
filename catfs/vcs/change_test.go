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
