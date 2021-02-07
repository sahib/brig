package bench

import (
	"bytes"
	"fmt"
	"io"
	"sort"

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

func NewMemInput(size uint64, isRandom bool) Input {
	return &MemInput{buf: benchData(size, isRandom)}
}

func (ni *MemInput) Reader() (io.Reader, error) {
	return bytes.NewReader(ni.buf), nil
}

func (ni *MemInput) Close() error {
	return nil
}

//////////

var (
	inputMap = map[string]func(size uint64) (Input, error){
		"ten":    func(size uint64) (Input, error) { return NewMemInput(size, false), nil },
		"random": func(size uint64) (Input, error) { return NewMemInput(size, true), nil },
	}
)

func InputByName(name string, size uint64) (Input, error) {
	newInput, ok := inputMap[name]
	if !ok {
		return nil, fmt.Errorf("no such input: %s", name)
	}

	return newInput(size)
}

func InputNames() []string {
	names := []string{}
	for name := range inputMap {
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}
