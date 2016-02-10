package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/disorganizer/brig/util"
	"github.com/golang/snappy"
	"io"
	"os"
)

var (
	ErrBadBlockIndex = errors.New("Invalid byte index while reading index.")
)

const (
	MaxBlockSize = 64 * 1024
)

func openFiles(from, to string) (*os.File, *os.File, error) {
	fdFrom, err := os.Open(from)
	if err != nil {
		return nil, nil, err
	}

	fdTo, err := os.OpenFile(to, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		fdFrom.Close()
		return nil, nil, err
	}

	return fdFrom, fdTo, nil
}

type snappyWriter struct {
	rawW          io.Writer
	zipW          io.Writer
	buf           *bytes.Buffer
	index         []BlockIndex
	headerWritten bool
	compression   uint64
}

type snappyReader struct {
	rawR      io.ReadSeeker
	zipR      io.Reader
	index     []BlockIndex
	headerBuf []byte
	tailBuf   []byte
}

type BlockIndex struct {
	fileOffset uint64
	zipOffset  uint64
	zipSize    uint64
}

func (bl *BlockIndex) marshal(buf []byte) {
	fmt.Println(bl.fileOffset, bl.zipOffset, bl.zipSize)
	binary.PutUvarint(buf[00:10], bl.fileOffset)
	binary.PutUvarint(buf[10:20], bl.zipOffset)
	binary.PutUvarint(buf[20:30], bl.zipSize)
}

func (bl *BlockIndex) unmarshal(buf []byte) error {
	var n int
	if bl.fileOffset, n = binary.Uvarint(buf[00:10]); n <= 0 {
		return ErrBadBlockIndex
	}
	if bl.zipOffset, n = binary.Uvarint(buf[10:20]); n <= 0 {
		return ErrBadBlockIndex
	}
	if bl.zipSize, n = binary.Uvarint(buf[20:30]); n <= 0 {
		return ErrBadBlockIndex
	}
	return nil
}

func (sr *snappyReader) Seek(offset int64, whence int) (int64, error) {
	return offset, nil
}

// Read a snappy compressed stream, with random access.
func (sr *snappyReader) Read(p []byte) (int, error) {

	// Do on first read when header buffer is empty.
	fmt.Println(len(sr.headerBuf))
	if len(sr.headerBuf) == 0 {

		if _, err := sr.rawR.Read(sr.headerBuf[:cap(sr.headerBuf)]); err != nil {
			fmt.Println(err)
		}
		if _, err := sr.rawR.Seek(-8, os.SEEK_END); err != nil {
			fmt.Println(err)
			return 0, err
		}
		// Read size of tail.
		var buf = make([]byte, 8)
		if n, err := sr.rawR.Read(buf); err != nil || n != 8 {
			fmt.Println(err)
			return n, err
		}

		tailSize, n := binary.Uvarint(buf)
		if n <= 0 {
			return 0, ErrBadBlockIndex
		}

		sr.tailBuf = make([]byte, tailSize)
		if _, err := sr.rawR.Seek(-(int64(tailSize) + 8), os.SEEK_END); err != nil {
			fmt.Println(err)
			return 0, err
		}
		if _, err := sr.rawR.Read(sr.tailBuf); err != nil {
			fmt.Println(err)
		}

		//Build Index
		for i := uint64(0); i < tailSize; i++ {
			b := BlockIndex{}
			b.unmarshal(sr.tailBuf)
			sr.index = append(sr.index, b)
			sr.tailBuf = sr.tailBuf[30:]
		}
		sr.rawR.Seek(32, os.SEEK_SET)
	}

	// curOff, err := sr.rawR.Seek(0, os.SEEK_CUR)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// Identify current block with getCurrentIndexBlock.
	// Go to the begining of the current block.
	// Read block in MaxBlockSize bytes buffer.
	// Read from buffer into p until full or eof.
	// TODO: Don't let snappy read tail.
	return sr.zipR.Read(p)
}

func (sw *snappyWriter) writeHeaderIfNeeded() error {
	if !sw.headerWritten {
		fmt.Println("writing header")
		buf := [32]byte{}
		binary.PutUvarint(buf[00:16], sw.compression)
		binary.PutUvarint(buf[16:32], MaxBlockSize)
		if _, err := sw.rawW.Write(buf[:]); err != nil {
			return err
		}
	}

	sw.headerWritten = true
	return nil
}

func (sw *snappyWriter) appendToBlockIndex(sizeCompressed int) {
	var fOff, zOff, zBlockSize = uint64(0), uint64(0), uint64(sizeCompressed)
	if len(sw.index) > 0 {
		var prevIdx = sw.index[len(sw.index)-1]
		fOff = prevIdx.fileOffset + MaxBlockSize
		zOff = prevIdx.zipOffset + uint64(sizeCompressed)
	}
	sw.index = append(sw.index, BlockIndex{fOff, zOff, zBlockSize})

}

func (sw *snappyWriter) flushBlock(flushSize int) (int, error) {

	// Compress and flush the current block.
	nc, err := sw.zipW.Write(sw.buf.Next(flushSize))
	if err != nil {
		fmt.Println("flushBlock:", err)
		return nc, err
	}

	// Build and update index for the current block.
	sw.appendToBlockIndex(nc)

	return nc, nil
}

// Write a snappy compressed stream with index.
func (sw *snappyWriter) Write(p []byte) (n int, err error) {

	if err := sw.writeHeaderIfNeeded(); err != nil {
		return 0, err
	}

	// Compress only MaxBlockSize equal chunks.
	for {
		n, _ := sw.buf.Write(p[:util.Min(len(p), MaxBlockSize)])
		fmt.Println(n)

		// Flush the current block.
		if sw.buf.Len() >= MaxBlockSize {
			if n, err := sw.flushBlock(MaxBlockSize); err != nil {
				return n, err
			}
			// Forget flushed input.
			p = p[n:]
			continue
		}
		break
	}

	// Fake bytes written, as expeted by some functions.
	return len(p), nil
}

func NewReader(r io.ReadSeeker) io.ReadSeeker {
	return &snappyReader{
		rawR:      r,
		zipR:      snappy.NewReader(io.Reader(r)),
		headerBuf: make([]byte, 0, 30),
	}
}

func NewWriter(w io.Writer) io.WriteCloser {
	return &snappyWriter{
		zipW:        snappy.NewWriter(w),
		rawW:        w,
		buf:         &bytes.Buffer{},
		compression: 1,
	}
}

func (sw *snappyWriter) Close() error {

	// Write header on empty files.
	sw.writeHeaderIfNeeded()

	// Write remaining bytes left in buffer.
	nc, err := sw.zipW.Write(sw.buf.Bytes())
	if err != nil {
		fmt.Println("Close():", err)
		return err
	}
	sw.appendToBlockIndex(nc)

	// Write compression index tail and close stream.
	indexSize := uint64(30 * len(sw.index))
	tailBuf := make([]byte, indexSize)
	tailBufStart := tailBuf
	if len(sw.index) > 0 {
		for _, blkidx := range sw.index {
			blkidx.marshal(tailBuf)
			tailBuf = tailBuf[30:]
		}
	}

	n, err := sw.rawW.Write(tailBufStart)
	if err != nil || uint64(n) != indexSize {
		fmt.Println("Close():", err, "n:", n, "idxSize:", indexSize, "TailBufLen:", len(tailBuf))
		return err
	}

	// Write index tail size at the end of stream.
	var tailSizeBuf = make([]byte, 8)
	binary.PutUvarint(tailSizeBuf, indexSize)
	sw.rawW.Write(tailBuf)

	cl, ok := sw.rawW.(io.Closer)
	if ok {
		return cl.Close()
	}

	return nil
}
