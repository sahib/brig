package compress

import (
	"errors"

	"github.com/bkaradzic/go-lz4"
	"github.com/golang/snappy"
)

var (
	ErrBadAlgo = errors.New("Invalid algorithm type")
)

type Algorithm interface {
	Encode([]byte) ([]byte, error)
	Decode([]byte) ([]byte, error)
}

type NoneAlgo struct{}
type SnappyAlgo struct{}
type LZ4Algo struct{}

var AlgoMap = map[AlgorithmType]Algorithm{
	AlgoNone:   NoneAlgo{},
	AlgoSnappy: SnappyAlgo{},
	AlgoLZ4:    LZ4Algo{},
}

// AlgoNone
func (_ NoneAlgo) Encode(src []byte) ([]byte, error) {
	return src, nil
}

func (_ NoneAlgo) Decode(src []byte) ([]byte, error) {
	return src, nil
}

// AlgoSnappy
func (_ SnappyAlgo) Encode(src []byte) ([]byte, error) {
	return snappy.Encode(src, src), nil

}

func (_ SnappyAlgo) Decode(src []byte) ([]byte, error) {
	return snappy.Decode(src, src)
}

// AlgoLZ4
func (_ LZ4Algo) Encode(src []byte) ([]byte, error) {
	return lz4.Encode(src, src)
}

func (_ LZ4Algo) Decode(src []byte) ([]byte, error) {
	return lz4.Decode(src, src)
}

func AlgorithmFromType(a AlgorithmType) (Algorithm, error) {
	if algo, ok := AlgoMap[a]; ok {
		return algo, nil
	}
	return nil, ErrBadAlgo
}
