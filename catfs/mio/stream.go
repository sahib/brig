package mio

import (
	"errors"
	"io"
	"io/ioutil"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/sahib/brig/catfs/mio/compress"
	"github.com/sahib/brig/catfs/mio/encrypt"
)

type Stream interface {
	io.Reader
	io.Seeker
	io.Closer
	io.WriterTo
}

// NewOutStream creates an OutStream piping data from brig to the outside.
// `key` is used to decrypt the data. The compression algorithm is read
// from the stream header.
func NewOutStream(r io.ReadSeeker, key []byte) (Stream, error) {
	rEnc, err := encrypt.NewReader(r, key)
	if err != nil {
		return nil, err
	}

	rZip := compress.NewReader(rEnc)
	return struct {
		io.Reader
		io.Seeker
		io.Closer
		io.WriterTo
	}{
		Reader:   rZip,
		Seeker:   rZip,
		WriterTo: rZip,
		Closer:   ioutil.NopCloser(rZip),
	}, nil
}

// NewInStream creates a new stream that pipes data into brig.
// The data is read from `r`, enrypted with `key` and compressed
// according to `compress`.
func NewInStream(r io.Reader, key []byte, algo compress.AlgorithmType) (io.Reader, error) {
	pr, pw := io.Pipe()

	// Setup the writer part:
	wEnc, encErr := encrypt.NewWriter(pw, key)
	if encErr != nil {
		return nil, encErr
	}

	wZip, zipErr := compress.NewWriter(wEnc, algo)
	if zipErr != nil {
		return nil, zipErr
	}

	// Suck the reader empty and move it to `wZip`.
	// Every write to wZip will be available as read in `pr`.
	go func() {
		// TODO: The error reporting is weird here.
		//       Can we get rid of the go()?
		var err error
		if _, copyErr := io.Copy(wZip, r); copyErr != nil {
			err = copyErr
		}

		if zipCloseErr := wZip.Close(); zipCloseErr != nil {
			err = zipCloseErr
		}

		if encCloseErr := wEnc.Close(); encCloseErr != nil {
			err = encCloseErr
		}

		if pwErr := pw.Close(); pwErr != nil {
			err = pwErr
		}

		if err != nil {
			// TODO: This need to be handled better.
			log.Warningf("Internal write error: %v", err)
		}
	}()

	return pr, nil
}

// limitedStream is a small wrapper around Stream,
// which allows truncating the stream at a certain size.
// It provides the same
type limitedStream struct {
	stream Stream
	pos    uint64
	size   uint64
}

func (ls *limitedStream) Read(buf []byte) (int, error) {
	isEOF := false
	if ls.pos+uint64(len(buf)) >= ls.size {
		buf = buf[:ls.size-ls.pos]
		isEOF = true
	}

	n, err := ls.stream.Read(buf)
	if err != nil {
		return n, err
	}

	if isEOF {
		err = io.EOF
	}

	return n, err
}

func (ls *limitedStream) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case os.SEEK_CUR:
		return ls.Seek(int64(ls.pos)+offset, os.SEEK_SET)
	case os.SEEK_END:
		return ls.Seek(offset-int64(ls.size), os.SEEK_SET)
	}

	newPos := int64(ls.pos) + offset
	if newPos < 0 {
		return -1, io.EOF
	}

	if newPos > int64(ls.size) {
		return -1, io.EOF
	}

	ls.pos = uint64(newPos)
	return ls.stream.Seek(newPos, os.SEEK_SET)
}

type limitWriter struct {
	w     io.Writer
	curr  uint64
	limit uint64
}

var (
	errOverLimit = errors.New("welp TODO")
)

func (lw *limitWriter) Write(buf []byte) (int, error) {
	isOverLimit := false
	left := lw.limit - lw.curr
	if left < uint64(len(buf)) {
		buf = buf[:left]
		isOverLimit = true
	}

	if len(buf) == 0 {
		if isOverLimit {
			return 0, errOverLimit
		}

		return 0, nil
	}

	n, err := lw.w.Write(buf)
	if err != nil {
		return n, err
	}

	if isOverLimit {
		return n, errOverLimit
	}

	return n, nil
}

func (ls *limitedStream) WriteTo(w io.Writer) (int64, error) {
	// We do not want to defeat the purpose of WriteTo here.
	// That's why we do the limit check in the writer part.
	lw := &limitWriter{
		limit: ls.size - ls.pos,
		curr:  ls.pos,
		w:     w,
	}

	n, err := ls.stream.WriteTo(lw)
	if err == errOverLimit {
		// silence errOverLimit, since it's more or less internal.
		return n, nil
	}

	return n, err
}

func (ls *limitedStream) Close() error {
	return ls.stream.Close()
}

// LimitStream is like io.LimitReader, but works for mio.Stream.
// It will not allow reading/seeking after the specified size.
func LimitStream(stream Stream, size uint64) Stream {
	return &limitedStream{
		stream: stream,
		pos:    0,
		size:   size,
	}
}
