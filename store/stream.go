package store

import (
	"github.com/disorganizer/brig/store/format"
	"github.com/disorganizer/brig/util/compress"
	"io"
	"os"
)

type Stream interface {
	io.Reader
	io.Writer
	io.Seeker
}

func NewFromPath(key []byte, path string) (Stream, error) {
	fd, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}

	return NewStream(key, fd, fd)
}

func NewStream(key []byte, r io.Reader, w io.Writer) (Stream, error) {
	wFin, err := format.NewEncryptedWriter(compress.NewWriter(w), key)
	if err != nil {
		return nil, err
	}

	rFin, err := format.NewEncryptedReader(compress.NewReader(r), key)
	if err != nil {
		return nil, err
	}

	return &ipfsStream{
		w: wFin,
		r: rFin,
	}, nil
}

type ipfsStream struct {
	r *format.EncryptedReader
	w *format.EncryptedWriter
}

func (i *ipfsStream) Read(buf []byte) (int, error) {
	return i.r.Read(buf)
}

func (i *ipfsStream) Write(buf []byte) (int, error) {
	return i.w.Write(buf)
}

func (i *ipfsStream) Seek(offset int64, whence int) (int64, error) {
	if _, err := i.r.Seek(offset, whence); err != nil {
		return 0, err
	}

	// TODO: Implement write-seek?
	return i.w.Seek(offset, whence)
}
