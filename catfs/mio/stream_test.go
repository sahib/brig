package mio

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"testing"

	"github.com/sahib/brig/catfs/mio/compress"
	"github.com/sahib/brig/catfs/mio/encrypt"
	"github.com/sahib/brig/catfs/mio/overlay"
	"github.com/sahib/brig/util/testutil"
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
	t.Parallel()

	s64k := int64(64 * 1024)
	sizes := []int64{
		0, 1, 10, s64k, s64k - 1, s64k + 1,
		s64k * 2, s64k * 1024,
	}

	for _, size := range sizes {
		regularData := testutil.CreateDummyBuf(size)
		randomData := testutil.CreateRandomDummyBuf(size, 42)

		for algo, _ := range compress.AlgoMap {
			prefix := fmt.Sprintf("%s-size%d-", algo, size)
			t.Run(prefix+"regular", func(t *testing.T) {
				t.Parallel()
				testWriteAndRead(t, regularData, algo)
			})
			t.Run(prefix+"random", func(t *testing.T) {
				t.Parallel()
				testWriteAndRead(t, randomData, algo)
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

	_, err = stream.Seek(0, io.SeekStart)
	require.Nil(t, err)

	limitStream := LimitStream(stream, 5)

	_, err = limitStream.Seek(5, io.SeekStart)
	require.Nil(t, err)

	_, err = limitStream.Seek(6, io.SeekStart)
	require.Equal(t, err, io.EOF)

	_, err = limitStream.Seek(5, io.SeekEnd)
	require.Nil(t, err)

	_, err = limitStream.Seek(6, io.SeekEnd)
	require.Equal(t, err, io.EOF)

	_, err = stream.Seek(0, io.SeekStart)
	require.Nil(t, err)

	limitStream = LimitStream(stream, 5)

	buf := &bytes.Buffer{}
	n, err := limitStream.WriteTo(buf)
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

func benchThroughputOp(b *testing.B, size int64, rfn func(io.ReadSeeker) io.Reader, wfn func(io.Writer) io.WriteCloser) {
	b.Run("read", func(b *testing.B) {
		benchThroughputReadOp(b, size, rfn, wfn)
	})

	b.Run("write", func(b *testing.B) {
		benchThroughputWriteOp(b, size, rfn, wfn)
	})
}

func benchThroughputReadOp(b *testing.B, size int64, rfn func(io.ReadSeeker) io.Reader, wfn func(io.Writer) io.WriteCloser) {
	var data []byte
	buf := &bytes.Buffer{}

	w := wfn(buf)
	_, err := io.Copy(w, bytes.NewReader(testutil.CreateDummyBuf(size)))
	if err != nil {
		b.Fatalf("data preparation failed: %v", err)
	}

	if err := w.Close(); err != nil {
		b.Fatalf("failed to close writer: %v", err)
	}

	data = buf.Bytes()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := testutil.DumbCopy(ioutil.Discard, rfn(bytes.NewReader(data)), false, false)
		if err != nil {
			b.Fatalf("copy failed: %v", err)
		}
	}
}

func benchThroughputWriteOp(b *testing.B, size int64, rfn func(io.ReadSeeker) io.Reader, wfn func(io.Writer) io.WriteCloser) {
	data := testutil.CreateDummyBuf(size)

	read := int64(0)

	for i := 0; i < b.N; i++ {
		w := wfn(ioutil.Discard)
		n, err := io.Copy(w, bytes.NewReader(data))
		if err != nil {
			b.Fatalf("data preparation failed: %v", err)
		}

		read += n

		if err := w.Close(); err != nil {
			b.Fatalf("failed to close writer: %v", err)
		}
	}
}

type dummyWriter struct{ w io.Writer }

func (df *dummyWriter) Write(data []byte) (int, error) { return df.w.Write(data) }
func (df *dummyWriter) Close() error                   { return nil }

type combinedClosers struct {
	cls []io.WriteCloser
}

func combineClosers(cls ...io.WriteCloser) io.WriteCloser {
	return &combinedClosers{cls}
}

func (cc *combinedClosers) Write(data []byte) (int, error) {
	return cc.cls[len(cc.cls)-1].Write(data)
}

func (cc *combinedClosers) Close() error {
	for i := len(cc.cls) - 1; i >= 0; i-- {
		cc.cls[i].Close()
	}

	return nil
}

func BenchmarkThroughputReader(b *testing.B) {
	size := int64(1024 * 1024)

	b.Run("baseline", func(b *testing.B) {
		benchThroughputOp(
			b, size,
			func(r io.ReadSeeker) io.Reader {
				return r
			},
			func(w io.Writer) io.WriteCloser {
				return &dummyWriter{w}
			},
		)
	})

	b.Run("layer", func(b *testing.B) {
		benchThroughputOp(
			b, size,
			func(r io.ReadSeeker) io.Reader {
				layer := overlay.NewLayer(r)
				layer.SetSize(1024 * 1024)
				return layer
			},
			func(w io.Writer) io.WriteCloser {
				// Put all the writes into an empty reader.
				layer := overlay.NewLayer(bytes.NewReader(nil))
				layer.SetSize(1024 * 1024)
				return layer
			},
		)
	})

	b.Run("encrypt", func(b *testing.B) {
		benchThroughputOp(
			b, size,
			func(r io.ReadSeeker) io.Reader {
				key := make([]byte, 32)
				renc, err := encrypt.NewReader(r, key)

				if err != nil {
					b.Fatalf("failed to setup enc reader: %v", err)
				}

				return renc
			},
			func(w io.Writer) io.WriteCloser {
				key := make([]byte, 32)
				wenc, err := encrypt.NewWriter(w, key)
				if err != nil {
					b.Fatalf("failed to setup enc writer: %v", err)
				}

				return wenc
			},
		)
	})

	compressAlgos := []compress.AlgorithmType{
		compress.AlgoLZ4,
		compress.AlgoSnappy,
		compress.AlgoNone,
	}

	for _, algo := range compressAlgos {
		name := fmt.Sprintf("zip-%s", algo)
		b.Run(name, func(b *testing.B) {
			benchThroughputOp(
				b, size,
				func(r io.ReadSeeker) io.Reader {
					return compress.NewReader(r)
				},
				func(w io.Writer) io.WriteCloser {
					wzip, err := compress.NewWriter(w, algo)
					if err != nil {
						b.Fatalf("failed to setup zip writer: %v", err)
					}

					return wzip
				},
			)
		})
	}

	for _, algo := range compressAlgos {
		name := fmt.Sprintf("srm-%s", algo)
		b.Run(name, func(b *testing.B) {
			benchThroughputOp(
				b, size,
				func(r io.ReadSeeker) io.Reader {
					key := make([]byte, 32)
					stream, err := NewOutStream(r, key)
					if err != nil {
						b.Fatalf("failed to create out stream: %v", err)
					}

					return stream
				},
				func(w io.Writer) io.WriteCloser {
					key := make([]byte, 32)

					// Setup the writer part:
					wEnc, err := encrypt.NewWriter(w, key)
					if err != nil {
						b.Fatalf("failed to setup wenc: %v", err)
					}

					wZip, err := compress.NewWriter(wEnc, algo)
					if err != nil {
						b.Fatalf("failed to setup wzip: %v", err)
					}

					return combineClosers(wEnc, wZip)
				},
			)
		})
	}
}
