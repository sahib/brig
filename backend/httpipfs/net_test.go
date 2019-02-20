package httpipfs

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	TestProtocol = "/brig/test/1.0"
)

var (
	TestMessage = []byte("Hello World!")
)

// TODO: Util to bootup ipfs dummy instances.

func testClientSide(t *testing.T) {
	nd, err := NewNode(5002)
	require.Nil(t, err)

	conn, err := nd.Dial(
		"QmUYz9dbqnYPyHCLUi7ghtiwFbdU93MQKFH4qg8iXHWcPV",
		TestProtocol,
	)
	require.Nil(t, err)

	defer func() {
		require.Nil(t, conn.Close())
	}()

	_, err = conn.Write(TestMessage)
	require.Nil(t, err)
}

func TestDialAndListen(t *testing.T) {
	nd, err := NewNode(5001)
	require.Nil(t, err)

	lst, err := nd.Listen(TestProtocol)
	require.Nil(t, err)
	defer func() {
		require.Nil(t, lst.Close())
	}()

	go testClientSide(t)

	conn, err := lst.Accept()
	require.Nil(t, err)

	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, conn)
	require.Nil(t, err)
	require.Equal(t, TestMessage, buf.Bytes())
}
