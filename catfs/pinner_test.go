package catfs

import (
	"io/ioutil"
	"os"
	"testing"

	c "github.com/sahib/brig/catfs/core"
	h "github.com/sahib/brig/util/hashlib"
	"github.com/stretchr/testify/require"
)

func withTempDirectory(t *testing.T, fn func(dir string)) {
	name, err := ioutil.TempDir("", "pin-cache-tests")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	fn(name)

	if err := os.RemoveAll(name); err != nil {
		t.Fatalf("Failed to remove temp dir: %v", err)
	}
}

func TestPinCache(t *testing.T) {
	c.WithDummyLinker(t, func(lkr *c.Linker) {
		withTempDirectory(t, func(dir string) {
			backend := NewMemFsBackend()

			pinCache, err := NewPinner(dir, lkr, backend)
			require.Nil(t, err)

			content := h.TestDummy(t, 1)
			require.Nil(t, pinCache.Remember(content, true, false))
			isPinned, isExplicit, err := pinCache.IsPinned(content)
			require.Nil(t, err)

			require.True(t, isPinned)
			require.False(t, isExplicit)

			require.Nil(t, pinCache.Remember(content, true, true))
			isPinned, isExplicit, err = pinCache.IsPinned(content)
			require.Nil(t, err)

			require.True(t, isPinned)
			require.True(t, isExplicit)

			require.Nil(t, pinCache.Close())
		})
	})
}
