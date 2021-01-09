// Package util implements small helper function that
// should be included in the stdlib in our opinion.
package util

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
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
		log.Warningf("Error on close `%v`: %v", c, err)
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

// NopWriteCloser returns a WriteCloser with a no-op Close method wrapping the
// provided Writer w.
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

// TimeoutReadWriter is io.ReadWriter capable of returning ErrTimeout
// if there was no result in a certain timeout period.
type TimeoutReadWriter struct {
	io.Writer
	io.Reader

	rtimeout time.Duration
	wtimeout time.Duration

	useDeadline bool
	rdeadline   time.Time
	wdeadline   time.Time
}

// ErrTimeout will be returned by Read/Write in case of a timeout.
var ErrTimeout = errors.New("I/O Timeout: Operation timed out")

func (rw *TimeoutReadWriter) io(p []byte, doRead bool) (n int, err error) {
	var deadline <-chan time.Time

	// Figoure out when it's too late:
	switch {
	case doRead && rw.useDeadline:
		deadline = time.After(rw.rdeadline.Sub(time.Now()))
	case doRead && !rw.useDeadline:
		deadline = time.After(rw.rtimeout)
	case !doRead && rw.useDeadline:
		deadline = time.After(rw.wdeadline.Sub(time.Now()))
	case !doRead && !rw.useDeadline:
		deadline = time.After(rw.wtimeout)
	}

	// Resever one element, so the go routine gets cleaned up
	// early even if the timeout already expired.
	done := make(chan bool, 1)
	go func() {
		if doRead {
			n, err = rw.Reader.Read(p)
		} else {
			n, err = rw.Writer.Write(p)
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

// SetReadDeadline sets an absolute time in the future where
// a read option should be canceled.
func (rw *TimeoutReadWriter) SetReadDeadline(d time.Time) error {
	rw.useDeadline = true
	rw.rdeadline = d
	return nil
}

// SetWriteDeadline sets an absolute time in the future where
// a write option should be canceled.
func (rw *TimeoutReadWriter) SetWriteDeadline(d time.Time) error {
	rw.useDeadline = true
	rw.wdeadline = d
	return nil
}

// SetDeadline sets an absolute time in the future where an I/O
// operation should be canceled.
func (rw *TimeoutReadWriter) SetDeadline(d time.Time) error {
	rw.SetWriteDeadline(d)
	rw.SetReadDeadline(d)
	return nil
}

// SetWriteTimeout sets an own timeout for writing.
func (rw *TimeoutReadWriter) SetWriteTimeout(d time.Duration) error {
	rw.wtimeout = d
	return nil
}

// SetReadTimeout sets an own timeout for reading.
func (rw *TimeoutReadWriter) SetReadTimeout(d time.Duration) error {
	rw.rtimeout = d
	return nil
}

// SetTimeout sets both the read and write timeout to `d`.
func (rw *TimeoutReadWriter) SetTimeout(d time.Duration) error {
	rw.rtimeout = d
	rw.wtimeout = d
	return nil
}

// NewTimeoutWriter wraps `w` and returns a io.Writer that times out
// after `d` elapsed with ErrTimeout if `w` didn't succeed in that time.
func NewTimeoutWriter(w io.Writer, d time.Duration) io.Writer {
	return &TimeoutReadWriter{Writer: w, wtimeout: d}
}

// NewTimeoutReader wraps `r` and returns a io.Reader that times out
// after `d` elapsed with ErrTimeout if `r` didn't succeed in that time.
func NewTimeoutReader(r io.Reader, d time.Duration) io.Reader {
	return &TimeoutReadWriter{Reader: r, rtimeout: d}
}

// NewTimeoutReadWriter wraps `rw` and returns a io.ReadWriter that times out
// after `d` elapsed with ErrTimeout if `rw` didn't succeed in that time.
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
		base := "More than one error happened:\n"
		for _, err := range es {
			base += "\t" + err.Error() + "\n"
		}

		return base
	}
}

// ToErr combines all errors in the list to a single error.
// If there were no errors, it returns nil.
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
	}

	return fmt.Sprintf("%v", data[:lo])
}

type limitWriter struct {
	wr  io.Writer
	sz  int64
	pos int64
}

// LimitWriter is like io.LimitReader but for an io.Writer
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

	// many go std functions require that all of `buf` was written,
	// or else they return with errShortWrite. Let's act like we
	// used all of it.
	return len(buf), nil
}

type prefixReader struct {
	data []byte
	curs int64
	r    io.Reader
}

func (pr *prefixReader) Read(buf []byte) (n int, err error) {
	nread := 0
	if pr.curs < 0 {
		return -1, fmt.Errorf("negative cursor")
	}

	if pr.curs < int64(len(pr.data)) {
		n := copy(buf, pr.data[pr.curs:])
		buf = buf[n:]
		pr.curs += int64(n)
		nread += n
	}

	if len(buf) == 0 {
		return nread, nil
	}

	n, err = pr.r.Read(buf)

	// err might be io.EOF, so progress.
	nread += n
	return nread, err
}

type seekablePrefixReader struct {
	prefixReader
	s io.Seeker
}

func (pr *seekablePrefixReader) Seek(offset int64, whence int) (int64, error) {
	// NOTE: pr.r shares the same offset space as pr does.
	//       All this Seek() does is preventing it from jumping it into the prefix space.
	//       It does this by calculating an absolute offset and limits the jump offset
	//       by the length of the prefix.
	switch whence {
	case io.SeekStart:
		newOff := Max64(int64(len(pr.data)), offset)
		if _, err := pr.s.Seek(newOff, whence); err != nil {
			return -1, err
		}

		pr.curs = offset
		return pr.curs, nil
	case io.SeekCurrent:
		newOff := Max64(int64(len(pr.data)), pr.curs+offset)
		if _, err := pr.s.Seek(newOff, io.SeekStart); err != nil {
			return -1, err
		}

		pr.curs += offset
		return pr.curs, nil
	case io.SeekEnd:
		size, err := pr.s.Seek(0, io.SeekEnd)
		if err != nil {
			return -1, err
		}

		newOff := Max64(int64(len(pr.data)), size+offset)
		if _, err := pr.s.Seek(newOff, io.SeekStart); err != nil {
			return -1, err
		}

		pr.curs = newOff
		return pr.curs, nil
	default:
		return -1, fmt.Errorf("invalid whence: %d", whence)
	}
}

// PeekHeader returns a new reader that will yield the very same data as `r`.
// It reads `size` bytes from `r` and returns it. The underlying implementation
// uses PrefixReader to prefix the stream with the header again.
func PeekHeader(rs io.ReadSeeker, size int64) ([]byte, io.ReadSeeker, error) {
	headerBuf := make([]byte, size)
	n, err := rs.Read(headerBuf)
	if err != nil && err != io.EOF {
		return nil, nil, err
	}

	headerBuf = headerBuf[:n]
	return headerBuf, &seekablePrefixReader{
		prefixReader: prefixReader{
			data: headerBuf,
			r:    rs,
		},
		s: rs,
	}, nil
}

// HeaderReader is an alternative to PeekHeader().
// It requires no seeking and buffers the header while reading the stream.
// The header will not be available when there was no Read().
type HeaderReader struct {
	r    io.Reader
	size uint64
	buf  []byte
}

// NewHeaderReader returns a new header reader.
func NewHeaderReader(r io.Reader, size uint64) *HeaderReader {
	return &HeaderReader{
		r:    r,
		size: size,
		buf:  []byte{},
	}
}

// Header returns the current header buffer. It's empty if no data was read
// from the stream yet. It's size will be smaller or equal the size you passed
// to NewHeaderReader.
func (hr *HeaderReader) Header() []byte {
	return hr.buf
}

// Peek reads the header of the stream and returns it.
// It does not alter the position of the stream, Read() will still
// return the header read by Peek().
//
// This will only work if no I/O was done on stream yet.
// In this case an error is returned. You can Peek() successfully at
// most once.
func (hr *HeaderReader) Peek() ([]byte, error) {
	if len(hr.buf) > 0 {
		// The stream was already read.
		return nil, errors.New("cannot peek, stream was read already")
	}

	buf := make([]byte, hr.size)
	n, err := io.ReadFull(hr.r, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, err
	}

	// Do not read the header anymore in Read():
	hr.size = 0
	hr.buf = buf[:n]

	// Concatenate the memory buffer to the stream:
	hr.r = &prefixReader{data: hr.buf, r: hr.r}
	return hr.buf, nil
}

func (hr *HeaderReader) Read(buf []byte) (int, error) {
	n, err := hr.r.Read(buf)
	if diff := int64(hr.size) - int64(len(hr.buf)); n > 0 && diff > 0 {
		if int64(n) < diff {
			diff = int64(n)
		}

		if len(hr.buf) == 0 {
			// only allocate header buffer when required:
			hr.buf = make([]byte, 0, hr.size)
		}

		hr.buf = append(hr.buf, buf[:diff]...)
	}

	return n, err
}

// CopyFile simply copies the file at `src` to `dst`.
// If `dst` already contains a file, it will be overwritten.
func CopyFile(src, dst string) error {
	srcFd, err := os.Open(src)
	if err != nil {
		return err
	}

	defer srcFd.Close()

	dstFd, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0700)
	if err != nil {
		return err
	}

	defer dstFd.Close()

	_, err = io.Copy(dstFd, srcFd)
	return err
}

// FindFreePort asks the operating system for a free port.
func FindFreePort() int {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0
	}

	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}

// URLToSchemeAndAddr is a helper that converts different URLS to
// a scheme and addr argument that can be passed to either net.Listen()
// or to net.Dial(). Some extras are supported.
func URLToSchemeAndAddr(s string) (string, string, error) {
	u, err := url.Parse(s)
	if err != nil {
		return "", "", err
	}

	switch u.Scheme {
	case "tcp":
		return u.Scheme, u.Host, nil
	case "unix":
		v := u.Query()
		if v.Get("abstract") == "true" {
			// See "man 7 unix", if the first byte is a 0 byte then the socket
			// is considered as abstract, i.e. no file in the filesystem.
			// The path just acts as unique identifier for the socket.
			// This makes cleaning up the path easier in case of crashes.
			b := []byte(u.Path)
			b[0] = 0
			u.Path = string(b)
		}

		// We can optionally specify an ID. That can be useful to
		// distinguish between different brig daemon on the same system.
		// (needed for the test-bed script for example)
		if id := v.Get("id"); id != "" {
			u.Path += id
		}

		return u.Scheme, u.Path, nil
	default:
		return "", "", fmt.Errorf("unsupported protocol: %v", u.Scheme)
	}
}
