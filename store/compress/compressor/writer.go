package main

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
	index []Record

	// Accumulator representing uncompressed offset.
	rawOff int64

	// Accumulator representing compressed offset.
	sizeAcc *util.SizeAccumulator

	// Holds trailer data.
	trailer *Trailer
}

func (w *writer) addToIndex() {
	a, b := w.rawOff, int64(w.sizeAcc.Size())
	w.index = append(w.index, Record{a, b})
}

func (w *writer) flushBuffer(flushSize int) (int, error) {
	// Add record with start offset of the current chunk.
	w.addToIndex()

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
	if algo == AlgoNone {
		return util.NopWriteCloser(w)
	}

	s := &util.SizeAccumulator{}
	return &writer{
		sizeAcc:  s,
		zipW:     snappy.NewWriter(io.MultiWriter(w, s)),
		rawW:     w,
		chunkBuf: &bytes.Buffer{},
		trailer:  &Trailer{},
	}
}

func (w *writer) Close() error {
	// Write remaining bytes left in buffer and update index.
	if _, err := w.flushBuffer(w.chunkBuf.Len()); err != nil {
		return err
	}
	w.addToIndex()

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
	w.trailer.marshal(trailerSizeBuf)

	if _, err := w.rawW.Write(trailerSizeBuf); err != nil {
		return err
	}

	if cl, ok := w.rawW.(io.Closer); ok {
		return cl.Close()
	}
	return nil
}
