package compress

import (
	"encoding/binary"
	"errors"
)

var (
	// ErrBadIndex is returned on invalid compression index.
	ErrBadIndex = errors.New("Broken compression index")
)

const (
	maxChunkSize   = 64 * 1024
	indexChunkSize = 16
	trailerSize    = 16
)

const (
	// AlgoNone represents a ,,uncompressed'' algorithm.
	AlgoNone = iota

	// AlgoSnappy represents the snappy compression algorithm:
	// https://en.wikipedia.org/wiki/Snappy_(software)
	AlgoSnappy

	//AlgoLZ4 represents the lz4 compression algorithm:
	// https://en.wikipedia.org/wiki/LZ4_(compression_algorithm)
	AlgoLZ4
)

// AlgorithmType user defined type to store the algorithm type.
type AlgorithmType byte

// record structure reprenents a offset mapping {uncompressed offset, compressedOffset}.
// A chunk of maxChunkSize is defined by two records. The size of a specific
// record can be determinated by a simple substitution of two record offsets.
type record struct {
	rawOff int64
	zipOff int64
}

// trailer holds basic information about the compressed file.
type trailer struct {
	algo      AlgorithmType
	chunksize uint32
	indexSize uint64
}

func (t *trailer) marshal(buf []byte) {
	binary.LittleEndian.PutUint32(buf[0:4], uint32(t.algo))
	binary.LittleEndian.PutUint32(buf[4:8], t.chunksize)
	binary.LittleEndian.PutUint64(buf[8:16], t.indexSize)
}

func (t *trailer) unmarshal(buf []byte) {
	t.algo = AlgorithmType(binary.LittleEndian.Uint32(buf[0:4]))
	t.chunksize = binary.LittleEndian.Uint32(buf[4:8])
	t.indexSize = binary.LittleEndian.Uint64(buf[8:16])
}

func (rc *record) marshal(buf []byte) {
	binary.LittleEndian.PutUint64(buf[0:8], uint64(rc.rawOff))
	binary.LittleEndian.PutUint64(buf[8:16], uint64(rc.zipOff))
}

func (rc *record) unmarshal(buf []byte) {
	rc.rawOff = int64(binary.LittleEndian.Uint64(buf[0:8]))
	rc.zipOff = int64(binary.LittleEndian.Uint64(buf[8:16]))
}
