package httpipfs

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"testing"

	"github.com/sahib/brig/util/testutil"
	"github.com/stretchr/testify/require"
)

func TestAddCatBasic(t *testing.T) {
	withIpfs(t, 1, func(t *testing.T, apiPort int) {
		nd, err := NewNode(apiPort)
		require.Nil(t, err)

		data := testutil.CreateDummyBuf(4096 * 1024)
		hash, err := nd.Add(bytes.NewReader(data))
		require.Nil(t, err)

		fmt.Println(hash)

		stream, err := nd.Cat(hash)
		require.Nil(t, err)

		echoData, err := ioutil.ReadAll(stream)
		require.Nil(t, err)
		require.Equal(t, data, echoData)
	})
}

func TestAddCatSize(t *testing.T) {
	withIpfs(t, 1, func(t *testing.T, apiPort int) {
		nd, err := NewNode(apiPort)
		require.Nil(t, err)

		data := testutil.CreateDummyBuf(4096 * 1024)
		hash, err := nd.Add(bytes.NewReader(data))
		require.Nil(t, err)

		stream, err := nd.Cat(hash)
		require.Nil(t, err)

		size, err := stream.Seek(0, io.SeekEnd)
		require.Nil(t, err)
		require.Equal(t, int64(len(data)), size)

		off, err := stream.Seek(0, io.SeekStart)
		require.Nil(t, err)
		require.Equal(t, int64(0), off)

		echoData, err := ioutil.ReadAll(stream)
		require.Nil(t, err)
		require.Equal(t, data, echoData)
	})
}
