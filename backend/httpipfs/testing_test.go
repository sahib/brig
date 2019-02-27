package httpipfs

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIpfsStartup(t *testing.T) {
	WithIpfs(t, 1, func(t *testing.T, apiPort int) {
		nd, err := NewNode(apiPort)
		require.Nil(t, err)

		hash, err := nd.Add(bytes.NewReader([]byte("hello")))
		require.Nil(t, err, fmt.Sprintf("%v", err))
		require.Equal(t, "QmWfVY9y3xjsixTgbd9AorQxH7VtMpzfx2HaWtsoUYecaX", hash.String())
	})
}

func TestDoubleIpfsStartup(t *testing.T) {
	WithDoubleIpfs(t, 1, func(t *testing.T, apiPortA, apiPortB int) {
		ndA, err := NewNode(apiPortA)
		require.Nil(t, err)

		ndB, err := NewNode(apiPortB)
		require.Nil(t, err)

		idA, err := ndA.Identity()
		require.Nil(t, err, fmt.Sprintf("%v", err))

		idB, err := ndB.Identity()
		require.Nil(t, err)

		require.NotEqual(t, idA.Addr, idB.Addr)
	})
}
