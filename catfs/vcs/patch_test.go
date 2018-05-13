package vcs

import (
	"testing"

	c "github.com/sahib/brig/catfs/core"
	"github.com/stretchr/testify/require"
)

func TestPatchMarshalling(t *testing.T) {
	c.WithDummyLinker(t, func(lkr *c.Linker) {
		head, err := lkr.Head()
		require.Nil(t, err)

		curr := c.MustTouch(t, lkr, "/x", 1)
		next := c.MustCommit(t, lkr, "hello")

		change1 := &Change{
			Mask:        ChangeTypeMove | ChangeTypeRemove,
			Head:        head,
			Next:        next,
			Curr:        curr,
			ReferToPath: "/something1",
		}

		c.MustModify(t, lkr, curr, 2)
		nextNext := c.MustCommit(t, lkr, "hello")

		change2 := &Change{
			Mask:        ChangeTypeAdd | ChangeTypeModify,
			Head:        next,
			Next:        nextNext,
			Curr:        curr,
			ReferToPath: "/something2",
		}

		patch := &Patch{
			From:    head,
			Changes: []*Change{change2, change1},
		}

		msg, err := patch.ToCapnp()
		require.Nil(t, err)

		newPatch := &Patch{}
		require.Nil(t, newPatch.FromCapnp(msg))

		require.Equal(t, patch, newPatch)
	})
}
