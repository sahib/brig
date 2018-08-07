package catfs

import (
	"bytes"
	"testing"

	c "github.com/sahib/brig/catfs/core"
	h "github.com/sahib/brig/util/hashlib"
	"github.com/stretchr/testify/require"
)

func TestPinMemCache(t *testing.T) {
	c.WithDummyLinker(t, func(lkr *c.Linker) {
		backend := NewMemFsBackend()
		pinner, err := NewPinner(lkr, backend)
		require.Nil(t, err)

		content := h.TestDummy(t, 1)
		require.Nil(t, pinner.remember(1, content, true, false))
		isPinned, isExplicit, err := pinner.IsPinned(1, content)
		require.Nil(t, err)

		require.True(t, isPinned)
		require.False(t, isExplicit)

		require.Nil(t, pinner.remember(1, content, true, true))
		isPinned, isExplicit, err = pinner.IsPinned(1, content)
		require.Nil(t, err)

		require.True(t, isPinned)
		require.True(t, isExplicit)

		require.Nil(t, pinner.Close())
	})
}

func TestPinRememberHashTwice(t *testing.T) {
	c.WithDummyLinker(t, func(lkr *c.Linker) {
		backend := NewMemFsBackend()
		pinner, err := NewPinner(lkr, backend)
		require.Nil(t, err)

		content := h.TestDummy(t, 1)
		require.Nil(t, pinner.remember(1, content, true, false))
		isPinned, isExplicit, err := pinner.IsPinned(1, content)
		require.Nil(t, err)
		require.True(t, isPinned)
		require.False(t, isExplicit)

		require.Nil(t, pinner.remember(2, content, true, true))
		isPinned, isExplicit, err = pinner.IsPinned(2, content)
		require.Nil(t, err)
		require.True(t, isPinned)
		require.True(t, isExplicit)

		require.Nil(t, pinner.remember(2, content, false, true))
		isPinned, isExplicit, err = pinner.IsPinned(2, content)
		require.Nil(t, err)
		require.False(t, isPinned)
		require.False(t, isExplicit)

		// old inode is still counted as pinned.
		isPinned, isExplicit, err = pinner.IsPinned(1, content)
		require.Nil(t, err)
		require.True(t, isPinned)
		require.False(t, isExplicit)

		require.Nil(t, pinner.Close())
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

func TestPinEntryMarshal(t *testing.T) {
	pinEntry := &pinCacheEntry{
		Inodes: map[uint64]bool{
			10: true,
			15: false,
			20: true,
		},
	}

	data, err := pinEnryToCapnpData(pinEntry)
	require.Nil(t, err)

	loadedPinEntry, err := capnpToPinCacheEntry(data)
	require.Nil(t, err)

	require.Equal(t, pinEntry, loadedPinEntry)
}
