package httpipfs

import (
	"bytes"
	"testing"

	"github.com/sahib/brig/util/testutil"
	"github.com/stretchr/testify/require"
)

func TestPinUnpin(t *testing.T) {
	nd, err := NewNode(5001)
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
}
