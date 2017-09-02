package cafs

import (
	"testing"

	"github.com/disorganizer/brig/cafs/db"
)

func setupBasicFile(t *testing.T, lkrSrc, lkrDst *Linker) {
	touchFile(t, lkrSrc, "/x.png", 1)
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
			withDummyKv(t, func(kvSrc db.Database) {
				withDummyKv(t, func(kvDst db.Database) {
					lkrSrc := NewLinker(kvSrc)
					lkrDst := NewLinker(kvDst)

					tc.setup(t, lkrSrc, lkrDst)

					syncer := NewSyncer(nil)
					if err := syncer.Sync(lkrSrc, lkrDst); err != nil {
						t.Fatalf("sync failed: %v", err)
					}

					tc.check(t, lkrSrc, lkrDst)
				})
			})
		})
	}
}
