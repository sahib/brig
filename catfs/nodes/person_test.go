package nodes

import (
	"testing"

	h "github.com/disorganizer/brig/util/hashlib"
	"github.com/stretchr/testify/require"
)

func TestPerson(t *testing.T) {
	person := NewPerson("berta", h.TestDummy(t, 1))
	data, err := person.ToBytes()
	if err != nil {
		t.Fatalf("Serializing person to bytes failed: %v", err)
	}

	newPerson, err := PersonFromBytes(data)
	if err != nil {
		t.Fatalf("Reloading person failed: %v", err)
	}

	require.Equal(t, person, newPerson, "person differs")
}
