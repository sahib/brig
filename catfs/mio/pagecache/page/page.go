package page

import (
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

	// Size needed to store a single extent.
	ExtentSize = 8
)

var (
	// not a real error, do not pass to outside.
	ErrCacheMiss = errors.New("cache miss")
)

// Extent marks a single write or
// several writes that were joined to one.
type Extent struct {
	OffLo, OffHi uint32
}

type Page struct {
	// Extents is a list describing where
	// `Data` contains valid data.
	Extents []Extent

	// Data is the data hold by the page.
	// It is allocated to Size+Meta bytes,
	// even when no data was used.
	Data []byte
}

func New(off uint32, write []byte) *Page {
	var extents []Extent
	if len(write) != Size {
		// special rule: if write fully occludes
		extents = append(extents, Extent{
			OffLo: off,
			OffHi: off + uint32(len(write)),
		})
	}

	// NOTE: We allocate more than we actually need in order to implement
	// AsBytes and FromBytes efficiently without further allocations.
	backing := make([]byte, Size+Meta)
	return &Page{
		Data:    backing[:Size],
		Extents: extents,
	}
}

func FromBytes(data []byte) (*Page, error) {
	if len(data) < Size {
		return nil, fmt.Errorf("page data smaller than mandatory size")
	}

	p := Page{Data: data[:Size]}

	extents := data[Size:]
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

func (p *Page) AsBytes() []byte {
	if cap(p.Data) < Size+Meta {
		// this is a programming error:
		panic(fmt.Sprintf("bug: page memory was allocated too small %d", cap(p.Data)))
	}

	pdata := p.Data[:Size+Meta]
	pmeta := pdata[Size:]

	for idx, extent := range p.Extents {
		off := idx * ExtentSize
		if off > Meta+ExtentSize {
			// NOTE: This is an inefficient allocation. It will occur only when
			// there are more than $(Meta/ExtentSize) distinct writes without a
			// single read of this page (a non-occluding read will unify all
			// extents). This is pretty unlikely to happen in normal
			// circumstances. If that happens it's a weird use case, so
			// allocate another 64 extents.
			pdata = append(pdata, make([]byte, ExtentSize*64)...)
			p.Data = pdata[:Size]
			pmeta = pdata[Size:]
		}

		binary.LittleEndian.PutUint32(pmeta[off+0:], uint32(extent.OffLo))
		binary.LittleEndian.PutUint32(pmeta[off+4:], uint32(extent.OffHi))
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
		return lo <= p.Extents[i].OffHi
	})

	maxExIdx := minExIdx + sort.Search(len(p.Extents[minExIdx:]), func(i int) bool {
		return hi <= p.Extents[i].OffLo
	})

	return minExIdx, maxExIdx
}

// OccludesStream will tell you if the page's cached contents
// fully occlude the underlying stream. Or in other words:
// If true, we do not need to read from the underlying stream.
func (p *Page) OccludesStream(pageOff, length uint32) bool {
	l := length
	minExIdx, maxExIdx := p.affectedExtentIdxs(pageOff, length)

	for idx := minExIdx; idx < maxExIdx && l > 0; idx++ {
		ex := p.Extents[idx]
		if idx > minExIdx && p.Extents[idx-1].OffHi != ex.OffLo {
			// non adjacent; there must be a gap.
			return false
		}

		l -= ex.OffHi - ex.OffLo
	}

	return l <= 0
}

// AddExtent adds newly written data in `write` to the page
// at `off` (relative to the page start!). off + len(write) may not
// exceed the page size! This is a programmer error.
//
// Internally, the data is copied to the page buffer and we keep
// note of the new data in an extent, possibly merging with existing
// ones. This is a relatively fast operation.
func (p *Page) AddExtent(off uint32, write []byte) {
	offPlusWrite := off + uint32(len(write))

	if offPlusWrite > uint32(len(p.Data)) {
		// this is a programmer error:
		panic(fmt.Sprintf("extent with write over page bound: %d", offPlusWrite))
	}

	// Copy the data to the requested part of the page.
	// Everything after is maintaining the extents.
	copy(p.Data[off:], write)

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
	p.Extents[minExIdx].OffHi = p.Extents[maxExIdx].OffHi
	copy(p.Extents[minExIdx+1:], p.Extents[maxExIdx:])
	p.Extents = p.Extents[:len(p.Extents)-(maxExIdx-minExIdx)]
}
