package catfs

import (
	"testing"

	c "github.com/sahib/brig/catfs/core"
	"github.com/stretchr/testify/require"
)

func TestRevParse(t *testing.T) {
	c.WithDummyLinker(t, func(lkr *c.Linker) {
		init, err := parseRev(lkr, "commit[0]")
		require.Nil(t, err)
		require.Equal(t, "init", init.Message())
	})
}
