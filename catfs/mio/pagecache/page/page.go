package page

// NOTE: I had quite often brain freeze while figuring out the indexing.
// If you do too, take a piece of paper and draw it.
// If you don't, congratulations. You're smarter than me.

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"sort"

	log "github.com/sirupsen/logrus"
)

const (
	// Size is the default size for a page.
	// Last page might be smaller.
	Size = 64 * 1024

	// Meta is the number of bytes we use
	// to store the extents of the page.
	// (4k is the typical page size on linux)
	Meta = 4 * 1024

	// ExtentSize needed to store a single extent.
	ExtentSize = 8
)

var (
	// ErrCacheMiss indicates that a page is missing from the cache.
	// Not a real error, but a sentinel to indicate this state.
	ErrCacheMiss = errors.New("cache miss")
)

// Extent marks a single write or
// several writes that were joined to one.
//
// The written data is in the range [lo, hi)
// where hi is not part of the write!
//
// In other words, when writing 16384 bytes
// at OffLo=0, then OffHi=16384, but the last
// valid bytes is at p.Data[OffHi-1]!
//
// This was chosen so you could say p.Data[OffLo:OffHi]
// and it would do what you would guess it would do.
type Extent struct {
	OffLo, OffHi uint32
}

func (e Extent) String() string {
	return fmt.Sprintf("[%d-%d)", e.OffLo, e.OffHi)
}

// Page is a single cached page
type Page struct {
	// Extents is a list describing where
	// `Data` contains valid data.
	Extents []Extent

	// Data is the data hold by the page.
	// It is allocated to Size+Meta bytes,
	// even when no data was used.
	Data []byte
}

func (p *Page) String() string {
	buf := &bytes.Buffer{}
	for idx, extent := range p.Extents {
		buf.WriteString(extent.String())
		if idx+1 != len(p.Extents) {
			buf.WriteString(", ")
		}
	}

	return fmt.Sprintf("<page %p %s>", p.Data, buf.String())
}

// New allocates a new page with an initial extent at `off`
// and with `write` as data. See also Overlay()
func New(off uint32, write []byte) *Page {
	// NOTE: We allocate more than we actually need in order to implement
	// AsBytes and FromBytes efficiently without further allocations.
	backing := make([]byte, Size+Meta)
	p := &Page{Data: backing[:Size]}
	p.Overlay(off, write)
	return p
}

// FromBytes reconstructs a page from the give data.
// Note that ownership over the data is taken, do not write
// to it anymore while using it as a page.
func FromBytes(data []byte) (*Page, error) {
	if len(data) < Size {
		return nil, fmt.Errorf("page data smaller than mandatory size")
	}

	p := Page{Data: data[:Size]}
	extents := data[Size:cap(data)]
	for idx := 0; idx < len(extents); idx += ExtentSize {
		if idx+ExtentSize > len(extents) {
			// sanity check: do not read after extents.
			continue
		}

		offLo := binary.LittleEndian.Uint32(extents[idx+0:])
		offHi := binary.LittleEndian.Uint32(extents[idx+4:])
		if offLo == 0 && offHi == 0 {
			// empty writes are invalid and serve as sentinel value
			// to tell us we read too far. No other extents to expect.
			break
		}

		if offLo == offHi {
			log.Warnf("page cache: loaded empty extent")
			continue
		}

		if offLo > offHi {
			log.Warnf("page cache: loaded invalid extent")
			continue
		}

		p.Extents = append(p.Extents, Extent{
			OffLo: offLo,
			OffHi: offHi,
		})
	}

	return &p, nil
}

// AsBytes encodes the extents at the end of the page data
// and returns the full sized page array.
func (p *Page) AsBytes() []byte {
	if cap(p.Data) < Size+Meta {
		// this is a programming error:
		panic(fmt.Sprintf("bug: page memory was allocated too small %d", cap(p.Data)))
	}

	pdata := p.Data[:Size+Meta]
	pmeta := pdata[Size:]

	for idx, extent := range p.Extents {
		off := idx * ExtentSize
		if off+ExtentSize >= cap(p.Data)-Size {
			// NOTE: This is an inefficient allocation/copy. It will occur only
			// when there are more than $(Meta/ExtentSize) distinct writes
			// without a single read of this page (a non-occluding read will
			// unify all extents). This is pretty unlikely to happen in normal
			// circumstances. If that happens it's a weird use case, so
			// allocate another 64 extents.
			pdata = append(pdata, make([]byte, ExtentSize*64)...)
			p.Data = pdata[:Size]
			pmeta = pdata[Size:cap(pdata)]
		}

		binary.LittleEndian.PutUint32(pmeta[off+0:], extent.OffLo)
		binary.LittleEndian.PutUint32(pmeta[off+4:], extent.OffHi)
	}

	return pdata
}

// affectedExtentIdxs() returns the indices of extents
// that would be affected when writing a new extent with
// the offsets [lo, hi].
//
// Consider the following cases, where "-" are areas
// with existing extents, "_" without and "|" denotes
// the area where we want to write newly. First extent
// is called E1, second E2 and so on.
//
// Case 1: => min=E2, max=E2 (does not hit any extent)
//
// ------__|--|___-------
//
// Case 2: => min=E2, max=E3 (partially hits an extent)
//
// ------__|-------|-----
//
// Case 3: => min=E2, max=E3 (fully inside one extent)
//
// ------________--|---|-
//
// Case 4: => min=len(extents), max=len(extents) (outside any extent)
//
// ------________--------  |-----|
func (p *Page) affectedExtentIdxs(lo, hi uint32) (int, int) {
	minExIdx := sort.Search(len(p.Extents), func(i int) bool {
		return lo < p.Extents[i].OffHi
	})

	maxExIdx := sort.Search(len(p.Extents), func(i int) bool {
		return hi <= p.Extents[i].OffLo
	})

	if minExIdx > maxExIdx {
		// this can happen if lo > hi.
		// (basically a programmer error)
		maxExIdx = minExIdx
	}

	return minExIdx, maxExIdx
}

// OccludesStream will tell you if the page's cached contents
// fully occlude the underlying stream. Or in other words:
// If true, we do not need to read from the underlying stream.
func (p *Page) OccludesStream(pageOff, length uint32) bool {
	l := int64(length)
	minExIdx, maxExIdx := p.affectedExtentIdxs(pageOff, pageOff+length)

	// TODO: Add test for:
	//    pageOff starts in extent
	//    goes to another extent with gap in between.
	for idx := minExIdx; idx < maxExIdx && l > 0; idx++ {
		ex := p.Extents[idx]
		if ex.OffHi < pageOff {
			continue
		}

		if ex.OffLo < pageOff {
			l -= int64(ex.OffHi - pageOff)
			continue
		}

		if idx > 0 && p.Extents[idx-1].OffHi != ex.OffLo {
			// non adjacent; there must be a gap.
			return false
		}

		l -= int64(ex.OffHi - ex.OffLo)
	}

	return l <= 0
}

// Overlay adds newly written data in `write` to the page
// at `off` (relative to the page start!). off + len(write) may not
// exceed the page size! This is a programmer error.
//
// Internally, the data is copied to the page buffer and we keep
// note of the new data in an extent, possibly merging with existing
// ones. This is a relatively fast operation.
func (p *Page) Overlay(off uint32, write []byte) {
	if len(write) == 0 {
		return
	}

	offPlusWrite := off + uint32(len(write))
	if offPlusWrite > uint32(len(p.Data)) {
		// this is a programmer error:
		panic(fmt.Sprintf("extent with write over page bound: %d", offPlusWrite))
	}

	// Copy the data to the requested part of the page.
	// Everything after is maintaining the extents.
	copy(p.Data[off:offPlusWrite], write)
	p.updateExtents(off, offPlusWrite)
}

func (p *Page) updateExtents(off, offPlusWrite uint32) {
	// base case: no extents yet:
	if len(p.Extents) == 0 {
		p.Extents = append(p.Extents, Extent{
			OffLo: off,
			OffHi: offPlusWrite,
		})
		return
	}

	// Find out where to insert the new extent.
	// Use binary search to find a range of extents
	// that are affected by this write.
	minExIdx, maxExIdx := p.affectedExtentIdxs(off, offPlusWrite)

	if minExIdx >= len(p.Extents) {
		// This means that no extent was affected because we wrote beyond any
		// existing extent. Append a new extent to the end of the list.
		p.Extents = append(p.Extents, Extent{
			OffLo: off,
			OffHi: offPlusWrite,
		})
		return
	}

	if minExIdx == maxExIdx {
		// write happens in "free space". No extent hit.
		if minExIdx > 0 && p.Extents[minExIdx-1].OffHi == off {
			// If the write happens to be right after another existing extent
			// then merge with it. Otherwise insert below.
			p.Extents[minExIdx-1].OffHi = offPlusWrite
			return
		}

		if maxExIdx < len(p.Extents) && p.Extents[maxExIdx].OffLo == offPlusWrite {
			// If the write happens to be right before another existing extent
			// then merge with it. Otherwise insert below.
			p.Extents[maxExIdx].OffLo = off
			return
		}

		// insert new extent in the middle of the slice.
		p.Extents = append(p.Extents, Extent{})
		copy(p.Extents[minExIdx+1:], p.Extents[minExIdx:])
		p.Extents[minExIdx] = Extent{
			OffLo: off,
			OffHi: offPlusWrite,
		}

		return
	}

	// Join all affected in the range to one single extent,
	// and move rest of extents further and cut to new size:
	newHi := p.Extents[maxExIdx-1].OffHi
	newLo := p.Extents[minExIdx].OffLo
	if newHi < offPlusWrite {
		newHi = offPlusWrite
	}

	if newLo > off {
		newLo = off
	}

	p.Extents[minExIdx].OffLo = newLo
	p.Extents[minExIdx].OffHi = newHi
	copy(p.Extents[minExIdx+1:], p.Extents[maxExIdx:])
	p.Extents = p.Extents[:len(p.Extents)-(maxExIdx-minExIdx)+1]
}

func minUint32(a, b uint32) uint32 {
	if a < b {
		return a
	}

	return b
}

// Underlay is like the "negative" of Overlay. It writes the data of `write`
// (starting at pageOff) to the underlying buffer where *no* extent is.
// It can be used to "cache" data from the underlying stream, but not
// overwriting any overlay. If OccludesStream() returns true for the same
// offsets, then Underlay() will be an (expensive) no-op.
func (p *Page) Underlay(pageOff uint32, write []byte) {
	pageOffPlusWrite := pageOff + uint32(len(write))
	if pageOff == pageOffPlusWrite {
		// zero underlay.
		return
	}

	cursor := write
	prevOff := pageOff
	for _, ex := range p.Extents {
		if ex.OffHi < pageOff {
			// Extent was before the desired write.
			// No need to consider this one.
			continue
		}

		if ex.OffLo < pageOff {
			// Extent started before pageOff,
			// but goes over it. We should not copy.
			// Instead "loose" the data of that extent.
			cutoff := minUint32(ex.OffHi-pageOff, uint32(len(cursor)))
			cursor = cursor[cutoff:]
			prevOff = ex.OffHi
			continue
		}

		toCopy := ex.OffLo - prevOff
		if toCopy > 0 {
			// Copy everything since last copy
			// to p.Data and jump over the data in cursor.
			copy(p.Data[prevOff:prevOff+toCopy], cursor)
		}

		cursor = cursor[minUint32(toCopy+ex.OffHi-ex.OffLo, uint32(len(cursor))):]
		prevOff = ex.OffHi
	}

	if prevOff < pageOffPlusWrite && len(cursor) > 0 {
		// Handle the case when the underlying write
		// goes beyond all extents or when there are
		// no extents at all.
		toCopy := pageOffPlusWrite - prevOff
		copy(p.Data[prevOff:prevOff+toCopy], cursor)
	}

	p.updateExtents(pageOff, pageOffPlusWrite)
}
