package mio

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/disorganizer/brig/catfs/mio/compress"
	"github.com/disorganizer/brig/util/testutil"
	"github.com/stretchr/testify/require"
)

var TestKey = []byte("01234567890ABCDE01234567890ABCDE")

type wrapReader struct {
	io.Reader
	io.Seeker
	io.Closer
	io.WriterTo
}

func testWriteAndRead(t *testing.T, raw []byte, algoType compress.AlgorithmType) {
	rawBuf := &bytes.Buffer{}
	if _, err := rawBuf.Write(raw); err != nil {
		t.Errorf("Huh, buf-write failed?")
		return
	}

	encStream, err := NewInStream(rawBuf, TestKey, algoType)
	if err != nil {
		t.Errorf("Creating encryption stream failed: %v", err)
		return
	}

	encrypted := &bytes.Buffer{}
	if _, err = io.Copy(encrypted, encStream); err != nil {
		t.Errorf("Reading encrypted data failed: %v", err)
		return
	}

	// Fake a close method:
	br := bytes.NewReader(encrypted.Bytes())

	r := wrapReader{
		Reader:   br,
		Seeker:   br,
		WriterTo: br,
		Closer:   ioutil.NopCloser(nil),
	}

	decStream, err := NewOutStream(r, TestKey)
	if err != nil {
		t.Errorf("Creating decryption stream failed: %v", err)
		return
	}

	decrypted := &bytes.Buffer{}
	if _, err = io.Copy(decrypted, decStream); err != nil {
		t.Errorf("Reading decrypted data failed: %v", err)
		return
	}

	if !bytes.Equal(decrypted.Bytes(), raw) {
		t.Errorf("Raw and decrypted is not equal => BUG.")
		t.Errorf("RAW:\n  %v", raw)
		t.Errorf("DEC:\n  %v", decrypted.Bytes())
		return
	}
}

func TestWriteAndRead(t *testing.T) {
	s64k := int64(64 * 1024)
	sizes := []int64{
		0, 1, 10, s64k, s64k - 1, s64k + 1,
		s64k * 2, s64k * 1024,
	}

	for _, size := range sizes {
		t.Logf("Testing stream at size %d", size)
		regularData := testutil.CreateDummyBuf(size)
		randomData := testutil.CreateRandomDummyBuf(size, 42)

		for algoType, _ := range compress.AlgoMap {
			testWriteAndRead(t, regularData, algoType)
			testWriteAndRead(t, randomData, algoType)
		}
	}
}

func TestLimitedStream(t *testing.T) {
	testData := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	r := bytes.NewReader(testData)

	// Fake a stream without all the encryption/compression fuzz.
	stream := struct {
		io.Reader
		io.Seeker
		io.Closer
		io.WriterTo
	}{
		Reader:   r,
		Seeker:   r,
		WriterTo: r,
		Closer:   ioutil.NopCloser(r),
	}

	for idx := 0; idx <= 10; idx++ {
		// Seek back to beginning:
		_, err := stream.Seek(0, os.SEEK_SET)
		require.Nil(t, err)

		smallStream := LimitStream(stream, uint64(idx))
		data, err := ioutil.ReadAll(smallStream)
		require.Nil(t, err)
		require.Equal(t, testData[:idx], data)
	}

	var err error

	_, err = stream.Seek(0, os.SEEK_SET)
	require.Nil(t, err)

	limitStream := LimitStream(stream, 5)

	_, err = limitStream.Seek(5, os.SEEK_SET)
	require.Nil(t, err)

	_, err = limitStream.Seek(6, os.SEEK_SET)
	require.Equal(t, err, io.EOF)

	_, err = limitStream.Seek(5, os.SEEK_END)
	require.Nil(t, err)

	_, err = limitStream.Seek(6, os.SEEK_END)
	require.Equal(t, err, io.EOF)

	_, err = stream.Seek(0, os.SEEK_SET)
	require.Nil(t, err)
	limitStream = LimitStream(stream, 5)

	buf := &bytes.Buffer{}
	n, err := limitStream.WriteTo(buf)
	require.Nil(t, err)
	require.Equal(t, n, int64(5))
	require.Equal(t, buf.Bytes(), testData[:5])

	buf.Reset()
	_, err = stream.Seek(0, os.SEEK_SET)
	require.Nil(t, err)
	limitStream = LimitStream(stream, 11)

	n, err = limitStream.WriteTo(buf)
	require.Nil(t, err)
	require.Equal(t, n, int64(10))
	require.Equal(t, buf.Bytes(), testData)
}
