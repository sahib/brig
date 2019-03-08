package vcs

import (
	"testing"

	c "github.com/sahib/brig/catfs/core"
	h "github.com/sahib/brig/util/hashlib"
	"github.com/stretchr/testify/require"
)

// Create a file in src and check
// that it's being synced to the dst side.
func setupBasicSrcFile(t *testing.T, lkrSrc, lkrDst *c.Linker) {
	c.MustTouch(t, lkrSrc, "/x.png", 1)
}

func checkBasicSrcFile(t *testing.T, lkrSrc, lkrDst *c.Linker) {
	xFile, err := lkrDst.LookupFile("/x.png")
	require.Nil(t, err)
	require.Equal(t, xFile.Path(), "/x.png")
	require.Equal(t, xFile.BackendHash(), h.TestDummy(t, 1))
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
	require.Equal(t, xFile.BackendHash(), h.TestDummy(t, 1))
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
	require.Equal(t, xSrcFile.BackendHash(), h.TestDummy(t, 1))

	xDstFile, err := lkrDst.LookupFile("/x.png")
	require.Nil(t, err)
	require.Equal(t, xDstFile.Path(), "/x.png")
	require.Equal(t, xDstFile.BackendHash(), h.TestDummy(t, 1))
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
	require.Equal(t, xSrcFile.BackendHash(), h.TestDummy(t, 42))

	xDstFile, err := lkrDst.LookupFile("/x.png")
	require.Nil(t, err)
	require.Equal(t, xDstFile.Path(), "/x.png")
	require.Equal(t, xDstFile.BackendHash(), h.TestDummy(t, 23))

	xConflictFile, err := lkrDst.LookupFile("/x.png.conflict.0")
	require.Nil(t, err)
	require.Equal(t, xConflictFile.Path(), "/x.png.conflict.0")
	require.Equal(t, xConflictFile.BackendHash(), h.TestDummy(t, 42))
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
	xDstGhost, err := lkrDst.LookupGhost("/x.png")
	require.Nil(t, err)

	require.Equal(t, xDstGhost.Path(), "/x.png")
	require.Equal(t, xDstGhost.BackendHash(), h.TestDummy(t, 42))

	yDstFile, err := lkrDst.LookupFile("/y.png")
	require.Nil(t, err)

	require.Equal(t, yDstFile.Path(), "/y.png")
	require.Equal(t, yDstFile.BackendHash(), h.TestDummy(t, 42))
}

////////

func setupEdgeMoveDirAndModifyChild(t *testing.T, lkrSrc, lkrDst *c.Linker) {
	// Syncing recursive empty dirs require detecting and adding them recursive.
	// This was buggy before, so prevent it from happening again.
	c.MustMkdir(t, lkrSrc, "/a")
	c.MustMkdir(t, lkrDst, "/a")
	c.MustCommit(t, lkrSrc, "added dirs src")
	c.MustCommit(t, lkrDst, "added dirs dst")
}

////////

func setupEdgeEmptyDir(t *testing.T, lkrSrc, lkrDst *c.Linker) {
	// Syncing recursive empty dirs require detecting and adding them recursive.
	// This was buggy before, so prevent it from happening again.
	c.MustMkdir(t, lkrSrc, "/empty/sub/blub")
}

func checkEdgeEmptyDir(t *testing.T, lkrSrc, lkrDst *c.Linker) {
	dir, err := lkrDst.LookupDirectory("/empty/sub/blub")
	require.Nil(t, err)
	require.Equal(t, dir.Path(), "/empty/sub/blub")
}

func TestSync(t *testing.T) {
	t.Parallel()

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
		}, {
			name:  "edge-empty-dir",
			setup: setupEdgeEmptyDir,
			check: checkEdgeEmptyDir,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c.WithLinkerPair(t, func(lkrSrc, lkrDst *c.Linker) {
				tc.setup(t, lkrSrc, lkrDst)
				// sync requires that all changes are committed.
				c.MustCommitIfPossible(t, lkrDst, "setup dst")
				c.MustCommitIfPossible(t, lkrSrc, "setup src")

				if err := Sync(lkrSrc, lkrDst, nil); err != nil {
					t.Fatalf("sync failed: %v", err)
				}

				tc.check(t, lkrSrc, lkrDst)
			})
		})
	}
}

func TestSyncMergeMarker(t *testing.T) {
	c.WithLinkerPair(t, func(lkrSrc, lkrDst *c.Linker) {
		c.MustTouchAndCommit(t, lkrSrc, "/x.png", 1)
		c.MustTouchAndCommit(t, lkrDst, "/y.png", 2)

		if err := Sync(lkrSrc, lkrDst, nil); err != nil {
			t.Fatalf("sync failed: %v", err)
		}

		dstHead, err := lkrDst.Head()
		require.Nil(t, err)

		srcHead, err := lkrSrc.Head()
		require.Nil(t, err)

		mergeUser, mergeHash := dstHead.MergeMarker()
		require.Equal(t, mergeUser, "src")
		require.Equal(t, mergeHash, srcHead.TreeHash())

		c.MustTouch(t, lkrSrc, "/a.png", 3)
		c.MustTouch(t, lkrDst, "/b.png", 4)

		diff, err := MakeDiff(lkrSrc, lkrDst, nil, nil, nil)
		require.Nil(t, err)

		require.Empty(t, diff.Conflict)
		require.Empty(t, diff.Ignored)
		require.Empty(t, diff.Merged)
		require.Empty(t, diff.Removed)

		require.Len(t, diff.Added, 1)
		require.Len(t, diff.Missing, 2)

		require.Equal(t, diff.Added[0].Path(), "/a.png")
		require.Equal(t, diff.Missing[0].Path(), "/b.png")
		require.Equal(t, diff.Missing[1].Path(), "/y.png")
	})
}

func TestSyncConflictMergeMarker(t *testing.T) {
	c.WithLinkerPair(t, func(lkrSrc, lkrDst *c.Linker) {
		c.MustTouchAndCommit(t, lkrSrc, "/x.png", 1)
		c.MustTouchAndCommit(t, lkrDst, "/x.png", 2)

		if err := Sync(lkrSrc, lkrDst, nil); err != nil {
			t.Fatalf("sync failed: %v", err)
		}

		dstHead, err := lkrDst.Head()
		require.Nil(t, err)

		srcHead, err := lkrSrc.Head()
		require.Nil(t, err)

		mergeUser, mergeHash := dstHead.MergeMarker()
		require.Equal(t, mergeUser, "src")
		require.Equal(t, mergeHash, srcHead.TreeHash())

		c.MustTouch(t, lkrSrc, "/a.png", 3)
		c.MustTouch(t, lkrDst, "/a.png", 4)

		diff, err := MakeDiff(lkrSrc, lkrDst, nil, nil, nil)
		require.Nil(t, err)

		require.Len(t, diff.Merged, 0)
		require.Len(t, diff.Ignored, 1)
		require.Len(t, diff.Conflict, 1)

		require.Empty(t, diff.Moved)
		require.Empty(t, diff.Missing)
		require.Empty(t, diff.Added)
		require.Empty(t, diff.Removed)

		// a.png is new and will conflict therefore.
		require.Equal(t, diff.Conflict[0].Dst.Path(), "/a.png")
		require.Equal(t, diff.Conflict[0].Src.Path(), "/a.png")

		// The previously created conflict file should count as missing.
		require.Equal(t, diff.Ignored[0].Path(), "/x.png.conflict.0")
	})
}

func TestSyncTwiceWithMovedFile(t *testing.T) {
	c.WithLinkerPair(t, func(lkrAli, lkrBob *c.Linker) {
		aliNd, _ := c.MustTouchAndCommit(t, lkrAli, "/ali-file", 1)
		bobNd, _ := c.MustTouchAndCommit(t, lkrBob, "/bob-file", 2)

		require.Nil(t, Sync(lkrAli, lkrBob, nil))
		require.Nil(t, Sync(lkrBob, lkrAli, nil))

		c.MustMove(t, lkrAli, aliNd, "/bali-bile")
		c.MustMove(t, lkrBob, bobNd, "/blob-lile")
		c.MustCommit(t, lkrAli, "moved file")

		diff, err := MakeDiff(lkrBob, lkrAli, nil, nil, nil)
		require.Nil(t, err)

		require.Len(t, diff.Added, 0)
		require.Len(t, diff.Removed, 0)
		require.Len(t, diff.Moved, 2)
	})
}

func TestSyncConflictStrategyEmbrace(t *testing.T) {
	c.WithLinkerPair(t, func(lkrSrc, lkrDst *c.Linker) {
		c.MustTouchAndCommit(t, lkrSrc, "/x.png", 1)
		c.MustTouchAndCommit(t, lkrDst, "/x.png", 2)

		cfg := &SyncOptions{
			ConflictStrategy: ConflictStragetyEmbrace,
		}

		diff, err := MakeDiff(lkrSrc, lkrDst, nil, nil, cfg)
		require.Nil(t, err)

		require.Len(t, diff.Conflict, 1)
		require.Empty(t, diff.Merged)
		require.Empty(t, diff.Ignored)
		require.Empty(t, diff.Moved)
		require.Empty(t, diff.Missing)
		require.Empty(t, diff.Added)
		require.Empty(t, diff.Removed)

		require.Nil(t, Sync(lkrSrc, lkrDst, cfg))

		srcX, err := lkrSrc.LookupFile("/x.png")
		require.Nil(t, err)
		dstX, err := lkrDst.LookupFile("/x.png")
		require.Nil(t, err)

		require.Equal(t, srcX.ContentHash(), dstX.ContentHash())
	})
}

func TestSyncReadOnlyFolders(t *testing.T) {
	c.WithLinkerPair(t, func(lkrSrc, lkrDst *c.Linker) {
		// Create a file on alice' side:
		c.MustTouchAndCommit(t, lkrSrc, "/public/x.png", 1)
		cfg := &SyncOptions{
			ReadOnlyFolders: map[string]bool{
				"/public": true,
			},
		}

		// Sync without a config - this is "bob's" side.
		// (he does not have any read-only folders)
		require.Nil(t, Sync(lkrSrc, lkrDst, nil))

		// Both alice and bob should have the same file/content:
		srcX, err := lkrSrc.LookupFile("/public/x.png")
		require.Nil(t, err)
		dstX, err := lkrDst.LookupFile("/public/x.png")
		require.Nil(t, err)
		require.Equal(t, srcX.ContentHash(), dstX.ContentHash())

		// bob modifies /public/x.png
		c.MustModify(t, lkrDst, dstX, 2)
		dstX, err = lkrDst.LookupFile("/public/x.png")
		require.Nil(t, err)

		// let alice sync back the change of bob:
		require.Nil(t, Sync(lkrDst, lkrSrc, cfg))

		srcX, err = lkrSrc.LookupFile("/public/x.png")
		require.Nil(t, err)

		require.NotEqual(t, srcX.ContentHash(), dstX.ContentHash())
		require.Equal(t, srcX.ContentHash(), h.TestDummy(t, byte(1)))
	})
}
