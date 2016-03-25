package compress

import (
	"io"
	"os"
	"sort"

	"github.com/disorganizer/brig/util"
	"github.com/golang/snappy"
)

// TODO: Tests schreiben (leere dateien, chunkgröße -1, +0, +1 etc.)
// TODO: os.Seek(0, os.CURR) möglichst beseitigen; mit normalen index ersetzen.
// TODO: Dokumentation schreiben.
// TODO: ReadFrom und WriteTo implementieren.
// TODO: Mehr Algorithmen anbieten (lz4, brotli?)
// TODO: In store/stream.go einbauen.
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
	c.size = util.Max64(c.size, c.writeOff)
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
	c.readOff = util.Min64(c.readOff, c.size)
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

	// Current seek offset in the compressed stream.
	rawSeekOffset int64

	// Current seek offset in the uncompressed stream.
	zipSeekOffset int64

	// Marker to identify initial read.
	isInitialRead bool
}

func (r *reader) Seek(destOff int64, whence int) (int64, error) {

	if whence == os.SEEK_END {
		if destOff > 0 {
			return 0, io.EOF
		}
		return r.Seek(r.index[len(r.index)-1].rawOff+destOff, os.SEEK_SET)
	}

	if whence == os.SEEK_CUR {
		return r.Seek(r.zipSeekOffset+destOff, os.SEEK_SET)
	}

	if destOff < 0 {
		return 0, io.EOF
	}

	if err := r.parseHeaderIfNeeded(); err != nil {
		return 0, err
	}

	// Check if given raw offset equals current offset.
	if r.zipSeekOffset == destOff {
		return destOff, nil
	}

	destRecord, _ := r.chunkLookup(destOff, true)
	currRecord, _ := r.chunkLookup(r.zipSeekOffset, true)

	r.rawSeekOffset = destRecord.zipOff
	r.zipSeekOffset = destOff

	//Don't re-read if offset is in current chunk.
	if currRecord.rawOff != destRecord.rawOff || !r.isInitialRead {
		if _, err := r.readZipChunk(); err != nil {
			return 0, err
		}
	}

	toRead := destOff - destRecord.rawOff
	if _, err := r.chunkBuf.Seek(toRead, os.SEEK_SET); err != nil {
		return 0, err
	}

	return destOff, nil
}

// Return start (prev offset) and end (curr offset) of the chunk currOff is
// located in. If currOff is 0, the startoffset of the first and second record is
// returned. If currOff is at the end of file the end offset of the last chunk
// is returned twice.  The difference between prev record and curr chunk is then
// equal to 0.
func (r *reader) chunkLookup(currOff int64, isRawOff bool) (*record, *record) {
	// Get smallest index that is before given currOff.
	i := sort.Search(len(r.index), func(i int) bool {
		if isRawOff {
			return r.index[i].rawOff > currOff
		} else {
			return r.index[i].zipOff > currOff
		}
	})

	// Beginning of the file, first chunk: prev offset is 0, curr offset is 1.
	if i == 0 {
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

	// Handle uncompressed stream.
	if r.trailer.algo == AlgoNone {
		if _, err := r.rawR.Seek(0, os.SEEK_SET); err != nil {
			return err
		}
		// No need to go further.
		return nil
	}

	// Handle compressed stream.
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
		r.index = append(r.index, currRecord)
		indexBuf = indexBuf[IndexChunkSize:]
	}

	// Set reader to beginning of file
	if _, err := r.rawR.Seek(0, os.SEEK_SET); err != nil {
		return err
	}
	r.rawSeekOffset = 0
	r.zipSeekOffset = 0
	return nil
}

// Read reads len(p) bytes from the compressed stream into p.
func (r *reader) Read(p []byte) (int, error) {
	if err := r.parseHeaderIfNeeded(); err != nil {
		return 0, err
	}

	// Handle uncompressed stream.
	if r.trailer.algo == AlgoNone {
		maxOff, errMax := r.maxOff(int64(len(p)))
		n, err := r.rawR.Read(p[:maxOff])
		r.zipSeekOffset += int64(n)
		if err != nil {
			return n, err
		}
		return n, errMax
	}

	// Handle stream using compression.
	read := 0
	for {
		if r.chunkBuf.Len() != 0 {
			n, err := r.chunkBuf.Read(p)
			if err != nil {
				return n, err
			}
			r.zipSeekOffset += int64(n)
			read += n
			p = p[n:]
		}

		if len(p) == 0 {
			break
		}

		if _, err := r.readZipChunk(); err != nil {
			return read, err
		}
	}

	return read, nil
}

//TODO: Save maxOffset in trailer and read from trailer instead of SEEK.
func (r *reader) maxOff(pSize int64) (int64, error) {
	// get current position.
	currOff, err := r.rawR.Seek(0, os.SEEK_CUR)
	if err != nil {
		return currOff, err
	}

	// get max offset (possition without trailer).
	maxOff, err := r.rawR.Seek(-TrailerSize, os.SEEK_END)
	if err != nil {
		return maxOff, err
	}

	// go back to current offset.
	_, err = r.rawR.Seek(currOff, os.SEEK_SET)
	if err != nil {
		return 0, err
	}

	// determinate max offset using.
	if pSize+currOff > maxOff {
		return maxOff - currOff, io.EOF
	}
	return pSize, nil
}

func (r *reader) readZipChunk() (int64, error) {
	// Get current position of the reader; offset of the compressed file.
	r.chunkBuf.Reset()

	// Get the start and end record of the chunk currOff is located in.
	prevRecord, currRecord := r.chunkLookup(r.rawSeekOffset, false)
	if currRecord == nil || prevRecord == nil {
		return 0, ErrBadIndex
	}

	// Determinate uncompressed chunksize; should only be 0 on empty file or at the end of file.
	chunkSize := currRecord.rawOff - prevRecord.rawOff
	if chunkSize == 0 {
		return 0, io.EOF
	}

	// Set reader to compressed offset.
	if _, err := r.rawR.Seek(prevRecord.zipOff, os.SEEK_SET); err != nil {
		return 0, err
	}

	n, err := io.CopyN(&r.chunkBuf, r.zipR, chunkSize)
	r.rawSeekOffset = currRecord.zipOff
	r.zipSeekOffset = prevRecord.rawOff
	r.isInitialRead = false
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
