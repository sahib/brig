package compress

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/disorganizer/brig/catfs/mio/chunkbuf"
)

type reader struct {
	// Underlying raw, compressed datastream.
	rawR io.ReadSeeker

	// Index with records which contain chunk offsets.
	index []record

	// Buffer holds currently read data; maxChunkSize.
	chunkBuf *chunkbuf.ChunkBuffer

	// Structure with parsed trailer.
	trailer *trailer

	// Current seek offset in the compressed stream.
	rawSeekOffset int64

	// Current seek offset in the uncompressed stream.
	zipSeekOffset int64

	// Marker to identify initial read.
	isInitialRead bool

	// Holds algorithm interface.
	algo Algorithm

	decodeBuf *bytes.Buffer
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

	if err := r.parseTrailerIfNeeded(); err != nil {
		return 0, err
	}

	if destOff < 0 {
		return 0, io.EOF
	}

	// Skip the header for seeking:
	// destOff += headerSize

	// Check if given raw offset equals current offset.
	if r.zipSeekOffset == destOff {
		return destOff, nil
	}

	destRecord, _ := r.chunkLookup(destOff, true)
	currRecord, _ := r.chunkLookup(r.zipSeekOffset, true)

	r.rawSeekOffset = destRecord.zipOff
	r.zipSeekOffset = destOff

	// Don't re-read if offset is in current chunk.
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

// Return start (prevRecord) and end (currRecord) of a chunk currOff is located
// in. If currOff is 0, the first and second record is returned. If currOff is
// at the end of file the end record (currRecord) is returned twice.  The offset
// difference (chunksize) between prevRecord and currRecord is then equal to 0.
func (r *reader) chunkLookup(currOff int64, isRawOff bool) (*record, *record) {
	// Get smallest index that is before given currOff.
	i := sort.Search(len(r.index), func(i int) bool {
		if isRawOff {
			return r.index[i].rawOff > currOff
		}
		return r.index[i].zipOff > currOff
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

func (r *reader) parseTrailerIfNeeded() error {
	if r.trailer != nil {
		return nil
	}

	// Attempt to read the front header:
	headerBuf := [headerSize]byte{}
	if _, err := r.rawR.Read(headerBuf[:]); err != nil {
		return err
	}

	header, err := readHeader(headerBuf[:])
	if err != nil {
		return err
	}

	// Goto end of file and read trailer buffer.
	if _, err := r.rawR.Seek(-trailerSize, os.SEEK_END); err != nil {
		return err
	}

	buf := [trailerSize]byte{}
	n, err := r.rawR.Read(buf[:])
	if err != nil {
		return err
	}

	if n != trailerSize {
		return fmt.Errorf("read trailer was too small: %d bytes", n)
	}

	r.trailer = &trailer{}
	r.trailer.unmarshal(buf[:])

	algo, err := AlgorithmFromType(header.algo)
	if err != nil {
		return err
	}
	r.algo = algo

	// Seek and read index into buffer.
	seekIdx := -(int64(r.trailer.indexSize) + trailerSize)
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
	for i := uint64(0); i < (r.trailer.indexSize / indexChunkSize); i++ {
		currRecord := record{}
		currRecord.unmarshal(indexBuf)

		if prevRecord.rawOff >= currRecord.rawOff {
			return ErrBadIndex
		}

		if prevRecord.zipOff >= currRecord.zipOff {
			return ErrBadIndex
		}
		r.index = append(r.index, currRecord)
		indexBuf = indexBuf[indexChunkSize:]
	}

	// Set reader to beginning of file
	if _, err := r.rawR.Seek(headerSize, os.SEEK_SET); err != nil {
		return err
	}

	r.rawSeekOffset = headerSize
	r.zipSeekOffset = 0
	return nil
}

func (r *reader) WriteTo(w io.Writer) (int64, error) {
	if err := r.parseTrailerIfNeeded(); err != nil {
		return 0, err
	}

	written := int64(0)

	n, cerr := io.Copy(w, r.chunkBuf)
	if cerr != nil {
		return n, cerr
	}
	written += n
	for {
		decData, rerr := r.readZipChunk()
		if rerr == io.EOF {
			return written, nil
		}

		if rerr != nil {
			return written, rerr
		}

		n, werr := w.Write(decData)
		written += int64(n)

		if werr != nil {
			return written, werr
		}
	}
}

// Read reads len(p) bytes from the compressed stream into p.
func (r *reader) Read(p []byte) (int, error) {
	if err := r.parseTrailerIfNeeded(); err != nil {
		return 0, err
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

func (r *reader) fixZipChunk() (int64, error) {
	// Get the start and end record of the chunk currOff is located in.
	prevRecord, currRecord := r.chunkLookup(r.rawSeekOffset, false)
	if currRecord == nil || prevRecord == nil {
		return 0, ErrBadIndex
	}

	// Determinate uncompressed chunksize; should only be 0 on empty file or at the end of file.
	chunkSize := currRecord.zipOff - prevRecord.zipOff
	if chunkSize == 0 {
		return 0, io.EOF
	}

	// Set reader to compressed offset.
	if _, err := r.rawR.Seek(prevRecord.zipOff, os.SEEK_SET); err != nil {
		return 0, err
	}

	r.rawSeekOffset = currRecord.zipOff
	r.zipSeekOffset = prevRecord.rawOff
	r.isInitialRead = false
	return chunkSize, nil
}

func (r *reader) readZipChunk() ([]byte, error) {
	// Get current position of the reader; offset of the compressed file.
	r.chunkBuf.Reset()
	chunkSize, err := r.fixZipChunk()
	if err != nil {
		return nil, err
	}

	r.decodeBuf.Reset()
	_, err = io.CopyN(r.decodeBuf, r.rawR, chunkSize)
	if err != nil {
		return nil, err
	}

	decData, err := r.algo.Decode(r.decodeBuf.Bytes())
	if err != nil {
		return nil, err
	}

	r.chunkBuf = chunkbuf.NewChunkBuffer(decData)
	return decData, nil
}

// NewReader returns a new ReadSeeker with compression support. As random access
// is the purpose of this layer, a ReadSeeker is required as parameter. The used
// compression algorithm is chosen based on trailer information.
func NewReader(r io.ReadSeeker) *reader {
	return &reader{
		rawR:      r,
		decodeBuf: &bytes.Buffer{},
		chunkBuf:  chunkbuf.NewChunkBuffer([]byte{}),
	}
}
