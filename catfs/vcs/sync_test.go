package vcs

import (
	"testing"

	c "github.com/disorganizer/brig/catfs/core"
	n "github.com/disorganizer/brig/catfs/nodes"
	h "github.com/disorganizer/brig/util/hashlib"
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

				syncer := NewSyncer(lkrSrc, lkrDst, nil)
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
						"resolve did not deliver the expected. Want %s, but got %s",
						expect.result,
						result,
					)
				}
			})
		})
	}
}

///////////////////////////
// HIGH LEVEL SYNC TESTS //
///////////////////////////

// Create a file in src and check
// that it's being synced to the dst side.
func setupBasicSrcFile(t *testing.T, lkrSrc, lkrDst *c.Linker) {
	c.MustTouch(t, lkrSrc, "/x.png", 1)
}

func checkBasicSrcFile(t *testing.T, lkrSrc, lkrDst *c.Linker) {
	xFile, err := lkrDst.LookupFile("/x.png")
	require.Nil(t, err)
	require.Equal(t, xFile.Path(), "/x.png")
	require.Equal(t, xFile.Content(), h.TestDummy(t, 1))
}

////////

// Only have the file on dst.
// Nothing should happen, since no pair can be found.
func setupBasicDstFile(t *testing.T, lkrSrc, lkrDst *c.Linker) {
	c.MustTouch(t, lkrDst, "/x.png", 1)
}

func checkBasicDstFile(t *testing.T, lkrSrc, lkrDst *c.Linker) {
	xFile, err := lkrDst.LookupFile("/x.png")
	require.Nil(t, err)
	require.Equal(t, xFile.Path(), "/x.png")
	require.Equal(t, xFile.Content(), h.TestDummy(t, 1))
}

////////

// Create the same file on both sides with the same content.
func setupBasicBothNoConflict(t *testing.T, lkrSrc, lkrDst *c.Linker) {
	c.MustTouch(t, lkrSrc, "/x.png", 1)
	c.MustTouch(t, lkrDst, "/x.png", 1)
}

func checkBasicBothNoConflict(t *testing.T, lkrSrc, lkrDst *c.Linker) {
	xSrcFile, err := lkrSrc.LookupFile("/x.png")
	require.Nil(t, err)
	require.Equal(t, xSrcFile.Path(), "/x.png")
	require.Equal(t, xSrcFile.Content(), h.TestDummy(t, 1))

	xDstFile, err := lkrDst.LookupFile("/x.png")
	require.Nil(t, err)
	require.Equal(t, xDstFile.Path(), "/x.png")
	require.Equal(t, xDstFile.Content(), h.TestDummy(t, 1))
}

////////

// Create the same file on both sides with different content.
// This should result in a conflict, resulting in conflict file.
func setupBasicBothConflict(t *testing.T, lkrSrc, lkrDst *c.Linker) {
	c.MustTouch(t, lkrSrc, "/x.png", 42)
	c.MustTouch(t, lkrDst, "/x.png", 23)
}

func checkBasicBothConflict(t *testing.T, lkrSrc, lkrDst *c.Linker) {
	xSrcFile, err := lkrSrc.LookupFile("/x.png")
	require.Nil(t, err)
	require.Equal(t, xSrcFile.Path(), "/x.png")
	require.Equal(t, xSrcFile.Content(), h.TestDummy(t, 42))

	xDstFile, err := lkrDst.LookupFile("/x.png")
	require.Nil(t, err)
	require.Equal(t, xDstFile.Path(), "/x.png")
	require.Equal(t, xDstFile.Content(), h.TestDummy(t, 23))

	xConflictFile, err := lkrDst.LookupFile("/x.png.conflict.0")
	require.Nil(t, err)
	require.Equal(t, xConflictFile.Path(), "/x.png.conflict.0")
	require.Equal(t, xConflictFile.Content(), h.TestDummy(t, 42))
}

////////

func setupBasicRemove(t *testing.T, lkrSrc, lkrDst *c.Linker) {
	// Create x.png on src and remove it after one commit:
	xFile := c.MustTouch(t, lkrSrc, "/x.png", 42)
	c.MustCommit(t, lkrSrc, "who let the x out")
	c.MustRemove(t, lkrSrc, xFile)

	// Create the same file on dst:
	c.MustTouch(t, lkrDst, "/x.png", 42)
}

func checkBasicRemove(t *testing.T, lkrSrc, lkrDst *c.Linker) {
	xDstFile, err := lkrDst.LookupGhost("/x.png")
	require.Nil(t, err)
	require.Equal(t, xDstFile.Path(), "/x.png")
}

////////

func setupBasicSrcMove(t *testing.T, lkrSrc, lkrDst *c.Linker) {
	// Create x.png on src and remove it after one commit:
	xFile := c.MustTouch(t, lkrSrc, "/x.png", 42)
	c.MustCommit(t, lkrSrc, "who let the x out")
	c.MustMove(t, lkrSrc, xFile, "/y.png")

	// Create the same file on dst:
	c.MustTouch(t, lkrDst, "/x.png", 42)
}

func checkBasicSrcMove(t *testing.T, lkrSrc, lkrDst *c.Linker) {
	// TODO: This test is recognized as conflict still.
	//       This is due to the way srcMask and dstMask is defined
	//       as conflict (added = conflict). Think about this more.
	// xDstFile, err := lkrDst.LookupFile("/x.png")
	// require.Nil(t, err)
	// require.Equal(t, xDstFile.Path(), "/x.png")
	// require.Equal(t, xDstFile.Content(), h.TestDummy(t, 23))
}

func TestSync(t *testing.T) {
	tcs := []struct {
		name  string
		setup func(t *testing.T, lkrSrc, lkrDst *c.Linker)
		check func(t *testing.T, lkrSrc, lkrDst *c.Linker)
	}{
		{
			name:  "basic-src-file",
			setup: setupBasicSrcFile,
			check: checkBasicSrcFile,
		}, {
			name:  "basic-dst-file",
			setup: setupBasicDstFile,
			check: checkBasicDstFile,
		}, {
			name:  "basic-both-file-no-conflict",
			setup: setupBasicBothNoConflict,
			check: checkBasicBothNoConflict,
		}, {
			name:  "basic-both-file-conflict",
			setup: setupBasicBothConflict,
			check: checkBasicBothConflict,
		}, {
			name:  "basic-src-remove",
			setup: setupBasicRemove,
			check: checkBasicRemove,
		}, {
			name:  "basic-src-move",
			setup: setupBasicSrcMove,
			check: checkBasicSrcMove,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			c.WithLinkerPair(t, func(lkrSrc, lkrDst *c.Linker) {
				tc.setup(t, lkrSrc, lkrDst)
				// sync requires that all changes are committed.
				c.MustCommitIfPossible(t, lkrDst, "setup dst")
				c.MustCommitIfPossible(t, lkrSrc, "setup src")

				syncer := NewSyncer(lkrSrc, lkrDst, nil)
				if err := syncer.Sync(); err != nil {
					t.Fatalf("sync failed: %v", err)
				}

				tc.check(t, lkrSrc, lkrDst)
			})
		})
	}
}
