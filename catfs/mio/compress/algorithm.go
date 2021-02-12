package compress

import (
	"errors"

	"github.com/golang/snappy"
	"github.com/klauspost/compress/zstd"
	"github.com/pierrec/lz4/v4"
)

var (
	// ErrBadAlgo is returned on a unsupported/unknown algorithm.
	ErrBadAlgo = errors.New("invalid algorithm type")
)

const (
	// AlgoUnknown represents an unknown algorithm.
	// When trying to use it an error will occur.
	AlgoUnknown = AlgorithmType(iota)

	// AlgoSnappy represents the snappy compression algorithm:
	// https://en.wikipedia.org/wiki/Snappy_(software)
	AlgoSnappy

	//AlgoLZ4 represents the lz4 compression algorithm:
	// https://en.wikipedia.org/wiki/LZ4_(compression_algorithm)
	AlgoLZ4

	// AlgoZstd represents the zstd compression algorithm:
	// https://de.wikipedia.org/wiki/Zstandard
	AlgoZstd
)

// AlgorithmType user defined type to store the algorithm type.
type AlgorithmType byte

// IsValid returns true if `at` is a valid algorithm type.
func (at AlgorithmType) IsValid() bool {
	switch at {
	case AlgoSnappy, AlgoLZ4, AlgoZstd:
		return true
	}

	return false
}

func (at AlgorithmType) String() string {
	name, ok := algoToString[at]
	if !ok {
		return "unknown"
	}

	return name
}

// Algorithm is the common interface for all supported algorithms.
type Algorithm interface {
	Encode(dst, src []byte) ([]byte, error)
	Decode(dst, src []byte) ([]byte, error)
	MaxEncodeBufferSize() int
}

type snappyAlgo struct{}

type lz4Algo struct {
	compressor *lz4.Compressor
}

type zstdAlgo struct{}

var (
	zstdWriter *zstd.Encoder
	zstdReader *zstd.Decoder
)

func init() {
	var err error

	// NOTE: zstd package allows us to use the same writer and reader
	//       stateless if we just use block encoding/decoding.
	//       This saves us some extra allocations.

	// TODO: configure compression level?
	zstdWriter, err = zstd.NewWriter(
		nil,
		zstd.WithEncoderLevel(zstd.SpeedDefault),
	)

	if err != nil {
		// configuring the writer wrong is a programmer error.
		panic(err)
	}

	// NOTE: reader should set max memory bound with WithDecoderMaxMemory.
	//       we can deduce it from maxChunkSize and protect against
	//       malicious inputs.
	zstdReader, err = zstd.NewReader(
		nil,
		zstd.WithDecoderMaxMemory(32*maxChunkSize),
	)
	if err != nil {
		// configuring the reader wrong is a programmer error.
		panic(err)
	}
}

var (
	// AlgoMap is a map of available algorithms.
	algoMap = map[AlgorithmType]func() Algorithm{
		AlgoSnappy: func() Algorithm {
			return snappyAlgo{}
		},
		AlgoLZ4: func() Algorithm {
			// TODO: we could configure compression level here.
			return &lz4Algo{
				compressor: &lz4.Compressor{},
			}
		},
		AlgoZstd: func() Algorithm {
			return zstdAlgo{}
		},
	}

	algoToString = map[AlgorithmType]string{
		AlgoSnappy: "snappy",
		AlgoLZ4:    "lz4",
		AlgoZstd:   "zstd",
	}
)

// AlgoSnappy
func (a snappyAlgo) Encode(dst, src []byte) ([]byte, error) {
	return snappy.Encode(dst, src), nil
}

func (a snappyAlgo) Decode(dst, src []byte) ([]byte, error) {
	return snappy.Decode(dst, src)
}

func (a snappyAlgo) MaxEncodeBufferSize() int {
	return snappy.MaxEncodedLen(maxChunkSize)
}

/////////////////////////

func (a *lz4Algo) Encode(dst, src []byte) ([]byte, error) {
	n, err := a.compressor.CompressBlock(src, dst)
	if err != nil {
		return dst[:n], err
	}

	// NOTE: n == 0 is returned when the data is not easy to compress
	// and the `dst` buf is too small to hold it. Since we always
	// supply a large enough buf this should not happen.
	return dst[:n], nil
}

func (a *lz4Algo) Decode(dst, src []byte) ([]byte, error) {
	n, err := lz4.UncompressBlock(src, dst)
	return dst[:n], err
}

func (a *lz4Algo) MaxEncodeBufferSize() int {
	return lz4.CompressBlockBound(maxChunkSize)
}

/////////////////////////

func (a zstdAlgo) Encode(dst, src []byte) ([]byte, error) {
	return zstdWriter.EncodeAll(src, dst[:0]), nil
}

func (a zstdAlgo) Decode(dst, src []byte) ([]byte, error) {
	return zstdReader.DecodeAll(src, dst[:0])
}

func (a zstdAlgo) MaxEncodeBufferSize() int {
	// TODO: Is there a better way to estimate?
	return maxChunkSize * 2
}

func algorithmFromType(a AlgorithmType) (Algorithm, error) {
	newAlgoFn, ok := algoMap[a]
	if !ok {
		return nil, ErrBadAlgo
	}

	return newAlgoFn(), nil
}
