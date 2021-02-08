package bench

import (
	"bytes"
	"fmt"
	"io"
	"sort"

	"github.com/sahib/brig/util/testutil"
)

type Verifier interface {
	io.Writer
	MissingBytes() int64
}

// Input generates input for a benchmark. It defines how the data looks that
// is fed to the streaming system.
type Input interface {
	Reader() (io.Reader, error)
	Verifier() (Verifier, error)
	Close() error
}

func benchData(size uint64, isRandom bool) []byte {
	if isRandom {
		return testutil.CreateRandomDummyBuf(int64(size), 23)
	}
	return testutil.CreateDummyBuf(int64(size))
}

//////////

type memVerifier struct {
	expect  []byte
	counter int64
}

func (m *memVerifier) Write(buf []byte) (int, error) {
	if int64(len(buf))+m.counter > int64(len(m.expect)) {
		return -1, fmt.Errorf("verify: got too much data")
	}

	slice := m.expect[m.counter : m.counter+int64(len(buf))]
	if !bytes.Equal(slice, buf) {
		return -1, fmt.Errorf("verify: data differs in block at %d", m.counter)
	}

	m.counter += int64(len(buf))

	// Just nod off the data and let GC do the rest.
	return len(buf), nil
}

func (m *memVerifier) MissingBytes() int64 {
	return int64(len(m.expect)) - m.counter
}

type memInput struct {
	buf []byte
}

func newMemInput(size uint64, isRandom bool) Input {
	return &memInput{buf: benchData(size, isRandom)}
}

func (ni *memInput) Reader() (io.Reader, error) {
	return bytes.NewReader(ni.buf), nil
}

func (ni *memInput) Verifier() (Verifier, error) {
	return &memVerifier{
		expect:  ni.buf,
		counter: 0,
	}, nil
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
