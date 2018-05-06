package catfs

import (
	"bytes"
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

func TestPinMemCache(t *testing.T) {
	c.WithDummyLinker(t, func(lkr *c.Linker) {
		withTempDirectory(t, func(dir string) {
			backend := NewMemFsBackend()
			pinner, err := NewPinner(dir, lkr, backend)
			require.Nil(t, err)

			content := h.TestDummy(t, 1)
			require.Nil(t, pinner.remember(content, true, false))
			isPinned, isExplicit, err := pinner.IsPinned(content)
			require.Nil(t, err)

			require.True(t, isPinned)
			require.False(t, isExplicit)

			require.Nil(t, pinner.remember(content, true, true))
			isPinned, isExplicit, err = pinner.IsPinned(content)
			require.Nil(t, err)

			require.True(t, isPinned)
			require.True(t, isExplicit)

			require.Nil(t, pinner.Close())
		})
	})
}

func TestPinNode(t *testing.T) {
	withDummyFS(t, func(fs *FS) {
		require.Nil(t, fs.Stage("/x", bytes.NewReader([]byte{1})))
		x, err := fs.lkr.LookupFile("/x")
		require.Nil(t, err)
		require.Nil(t, fs.pinner.PinNode(x, false))

		isPinned, isExplicit, err := fs.pinner.IsNodePinned(x)
		require.Nil(t, err)
		require.True(t, isPinned)
		require.False(t, isExplicit)

		require.Nil(t, fs.pinner.PinNode(x, true))
		isPinned, isExplicit, err = fs.pinner.IsNodePinned(x)
		require.Nil(t, err)
		require.True(t, isPinned)
		require.True(t, isExplicit)

		// Downgrade unpin(false) when explicit => no change.
		require.Nil(t, fs.pinner.UnpinNode(x, false))
		isPinned, isExplicit, err = fs.pinner.IsNodePinned(x)
		require.Nil(t, err)
		require.True(t, isPinned)
		require.True(t, isExplicit)

		require.Nil(t, fs.pinner.UnpinNode(x, true))
		isPinned, isExplicit, err = fs.pinner.IsNodePinned(x)
		require.Nil(t, err)
		require.False(t, isPinned)
		require.False(t, isExplicit)
	})
}
