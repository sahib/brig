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
	ErrBadBlockIndex = errors.New("Invalid byte index while reading index")
)

const (
	MaxBlockSize   = 64 * 1024
	HeaderBufSize  = 32
	IndexBlockSize = 30
)

func openFiles(from, to string) (*os.File, *os.File, error) {
	fdFrom, err := os.Open(from)
	if err != nil {
		return nil, nil, err
	}

	fdTo, err := os.OpenFile(to, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		fdFrom.Close()
		return nil, nil, err
	}

	return fdFrom, fdTo, nil
}

type snappyWriter struct {
	sizeAcc       *util.SizeAccumulator
	rawW          io.Writer
	zipW          io.Writer
	buf           *bytes.Buffer
	index         []BlockIndex
	headerWritten bool
	compression   uint64
}

type snappyReader struct {
	rawR       io.ReadSeeker
	zipR       io.Reader
	index      []BlockIndex
	fileEndOff int64
	headerBuf  []byte
	tailBuf    []byte
	readBuf    *bytes.Buffer
}

type BlockIndex struct {
	fileOffset uint64
	zipOffset  uint64
	zipSize    uint64
}

func (bl *BlockIndex) marshal(buf []byte) {
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

func (sr *snappyReader) getCurrentIndexBlock(curIndex uint64) uint64 {
	// offset | zipOffset | zipSize
	if len(sr.index) == 0 {
		panic("Index len is zero.")
	}

	fileIdx, zipIdx := uint64(0), uint64(0)
	for _, BlockIndex := range sr.index {
		zipIdx += BlockIndex.zipSize
		if zipIdx >= curIndex {
			return BlockIndex.zipOffset
		}
		fileIdx += MaxBlockSize
	}
	return 0
}

// Read a snappy compressed stream, with random access.
func (sr *snappyReader) Read(p []byte) (int, error) {

	// Do on first read when header buffer is empty.
	if len(sr.headerBuf) == 0 {

		n, err := sr.rawR.Read(sr.headerBuf[:cap(sr.headerBuf)])
		if err != nil {
			fmt.Println(err)
			return 0, err
		}

		_, err = sr.rawR.Seek(-8, os.SEEK_END)
		if err != nil {
			fmt.Println(err)
			return 0, err
		}

		// Read size of tail.
		var buf = make([]byte, 8)
		bytes, err := sr.rawR.Read(buf)
		if err != nil || bytes != 8 {
			return n, err
		}

		tailSize, n := binary.Uvarint(buf)
		if n <= 0 {
			return 0, ErrBadBlockIndex
		}

		sr.tailBuf = make([]byte, tailSize)
		sr.fileEndOff, err = sr.rawR.Seek(-(int64(tailSize) + 8), os.SEEK_END)
		if err != nil {
			fmt.Println(err)
			return 0, err
		}

		if _, err := sr.rawR.Read(sr.tailBuf); err != nil {
			fmt.Println(err)
			return 0, err
		}

		//Build Index
		for i := uint64(0); i < (tailSize / IndexBlockSize); i++ {
			b := BlockIndex{}

			b.unmarshal(sr.tailBuf)

			sr.index = append(sr.index, b)
			sr.tailBuf = sr.tailBuf[IndexBlockSize:]
		}
		sr.rawR.Seek(HeaderBufSize, os.SEEK_SET)
	}

	curOff, err := sr.rawR.Seek(0, os.SEEK_CUR)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}

	curZipIdx := sr.getCurrentIndexBlock(uint64(curOff))
	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	newOffset, err := sr.rawR.Seek(int64(curZipIdx), os.SEEK_SET)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}

	for newOffset <= sr.fileEndOff {

		endMarker := util.UMin(uint(newOffset+MaxBlockSize), uint(sr.fileEndOff))
		n, err := io.CopyN(sr.readBuf, sr.zipR, int64(endMarker))
		if err != nil {
			fmt.Println(err, n)
			return 0, err
		}

		newOffset += n
		nb, _ := sr.readBuf.Read(p)
		if nb == 0 {
			break
		}

	}
	return len(p), nil
}

func (sw *snappyWriter) writeHeaderIfNeeded() error {

	if !sw.headerWritten {
		buf := [32]byte{}
		binary.PutUvarint(buf[00:16], sw.compression)
		binary.PutUvarint(buf[16:32], MaxBlockSize)
		_, err := sw.rawW.Write(buf[:])
		if err != nil {
			return err
		}
	}

	sw.headerWritten = true
	return nil
}

func (sw *snappyWriter) appendToBlockIndex() {
	var fOff, zOff, zBlockSize = uint64(0), uint64(0), uint64(sw.sizeAcc.Size())
	if len(sw.index) > 0 {
		var prevIdx = sw.index[len(sw.index)-1]
		fOff = prevIdx.fileOffset + MaxBlockSize
		zOff = prevIdx.zipOffset + zBlockSize
	}
	sw.index = append(sw.index, BlockIndex{fOff, zOff, zBlockSize})
	sw.sizeAcc.Reset()

}

func (sw *snappyWriter) flushBlock(flushSize int) (int, error) {

	// Compress and flush the current block.
	nc, err := sw.zipW.Write(sw.buf.Next(flushSize))
	if err != nil {
		fmt.Println("flushBlock:", err)
		return nc, err
	}

	// Build and update index for the current block.
	sw.appendToBlockIndex()

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
		headerBuf: make([]byte, 0, HeaderBufSize),
		readBuf:   &bytes.Buffer{},
	}
}

func NewWriter(w io.Writer) io.WriteCloser {
	s := &util.SizeAccumulator{}
	return &snappyWriter{

		sizeAcc:     s,
		zipW:        snappy.NewWriter(io.MultiWriter(w, s)),
		rawW:        w,
		buf:         &bytes.Buffer{},
		compression: 1,
	}
}

func (sw *snappyWriter) Close() error {

	// Write header on empty files.
	sw.writeHeaderIfNeeded()

	// Write remaining bytes left in buffer.
	_, err := sw.zipW.Write(sw.buf.Bytes())

	if err != nil {
		fmt.Println("Close():", err)
		return err
	}
	sw.appendToBlockIndex()

	// Write compression index tail and close stream.
	indexSize := uint64(IndexBlockSize * len(sw.index))
	tailBuf := make([]byte, indexSize)
	tailBufStart := tailBuf
	if len(sw.index) > 0 {
		for _, blkidx := range sw.index {
			blkidx.marshal(tailBuf)
			tailBuf = tailBuf[IndexBlockSize:]
		}
	}

	n, err := sw.rawW.Write(tailBufStart)
	if err != nil || uint64(n) != indexSize {
		return err
	}

	// Write index tail size at the end of stream.
	var tailSizeBuf = make([]byte, 8)
	binary.PutUvarint(tailSizeBuf, indexSize)
	n, err = sw.rawW.Write(tailSizeBuf)
	if err != nil {
		fmt.Println("Error writing tailSizeBuf:", err)
		return err
	}

	cl, ok := sw.rawW.(io.Closer)
	if ok {
		return cl.Close()
	}

	return nil
}
