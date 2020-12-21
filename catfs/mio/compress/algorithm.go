package compress

import (
	"errors"

	"github.com/bkaradzic/go-lz4"
	"github.com/golang/snappy"
)

var (
	// ErrBadAlgo is returned on a unsupported/unknown algorithm.
	ErrBadAlgo = errors.New("Invalid algorithm type")
)

// Algorithm is the common interface for all supported algorithms.
type Algorithm interface {
	Encode([]byte) ([]byte, error)
	Decode([]byte) ([]byte, error)
}

type noneAlgo struct{}
type snappyAlgo struct{}
type lz4Algo struct{}
type zstdAlgo struct{}

var (
	// AlgoMap is a map of available algorithms.
	AlgoMap = map[AlgorithmType]Algorithm{
		AlgoNone:   noneAlgo{},
		AlgoSnappy: snappyAlgo{},
		AlgoLZ4:    lz4Algo{},
	}

	algoToString = map[AlgorithmType]string{
		AlgoNone:   "none",
		AlgoSnappy: "snappy",
		AlgoLZ4:    "lz4",
	}

	stringToAlgo = map[string]AlgorithmType{
		"none":   AlgoNone,
		"snappy": AlgoSnappy,
		"lz4":    AlgoLZ4,
	}
)

// AlgoNone
func (a noneAlgo) Encode(src []byte) ([]byte, error) {
	return src, nil
}

func (a noneAlgo) Decode(src []byte) ([]byte, error) {
	return src, nil
}

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

// AlgoZstd
func (a zstdAlgo) Encode(src []byte) ([]byte, error) {
	return lz4.Encode(nil, src)
}

func (a zstdAlgo) Decode(src []byte) ([]byte, error) {
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
