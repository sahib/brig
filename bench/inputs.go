package bench

import (
	"bytes"
	"fmt"
	"io"
	"sort"

	"github.com/sahib/brig/util/testutil"
)

// Input generates input for a benchmark. It defines how the data looks that
// is fed to the streaming system.
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

type memInput struct {
	buf []byte
}

func newMemInput(size uint64, isRandom bool) Input {
	return &memInput{buf: benchData(size, isRandom)}
}

func (ni *memInput) Reader() (io.Reader, error) {
	return bytes.NewReader(ni.buf), nil
}

func (ni *memInput) Close() error {
	return nil
}

//////////

var (
	inputMap = map[string]func(size uint64) (Input, error){
		"ten":    func(size uint64) (Input, error) { return newMemInput(size, false), nil },
		"random": func(size uint64) (Input, error) { return newMemInput(size, true), nil },
	}
)

// InputByName fetches the input by it's name and returns an input
// that will produce data with `size` bytes.
func InputByName(name string, size uint64) (Input, error) {
	newInput, ok := inputMap[name]
	if !ok {
		return nil, fmt.Errorf("no such input: %s", name)
	}

	return newInput(size)
}

// InputNames returns the sorted list of all possible inputs.
func InputNames() []string {
	names := []string{}
	for name := range inputMap {
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}
