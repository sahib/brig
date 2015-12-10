package main

import (
	"io"
	"os"

	lz4 "github.com/bkaradzic/go-lz4"
)

type Lz4Writer struct {
	Writer        io.Writer
	MaxBufferSize uint32
	HeaderWritten bool
}

// GenerateLz4Header creates a valid LZ4 1.5 Header.
// https://docs.google.com/document/d/
// 1cl8N1bmkTdIpPLtnlzbBSFAdUeyNo5fwfHbHU7VRNWY/edit?pli=1
func GenerateLz4Header() []byte {

	header := []byte{
		// LZ4 Magic number, Little Endian (4 Bytes):
		0x04, 0x22, 0x4d, 0x18,
		// Frame Descriptor (3-11 Bytes):
		0x0, 0x1, 0x3,
	}
	return header
}

func (w *Lz4Writer) writeHeader() {
	if _, err := w.Writer.Write(GenerateLz4Header()); err != nil {
		panic(err)
	}
	w.HeaderWritten = true
}

func (w *Lz4Writer) Write(data []byte) (int, error) {
	dst := make([]byte, len(data))
	compressedBytes, err := lz4.Encode(dst, data)
	if err != nil {
		return 0, err
	}

	if !w.HeaderWritten {
		w.writeHeader()
	}

	bytes, err := w.Writer.Write(compressedBytes)
	return bytes, err
}

func main() {
	fdFrom, _ := os.OpenFile(os.Args[1], os.O_RDONLY, 0644)
	defer fdFrom.Close()
	fdTo, _ := os.OpenFile(os.Args[2], os.O_CREATE|os.O_WRONLY, 0644)
	defer fdTo.Close()
	writer := &Lz4Writer{
		Writer:        fdTo,
		MaxBufferSize: lz4.MaxInputSize,
	}

	buffer := make([]byte, 4096)
	for {
		read, err := fdFrom.Read(buffer)
		if err != nil {
			break
		}
		writer.Write(buffer[:read])
	}
}
