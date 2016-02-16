package main

import (
	"encoding/binary"
	"errors"
)

var (
	ErrBadIndex = errors.New("Broken compression index.")
)

const (
	MaxBlockSize   = 64 * 1024
	IndexBlockSize = 16
	TrailerSize       = 16
)

const (
	AlgoNone = iota
	AlgoSnappy
	AlgoLZ4
	// TODO: AlgoLZ4?
)

type Algorithm byte

// Ist Block ein exportierter Typ?
type Block struct {
	rawOff int64
	zipOff int64
}

type Trailer struct {
	algo      Algorithm
	blocksize uint32
	indexSize uint64
}

func (t *Trailer) marshal(buf []byte) {
	binary.LittleEndian.PutUint32(buf[0:4], uint32(t.algo))
	binary.LittleEndian.PutUint32(buf[4:8], t.blocksize)
	binary.LittleEndian.PutUint64(buf[8:], t.indexSize)
}

func (t *Trailer) unmarshal(buf []byte) {
	t.algo = Algorithm(binary.LittleEndian.Uint32(buf[0:4]))
	t.blocksize = binary.LittleEndian.Uint32(buf[4:8])
	t.indexSize = binary.LittleEndian.Uint64(buf[8:])
}

func (bl *Block) marshal(buf []byte) {
	binary.LittleEndian.PutUint64(buf[0:8], uint64(bl.rawOff))
	binary.LittleEndian.PutUint64(buf[8:16], uint64(bl.zipOff))
}

func (bl *Block) unmarshal(buf []byte) {
	bl.rawOff = int64(binary.LittleEndian.Uint64(buf[0:8]))
	bl.zipOff = int64(binary.LittleEndian.Uint64(buf[8:16]))
}
