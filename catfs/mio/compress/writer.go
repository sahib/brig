package compress

import (
	"bytes"
	"fmt"
	"io"

	"github.com/sahib/brig/util"
)

// Writer implements a compression writer.
type Writer struct {
	// Underlying raw, uncompressed data stream.
	rawW io.Writer

	// Buffers data into maxChunkSize chunks.
	chunkBuf *bytes.Buffer

	// Index with records which contain chunk offsets.
	index []record

	// Accumulator representing uncompressed offset.
	rawOff int64

	// Accumulator representing compressed offset.
	zipOff int64

	// Holds trailer data.
	trailer *trailer

	// Holds algorithm interface.
	algo Algorithm

	// Type of the algorithm
	algoType AlgorithmType

	// Becomes true after the first write.
	headerWritten bool
}

func (w *Writer) addRecordToIndex() {
	w.index = append(w.index, record{w.rawOff, w.zipOff})
}

func (w *Writer) flushBuffer(data []byte) error {
	if len(data) <= 0 {
		return nil
	}

	// Add record with start offset of the current chunk.
	w.addRecordToIndex()

	// Compress and flush the current chunk.
	encData, err := w.algo.Encode(data)
	if err != nil {
		return err
	}

	n, err := w.rawW.Write(encData)
	if err != nil {
		return err
	}

	// Update offset for the current chunk.
	w.rawOff += int64(len(data))
	w.zipOff += int64(n)
	return nil
}

func (w *Writer) writeHeaderIfNeeded() error {
	if w.headerWritten {
		return nil
	}

	if _, err := w.rawW.Write(makeHeader(w.algoType, currentVersion)); err != nil {
		return err
	}

	w.headerWritten = true
	w.zipOff += headerSize
	return nil
}

// ReadFrom implements io.ReaderFrom
func (w *Writer) ReadFrom(r io.Reader) (n int64, err error) {
	fmt.Println("READ ME")
	read := 0

	buf := make([]byte, maxChunkSize)

	if err := w.writeHeaderIfNeeded(); err != nil {
		return 0, err
	}

	for {
		// Only last block may be < maxChunkSize.
		// So we need to make sure to fill the buffer as best we can.
		n, rerr := io.ReadFull(r, buf)
		read += n
		if rerr != nil && rerr != io.ErrUnexpectedEOF && rerr != io.EOF {
			return int64(read), rerr
		}

		werr := w.flushBuffer(buf[:n])
		if werr != nil && werr != io.EOF {
			return int64(read), werr
		}

		if werr == io.EOF || rerr == io.EOF || rerr == io.ErrUnexpectedEOF {
			return int64(read), nil
		}
	}
}

func (w *Writer) Write(p []byte) (n int, err error) {
	if err := w.writeHeaderIfNeeded(); err != nil {
		return 0, err
	}

	written := len(p)
	// Compress only maxChunkSize equal chunks.
	for {
		n, _ := w.chunkBuf.Write(p[:util.Min(len(p), maxChunkSize)])

		if w.chunkBuf.Len() < maxChunkSize {
			break
		}

		if err := w.flushBuffer(w.chunkBuf.Next(maxChunkSize)); err != nil {
			return 0, err
		}
		p = p[n:]
	}
	return written, nil
}

// NewWriter returns a WriteCloser with compression support.
func NewWriter(w io.Writer, algoType AlgorithmType) (*Writer, error) {
	algo, err := AlgorithmFromType(algoType)
	if err != nil {
		return nil, err
	}
	return &Writer{
		rawW:     w,
		algo:     algo,
		algoType: algoType,
		chunkBuf: &bytes.Buffer{},
		trailer:  &trailer{},
	}, nil
}

// Close cleans up internal resources.
// Make sure to call close always since it might write data.
func (w *Writer) Close() error {
	if err := w.writeHeaderIfNeeded(); err != nil {
		return err
	}

	// Write remaining bytes left in buffer and update index.
	if err := w.flushBuffer(w.chunkBuf.Bytes()); err != nil {
		return err
	}
	w.addRecordToIndex()

	// Handle trailer of uncompressed file.
	// Write compression index trailer and close stream.
	w.trailer.indexSize = uint64(indexChunkSize * len(w.index))
	indexBuf := make([]byte, w.trailer.indexSize)
	indexBufStartOff := indexBuf
	for _, record := range w.index {
		record.marshal(indexBuf)
		indexBuf = indexBuf[indexChunkSize:]
	}

	if n, err := w.rawW.Write(indexBufStartOff); err != nil || uint64(n) != w.trailer.indexSize {
		return err
	}

	// Write trailer buffer (algo, chunksize, indexsize)
	// at the end of file and close the stream.
	trailerSizeBuf := make([]byte, trailerSize)
	w.trailer.marshal(trailerSizeBuf)

	if _, err := w.rawW.Write(trailerSizeBuf); err != nil {
		return err
	}
	return nil
}
