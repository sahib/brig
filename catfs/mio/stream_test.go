package mio

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/sahib/brig/catfs/mio/compress"
	"github.com/sahib/brig/repo/hints"
	"github.com/sahib/brig/util/testutil"
	"github.com/stretchr/testify/require"
)

var testKey = []byte("01234567890ABCDE01234567890ABCDE")

func testWriteAndRead(
	t *testing.T,
	raw []byte,
	hint hints.Hint,
) {
	rawBuf := &bytes.Buffer{}
	if _, err := rawBuf.Write(raw); err != nil {
		t.Errorf("Huh, buf-write failed?")
		return
	}

	encStream, isRaw, err := NewInStream(
		rawBuf,
		gofakeit.Name(),
		testKey,
		hint,
	)
	if err != nil {
		t.Errorf("creating encryption stream failed: %v", err)
		return
	}

	encrypted := &bytes.Buffer{}
	if _, err = io.Copy(encrypted, encStream); err != nil {
		t.Errorf("reading encrypted data failed: %v", err)
		return
	}

	// Fake a close method:
	br := bytes.NewReader(encrypted.Bytes())

	r := stream{
		Reader:   br,
		Seeker:   br,
		WriterTo: br,
		Closer:   ioutil.NopCloser(nil),
	}

	decStream, err := NewOutStream(r, isRaw, testKey)
	if err != nil {
		t.Errorf("creating decryption stream failed: %v", err)
		return
	}

	decrypted := &bytes.Buffer{}
	if _, err = io.Copy(decrypted, decStream); err != nil {
		t.Errorf("reading decrypted data failed: %v", err)
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
	t.Parallel()

	s64k := int64(64 * 1024)
	sizes := []int64{
		0,
		1,
		10,
		s64k,
		s64k - 1,
		s64k + 1,
		s64k * 2,
		s64k * 1024,
	}

	for _, size := range sizes {
		regularData := testutil.CreateDummyBuf(size)
		randomData := testutil.CreateRandomDummyBuf(size, 42)

		for _, hint := range hints.AllPossibleHints() {
			prefix := fmt.Sprintf("%s-size-%d-", hint, size)
			t.Run(prefix+"regular", func(t *testing.T) {
				t.Parallel()
				testWriteAndRead(t, regularData, hint)
			})
			t.Run(prefix+"random", func(t *testing.T) {
				t.Parallel()
				testWriteAndRead(t, randomData, hint)
			})
		}
	}
}

func TestLimitedStream(t *testing.T) {
	t.Parallel()

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
		_, err := stream.Seek(0, io.SeekStart)
		require.Nil(t, err)

		smallStream := LimitStream(stream, uint64(idx))
		data, err := ioutil.ReadAll(smallStream)
		require.Nil(t, err)
		require.Equal(t, testData[:idx], data)
	}

	var err error

	// Reset and do some special torturing:
	_, err = stream.Seek(0, io.SeekStart)
	require.Nil(t, err)

	limitStream := LimitStream(stream, 5)

	n, err := limitStream.Seek(4, io.SeekStart)
	require.Nil(t, err)
	require.Equal(t, int64(4), n)

	n, err = limitStream.Seek(6, io.SeekStart)
	require.True(t, err == io.EOF)

	n, err = limitStream.Seek(-5, io.SeekEnd)
	require.Nil(t, err)
	require.Equal(t, int64(0), n)

	_, err = limitStream.Seek(-6, io.SeekEnd)
	require.True(t, err == io.EOF)

	_, err = stream.Seek(0, io.SeekStart)
	require.Nil(t, err)

	limitStream = LimitStream(stream, 5)

	buf := &bytes.Buffer{}
	n, err = limitStream.WriteTo(buf)
	require.Nil(t, err)
	require.Equal(t, n, int64(10))
	require.Equal(t, buf.Bytes(), testData[:5])

	buf.Reset()
	_, err = stream.Seek(0, io.SeekStart)
	require.Nil(t, err)
	limitStream = LimitStream(stream, 11)

	n, err = limitStream.WriteTo(buf)
	require.Nil(t, err)
	require.Equal(t, n, int64(10))
	require.Equal(t, buf.Bytes(), testData)
}

func TestLimitStreamSize(t *testing.T) {
	// Size taken from a dummy file that showed this bug:
	data := testutil.CreateDummyBuf(6041)
	packData, err := compress.Pack(data, compress.AlgoSnappy)
	require.Nil(t, err)

	rZip := compress.NewReader(bytes.NewReader(packData))
	stream := struct {
		io.Reader
		io.Seeker
		io.Closer
		io.WriterTo
	}{
		Reader:   rZip,
		Seeker:   rZip,
		WriterTo: rZip,
		Closer:   ioutil.NopCloser(rZip),
	}

	r := LimitStream(stream, uint64(len(data)))

	size, err := r.Seek(0, io.SeekEnd)
	require.Nil(t, err)
	require.Equal(t, int64(len(data)), size)

	off, err := r.Seek(0, io.SeekStart)
	require.Nil(t, err)
	require.Equal(t, int64(0), off)

	buf := &bytes.Buffer{}
	n, err := io.Copy(buf, r)
	require.Nil(t, err)
	require.Equal(t, int64(len(data)), n)
	require.Equal(t, data, buf.Bytes())
}

func TestStreamSizeBySeek(t *testing.T) {
	buf := &bytes.Buffer{}
	data := testutil.CreateDummyBuf(6041 * 1024)
	encStream, isRaw, err := NewInStream(
		bytes.NewReader(data),
		"",
		testKey,
		hints.Default(),
	)
	require.Nil(t, err)

	_, err = io.Copy(buf, encStream)
	require.Nil(t, err)

	stream, err := NewOutStream(
		bytes.NewReader(buf.Bytes()),
		isRaw,
		testKey,
	)
	require.Nil(t, err)

	n, err := stream.Seek(0, io.SeekEnd)
	require.Nil(t, err)
	require.Equal(t, int64(len(data)), n)

	n, err = stream.Seek(0, io.SeekStart)
	require.Nil(t, err)
	require.Equal(t, int64(0), n)

	outBuf := &bytes.Buffer{}
	n, err = io.Copy(outBuf, stream)
	require.Nil(t, err)
	require.Equal(t, int64(len(data)), n)
	require.Equal(t, outBuf.Bytes(), data)
}
