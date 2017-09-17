package mio

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/disorganizer/brig/catfs/mio/compress"
	"github.com/disorganizer/brig/catfs/mio/encrypt"
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
func NewOutStream(r Stream, key []byte) (Stream, error) {
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
			fmt.Println("TODO: Internal write err", err)
		}
	}()

	return pr, nil
}
