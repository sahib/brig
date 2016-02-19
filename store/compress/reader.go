package compress

import (
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/golang/snappy"
)

// TODO: Tests schreiben (leere dateien, chunkgröße -1, +0, +1 etc.)
// TODO: linter durchlaufen lassen.

type chunkBuffer struct {
	buf      [MaxChunkSize]byte
	readOff  int64
	writeOff int64
	size     int64
}

func (c *chunkBuffer) Write(p []byte) (int, error) {
	n := copy(c.buf[c.writeOff:MaxChunkSize], p)
	c.writeOff += int64(n)
	if c.writeOff > c.size {
		c.size = c.writeOff
	}
	return n, nil
}

func (c *chunkBuffer) Reset() {
	c.readOff = 0
	c.writeOff = 0
	c.size = 0
}

func (c *chunkBuffer) Len() int {
	return int(c.size - c.readOff)
}

func (c *chunkBuffer) Read(p []byte) (int, error) {
	n := copy(p, c.buf[c.readOff:c.size])
	c.readOff += int64(n)
	if n == 0 {
		return n, io.EOF
	}
	return n, nil
}

func (c *chunkBuffer) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case os.SEEK_CUR:
		c.readOff += offset
	case os.SEEK_END:
		c.readOff = c.size + offset
	case os.SEEK_SET:
		c.readOff = offset
	}

	if c.readOff > c.size {
		c.readOff = c.size
	}
	c.writeOff = c.readOff
	return c.readOff, nil
}

func newChunkBuffer() chunkBuffer {
	return chunkBuffer{}
}

type reader struct {
	// Underlying raw, compressed datastream.
	rawR io.ReadSeeker

	// Decompression layer, reader is based on chosen algorithm.
	zipR io.Reader

	// Index with records which contain chunk offsets.
	index []record

	// Buffer holds currently read data; MaxChunkSize.
	chunkBuf chunkBuffer

	// Structure with parsed trailer.
	trailer *trailer
}

func (r *reader) Seek(rawOff int64, whence int) (int64, error) {
	if err := r.parseHeaderIfNeeded(); err != nil {
		return 0, err
	}

	if whence == os.SEEK_END {
		if rawOff > 0 {
			return 0, io.EOF
		}
		return r.Seek(r.index[len(r.index)-1].rawOff+rawOff, os.SEEK_SET)
	}

	if whence == os.SEEK_CUR {
		currPos, err := r.rawR.Seek(0, os.SEEK_CUR)
		if err != nil {
			return currPos, err
		}
		return r.Seek(currPos+rawOff, os.SEEK_SET)
	}

	// Check if given raw offset equals current offset.
	currRawOff, err := r.rawR.Seek(0, os.SEEK_CUR)
	if err != nil || currRawOff == rawOff {
		return currRawOff, err
	}

	currRecord, _ := r.chunkLookup(currRawOff)
	prevRecord, _ := r.chunkLookup(rawOff)
	if _, err := r.rawR.Seek(prevRecord.zipOff, os.SEEK_SET); err != nil {
		return 0, err
	}

	// Don't re-read if offset is in current chunk.
	if currRecord.rawOff == prevRecord.rawOff {
		if _, err := r.readChunk(); err != nil {
			return 0, err
		}
	}

	toRead := rawOff - prevRecord.rawOff
	if _, err := r.chunkBuf.Seek(toRead, os.SEEK_SET); err != nil {
		return 0, err
	}

	return rawOff, nil
}

// Return start (prev offset) and end (curr offset) of the chunk currOff is
// located in. If currOff is 0, the startoffset of the first and second record is
// returned. If currOff is at the end of file the end offset of the last chunk
// is returned twice.  The difference between prev record and curr chunk is then
// equal to 0.
func (r *reader) chunkLookup(currOff int64) (*record, *record) {
	i := sort.Search(len(r.index), func(i int) bool {
		return r.index[i].zipOff > currOff
	})

	// Beginning of the file, first chunk: prev offset is 0, curr offset is 1
	if i == 0 {
		fmt.Println("INDEX:", r.index, i, currOff)
		return &r.index[i], &r.index[i+1]
	}

	// End of the file, last chunk: prev and curr offset is the last index.
	if i == len(r.index) {
		return &r.index[i-1], &r.index[i-1]
	}
	return &r.index[i-1], &r.index[i]
}

func (r *reader) parseHeaderIfNeeded() error {
	if r.trailer != nil {
		return nil
	}

	// Goto end of file and read trailer buffer.
	if _, err := r.rawR.Seek(-TrailerSize, os.SEEK_END); err != nil {
		return err
	}

	buf := [TrailerSize]byte{}
	if n, err := r.rawR.Read(buf[:]); err != nil || n != TrailerSize {
		return err
	}
	r.trailer = &trailer{}
	r.trailer.unmarshal(buf[:])

	// Seek and read index into buffer.
	seekIdx := -(int64(r.trailer.indexSize) + TrailerSize)
	if _, err := r.rawR.Seek(seekIdx, os.SEEK_END); err != nil {
		return err
	}
	indexBuf := make([]byte, r.trailer.indexSize)
	if _, err := r.rawR.Read(indexBuf); err != nil {
		return err
	}

	// Build index with records. A record encapsulates a raw offset and the
	// compressed offset it is mapped to.
	prevRecord := record{-1, -1}
	for i := uint64(0); i < (r.trailer.indexSize / IndexChunkSize); i++ {
		currRecord := record{}
		currRecord.unmarshal(indexBuf)

		if prevRecord.rawOff >= currRecord.rawOff {
			return ErrBadIndex
		}

		if prevRecord.zipOff >= currRecord.zipOff {
			return ErrBadIndex
		}
		fmt.Println(i, currRecord)
		r.index = append(r.index, currRecord)
		indexBuf = indexBuf[IndexChunkSize:]
	}

	// Set reader to beginning of file
	if _, err := r.rawR.Seek(0, os.SEEK_SET); err != nil {
		return err
	}
	return nil
}

// Read reads len(p) bytes from the compressed stream into p.
func (r *reader) Read(p []byte) (int, error) {
	if err := r.parseHeaderIfNeeded(); err != nil {
		return 0, err
	}

	read := 0
	for {
		if r.chunkBuf.Len() != 0 {
			n, err := r.chunkBuf.Read(p)
			if err != nil {
				return n, err
			}
			read += n
			p = p[n:]
		}

		if len(p) == 0 {
			break
		}

		if _, err := r.readChunk(); err != nil {
			return read, err
		}
	}

	return read, nil
}

func (r *reader) readChunk() (int64, error) {
	// Get current position of the reader; offset of the compressed file.
	r.chunkBuf.Reset()
	currOff, err := r.rawR.Seek(0, os.SEEK_CUR)
	if err != nil {
		return 0, err
	}

	// Get the start and end record of the chunk currOff is located in.
	prevRecord, currRecord := r.chunkLookup(currOff)
	if currRecord == nil || prevRecord == nil {
		return 0, ErrBadIndex
	}

	// Determinate uncompressed chunksize; should only be 0 on empty file or at the end of file.
	chunkSize := currRecord.rawOff - prevRecord.rawOff
	if chunkSize == 0 {
		return 0, io.EOF
	}

	// Set reader to compressed offset.
	if _, err = r.rawR.Seek(prevRecord.zipOff, os.SEEK_SET); err != nil {
		return 0, err
	}

	n, err := io.CopyN(&r.chunkBuf, r.zipR, chunkSize)
	return n, err
}

// Return a new ReadSeeker with compression support. As random access is the
// purpose of this layer, a ReadSeeker is required as parameter. The used
// compression algorithm is chosen based on trailer information.
func NewReader(r io.ReadSeeker) io.ReadSeeker {
	return &reader{
		rawR:     r,
		zipR:     snappy.NewReader(r),
		chunkBuf: newChunkBuffer(),
	}
}
