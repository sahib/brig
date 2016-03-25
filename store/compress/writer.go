package compress

import (
	"bytes"
	"io"

	"github.com/disorganizer/brig/util"
	"github.com/golang/snappy"
)

type writer struct {
	// Underlying raw, uncompressed data stream.
	rawW io.Writer

	// Compression layer.
	zipW io.Writer

	// Buffers data into MaxChunkSize chunks.
	chunkBuf *bytes.Buffer

	// Index with records which contain chunk offsets.
	index []record

	// Accumulator representing uncompressed offset.
	rawOff int64

	// Accumulator representing compressed offset.
	sizeAcc *util.SizeAccumulator

	// Holds trailer data.
	trailer *trailer
}

func (w *writer) addToIndex() {
	a, b := w.rawOff, int64(w.sizeAcc.Size())
	w.index = append(w.index, record{a, b})
}

func (w *writer) flushBuffer(flushSize int) (int, error) {
	// Add record with start offset of the current chunk.
	w.addToIndex()

	w.zipW = snappy.NewWriter(io.MultiWriter(w.rawW, w.sizeAcc))

	// Compress and flush the current chunk.
	rawN, err := w.zipW.Write(w.chunkBuf.Next(flushSize))
	if err != nil {
		return rawN, err
	}

	// Update offset for the current chunk. The compressed data
	// offset is updated in background using a SizeAccumulator
	// in combination with a MultiWriter.
	w.rawOff += int64(rawN)
	return rawN, nil
}

func (w *writer) Write(p []byte) (n int, err error) {
	// Handle uncompressed stream.
	if w.trailer.algo == AlgoNone {
		n, err := w.rawW.Write(p)
		if err != nil {
			return n, err
		}
		w.rawOff += int64(n)
		return n, nil
	}

	// Handle compressed stream.
	written := len(p)
	// Compress only MaxChunkSize equal chunks.
	for {
		n, _ := w.chunkBuf.Write(p[:util.Min(len(p), MaxChunkSize)])

		if w.chunkBuf.Len() < MaxChunkSize {
			break
		}

		if _, err := w.flushBuffer(MaxChunkSize); err != nil {
			return 0, err
		}
		p = p[n:]
	}
	return written, nil
}

// Return a WriteCloser with compression support.
func NewWriter(w io.Writer, algo Algorithm) io.WriteCloser {
	s := &util.SizeAccumulator{}
	return &writer{
		sizeAcc:  s,
		rawW:     w,
		chunkBuf: &bytes.Buffer{},
		trailer:  &trailer{algo: algo},
	}
}

func (w *writer) Close() error {
	// Handle trailer of uncompressed file.
	if w.trailer.algo == AlgoNone {
		var trailerSizeBuf = make([]byte, TrailerSize)
		w.trailer.maxFileOffset = uint64(w.rawOff)
		w.trailer.marshal(trailerSizeBuf)
		_, err := w.rawW.Write(trailerSizeBuf)
		if err != nil {
			return err
		}

		return nil
	}

	// Write remaining bytes left in buffer and update index.
	if _, err := w.flushBuffer(w.chunkBuf.Len()); err != nil {
		return err
	}
	w.addToIndex()

	// Handle trailer of uncompressed file.
	// Write compression index trailer and close stream.
	w.trailer.indexSize = uint64(IndexChunkSize * len(w.index))
	indexBuf := make([]byte, w.trailer.indexSize)
	indexBufStartOff := indexBuf

	for _, record := range w.index {
		record.marshal(indexBuf)
		indexBuf = indexBuf[IndexChunkSize:]
	}

	if n, err := w.rawW.Write(indexBufStartOff); err != nil || uint64(n) != w.trailer.indexSize {
		return err
	}

	// Write trailer buffer (algo, chunksize, indexsize)
	// at the end of file and close the stream.
	var trailerSizeBuf = make([]byte, TrailerSize)
	w.trailer.maxFileOffset = uint64(w.rawOff)
	w.trailer.marshal(trailerSizeBuf)

	if _, err := w.rawW.Write(trailerSizeBuf); err != nil {
		return err
	}

	return nil
}
