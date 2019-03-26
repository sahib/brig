package httpipfs

import (
	"bytes"
	"testing"

	h "github.com/sahib/brig/util/hashlib"
	"github.com/sahib/brig/util/testutil"
	"github.com/stretchr/testify/require"
)

func TestPinUnpin(t *testing.T) {
	WithIpfs(t, 1, func(t *testing.T, ipfsPath string) {
		nd, err := NewNode(ipfsPath, "")
		require.Nil(t, err)

		data := testutil.CreateDummyBuf(4096 * 1024)
		hash, err := nd.Add(bytes.NewReader(data))
		require.Nil(t, err)

		isPinned, err := nd.IsPinned(hash)
		require.Nil(t, err)
		require.True(t, isPinned)

		require.Nil(t, nd.Unpin(hash))

		isPinned, err = nd.IsPinned(hash)
		require.Nil(t, err)
		require.False(t, isPinned)

		require.Nil(t, nd.Pin(hash))

		isPinned, err = nd.IsPinned(hash)
		require.Nil(t, err)
		require.True(t, isPinned)
	})
}

func TestIsCached(t *testing.T) {
	WithIpfs(t, 1, func(t *testing.T, ipfsPath string) {
		nd, err := NewNode(ipfsPath, "")
		require.Nil(t, err)

		hash, err := nd.Add(bytes.NewReader([]byte{1, 2, 3}))
		require.Nil(t, err)

		isCached, err := nd.IsCached(hash)
		require.Nil(t, err)
		require.True(t, isCached)

		// Let's just hope this hash does not exist locally:
		dummyHash, err := h.FromB58String("QmanyEbg6appBzzGaGMZm9NKqPVCbrWaB8ayGDerWh6aMB")
		require.Nil(t, err)

		isCached, err = nd.IsCached(dummyHash)
		require.Nil(t, err)
		require.False(t, isCached)
	})
}
