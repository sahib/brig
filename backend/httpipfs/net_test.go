package httpipfs

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	TestProtocol = "/brig/test/1.0"
)

var (
	TestMessage = []byte("Hello World!")
)

func testClientSide(t *testing.T, portB int, addr string) {
	nd, err := NewNode(portB)
	require.Nil(t, err)

	conn, err := nd.Dial(addr, TestProtocol)
	require.Nil(t, err)

	defer func() {
		require.Nil(t, conn.Close())
	}()

	_, err = conn.Write(TestMessage)
	require.Nil(t, err)
}

func TestDialAndListen(t *testing.T) {
	WithDoubleIpfs(t, 1, func(t *testing.T, portA, portB int) {
		nd, err := NewNode(portA)
		require.Nil(t, err)

		lst, err := nd.Listen(TestProtocol)
		require.Nil(t, err)
		defer func() {
			require.Nil(t, lst.Close())
		}()

		id, err := nd.Identity()
		require.Nil(t, err)

		go testClientSide(t, portB, id.Addr)

		conn, err := lst.Accept()
		require.Nil(t, err)

		buf := &bytes.Buffer{}
		_, err = io.Copy(buf, conn)
		require.Nil(t, err)
		require.Equal(t, TestMessage, buf.Bytes())
	})
}

func TestPing(t *testing.T) {
	WithDoubleIpfs(t, 1, func(t *testing.T, portA, portB int) {
		ndA, err := NewNode(portA)
		require.Nil(t, err)

		idA, err := ndA.Identity()
		require.Nil(t, err)

		pinger, err := ndA.Ping(idA.Addr)
		require.Nil(t, err)

		defer func() {
			require.Nil(t, pinger.Close())
		}()

		for idx := 0; idx < 60; idx++ {
			if pinger.Err() != ErrWaiting {
				break
			}

			time.Sleep(1 * time.Second)
		}

		require.Nil(t, pinger.Err())
		require.True(t, pinger.Roundtrip() < time.Second)
		require.True(t, time.Since(pinger.LastSeen()) < time.Second)
	})
}
