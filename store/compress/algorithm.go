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

var algoMap = map[AlgorithmType]Algorithm{
	AlgoNone:   noneAlgo{},
	AlgoSnappy: snappyAlgo{},
	AlgoLZ4:    lz4Algo{},
}

// AlgoNone
func (a noneAlgo) Encode(src []byte) ([]byte, error) {
	return src, nil
}

func (a noneAlgo) Decode(src []byte) ([]byte, error) {
	return src, nil
}

// AlgoSnappy
func (a snappyAlgo) Encode(src []byte) ([]byte, error) {
	return snappy.Encode(src, src), nil

}

func (a snappyAlgo) Decode(src []byte) ([]byte, error) {
	return snappy.Decode(src, src)
}

// AlgoLZ4
func (a lz4Algo) Encode(src []byte) ([]byte, error) {
	return lz4.Encode(src, src)
}

func (a lz4Algo) Decode(src []byte) ([]byte, error) {
	return lz4.Decode(src, src)
}

// AlgorithmFromType returns a interface to the given AlgorithmType.
func AlgorithmFromType(a AlgorithmType) (Algorithm, error) {
	if algo, ok := algoMap[a]; ok {
		return algo, nil
	}
	return nil, ErrBadAlgo
}
