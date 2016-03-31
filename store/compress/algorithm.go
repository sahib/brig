package compress

import (
	"github.com/golang/snappy"
	"io"
)

type Algorithm interface {
	WrapWriter(w io.Writer) io.Writer
	WrapReader(r io.ReadSeeker) io.Reader
}

type NoneAlgo struct {
}

type SnappyAlgo struct {
}

var AlgoMap = map[AlgorithmType]Algorithm{
	AlgoNone:   &NoneAlgo{},
	AlgoSnappy: &SnappyAlgo{},
}

func (na *NoneAlgo) WrapWriter(w io.Writer) io.Writer {
	return w
}

func (na *NoneAlgo) WrapReader(r io.ReadSeeker) io.Reader {
	return r
}

func (na *SnappyAlgo) WrapWriter(w io.Writer) io.Writer {
	return snappy.NewWriter(w)
}

func (na *SnappyAlgo) WrapReader(r io.ReadSeeker) io.Reader {
	return snappy.NewReader(r)
}

func wrapWriter(w io.Writer, algo AlgorithmType) io.Writer {
	if algoInf, ok := AlgoMap[algo]; ok {
		return algoInf.WrapWriter(w)
	}
	return nil
}

func wrapReader(r io.ReadSeeker, algo AlgorithmType) io.Reader {
	if algoInf, ok := AlgoMap[algo]; ok {
		return algoInf.WrapReader(r)
	}
	return nil
}
