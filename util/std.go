// Package util implements small helper function that
// should be included in the stdlib in our opinion.
package util

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"

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

// Max returns the maximum of a and b.
func Max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

// Min64 returns the minimum of a and b.
func Min64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// Max64 returns the maximum of a and b.
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

type TimeoutReadWriter struct {
	io.Writer
	io.Reader

	rtimeout time.Duration
	wtimeout time.Duration

	useDeadline bool
	rdeadline   time.Time
	wdeadline   time.Time
}

var ErrTimeout = errors.New("I/O Timeout: Operation timed out")

func (w *TimeoutReadWriter) io(p []byte, doRead bool) (n int, err error) {
	var deadline <-chan time.Time

	// Figoure out when it's too late:
	switch {
	case doRead && w.useDeadline:
		deadline = time.After(w.rdeadline.Sub(time.Now()))
	case doRead && !w.useDeadline:
		deadline = time.After(w.rtimeout)
	case !doRead && w.useDeadline:
		deadline = time.After(w.wdeadline.Sub(time.Now()))
	case !doRead && !w.useDeadline:
		deadline = time.After(w.wtimeout)
	}

	// Resever one element, so the go routine gets cleaned up
	// early even if the timeout already expired.
	done := make(chan bool, 1)
	go func() {
		if doRead {
			n, err = w.Reader.Read(p)
		} else {
			n, err = w.Writer.Write(p)
		}
		done <- true
	}()

	// Wait for something to happen:
	select {
	case <-done:
		return
	case <-deadline:
		return 0, ErrTimeout
	}
}

func (rw *TimeoutReadWriter) Read(p []byte) (n int, err error) {
	return rw.io(p, true)
}

func (rw *TimeoutReadWriter) Write(p []byte) (n int, err error) {
	return rw.io(p, false)
}

func (rw *TimeoutReadWriter) SetReadDeadline(d time.Time) error {
	rw.useDeadline = true
	rw.rdeadline = d
	return nil
}

func (rw *TimeoutReadWriter) SetWriteDeadline(d time.Time) error {
	rw.useDeadline = true
	rw.wdeadline = d
	return nil
}

func (rw *TimeoutReadWriter) SetDeadline(d time.Time) error {
	rw.SetWriteDeadline(d)
	rw.SetReadDeadline(d)
	return nil
}

func (rw *TimeoutReadWriter) SetWriteTimeout(d time.Duration) error {
	rw.wtimeout = d
	return nil
}

func (rw *TimeoutReadWriter) SetReadTimeout(d time.Duration) error {
	rw.rtimeout = d
	return nil
}

func (rw *TimeoutReadWriter) SetTimeout(d time.Duration) error {
	rw.rtimeout = d
	rw.wtimeout = d
	return nil
}

// TimeoutReadWriter wraps `w` and returns a io.Writer that times out
// after `d` elapsed with ErrTimeout if `w` didn't succeed in that time.
func NewTimeoutWriter(w io.Writer, d time.Duration) io.Writer {
	return &TimeoutReadWriter{Writer: w, wtimeout: d}
}

func NewTimeoutReader(r io.Reader, d time.Duration) io.Reader {
	return &TimeoutReadWriter{Reader: r, rtimeout: d}
}

func NewTimeoutReadWriter(rw io.ReadWriter, d time.Duration) *TimeoutReadWriter {
	return &TimeoutReadWriter{
		Reader: rw, Writer: rw,
		rtimeout: d, wtimeout: d,
	}
}

// Errors is a list of errors that render to one single message
type Errors []error

func (es Errors) Error() string {
	switch len(es) {
	case 0:
		return ""
	case 1:
		return es[0].Error()
	default:
		base := "More than one error happended:\n"
		for _, err := range es {
			base += "\t" + err.Error() + "\n"
		}

		return base
	}
}

func (es Errors) ToErr() error {
	if len(es) > 0 {
		return es
	}
	return nil
}

// OmitBytes converts a byte slice into a string representation that
// omits data in the middle if necessary. It is useful for testing
// and printing user information. `lim` is the number of bytes
//
// Example:
//
// OmitBytes([]byte{1,2,3,4}, 2)
// -> [1 ... 2]
// OmitBytes([]byte{1,2,3,4}, 4)
// -> [1, 2, 3, 4]
//
func OmitBytes(data []byte, lim int) string {
	lo := lim
	if lo > len(data) {
		lo = len(data)
	}

	hi := len(data) - lim
	if hi < 0 {
		hi = len(data)
	}

	if len(data[hi:]) > 0 {
		return fmt.Sprintf("%v ... %v", data[:lo], data[hi:])
	} else {
		return fmt.Sprintf("%v", data[:lo])
	}
}

type limitWriter struct {
	wr  io.Writer
	sz  int64
	pos int64
}

func LimitWriter(w io.Writer, sz int64) io.Writer {
	return &limitWriter{
		wr: w,
		sz: sz,
	}
}

func (lw *limitWriter) Write(buf []byte) (int, error) {
	if lw.pos >= lw.sz {
		return len(buf), nil
	}

	n := Min64(lw.sz-lw.pos, int64(len(buf)))
	lw.pos += n

	_, err := lw.wr.Write(buf[:n])
	if err != nil {
		return -1, err
	}

	return len(buf), nil
}

type prefixReader struct {
	data []byte
	curs int
	r    io.Reader
}

func (pr *prefixReader) Read(buf []byte) (n int, err error) {
	nread := 0
	if pr.curs < len(pr.data) {
		n := copy(buf, pr.data[pr.curs:])
		buf = buf[n:]
		pr.curs += n
		nread += n
	}

	if len(buf) == 0 {
		return nread, nil
	}

	n, err = pr.r.Read(buf)
	nread += n
	return nread, err
}

func PrefixReader(data []byte, r io.Reader) io.Reader {
	return &prefixReader{data: data, r: r}
}
