package store

import (
	"fmt"
	"io"
	"os"

	// "github.com/disorganizer/brig/store/compress"
	"github.com/disorganizer/brig/store/encrypt"
)

type Reader interface {
	io.Reader
	io.Seeker
	io.Closer
}

func NewIpfsReader(key []byte, r io.ReadSeeker) (Reader, error) {
	rEnc, err := encrypt.NewReader(r, key)
	if err != nil {
		return nil, err
	}

	return rEnc, nil
	// return compress.NewReader(rEnc), nil
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
func NewFileReader(key []byte, r io.Reader) (io.Reader, error) {
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
			wEnc.Close()
			pw.Close()
		}()

		if _, err := io.Copy(wEnc, r); err != nil {
			// TODO: Warn or pass to outside?
			fmt.Println("FUCK", err)
		}
	}()

	return pr, nil
}
