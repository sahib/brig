package vcs

import (
	"testing"

	c "github.com/sahib/brig/catfs/core"
	n "github.com/sahib/brig/catfs/nodes"
	"github.com/stretchr/testify/require"
)

type expect struct {
	dstMergeCmt *n.Commit
	srcMergeCmt *n.Commit

	srcFile *n.File
	dstFile *n.File

	err    error
	result bool
}

func setupResolveBasicNoConflict(t *testing.T, lkrSrc, lkrDst *c.Linker) *expect {
	src, _ := c.MustTouchAndCommit(t, lkrSrc, "/x.png", 1)
	dst, _ := c.MustTouchAndCommit(t, lkrDst, "/x.png", 2)

	return &expect{
		dstMergeCmt: nil,
		srcMergeCmt: nil,
		srcFile:     src,
		dstFile:     dst,
		err:         nil,
		result:      false,
	}
}

func TestHasConflicts(t *testing.T) {
	tcs := []struct {
		name  string
		setup func(t *testing.T, lkrSrc, lkrDst *c.Linker) *expect
	}{
		{
			name:  "basic-no-conflict-file",
			setup: setupResolveBasicNoConflict,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			c.WithLinkerPair(t, func(lkrSrc, lkrDst *c.Linker) {
				expect := tc.setup(t, lkrSrc, lkrDst)

				syncer, err := newResolver(lkrSrc, lkrDst, nil, nil, nil)
				require.Nil(t, err)

				if err := syncer.cacheLastCommonMerge(); err != nil {
					t.Fatalf("Failed to find last common merge.")
				}

				require.Equal(
					t,
					expect.dstMergeCmt,
					syncer.dstMergeCmt,
					"dst merge marker",
				)
				require.Equal(
					t,
					expect.srcMergeCmt,
					syncer.srcMergeCmt,
					"src merge marker",
				)

				result, _, _, err := syncer.hasConflicts(
					expect.srcFile,
					expect.dstFile,
				)
				if expect.err != err {
					t.Fatalf(
						"Resolve failed with wrong error: %v (want %v)",
						err, expect.err)
				}

				if expect.result == result {
					t.Fatalf(
						"resolve did not deliver the expected. Want %v, but got %v",
						expect.result,
						result,
					)
				}
			})
		})
	}
}
