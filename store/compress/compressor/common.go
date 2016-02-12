package main

import (
	"encoding/binary"
	"errors"
)

var (
	ErrBadBlockIndex = errors.New("Invalid byte index while reading index.")
)

const (
	MaxBlockSize   = 64 * 1024
	HeaderBufSize  = 8
	IndexBlockSize = 16
)

const (
	AlgoNone = iota
	AlgoSnappy
)

type Algorithm byte

type Block struct {
	rawOff int64
	zipOff int64
}

func (bl *Block) marshal(buf []byte) {
	binary.LittleEndian.PutUint64(buf[0:8], uint64(bl.rawOff))
	binary.LittleEndian.PutUint64(buf[8:16], uint64(bl.zipOff))
}

func (bl *Block) unmarshal(buf []byte) {
	bl.rawOff = int64(binary.LittleEndian.Uint64(buf[0:8]))
	bl.zipOff = int64(binary.LittleEndian.Uint64(buf[8:16]))
}
