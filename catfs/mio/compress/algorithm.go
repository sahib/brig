package compress

import (
	"errors"

	"github.com/bkaradzic/go-lz4"
	"github.com/golang/snappy"
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
)

// AlgorithmType user defined type to store the algorithm type.
type AlgorithmType byte

// IsValid returns true if `at` is a valid algorithm type.
func (at AlgorithmType) IsValid() bool {
	switch at {
	case AlgoSnappy, AlgoLZ4:
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
	Encode([]byte) ([]byte, error)
	Decode([]byte) ([]byte, error)
}

type snappyAlgo struct{}
type lz4Algo struct{}

var (
	// AlgoMap is a map of available algorithms.
	AlgoMap = map[AlgorithmType]Algorithm{
		AlgoSnappy: snappyAlgo{},
		AlgoLZ4:    lz4Algo{},
	}

	// TODO: still needded?
	algoToString = map[AlgorithmType]string{
		AlgoSnappy: "snappy",
		AlgoLZ4:    "lz4",
	}

	// TODO: still needed?
	stringToAlgo = map[string]AlgorithmType{
		"snappy": AlgoSnappy,
		"lz4":    AlgoLZ4,
	}
)

// AlgoSnappy
func (a snappyAlgo) Encode(src []byte) ([]byte, error) {
	return snappy.Encode(nil, src), nil

}

func (a snappyAlgo) Decode(src []byte) ([]byte, error) {
	return snappy.Decode(nil, src)
}

// AlgoLZ4
func (a lz4Algo) Encode(src []byte) ([]byte, error) {
	return lz4.Encode(nil, src)
}

func (a lz4Algo) Decode(src []byte) ([]byte, error) {
	return lz4.Decode(nil, src)
}

// AlgorithmFromType returns a interface to the given AlgorithmType.
func AlgorithmFromType(a AlgorithmType) (Algorithm, error) {
	if algo, ok := AlgoMap[a]; ok {
		return algo, nil
	}
	return nil, ErrBadAlgo
}

// AlgoToString converts a algorithm type to a string.
func AlgoToString(a AlgorithmType) string {
	algo, ok := algoToString[a]
	if !ok {
		return "unknown algorithm"
	}
	return algo
}

// AlgoFromString tries to convert a string to AlgorithmType
func AlgoFromString(s string) (AlgorithmType, error) {
	algoType, ok := stringToAlgo[s]
	if !ok {
		return 0, errors.New("Invalid algorithm name")
	}
	return algoType, nil
}
