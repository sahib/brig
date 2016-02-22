// Package util implements small helper function that
// should be included in the stdlib in our opinion.
package util

import (
	"bytes"
	"io"
	"os"
	"sync"
	"sync/atomic"

	log "github.com/Sirupsen/logrus"
)

// Empty is just an empty struct.
// Empty{} reads nicer than struct{}{}
type Empty struct{}

// Min returns the minimum of a and b.
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Min64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// Max returns the maximum of a and b.
func Max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func Max64(a, b int64) int64 {
	if a < b {
		return b
	}
	return a
}

// Clamp limits x to the range [lo, hi]
func Clamp(x, lo, hi int) int {
	return Max(lo, Min(x, hi))
}

// UMin returns the unsigned minimum of a and b
func UMin(a, b uint) uint {
	if a < b {
		return a
	}
	return b
}

// UMax returns the unsigned minimum of a and b
func UMax(a, b uint) uint {
	if a < b {
		return b
	}
	return a
}

// UClamp limits x to the range [lo, hi]
func UClamp(x, lo, hi uint) uint {
	return UMax(lo, UMin(x, hi))
}

// Closer closes c. If that fails, it will log the error.
// The intended usage is for convinient defer calls only!
// It gives only little knowledge about where the error is,
// but it's slightly better than a bare defer xyz.Close()
func Closer(c io.Closer) {
	if err := c.Close(); err != nil {
		log.Errorf("Error on close `%v`: %v", c, err)
	}
}

// Touch works like the unix touch(1)
func Touch(path string) error {
	fd, err := os.Create(path)
	if err != nil {
		return err
	}

	return fd.Close()
}

// SizeAccumulator is a io.Writer that simply counts
// the amount of bytes that has been written to it.
// It's useful to count the received bytes from a reader
// in conjunction with a io.TeeReader
//
// Example usage without error handling:
//
//   s := &SizeAccumulator{}
//   teeR := io.TeeReader(r, s)
//   io.Copy(os.Stdout, teeR)
//   fmt.Printf("Wrote %d bytes to stdout\n", s.Size())
//
type SizeAccumulator struct {
	size uint64
}

// Write simply increments the internal size count without any IO.
// It can be safely called from any go routine.
func (s *SizeAccumulator) Write(buf []byte) (int, error) {
	atomic.AddUint64(&s.size, uint64(len(buf)))
	return len(buf), nil
}

// Size returns the cumulated written bytes.
// It can be safely called from any go routine.
func (s *SizeAccumulator) Size() uint64 {
	return atomic.LoadUint64(&s.size)
}

// Reset resets the size counter to 0.
func (s *SizeAccumulator) Reset() {
	atomic.StoreUint64(&s.size, 0)
}

// NopCloser returns a WriteCloser with a no-op Close method wrapping
// the provided Writer w.
func NopWriteCloser(w io.Writer) io.WriteCloser {
	return nopCloser{w}
}

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }

type syncReadWriter struct {
	io.ReadWriter
	sync.Mutex
}

func (s *syncReadWriter) Write(buf []byte) (int, error) {
	s.Lock()
	defer s.Unlock()

	return s.ReadWriter.Write(buf)
}

func (s *syncReadWriter) Read(buf []byte) (int, error) {
	s.Lock()
	defer s.Unlock()

	return s.ReadWriter.Read(buf)
}

// SyncedReadWriter returns a io.ReadWriter that protects each call
// to Read() and Write() with a sync.Mutex.
func SyncedReadWriter(w io.ReadWriter) io.ReadWriter {
	return &syncReadWriter{ReadWriter: w}
}

// SyncBuffer is a bytes.Buffer that protects each call
// to Read() and Write() with a sync.RWMutex, i.e. parallel
// access to Read() is possible, but blocks when doing a Write().
type SyncBuffer struct {
	sync.RWMutex
	buf bytes.Buffer
}

func (b *SyncBuffer) Read(p []byte) (int, error) {
	b.Lock()
	defer b.Unlock()

	return b.buf.Read(p)
}

func (b *SyncBuffer) Write(p []byte) (int, error) {
	b.Lock()
	defer b.Unlock()

	return b.buf.Write(p)
}
