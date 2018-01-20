package vcs

import (
	"fmt"
	"testing"

	c "github.com/sahib/brig/catfs/core"
	"github.com/stretchr/testify/require"
)

func setupDiffBasicSrcFile(t *testing.T, lkrSrc, lkrDst *c.Linker) {
	c.MustTouch(t, lkrSrc, "/x.png", 1)
}

func checkDiffBasicSrcFile(t *testing.T, lkrSrc, lkrDst *c.Linker, diff *Diff) {
	fmt.Println("diff")
	fmt.Println(diff)
}

///////////////

func TestDiff(t *testing.T) {
	tcs := []struct {
		name  string
		setup func(t *testing.T, lkrSrc, lkrDst *c.Linker)
		check func(t *testing.T, lkrSrc, lkrDst *c.Linker, diff *Diff)
	}{
		{
			name:  "basic-src-file",
			setup: setupDiffBasicSrcFile,
			check: checkDiffBasicSrcFile,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			c.WithLinkerPair(t, func(lkrSrc, lkrDst *c.Linker) {
				tc.setup(t, lkrSrc, lkrDst)
				c.MustCommitIfPossible(t, lkrDst, "setup dst")
				c.MustCommitIfPossible(t, lkrSrc, "setup src")

				srcStatus, err := lkrSrc.Status()
				require.Nil(t, err)

				srcHead, err := lkrSrc.Head()
				require.Nil(t, err)

				// dstStatus, err := lkrDst.Status()
				// require.Nil(t, err)

				diff, err := MakeDiff(lkrSrc, lkrSrc, srcStatus, srcHead, nil)
				if err != nil {
					t.Fatalf("diff failed: %v", err)
				}

				tc.check(t, lkrSrc, lkrDst, diff)
			})
		})
	}
}
