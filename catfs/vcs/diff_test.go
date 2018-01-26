package vcs

import (
	"testing"

	c "github.com/sahib/brig/catfs/core"
	"github.com/stretchr/testify/require"
)

func setupDiffBasicSrcFile(t *testing.T, lkrSrc, lkrDst *c.Linker) {
	c.MustTouch(t, lkrSrc, "/x.png", 1)
}

func checkDiffBasicSrcFileForward(t *testing.T, lkrSrc, lkrDst *c.Linker, diff *Diff) {
	require.Empty(t, diff.Removed)
	require.Empty(t, diff.Conflict)
	require.Empty(t, diff.Ignored)
	require.Empty(t, diff.Merged)

	require.Len(t, diff.Added, 1)
	require.Equal(t, "/x.png", diff.Added[0].Path())
}

func checkDiffBasicSrcFileBackward(t *testing.T, lkrSrc, lkrDst *c.Linker, diff *Diff) {
	require.Empty(t, diff.Added)
	require.Empty(t, diff.Conflict)
	require.Empty(t, diff.Ignored)
	require.Empty(t, diff.Merged)

	require.Len(t, diff.Removed, 1)
	require.Equal(t, "/x.png", diff.Removed[0].Path())
}

///////////////

func TestDiff(t *testing.T) {
	tcs := []struct {
		name          string
		setup         func(t *testing.T, lkrSrc, lkrDst *c.Linker)
		checkForward  func(t *testing.T, lkrSrc, lkrDst *c.Linker, diff *Diff)
		checkBackward func(t *testing.T, lkrSrc, lkrDst *c.Linker, diff *Diff)
	}{
		{
			name:          "basic-src-file",
			setup:         setupDiffBasicSrcFile,
			checkForward:  checkDiffBasicSrcFileForward,
			checkBackward: checkDiffBasicSrcFileBackward,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			c.WithLinkerPair(t, func(lkrSrc, lkrDst *c.Linker) {
				c.MustTouch(t, lkrSrc, "/README", 42)
				c.MustTouch(t, lkrDst, "/README", 42)

				c.MustCommitIfPossible(t, lkrDst, "setup dst")
				c.MustCommitIfPossible(t, lkrSrc, "setup src")

				tc.setup(t, lkrSrc, lkrDst)

				srcHead, err := lkrSrc.Head()
				require.Nil(t, err)

				srcStatus, err := lkrSrc.Status()
				require.Nil(t, err)

				diff, err := MakeDiff(lkrSrc, lkrSrc, srcStatus, srcHead, nil)
				if err != nil {
					t.Fatalf("diff forward failed: %v", err)
				}

				tc.checkForward(t, lkrSrc, lkrDst, diff)

				diff, err = MakeDiff(lkrSrc, lkrSrc, srcHead, srcStatus, nil)
				if err != nil {
					t.Fatalf("diff backward failed: %v", err)
				}

				tc.checkBackward(t, lkrSrc, lkrDst, diff)

				// Checking the same commit should always result into an empty diff:
				// We could of course cheat and check the hash to be equal,
				// but this is helpful to validate the implementation.
				diff, err = MakeDiff(lkrSrc, lkrSrc, srcHead, srcHead, nil)
				if err != nil {
					t.Fatalf("diff equal failed: %v", err)
				}

				require.Empty(t, diff.Added)
				require.Empty(t, diff.Removed)
				require.Empty(t, diff.Conflict)
				require.Empty(t, diff.Ignored)
				require.Empty(t, diff.Merged)
			})
		})
	}
}
