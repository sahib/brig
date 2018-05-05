package ipfs

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPin(t *testing.T) {
	WithIpfs(t, func(alice *Node) {
		data := []byte{1, 2, 3}
		h, err := alice.Add(bytes.NewReader(data))
		require.Nil(t, err)

		require.Nil(t, alice.Pin(h))
		isPinned, err := alice.IsPinned(h)
		require.Nil(t, err)
		require.True(t, isPinned)

		require.Nil(t, alice.Unpin(h))
		isPinned, err = alice.IsPinned(h)
		require.Nil(t, err)
		require.False(t, isPinned)

		require.Nil(t, alice.Pin(h))
		isPinned, err = alice.IsPinned(h)
		require.Nil(t, err)
		require.True(t, isPinned)
	})
}
