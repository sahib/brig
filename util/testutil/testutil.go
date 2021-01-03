package testutil

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"testing"
)

// CreateDummyBuf creates a byte slice that is `size` big.
// It's filled with the repeating numbers [0...254].
func CreateDummyBuf(size int64) []byte {
	buf := make([]byte, size)

	for i := int64(0); i < size; i++ {
		// Be evil and stripe the data, %255 is not an mistake:
		buf[i] = byte(i % 255)
	}

	return buf
}

// CreateRandomDummyBuf creates data that is evenly distributed
// and therefore notirously hard to compress.
func CreateRandomDummyBuf(size, seed int64) []byte {
	src := rand.NewSource(seed)
	buf := make([]byte, size)

	for i := int64(0); i < size; i++ {
		buf[i] = byte(src.Int63() % 256)
	}

	return buf
}

// CreateFile creates a temporary file in the systems tmp-folder.
// The file will be `size` bytes big, filled with content from CreateDummyBuf.
func CreateFile(size int64) string {
	fd, err := ioutil.TempFile("", "brig_test")
	if err != nil {
		panic("Cannot create temp file")
	}

	blockSize := int64(1 * 1024 * 1024)
	buf := CreateDummyBuf(blockSize)

	for size > 0 {
		take := size
		if size > int64(len(buf)) {
			take = int64(len(buf))
		}

		_, err := fd.Write(buf[:take])
		if err != nil {
			panic(err)
		}

		size -= blockSize
	}

	if err := fd.Close(); err != nil {
		return ""
	}

	return fd.Name()
}

// Remover removes all files in paths recursively and errors when it fails.
// It is no error if there's nothing to delete. It's useful in defer statements.
func Remover(t *testing.T, paths ...string) {
	for _, path := range paths {
		if err := os.RemoveAll(path); err != nil {
			t.Errorf("removing temp directory failed: %v", err)
		}
	}
}

// DumbCopy works like io.Copy but may be instructed to not use WriteTo or ReadFrom
func DumbCopy(dst io.Writer, src io.Reader, useReadFrom, useWriteTo bool) (written int64, err error) {
	// If the reader has a WriteTo method, use it to do the copy.
	// Avoids an allocation and a copy.
	if wt, ok := src.(io.WriterTo); ok && useWriteTo {
		return wt.WriteTo(dst)
	}
	// Similarly, if the writer has a ReadFrom method, use it to do the copy.
	if rt, ok := dst.(io.ReaderFrom); ok && useReadFrom {
		return rt.ReadFrom(src)
	}

	buf := make([]byte, 32*1024)

	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er == io.EOF {
			break
		}
		if er != nil {
			err = er
			break
		}
	}
	return written, err
}

// RandomLocalListener returns a net.Listener that is listening on
// a random free port. You should close it when done.
func RandomLocalListener() (net.Listener, error) {
	// Asking for a port and then trying to bind it is slightly racy.
	// Protect against that by retrying a bit.
	for retries := 0; retries < 10; retries++ {
		lst, err := net.Listen("tcp", ":0")
		if err != nil {
			continue
		}

		return lst, nil
	}

	return nil, fmt.Errorf("too many retries")
}

// TenReader is an io.Reader that produces the sequence [0-9] over and over.
type TenReader struct {
	idx int64
}

func (tr *TenReader) Read(buf []byte) (int, error) {
	for bufIdx := range buf {
		buf[bufIdx] = byte(tr.idx%10) + '0'
		tr.idx++
	}

	return len(buf), nil
}

////////////////

// TenWriter is similar to ioutil.Discard,
// but checks that the incoming is the repeating sequence [0-9]
//
// Meant to be used as assert in conjunction with TenReader.
type TenWriter struct {
	idx int64
}

func (tw *TenWriter) Write(buf []byte) (int, error) {
	for bufIdx := range buf {
		expected := byte(tw.idx%10) + '0'
		if got := buf[bufIdx]; got != expected {
			return bufIdx, fmt.Errorf(
				"ten-writer error at index %d: want '%c', got: '%c'",
				tw.idx,
				expected,
				got,
			)
		}

		tw.idx++
	}

	return len(buf), nil
}

////////////////

type rr struct {
	r          io.Reader
	maxBufSize int
	randomize  bool
	rng        *rand.Rand
}

func (rr *rr) Read(buf []byte) (int, error) {
	if len(buf) > rr.maxBufSize {
		buf = buf[:rr.maxBufSize]
	}

	if rr.randomize {
		randomLen := rr.rng.Int() % (rr.maxBufSize + 1)
		if len(buf) > randomLen {
			buf = buf[:randomLen]
		}
	}

	return rr.r.Read(buf)
}

// RandomizeReads returns `r` modified, so that calls to Read()
// will return at most maxBufSize, no matter how big the buffer is.
// If `randomize` is true, the max bytes read are varied additional
// in the range [0, maxBufSize]
func RandomizeReads(r io.Reader, maxBufSize int, randomize bool) io.Reader {
	return &rr{
		r:          r,
		maxBufSize: maxBufSize,
		randomize:  randomize,
		rng:        rand.New(rand.NewSource(0xdeadbeef)),
	}
}
