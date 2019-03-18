package httpipfs

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIpfsStartup(t *testing.T) {
	WithIpfs(t, 1, func(t *testing.T, ipfsPath string) {
		nd, err := NewNode(ipfsPath, "")
		require.Nil(t, err)

		hash, err := nd.Add(bytes.NewReader([]byte("hello")))
		require.Nil(t, err, fmt.Sprintf("%v", err))
		require.Equal(t, "QmWfVY9y3xjsixTgbd9AorQxH7VtMpzfx2HaWtsoUYecaX", hash.String())
	})
}

func TestDoubleIpfsStartup(t *testing.T) {
	WithDoubleIpfs(t, 1, func(t *testing.T, ipfsPathA, ipfsPathB string) {
		ndA, err := NewNode(ipfsPathA, "")
		require.Nil(t, err)

		ndB, err := NewNode(ipfsPathB, "")
		require.Nil(t, err)

		idA, err := ndA.Identity()
		require.Nil(t, err, fmt.Sprintf("%v", err))

		idB, err := ndB.Identity()
		require.Nil(t, err)

		require.NotEqual(t, idA.Addr, idB.Addr)
	})
}
