package hashlib

import (
	"testing"
)

func TestHashWriter(t *testing.T) {
	data := []byte{1, 2, 3, 4}

	hw1 := NewHashWriter()
	hw1.Write(data[0:2])
	hw1.Write(data[2:4])

	hw2 := NewHashWriter()
	hw2.Write(data[0:3])
	hw2.Write(data[3:4])

	// The hashes should be the same, even though the order in which
	// we feed the data is different. This forbids using things like XOR
	// for combining hashes of blocks.
	if !hw1.Finalize().Equal(hw2.Finalize()) {
		t.Fatalf("hashes differ due to different feed order")
	}
}
