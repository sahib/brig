package bench

import (
	"bytes"
	"fmt"
	"io"

	"github.com/sahib/brig/util/testutil"
)

type Input interface {
	Reader() (io.Reader, error)
	Close() error
}

//////////

type NullInput struct {
	buf []byte
}

func NewNullInput(size uint64, isRandom bool) *NullInput {
	var buf []byte

	if isRandom {
		buf = testutil.CreateRandomDummyBuf(int64(size), 23)
	} else {
		buf = testutil.CreateDummyBuf(int64(size))
	}

	return &NullInput{buf: buf}
}

func (ni *NullInput) Reader() (io.Reader, error) {
	return bytes.NewReader(ni.buf), nil
}

func (ni *NullInput) Close() error {
	return nil
}

//////////

func InputByName(name string, size uint64, isRandom bool) (Input, error) {
	switch name {
	case "null":
		return NewNullInput(size, isRandom), nil
	default:
		return nil, fmt.Errorf("no such input: %s", name)
	}
}
