package page

import (
	"testing"

	"github.com/sahib/brig/util/testutil"
	"github.com/stretchr/testify/require"
)

func TestPageAffectedIndices(t *testing.T) {
	e1 := Extent{
		OffLo: 0,
		OffHi: Size / 4,
	}

	e2 := Extent{
		OffLo: 2 * Size / 4,
		OffHi: 3 * Size / 4,
	}

	e3 := Extent{
		OffLo: 3 * Size / 4,
		OffHi: 4 * Size / 4,
	}

	p := Page{
		Data:    make([]byte, Size+Meta),
		Extents: []Extent{e1, e2, e3},
	}

	for idx := uint32(1); idx < e1.OffHi; idx++ {
		l, h := p.affectedExtentIdxs(0, uint32(idx))

		// result means: affects extents[0] until
		// (excluding) extents[1]
		require.Equal(t, 0, l, idx)
		require.Equal(t, 1, h, idx)
	}

	for idx := e1.OffHi; idx < e2.OffLo; idx++ {
		l, h := p.affectedExtentIdxs(e1.OffHi, uint32(idx))

		// result means: no extent was found, the array
		// of affected extents is empty.
		require.Equal(t, 1, l, idx)
		require.Equal(t, 1, h, idx)
	}

	for idx := e2.OffLo + 1; idx < e3.OffLo; idx++ {
		l, h := p.affectedExtentIdxs(idx, idx+1)

		// result means: no extent was found, the array
		// of affected extents is empty.
		require.Equal(t, 1, l, idx)
		require.Equal(t, 2, h, idx)
	}

	// No for loop needed for last case:
	l, h := p.affectedExtentIdxs(Size, Size+1)
	require.Equal(t, 3, l)
	require.Equal(t, 3, h)
}

func TestPageSerializeDeserialize(t *testing.T) {
	expected := New(0, testutil.CreateDummyBuf(Size))

	data := expected.AsBytes()
	got, err := FromBytes(data)
	require.NoError(t, err)

	require.Equal(t, expected.Data, got.Data)
	require.Equal(t, expected.Extents, got.Extents)
}

func TestPageSerializeWithManyWrites(t *testing.T) {
	// Simulate a really pathological case where we have tons of small writes.
	// They will cause AsBytes() to increase the backing buffer when
	// serializing. This will be a performance hit, but is at least correct.
	expected := New(0, testutil.CreateDummyBuf(10))
	for idx := 0; idx < 1024; idx++ {
		off := uint32(11 + (idx * 11))
		expected.Overlay(off, testutil.CreateDummyBuf(10))
	}

	// 1025: initial extent existed:
	require.Len(t, expected.Extents, 1025)

	data := expected.AsBytes()
	got, err := FromBytes(data)
	require.NoError(t, err)

	require.Equal(t, expected.Data, got.Data)
	require.Equal(t, expected.Extents, got.Extents)
}

func TestPageOccludeStream(t *testing.T) {
	// page with one extent:
	p := New(0, testutil.CreateDummyBuf(Size/4))

	require.False(t, p.OccludesStream(0, Size))
	require.False(t, p.OccludesStream(0, Size/4+1))
	require.True(t, p.OccludesStream(0, Size/4))
	require.True(t, p.OccludesStream(0, Size/4-1))

	p.Overlay(2*Size/4, testutil.CreateDummyBuf(Size/4))

	require.False(t, p.OccludesStream(2*Size/4, Size/4+1))
	require.False(t, p.OccludesStream(0, 3*Size/4+1))
	require.True(t, p.OccludesStream(2*Size/4, Size/4))
	require.True(t, p.OccludesStream(2*Size/4, Size/4-1))
}

func TestPageAddExtent(t *testing.T) {
	// page with one extent:
	p := New(0, testutil.CreateDummyBuf(Size/4))

	// This matches the extents in TestPageAffectedIndices:
	// (first extent touches existing one!)
	p.Overlay(100, testutil.CreateDummyBuf(Size/4))
	p.Overlay(2*Size/4, testutil.CreateDummyBuf(Size/4))
	p.Overlay(3*Size/4, testutil.CreateDummyBuf(Size/4))

	require.Len(t, p.Extents, 3)
	require.Equal(
		t,
		[]Extent{{0, Size/4 + 100}, {2 * Size / 4, 3 * Size / 4}, {3 * Size / 4, Size}},
		p.Extents,
	)

	require.Panics(t, func() {
		// Write beyond the extents:
		p.Overlay(Size, testutil.CreateDummyBuf(1))
	})

	// Write an extent in free space, right after another one:
	// It should detect this and merge with it.
	p.Overlay(Size/4+100, testutil.CreateDummyBuf(20))
	require.Len(t, p.Extents, 3)
	require.Equal(t, Extent{0, Size/4 + 120}, p.Extents[0])

	// Write an extent in free space (not adjacent):
	p.Overlay(Size/4+200, testutil.CreateDummyBuf(30))
	require.Len(t, p.Extents, 4)
	require.Equal(t, Extent{Size/4 + 200, Size/4 + 230}, p.Extents[1])

	// Write an extent that covers everything,
	// should reduce to a single one:
	p.Overlay(0, testutil.CreateDummyBuf(Size))
	require.Len(t, p.Extents, 1)
	require.Equal(t, Extent{0, Size}, p.Extents[0])
	require.Equal(t, testutil.CreateDummyBuf(Size), p.Data)

	// Try to add an empty extent, it should not do anything.
	p.Overlay(0, []byte{})
	require.Len(t, p.Extents, 1)
}

func TestPageAddExtentRegression(t *testing.T) {
	// I forgot to adjust the lower extent bound:
	p := New(3*Size/4, testutil.CreateDummyBuf(Size/4))
	p.Overlay(0, testutil.CreateDummyBuf(Size))
	require.Len(t, p.Extents, 1)
	require.Equal(t, uint32(0), p.Extents[0].OffLo)
	require.Equal(t, uint32(Size), p.Extents[0].OffHi)
}

func TestPageUnderlayFull(t *testing.T) {
	underlay := testutil.CreateRandomDummyBuf(Size, 23)
	overlay := testutil.CreateDummyBuf(Size / 4)

	p := New(2*Size/4, overlay)
	p.Underlay(0, underlay)

	copy(underlay[2*Size/4:], overlay)
	require.Equal(t, len(underlay), len(p.Data))
	require.Equal(t, underlay, p.Data)
}

func TestPageUnderlayPartial(t *testing.T) {
	underlay := testutil.CreateRandomDummyBuf(2*Size/4, 23)
	p := New(2*Size/4, testutil.CreateDummyBuf(Size/4))

	// That overlay should be ignored:
	p.Overlay(Size/16, testutil.CreateDummyBuf(Size/32))

	// This overlay shadows the underlay:
	p.Overlay(Size/8, testutil.CreateDummyBuf(Size/4))

	// Now underlay it. Should only write things
	// between 3*Size/8 and Size/2, everything else is shadowed.
	p.Underlay(Size/4, underlay)

	// Construct our expectation:
	expected := make([]byte, Size)
	copy(expected[Size/4:], underlay)
	copy(expected[Size/16:], testutil.CreateDummyBuf(Size/32))
	copy(expected[Size/8:], testutil.CreateDummyBuf(Size/4))
	copy(expected[2*Size/4:], testutil.CreateDummyBuf(Size/4))

	require.Equal(t, expected, p.Data)
}

func TestPageUnderlayLeftover(t *testing.T) {
	underlay := testutil.CreateRandomDummyBuf(1*Size/4, 23)
	overlay := testutil.CreateDummyBuf(3 * Size / 4)
	p := New(0, overlay)

	// should do nothing!
	p.Underlay(0, underlay)

	// Construct our expectation:
	expected := make([]byte, Size)
	copy(expected, overlay)
	require.Equal(t, expected, p.Data)
}
