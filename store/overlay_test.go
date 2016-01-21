package store

import (
	"fmt"
	"testing"
)

func makeMod(off, size int) *Modification {
	s := make([]byte, size)
	for i := 0; i < size; i++ {
		s[i] = byte(off + i)
	}

	return &Modification{off, s}
}

func TestMerge(t *testing.T) {
	i := &IntervalIndex{}
	i.Add(makeMod(0, 10))
	i.Add(makeMod(15, 5))
	i.Add(makeMod(10, 5))
	i.Add(makeMod(9, 8))
	i.Add(makeMod(90, 8))

	check := func(m *Modification, lo, hi int) bool {
		if int(m.data[0]) != m.offset {
			t.Errorf("Offset and first element do not match.")
			t.Errorf("Off: %v Data: %v", m.offset, m.data)
			return false
		}

		for i := lo; i < hi; i++ {
			if int(m.data[i-lo]) != i {
				t.Errorf("Merge hickup: %v != %v", m.data[i-lo], i)
				return false
			}
		}

		return true
	}

	// First three intervals should be merged to one:
	if !check(i.r[0].(*Modification), 0, 20) {
		return
	}

	// Last one should be totally untouched:
	if !check(i.r[1].(*Modification), 90, 98) {
		return
	}
}
