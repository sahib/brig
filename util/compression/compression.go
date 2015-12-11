package main

import (
	"github.com/golang/snappy"
	"io"
	"os"
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

// NewReader returns a new compression Reader.
func NewReader(r io.Reader) io.Reader {
	return snappy.NewReader(r)
}

// NewWriter returns a new compression Writer.
func NewWriter(w io.Writer) io.Writer {
	return snappy.NewWriter(w)
}

func main() {

	if os.Args[1] == "d" {
		DecompressFile(os.Args[2], os.Args[3])
	}

	if os.Args[1] == "c" {
		CompressFile(os.Args[2], os.Args[3])
	}
}
