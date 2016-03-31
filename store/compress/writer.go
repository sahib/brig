package compress

import (
	"bytes"
	"io"

	"github.com/disorganizer/brig/util"
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

func (w *writer) flushBuffer(data []byte) error {

	if len(data) <= 0 {
		return nil
	}

	// Add record with start offset of the current chunk.
	w.addToIndex()

	w.zipW = wrapWriter(io.MultiWriter(w.rawW, w.sizeAcc), w.trailer.algo)

	// Compress and flush the current chunk.
	rawN, err := w.zipW.Write(data)
	if err != nil {
		return err
	}

	// Update offset for the current chunk. The compressed data
	// offset is updated in background using a SizeAccumulator
	// in combination with a MultiWriter.
	w.rawOff += int64(rawN)
	return nil
}

func (w *writer) ReadFrom(r io.Reader) (n int64, err error) {
	read := 0
	buf := [MaxChunkSize]byte{}
	for {
		n, rerr := r.Read(buf[:])
		read += n
		if rerr != nil && rerr != io.EOF {
			return int64(read), rerr
		}

		werr := w.flushBuffer(buf[:n])
		if werr != nil && werr != io.EOF {
			return int64(read), werr
		}
		if werr == io.EOF || rerr == io.EOF {
			return int64(read), nil
		}
	}
}

func (w *writer) Write(p []byte) (n int, err error) {
	written := len(p)
	// Compress only MaxChunkSize equal chunks.
	for {
		n, _ := w.chunkBuf.Write(p[:util.Min(len(p), MaxChunkSize)])

		if w.chunkBuf.Len() < MaxChunkSize {
			break
		}

		if err := w.flushBuffer(w.chunkBuf.Next(MaxChunkSize)); err != nil {
			return 0, err
		}
		p = p[n:]
	}
	return written, nil
}

// Return a WriteCloser with compression support.
func NewWriter(w io.Writer, algo AlgorithmType) io.WriteCloser {
	s := &util.SizeAccumulator{}
	return &writer{
		sizeAcc:  s,
		rawW:     w,
		chunkBuf: &bytes.Buffer{},
		trailer:  &trailer{algo: algo},
	}
}

func (w *writer) Close() error {
	// Write remaining bytes left in buffer and update index.
	if err := w.flushBuffer(w.chunkBuf.Bytes()); err != nil {
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
	w.trailer.marshal(trailerSizeBuf)

	if _, err := w.rawW.Write(trailerSizeBuf); err != nil {
		return err
	}

	return nil
}
