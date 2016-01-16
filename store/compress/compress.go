package compress

import (
	"io"
	"os"

	"github.com/golang/snappy"
)

// Compress the file at src to dst.
func CompressFile(src, dst string) (int64, error) {
	fdFrom, err := os.OpenFile(src, os.O_RDONLY, 0644)
	if err != nil {
		return 0, err
	}
	defer fdFrom.Close()

	fdTo, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return 0, err
	}
	defer fdTo.Close()

	return Compress(fdFrom, fdTo)
}

// Decompress the file at src to dst.
func DecompressFile(src, dst string) (int64, error) {
	fdFrom, err := os.OpenFile(src, os.O_RDONLY, 0644)
	if err != nil {
		return 0, err
	}
	defer fdFrom.Close()

	fdTo, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return 0, err
	}
	defer fdTo.Close()

	return Decompress(fdFrom, fdTo)
}

// Compress represents a layer stream compression.
// As input src and dst io.Reader/io.Writer is expected.
func Compress(src io.Reader, dst io.Writer) (int64, error) {
	return io.Copy(snappy.NewWriter(dst), src)
}

// Decompress represents a layer for stream decompression.
// As input src and dst io.Reader/io.Writer is expected.
func Decompress(src io.Reader, dst io.Writer) (int64, error) {
	return io.Copy(dst, snappy.NewReader(src))
}

type reader struct {
	rRaw          io.Reader
	rZip          io.Reader
	wasCompressed bool
	readHeader    bool
}

func (r *reader) Read(buf []byte) (int, error) {
	if !r.readHeader {
		r.readHeader = true

		marker := make([]byte, 1)
		if _, err := r.rRaw.Read(marker); err != nil {
			return 0, err
		}

		r.wasCompressed = marker[0] > 0
	}

	if r.wasCompressed {
		return r.rZip.Read(buf)
	}

	return r.rRaw.Read(buf)
}

// NewReader returns a new compression Reader.
func NewReader(r io.Reader) io.Reader {
	return &reader{
		rRaw: r,
		rZip: snappy.NewReader(r),
	}
}

type writer struct {
	wRaw          io.Writer
	wZip          io.Writer
	headerWritten bool
}

func (w *writer) Write(buf []byte) (int, error) {
	if !w.headerWritten {
		w.headerWritten = true

		if _, err := w.wRaw.Write([]byte{1}); err != nil {
			return 0, err
		}
	}

	return w.wZip.Write(buf)
}

// NewWriter returns a new compression Writer.
func NewWriter(w io.Writer) io.Writer {
	return &writer{
		wRaw: w,
		wZip: snappy.NewWriter(w),
	}
}
