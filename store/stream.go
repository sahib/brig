package store

import (
	"io"
	"io/ioutil"

	"github.com/disorganizer/brig/store/compress"
	"github.com/disorganizer/brig/store/encrypt"
)

// TODO: needed?
// Reader accumulates all interface brig requests from a stream.
type Reader interface {
	io.Reader
	io.Seeker
	io.Closer
	io.WriterTo
}

type reader struct {
	io.Reader
	io.Seeker
	io.Closer
	io.WriterTo
}

// NewIpfsReader wraps the raw-data reader `r` and returns a Reader
// that yields the clear data, if `key` is correct.
func NewIpfsReader(key []byte, r Reader) (Reader, error) {
	rEnc, err := encrypt.NewReader(r, key)
	if err != nil {
		return nil, err
	}

	rZip := compress.NewReader(rEnc)

	return reader{
		Reader:   rZip,
		Seeker:   rZip,
		WriterTo: rZip,
		Closer:   ioutil.NopCloser(rZip),
	}, nil
}

// NewFileReader reads an unencrypted, uncompressed file and
// returns a reader that will yield the data we feed to ipfs.
func NewFileReader(key []byte, r io.Reader) (outR io.Reader, outErr error) {
	pr, pw := io.Pipe()

	// Setup the writer part:
	wEnc, err := encrypt.NewWriter(pw, key)
	if err != nil {
		return nil, err
	}
	// TODO: Paremetrize algo et cetera.
	wZip, err := compress.NewWriter(wEnc, compress.AlgoSnappy)
	if err != nil {
		return nil, err
	}

	// Suck the reader empty and move it to `wZip`.
	// Every write to wZip will be available as read in `pr`.
	go func() {
		defer func() {
			if err := wZip.Close(); err != nil {
				outErr = err
			}

			if err := wEnc.Close(); err != nil {
				outErr = err
			}

			if err := pw.Close(); err != nil {
				outErr = err

			}
		}()

		if _, err := io.Copy(wZip, r); err != nil {
			outErr = err
		}
	}()

	return pr, nil
}
