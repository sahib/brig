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

func benchData(size uint64, isRandom bool) []byte {
	if isRandom {
		return testutil.CreateRandomDummyBuf(int64(size), 23)
	}
	return testutil.CreateDummyBuf(int64(size))
}

//////////

type MemInput struct {
	buf []byte
}

func NewMemInput(size uint64, isRandom bool) *MemInput {
	return &MemInput{buf: benchData(size, isRandom)}
}

func (ni *MemInput) Reader() (io.Reader, error) {
	return bytes.NewReader(ni.buf), nil
}

func (ni *MemInput) Close() error {
	return nil
}

//////////

func InputByName(name string, size uint64) (Input, error) {
	switch name {
	case "ten":
		return NewMemInput(size, false), nil
	case "random":
		return NewMemInput(size, true), nil
	default:
		return nil, fmt.Errorf("no such input: %s", name)
	}
}
