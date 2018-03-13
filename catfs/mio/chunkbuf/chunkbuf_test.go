package chunkbuf

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/sahib/brig/util/testutil"
	"github.com/stretchr/testify/require"
)

func TestChunkBufBasic(t *testing.T) {
	data := testutil.CreateDummyBuf(1024)
	buf := NewChunkBuffer(data)

	copiedData, err := ioutil.ReadAll(buf)
	require.Nil(t, err)
	require.Equal(t, data, copiedData)
}

func TestChunkBufEOF(t *testing.T) {
	data := testutil.CreateDummyBuf(1024)
	buf := NewChunkBuffer(data)

	cache := make([]byte, 2048)
	n, err := buf.Read(cache)
	require.True(t, err == io.EOF)
	require.Equal(t, n, 1024)
	require.Nil(t, buf.Close())
}

func TestChunkBufWriteTo(t *testing.T) {
	data := testutil.CreateDummyBuf(1024)
	buf := NewChunkBuffer(data)

	stdBuf := &bytes.Buffer{}
	n, err := buf.WriteTo(stdBuf)
	require.Nil(t, err)
	require.Equal(t, int64(n), int64(1024))
	require.Equal(t, data, stdBuf.Bytes())
}

func TestChunkBufSeek(t *testing.T) {
	data := testutil.CreateDummyBuf(1024)
	buf := NewChunkBuffer(data)

	var err error
	var n int

	cache := make([]byte, 128)
	n, err = buf.Read(cache)
	require.Nil(t, err)
	require.Equal(t, n, 128)
	require.Equal(t, cache[:n], data[:n])

	jumpedTo, err := buf.Seek(256, io.SeekStart)
	require.Nil(t, err)
	require.Equal(t, int64(jumpedTo), int64(256))

	cache = make([]byte, 128)
	n, err = buf.Read(cache)
	require.Nil(t, err)
	require.Equal(t, n, 128)
	require.Equal(t, cache[:n], data[256:n+256])

	// read advanced by 128, add 128 to go to 512
	jumpedTo, err = buf.Seek(128, io.SeekCurrent)
	require.Nil(t, err)
	require.Equal(t, int64(jumpedTo), int64(512))

	cache = make([]byte, 128)
	n, err = buf.Read(cache)
	require.Nil(t, err)
	require.Equal(t, n, 128)
	require.Equal(t, cache[:n], data[512:n+512])

	// read advanced by 128, add 128 to go to 512
	jumpedTo, err = buf.Seek(-128, io.SeekEnd)
	require.Nil(t, err)
	require.Equal(t, int64(jumpedTo), int64(896))

	cache = make([]byte, 128)
	n, err = buf.Read(cache)
	require.Nil(t, err)
	require.Equal(t, n, 128)
	require.Equal(t, cache[:n], data[896:n+896])
}

func TestChunkBufWrite(t *testing.T) {
	data := testutil.CreateDummyBuf(1024)
	ref := testutil.CreateDummyBuf(1024)
	buf := NewChunkBuffer(data)

	ref[0] = 1
	ref[1] = 2
	ref[2] = 3

	n, err := buf.Write([]byte{1, 2, 3})
	require.Nil(t, err)
	require.Equal(t, n, 3)

	jumpedTo, err := buf.Seek(-1, io.SeekEnd)
	require.Nil(t, err)
	require.Equal(t, int64(jumpedTo), int64(1023))

	ref[1023] = 255

	n, err = buf.Write([]byte{255, 255, 255})
	require.Nil(t, err)
	require.Equal(t, n, 1)

	jumpedTo, err = buf.Seek(0, io.SeekStart)
	require.Nil(t, err)
	require.Equal(t, int64(jumpedTo), int64(0))

	stdBuf := &bytes.Buffer{}
	nWriteTo, err := buf.WriteTo(stdBuf)
	require.Nil(t, err)
	require.Equal(t, nWriteTo, int64(1024))

	require.Equal(t, stdBuf.Bytes(), ref)
}
