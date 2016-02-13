package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/golang/snappy"
)

// TODO: Tests schreiben.
// TODO: linter durchlaufen lassen.

type reader struct {
	rawR         io.ReadSeeker
	zipR         io.Reader
	index        []Block
	fileEndOff   int64
	headerBuf    []byte
	tailBuf      []byte
	readBuf      *bytes.Buffer
	headerParsed bool
	blockStarted bool
	blockSize    int64
}

func (r *reader) Seek(offset int64, whence int) (int64, error) {
	return offset, nil
}

// Optimierung: Nutze binäre suche um korrekten index zu finden.
func (r *reader) currBlock(currOff int64) int64 {
	prevZipOff := int64(0)
	for _, block := range r.index {
		if block.zipOff > currOff {
			break
		}
		prevZipOff = block.zipOff
	}
	return prevZipOff + HeaderBufSize
}

func (r *reader) parseHeaderIfNeeded() error {
	if r.headerParsed {
		return nil
	}

	if _, err := r.rawR.Read(r.headerBuf[:cap(r.headerBuf)]); err != nil {
		fmt.Println(err)
		return err
	}

	if _, err := r.rawR.Seek(-8, os.SEEK_END); err != nil {
		return err
	}

	// Read size of tail.
	buf := [8]byte{}
	if n, err := r.rawR.Read(buf[:]); err != nil || n != 8 {
		return err
	}

	tailSize := binary.LittleEndian.Uint64(buf[:])
	r.tailBuf = make([]byte, tailSize)
	var err error
	seekIdx := -(int64(tailSize) + 8)
	if r.fileEndOff, err = r.rawR.Seek(seekIdx, os.SEEK_END); err != nil {
		fmt.Println(err)
		return err
	}

	if _, err := r.rawR.Read(r.tailBuf); err != nil {
		fmt.Println(err)
		return err
	}

	//Build Index
	for i := uint64(0); i < (tailSize / IndexBlockSize); i++ {
		b := Block{}
		b.unmarshal(r.tailBuf)
		r.index = append(r.index, b)
		r.tailBuf = r.tailBuf[IndexBlockSize:]
	}
	if _, err := r.rawR.Seek(HeaderBufSize, os.SEEK_SET); err != nil {
		return err
	}
	r.headerParsed = true

	return nil
}

// Macht fast dasselbe wie currBlock? Ineffizient/unnötig?
func (r *reader) rawBlockSize(currOff int64) int64 {
	prevOff, nextOff := int64(0), int64(0)
	for _, block := range r.index {
		nextOff = block.rawOff
		if block.rawOff >= currOff && block.rawOff != 0 {
			break
		}
		prevOff = block.rawOff
	}
	return nextOff - prevOff
}

func (r *reader) Read(p []byte) (int, error) {
	if err := r.parseHeaderIfNeeded(); err != nil {
		fmt.Println(err)
		return 0, err
	}

	read := 0
	if r.readBuf.Len() != 0 {
		n, err := r.readBlockBuffered(p)
		if err != nil {
			return n, err
		}
		read += n
	} else {
		currZipOff, err := r.startReadOffset()
		if err != nil {
			return 0, err
		}

		r.blockSize = r.rawBlockSize(currZipOff)
		fmt.Println("blockSize:", r.blockSize)
		if r.blockSize == 0 {
			// TODO?!
			return 0, io.EOF
		}
	}

	n, err := r.readBlockBuffered(p[read:])
	if err != nil {
		return n, err
	}

	read += n
	fmt.Println("read", read, len(p))
	return read, nil
}

func (r *reader) startReadOffset() (int64, error) {
	// Get current raw position
	curOff, err := r.rawR.Seek(0, os.SEEK_CUR)
	fmt.Println("CurrentOff:", curOff)
	if err != nil {
		return 0, err
	}

	// Get zip offset and set cursor to that position
	currZipOff := r.currBlock(curOff)
	fmt.Println("CurrentZIPOff:", currZipOff)
	if _, err = r.rawR.Seek(currZipOff, os.SEEK_SET); err != nil {
		fmt.Println(err)
		return 0, err
	}
	r.blockStarted = true
	return currZipOff, nil
}

func (r *reader) readBlockBuffered(p []byte) (int, error) {
	n, err := io.CopyN(r.readBuf, r.zipR, r.blockSize)
	r.blockSize -= n
	if err != nil {
		return 0, err
	}

	// Nothing to read, block finished.
	if n == 0 {
		r.blockStarted = false
	}

	nb, err := r.readBuf.Read(p) // Schreibe N in p
	if err != nil {
		return 0, err
	}
	fmt.Println("buff", nb)
	return nb, nil
}

func NewReader(r io.ReadSeeker) io.ReadSeeker {
	return &reader{
		rawR:      r,
		zipR:      snappy.NewReader(r),
		headerBuf: make([]byte, 0, HeaderBufSize),
		readBuf:   &bytes.Buffer{},
	}
}
