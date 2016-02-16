package main

import (
	"bytes"
	"fmt"
	"io"

	"github.com/disorganizer/brig/util"
	"github.com/golang/snappy"
)

type writer struct {
	sizeAcc  *util.SizeAccumulator
	rawW     io.Writer
	zipW     io.Writer
	chunkBuf *bytes.Buffer
	index    []Block
	rawOff   int64
	trailer  *Trailer
}

func (w *writer) addToIndex() {
	a, b := w.rawOff, int64(w.sizeAcc.Size())
	fmt.Println(a, b)
	w.index = append(w.index, Block{a, b})
}

func (w *writer) flushBuffer(flushSize int) (int, error) {
	w.addToIndex()
	// Compress and flush the current block.
	rawN, err := w.zipW.Write(w.chunkBuf.Next(flushSize))
	if err != nil {
		return rawN, err
	}

	// Build and update index for the current block.
	w.rawOff += int64(rawN)
	return rawN, nil
}

func (w *writer) Write(p []byte) (n int, err error) {
	pSize := len(p)
	// Compress only MaxBlockSize equal chunks.
	for {
		n, _ := w.chunkBuf.Write(p[:util.Min(len(p), MaxBlockSize)])

		if w.chunkBuf.Len() < MaxBlockSize {
			break
		}

		if _, err := w.flushBuffer(MaxBlockSize); err != nil {
			return 0, err
		}
		p = p[n:]
	}
	return pSize, nil
}

func NewWriter(w io.Writer, algo Algorithm) io.WriteCloser {
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

	// Write remaining bytes left in buffer.
	if _, err := w.flushBuffer(w.chunkBuf.Len()); err != nil {

		fmt.Println("Close():", err)
		return err
	}
	w.addToIndex()

	// Write compression index trailer and close stream.
	w.trailer.indexSize = uint64(IndexBlockSize * len(w.index))

	// TODO: Variablen bezeichnungen noch etwas aufräumen?
	trailerBuf := make([]byte, w.trailer.indexSize)
	trailerBufStart := trailerBuf
	for _, blkidx := range w.index {
		blkidx.marshal(trailerBuf)
		trailerBuf = trailerBuf[IndexBlockSize:]
	}

	if n, err := w.rawW.Write(trailerBufStart); err != nil || uint64(n) != w.trailer.indexSize {
		return err
	}

	// Write index trailer size at the end of stream.
	var trailerSizeBuf = make([]byte, TrailerSize)
	w.trailer.marshal(trailerSizeBuf)
	//binary.LittleEndian.PutUint32(trailerSizeBuf[0:4], uint32(w.trailer.algo))
	//binary.LittleEndian.PutUint32(trailerSizeBuf[4:8], MaxBlockSize)
	//binary.LittleEndian.PutUint64(trailerSizeBuf[8:], w.trailer.indexSize)
	if _, err := w.rawW.Write(trailerSizeBuf); err != nil {
		fmt.Println("Error writing trailerSizeBuf:", err)
		return err
	}

	if cl, ok := w.rawW.(io.Closer); ok {
		return cl.Close()
	}

	return nil
}
