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

		require.Nil(t, alice.Pin(h, true))
		isPinned, isExplicit, err := alice.IsPinned(h)
		require.Nil(t, err)
		require.True(t, isPinned)
		require.True(t, isExplicit)

		require.Nil(t, alice.Pin(h, true))
		isPinned, isExplicit, err = alice.IsPinned(h)
		require.Nil(t, err)
		require.True(t, isPinned)
		require.True(t, isExplicit)

		require.Nil(t, alice.Unpin(h, true))

		isPinned, isExplicit, err = alice.IsPinned(h)
		require.Nil(t, err)
		require.False(t, isPinned)
		require.False(t, isExplicit)
	})
}

func TestPinUpgrade(t *testing.T) {
	// See if we can "upgrade" a pin from implicit to explicit.
	WithIpfs(t, func(alice *Node) {
		data := []byte{1, 2, 3}
		h, err := alice.Add(bytes.NewReader(data))
		require.Nil(t, err)

		require.Nil(t, alice.Pin(h, false))
		isPinned, isExplicit, err := alice.IsPinned(h)
		require.Nil(t, err)
		require.True(t, isPinned)
		require.False(t, isExplicit)

		require.Nil(t, alice.Pin(h, true))
		isPinned, isExplicit, err = alice.IsPinned(h)
		require.Nil(t, err)
		require.True(t, isPinned)
		require.True(t, isExplicit)

		// See if the upgrade was not overwritten.
		require.Nil(t, alice.Pin(h, false))
		isPinned, isExplicit, err = alice.IsPinned(h)
		require.Nil(t, err)
		require.True(t, isPinned)
		require.True(t, isExplicit)

		// Implicit unpin should not hurt the file:
		require.Nil(t, alice.Unpin(h, false))
		isPinned, isExplicit, err = alice.IsPinned(h)
		require.Nil(t, err)
		require.True(t, isPinned)
		require.True(t, isExplicit)

		// Implicit unpin should not hurt the file:
		require.Nil(t, alice.Unpin(h, true))
		isPinned, isExplicit, err = alice.IsPinned(h)
		require.Nil(t, err)
		require.False(t, isPinned)
		require.False(t, isExplicit)
	})
}
