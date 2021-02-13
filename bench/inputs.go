package bench

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"sort"

	"github.com/sahib/brig/util/testutil"
)

// Verifier is a io.Writer that should be used for benchmarks
// that read encoded data. It verifies that the data is actually
// correct in the sense that it is equal to the original input.
type Verifier interface {
	io.Writer

	// MissingBytes returns the diff of bytes to the original input.
	// This number can be negative when too much data was written.
	// Only 0 is a valid value after the benchmark finished.
	MissingBytes() int64
}

// Input generates input for a benchmark. It defines how the data looks that
// is fed to the streaming system.
type Input interface {
	Reader(seed uint64) (io.Reader, error)
	Size() int64
	Verifier() (Verifier, error)
	Close() error
}

func benchData(size uint64, name string) []byte {
	switch name {
	case "random":
		return testutil.CreateRandomDummyBuf(int64(size), 23)
	case "ten":
		return testutil.CreateDummyBuf(int64(size))
	case "mixed":
		return testutil.CreateMixedDummyBuf(int64(size), 42)
	default:
		return nil
	}
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

func newMemInput(size uint64, name string) Input {
	return &memInput{buf: benchData(size, name)}
}

func (ni *memInput) Reader(seed uint64) (io.Reader, error) {
	// Put a few bytes difference at the start to make the complete
	// stream different than the last seed. This is here to avoid
	// that consequent runs of a benchmark get speed ups because
	// they can cache inputs.
	binary.LittleEndian.PutUint64(ni.buf, seed)
	return bytes.NewReader(ni.buf), nil
}

func (ni *memInput) Verifier() (Verifier, error) {
	return &memVerifier{
		expect:  ni.buf,
		counter: 0,
	}, nil
}

func (ni *memInput) Size() int64 {
	return int64(len(ni.buf))
}

func (ni *memInput) Close() error {
	return nil
}

//////////

var (
	inputMap = map[string]func(size uint64) (Input, error){
		"ten": func(size uint64) (Input, error) {
			return newMemInput(size, "ten"), nil
		},
		"random": func(size uint64) (Input, error) {
			return newMemInput(size, "random"), nil
		},
		"mixed": func(size uint64) (Input, error) {
			return newMemInput(size, "mixed"), nil
		},
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
