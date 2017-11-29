package chunkbuf

import (
	"io"
	"os"

	"github.com/disorganizer/brig/util"
)

// TODO: This has no tests yet (only indirect by compress)

// ChunkBuffer represents a custom buffer struct with Read/Write and Seek support.
type ChunkBuffer struct {
	buf      []byte
	readOff  int64
	writeOff int64
	size     int64
}

const (
	maxChunkSize = 64 * 1024
)

func (c *ChunkBuffer) Write(p []byte) (int, error) {
	n := copy(c.buf[c.writeOff:maxChunkSize], p)
	c.writeOff += int64(n)
	c.size = util.Max64(c.size, c.writeOff)
	return n, nil
}

func (c *ChunkBuffer) Reset() {
	c.readOff = 0
	c.writeOff = 0
	c.size = 0
}

func (c *ChunkBuffer) Len() int {
	return int(c.size - c.readOff)
}

func (c *ChunkBuffer) Read(p []byte) (int, error) {
	n := copy(p, c.buf[c.readOff:c.size])
	c.readOff += int64(n)
	if n == 0 {
		return n, io.EOF
	}
	return n, nil
}

func (c *ChunkBuffer) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case os.SEEK_CUR:
		c.readOff += offset
	case os.SEEK_END:
		c.readOff = c.size + offset
	case os.SEEK_SET:
		c.readOff = offset
	}
	c.readOff = util.Min64(c.readOff, c.size)
	c.writeOff = c.readOff
	return c.readOff, nil
}

// Close is a no-op only existing to fulfill io.Closer
func (c *ChunkBuffer) Close() error {
	return nil
}

func (c *ChunkBuffer) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(c.buf[c.readOff:])
	if err != nil {
		return 0, err
	}

	c.readOff += int64(n)
	return int64(n), nil
}

// NewChunkBuffer returns a ChunkBuffer with the given data. if data is nil a
// ChunkBuffer with 64k is returned.
func NewChunkBuffer(data []byte) *ChunkBuffer {
	if data == nil {
		data = make([]byte, maxChunkSize)
	}

	return &ChunkBuffer{buf: data, size: int64(len(data))}
}
