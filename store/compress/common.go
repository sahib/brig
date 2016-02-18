package compress

import (
	"encoding/binary"
	"errors"
)

var (
	ErrBadIndex = errors.New("Broken compression index.")
)

const (
	MaxChunkSize   = 64 * 1024
	IndexChunkSize = 16
	TrailerSize    = 16
)

const (
	AlgoNone = iota
	AlgoSnappy
	AlgoLZ4
	// TODO: AlgoLZ4?
)

type Algorithm byte

type record struct {
	rawOff int64
	zipOff int64
}

type trailer struct {
	algo      Algorithm
	chunksize uint32
	indexSize uint64
}

func (t *trailer) marshal(buf []byte) {
	binary.LittleEndian.PutUint32(buf[0:4], uint32(t.algo))
	binary.LittleEndian.PutUint32(buf[4:8], t.chunksize)
	binary.LittleEndian.PutUint64(buf[8:], t.indexSize)
}

func (t *trailer) unmarshal(buf []byte) {
	t.algo = Algorithm(binary.LittleEndian.Uint32(buf[0:4]))
	t.chunksize = binary.LittleEndian.Uint32(buf[4:8])
	t.indexSize = binary.LittleEndian.Uint64(buf[8:])
}

func (rc *record) marshal(buf []byte) {
	binary.LittleEndian.PutUint64(buf[0:8], uint64(rc.rawOff))
	binary.LittleEndian.PutUint64(buf[8:16], uint64(rc.zipOff))
}

func (rc *record) unmarshal(buf []byte) {
	rc.rawOff = int64(binary.LittleEndian.Uint64(buf[0:8]))
	rc.zipOff = int64(binary.LittleEndian.Uint64(buf[8:16]))
}
