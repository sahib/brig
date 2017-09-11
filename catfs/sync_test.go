package catfs

import (
	"testing"

	n "github.com/disorganizer/brig/catfs/nodes"
	"github.com/stretchr/testify/require"
)

type expect struct {
	dstMergeCmt *n.Commit
	srcMergeCmt *n.Commit

	srcFile *n.File
	dstFile *n.File

	err error
}

func setupResolveBasicNoConflict(t *testing.T, lkrSrc, lkrDst *Linker) *expect {
	src, _ := mustTouchAndCommit(t, lkrSrc, "/x.png", 1)
	dst, _ := mustTouchAndCommit(t, lkrDst, "/x.png", 2)

	return &expect{
		dstMergeCmt: nil,
		srcMergeCmt: nil,
		srcFile:     src,
		dstFile:     dst,
		err:         ErrConflict,
	}
}

func TestResolve(t *testing.T) {
	tcs := []struct {
		name  string
		setup func(t *testing.T, lkrSrc, lkrDst *Linker) *expect
	}{
		{
			name:  "basic-no-conflict-file",
			setup: setupResolveBasicNoConflict,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			withLinkerPair(t, func(lkrSrc, lkrDst *Linker) {
				expect := tc.setup(t, lkrSrc, lkrDst)

				syncer := NewSyncer(lkrSrc, lkrDst, nil)
				if err := syncer.cacheLastCommonMerge(); err != nil {
					t.Fatalf("Failed to find last common merge.")
				}

				require.Equal(t, expect.dstMergeCmt, syncer.dstMergeCmt, "dst merge marker")
				require.Equal(t, expect.srcMergeCmt, syncer.srcMergeCmt, "src merge marker")

				err := syncer.resolve(expect.srcFile, expect.dstFile)
				if expect.err != err {
					t.Fatalf("Resolve failed with wrong error: %v (want %v)", err, expect.err)
				}
			})
		})
	}
}

///////////////////////////
// HIGH LEVEL SYNC TESTS //
///////////////////////////

func setupBasicFile(t *testing.T, lkrSrc, lkrDst *Linker) {
	mustTouch(t, lkrSrc, "/x.png", 1)
}

func checkBasicFile(t *testing.T, lkrSrc, lkrDst *Linker) {
	// TODO: Really implement checks.
}

func TestSync(t *testing.T) {
	tcs := []struct {
		name  string
		setup func(t *testing.T, lkrSrc, lkrDst *Linker)
		check func(t *testing.T, lkrSrc, lkrDst *Linker)
	}{
		{
			name:  "basic-file",
			setup: setupBasicFile,
			check: checkBasicFile,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			withLinkerPair(t, func(lkrSrc, lkrDst *Linker) {
				tc.setup(t, lkrSrc, lkrDst)

				syncer := NewSyncer(lkrSrc, lkrDst, nil)
				if err := syncer.Sync(); err != nil {
					t.Fatalf("sync failed: %v", err)
				}

				tc.check(t, lkrSrc, lkrDst)
			})
		})
	}
}
