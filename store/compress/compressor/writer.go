package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/disorganizer/brig/util"
	"github.com/golang/snappy"
)

type writer struct {
	sizeAcc       *util.SizeAccumulator
	rawW          io.Writer
	zipW          io.Writer
	chunkBuf      *bytes.Buffer
	index         []Block
	headerWritten bool
	rawOff        int64
	algorithm     Algorithm
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

		// Flush the current block.
		//fmt.Println("buflen und max:", w.chunkBuf.Len(), MaxBlockSize)
		if w.chunkBuf.Len() >= MaxBlockSize {
			if _, err := w.flushBuffer(MaxBlockSize); err != nil {
				fmt.Println(err)
				return 0, err
			}
			// Forget flushed input.
			//fmt.Println("p1", len(p), n)
			p = p[n:]
			//fmt.Println("p2", len(p), n)
			continue
		}
		break
	}

	// Fake bytes written, as expeted by some functions.
	return pSize, nil
}

//TODO: Make algorithm a function parameter.
func NewWriter(w io.Writer) io.WriteCloser {
	s := &util.SizeAccumulator{}
	return &writer{
		sizeAcc:   s,
		zipW:      snappy.NewWriter(io.MultiWriter(w, s)),
		rawW:      w,
		chunkBuf:  &bytes.Buffer{},
		algorithm: AlgoSnappy,
	}
}

func (w *writer) Close() error {

	// Write remaining bytes left in buffer.
	if _, err := w.flushBuffer(w.chunkBuf.Len()); err != nil {

		fmt.Println("Close():", err)
		return err
	}
	w.addToIndex()

	// Write compression index tail and close stream.
	indexSize := uint64(IndexBlockSize * len(w.index))

	tailBuf := make([]byte, indexSize)
	tailBufStart := tailBuf
	for _, blkidx := range w.index {
		blkidx.marshal(tailBuf)
		tailBuf = tailBuf[IndexBlockSize:]
	}

	if n, err := w.rawW.Write(tailBufStart); err != nil || uint64(n) != indexSize {
		return err
	}

	// Write index tail size at the end of stream.
	var tailSizeBuf = make([]byte, TailSize)
	binary.LittleEndian.PutUint32(tailSizeBuf[0:4], uint32(w.algorithm))
	binary.LittleEndian.PutUint32(tailSizeBuf[4:8], MaxBlockSize)
	binary.LittleEndian.PutUint64(tailSizeBuf[8:], indexSize)
	if _, err := w.rawW.Write(tailSizeBuf); err != nil {
		fmt.Println("Error writing tailSizeBuf:", err)
		return err
	}

	if cl, ok := w.rawW.(io.Closer); ok {
		return cl.Close()
	}

	return nil
}
