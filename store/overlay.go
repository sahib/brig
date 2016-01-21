package store

import "sort"

// Interval represents a 2D range of integers.
type Interval interface {
	// Range returns the minimum and maximum of the interval.
	// Minimum value is inclusive, maximum value exclusive.
	// In other notation: [min, max)
	Range() (int, int)

	// Merge merges the interval `i` to this interval.
	// The range borders should be fixed accordingly,
	// so that [min(i.min, self.min), max(i.max, self.max)] applies.
	Merge(i Interval)
}

// Modification represents a single write
type Modification struct {
	// Offset where the modification started:
	offset int

	// Data that was changed:
	// This might be changed to a mmap'd byte slice later.
	data []byte
}

// Range returns the fitting integer interval
func (n *Modification) Range() (int, int) {
	return n.offset, n.offset + len(n.data)
}

// Merge adds the data of another interval where they intersect.
// The overlapping parts are taken from `n` always.
// Note: `i` shall not be used after calling Merge.
func (n *Modification) Merge(i Interval) {
	// Interracial merges are forbidden :-(
	other, ok := i.(*Modification)
	if !ok {
		return
	}

	oMin, oMax := other.Range()
	nMin, nMax := n.Range()

	// Check if the intervals overlap.
	// If not, there's nothing left to do.
	if nMin > oMax || oMin > nMax {
		return
	}

	// Prepend non-overlapping data from `other`:
	if nMin > oMin {
		n.data = append(other.data[:(nMin-oMin)], n.data...)
		n.offset = other.offset
	}

	// Append non-overlapping data from `other`:
	if nMax < oMax {
		// Append other.data[(other.Max - n.Max):]
		n.data = append(n.data, other.data[(oMax-nMax-1):]...)
	}

	// Make sure old data gets invalidated quickly:
	other.data = nil
}

// IntervalIndex represents a continous array of sorted intervals.
// When adding intervals to the index, it will merge them overlapping areas.
// Holes between the intervals are allowed.
type IntervalIndex struct {
	r []Interval
}

// cut deletes the a[i:j] from a and returns the new slice.
func cut(a []Interval, i, j int) []Interval {
	copy(a[i:], a[j:])
	for k, n := len(a)-j+i, len(a); k < n; k++ {
		a[k] = nil
	}
	return a[:len(a)-j+i]
}

// insert squeezes `x` at a[j] and moves the reminding elements.
// Returns the modified slice.
func insert(a []Interval, i int, x Interval) []Interval {
	a = append(a, nil)
	copy(a[i+1:], a[i:])
	a[i] = x
	return a
}

// Add inserts a single interval to the index.
// If it overlaps with existing intervals, it's data will
// take priority over other intervals.
func (ivl *IntervalIndex) Add(n Interval) {
	Min, Max := n.Range()
	if Max < Min {
		panic("Max > Min!")
	}

	// Initial case: Add as single element.
	if ivl.r == nil {
		ivl.r = []Interval{n}
		return
	}

	// Finding the lowest fitting interval:
	minIdx := sort.Search(len(ivl.r), func(i int) bool {
		_, iMax := ivl.r[i].Range()
		return Min <= iMax
	})

	maxIdx := sort.Search(len(ivl.r), func(i int) bool {
		iMin, _ := ivl.r[i].Range()
		return Max <= iMin
	})

	// New interval is bigger than all others:
	if minIdx >= len(ivl.r) {
		ivl.r = append(ivl.r, n)
		return
	}

	// New range fits nicely in; just insert it inbetween:
	if minIdx == maxIdx {
		ivl.r = insert(ivl.r, minIdx, n)
		return
	}

	// Something in between. Merge to continuous interval:
	for i := minIdx; i < maxIdx; i++ {
		n.Merge(ivl.r[i])
	}

	// Delete old unmerged intervals and substitute with merged:
	ivl.r[minIdx] = n
	ivl.r = cut(ivl.r, minIdx+1, maxIdx)
}
