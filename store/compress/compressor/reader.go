package main

import (
	"bytes"
	"io"
	"os"
	"sort"

	"github.com/golang/snappy"
)

// TODO: Tests schreiben (leere dateien, chunkgröße -1, +0, +1 etc.)
// TODO: linter durchlaufen lassen.
// TODO: Seek.

type reader struct {
	// Underlying raw, compressed datastream.
	rawR io.ReadSeeker

	// Decompression layer, reader is based on chosen algorithm.
	zipR io.Reader

	// Index with records which contain chunk offsets.
	index []Record

	// Buffer holds currently read data; MaxChunkSize.
	readBuf *bytes.Buffer

	// Structure with parsed trailer.
	trailer *Trailer
}

func (r *reader) Seek(offset int64, whence int) (int64, error) {
	return offset, nil
}

// Return start (prev offset) and end (curr offset) of the chunk currOff is
// located in. If currOff is 0, the startoffset of the first and second record is
// returned. If currOff is at the end of file the end offset of the last chunk
// is returned twice.  The difference between prev record and curr chunk is then
// equal to 0.
func (r *reader) chunkLookup(currOff int64) (*Record, *Record) {
	i := sort.Search(len(r.index), func(i int) bool {
		return r.index[i].zipOff > currOff
	})

	// Beginning of the file, first chunk: prev offset is 0, curr offset is 1
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
	r.trailer = &Trailer{}
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

	// Build index with Records. A record encapsulates a raw offset and the
	// compressed offset it is mapped to.
	prevRecord := Record{-1, -1}
	for i := uint64(0); i < (r.trailer.indexSize / IndexChunkSize); i++ {
		currRecord := Record{}
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
	return nil
}

// Read reads len(p) bytes from the compressed stream into p.
func (r *reader) Read(p []byte) (int, error) {
	if err := r.parseHeaderIfNeeded(); err != nil {
		return 0, err
	}

	read := 0
	for {
		if r.readBuf.Len() != 0 {
			n := copy(p, r.readBuf.Next(len(p)))
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

	return io.CopyN(r.readBuf, r.zipR, chunkSize)
}

// Return a new ReadSeeker with compression support. As random access is the
// purpose of this layer, a ReadSeeker is required as parameter. The used
// compression algorithm is chosen based on trailer information.
func NewReader(r io.ReadSeeker) io.ReadSeeker {
	return &reader{
		rawR:    r,
		zipR:    snappy.NewReader(r),
		readBuf: &bytes.Buffer{},
	}
}
