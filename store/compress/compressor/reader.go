package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/golang/snappy"
)

// TODO: Tests schreiben (leere dateien, blockgröße -1, +0, +1 etc.)
// TODO: linter durchlaufen lassen.
// TODO: Seek.

type reader struct {
	rawR  io.ReadSeeker
	zipR  io.Reader
	index []Block
	// TODO: Noch benötigt?
	fileEndOff   int64
	tailBuf      []byte
	readBuf      *bytes.Buffer
	headerParsed bool
	trailer      *Trailer
}

func (r *reader) Seek(offset int64, whence int) (int64, error) {
	return offset, nil
}

// Return start (prev offset) and end (curr offset) of the block currOff is
// located in. If currOff is 0, the startoffset of the first and second block is
// returned. If currOff is at the end of file the end offset of the last block
// is returned twice.  The difference between prev block and curr block is then
// equal to 0.
func (r *reader) blockLookup(currOff int64) (*Block, *Block) {
	i := sort.Search(len(r.index), func(i int) bool {
		return r.index[i].zipOff > currOff
	})
	// Beginning of the file, first block: prev offset is 0, curr offset is 1
	if i == 0 {
		return &r.index[i], &r.index[i+1]
	}
	// End of the file, last block: prev and curr offset is the last index.
	if i == len(r.index) {
		return &r.index[i-1], &r.index[i-1]
	}
	return &r.index[i-1], &r.index[i]
}

//TODO: Clean code.
func (r *reader) parseHeaderIfNeeded() error {
	if r.headerParsed {
		return nil
	}

	if _, err := r.rawR.Seek(-TailSize, os.SEEK_END); err != nil {
		return err
	}

	buf := [TailSize + 16]byte{}
	if n, err := r.rawR.Read(buf[:]); err != nil || n != TailSize {
		return err
	}
	r.trailer.unmarshal(buf[:])
	r.tailBuf = make([]byte, r.trailer.indexSize)

	var err error
	seekIdx := -(int64(r.trailer.indexSize) + TailSize)
	if r.fileEndOff, err = r.rawR.Seek(seekIdx, os.SEEK_END); err != nil {
		fmt.Println(err)
		return err
	}

	if _, err := r.rawR.Read(r.tailBuf); err != nil {
		fmt.Println(err)
		return err
	}

	//Build Index
	prevBlock := Block{-1, -1}
	for i := uint64(0); i < (r.trailer.indexSize / IndexBlockSize); i++ {
		currBlock := Block{}
		currBlock.unmarshal(r.tailBuf)

		if prevBlock.rawOff >= currBlock.rawOff && prevBlock.zipOff >= currBlock.zipOff {
			return ErrBadIndex
		}

		r.index = append(r.index, currBlock)
		r.tailBuf = r.tailBuf[IndexBlockSize:]
	}
	if _, err := r.rawR.Seek(0, os.SEEK_SET); err != nil {
		return err
	}
	r.headerParsed = true

	return nil
}

func (r *reader) Read(p []byte) (int, error) {
	if err := r.parseHeaderIfNeeded(); err != nil {
		fmt.Println(err)
		return 0, err
	}

	read := 0
	for {
		fmt.Println("READBUF LEN:", r.readBuf.Len())
		if r.readBuf.Len() != 0 {
			n := copy(p, r.readBuf.Next(len(p)))
			read += n
			p = p[n:]
			fmt.Println("P:", len(p))
		}
		if len(p) == 0 {
			break
		}

		if _, err := r.readBlockBuffered(); err != nil {
			fmt.Println("EOF?", read, err)
			return read, err
		}
	}
	fmt.Println("END:", read)
	return read, nil
}

func (r *reader) readBlockBuffered() (int64, error) {
	// Get current raw position
	currOff, err := r.rawR.Seek(0, os.SEEK_CUR)
	if err != nil {
		return 0, err
	}

	// Get compressed offset and set cursor to that position.
	prevBlock, currBlock := r.blockLookup(currOff)
	fmt.Println("#########################", currOff, prevBlock, currBlock)
	if currBlock == nil || prevBlock == nil {
		return 0, ErrBadIndex
	}

	// Blocksize should only be 0 on empty file or at the end of file.
	blockSize := currBlock.rawOff - prevBlock.rawOff
	if blockSize == 0 {
		return 0, io.EOF
	}

	currZipOff := prevBlock.zipOff
	if _, err = r.rawR.Seek(currZipOff, os.SEEK_SET); err != nil {
		return 0, err
	}

	return io.CopyN(r.readBuf, r.zipR, blockSize)
}

func NewReader(r io.ReadSeeker) io.ReadSeeker {
	return &reader{
		rawR:    r,
		zipR:    snappy.NewReader(r),
		readBuf: &bytes.Buffer{},
		trailer: &Trailer{},
	}
}
