package mio

import (
	"io"
	"io/ioutil"

	"github.com/sahib/brig/catfs/mio/compress"
	"github.com/sahib/brig/catfs/mio/encrypt"
	"github.com/sahib/brig/util"
	log "github.com/sirupsen/logrus"
)

// Stream is a stream coming from the backend.
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

// NewInStream creates a new stream that pipes data into ipfs.
// The data is read from `r`, encrypted with `key` and compressed with `algo`.
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
		if _, err := io.Copy(wZip, r); err != nil {
			// Continue closing the fds; no return.
			log.Warningf("internal write error: %v", err)
		}

		if err := wZip.Close(); err != nil {
			// Continue closing the others:
			log.Warningf("internal close zip error: %v", err)
		}

		if err := wEnc.Close(); err != nil {
			// Continue closing the others:
			log.Warningf("internal close enclayer error: %v", err)
		}

		if err := pw.Close(); err != nil {
			log.Warningf("internal close pipe error: %v", err)
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
	case io.SeekCurrent:
		return ls.Seek(int64(ls.pos)+offset, io.SeekStart)
	case io.SeekEnd:
		ls.pos = 0
		return ls.Seek(int64(ls.size)+offset, io.SeekStart)
	case io.SeekStart:
		ls.pos = 0
	}

	newPos := int64(ls.pos) + offset
	if newPos < 0 {
		return -1, io.EOF
	}

	if newPos > int64(ls.size) {
		return int64(ls.size), io.EOF
	}

	ls.pos = uint64(newPos)
	return ls.stream.Seek(newPos, io.SeekStart)
}

func (ls *limitedStream) WriteTo(w io.Writer) (int64, error) {
	// We do not want to defeat the purpose of WriteTo here.
	// That's why we do the limit check in the writer part.
	return ls.stream.WriteTo(util.LimitWriter(w, int64(ls.size-ls.pos)))
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
