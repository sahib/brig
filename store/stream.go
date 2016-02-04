package store

import (
	"io"
	"os"

	// "github.com/disorganizer/brig/store/compress"
	"github.com/disorganizer/brig/store/encrypt"
)

// Reader accumulates all interface brig requests from a stream.
type Reader interface {
	io.Reader
	io.Seeker
	io.Closer
}

// NewIpfsReader wraps the raw-data reader `r` and returns a Reader
// that yields the clear data, if `key` is correct.
func NewIpfsReader(key []byte, r io.ReadSeeker) (Reader, error) {
	rEnc, err := encrypt.NewReader(r, key)
	if err != nil {
		return nil, err
	}

	// TODO: Bring back compression.
	return rEnc, nil
}

// NewFileReaderFromPath is a shortcut for reading a file from disk
// and returning ipfs-conforming data.
func NewFileReaderFromPath(key []byte, path string) (io.Reader, error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return NewFileReader(key, fd)
}

// NewFileReader reads an unencrypted, uncompressed file and
// returns a reader that will yield the data we feed to ipfs.
func NewFileReader(key []byte, r io.Reader) (outR io.Reader, outErr error) {
	pr, pw := io.Pipe()

	// Setup the writer part:
	// TODO: When using compression, pass as `true` to NewWriter.
	wEnc, err := encrypt.NewWriter(pw, key, false)
	if err != nil {
		return nil, err
	}

	// TODO: Implement seeking compression.
	// wZip := compress.NewWriter(wEnc)

	// Suck the reader empty and move it to `wZip`.
	// Every write to wZip will be available as read in `pr`.
	go func() {
		defer func() {
			if err := wEnc.Close(); err != nil {
				outErr = err
			}

			if err := pw.Close(); err != nil {
				outErr = err

			}
		}()

		if _, err := io.Copy(wEnc, r); err != nil {
			outErr = err
		}
	}()

	return pr, nil
}
