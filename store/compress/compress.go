package compress

import (
	"io"
	"os"

	"github.com/golang/snappy"
)

func openFiles(from, to string) (*os.File, *os.File, error) {
	fdFrom, err := os.Open(from)
	if err != nil {
		return nil, nil, err
	}

	fdTo, err := os.OpenFile(to, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		fdFrom.Close()
		return nil, nil, err
	}

	return fdFrom, fdTo, nil
}

// CopyCompressed reads the file at `src` and writes it compressed to `dst`.
func CopyCompressed(src, dst string) (n int64, outErr error) {
	fdFrom, fdTo, err := openFiles(src, dst)
	if err != nil {
		return 0, err
	}

	defer func() {
		// Only fdTo needs to be closed, Decrypt closes fdFrom.
		if err := fdFrom.Close(); err != nil {
			outErr = err
		}
		if err := fdTo.Close(); err != nil {
			outErr = err
		}
	}()

	return Compress(fdFrom, fdTo)
}

// CopyDecompressed reads the compressed file at `src` and writes the clear file
// at `dst`.
func CopyDecompressed(src, dst string) (n int64, outErr error) {
	fdFrom, fdTo, err := openFiles(src, dst)
	if err != nil {
		return 0, err
	}

	defer func() {
		// Only fdTo needs to be closed, Decrypt closes fdFrom.
		if err := fdFrom.Close(); err != nil {
			outErr = err
		}
		if err := fdTo.Close(); err != nil {
			outErr = err
		}
	}()

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
