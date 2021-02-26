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

	// Overhead is the number of bytes we use
	// to store the extents of the page.
	Meta = 6 * 1024

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
	OffLo, OffHi int32
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

func New(off int32, write []byte) *Page {
	var extents []Extent
	if len(write) != Size {
		// special rule: if write fully occludes
		extents = append(extents, Extent{
			OffLo: off,
			OffHi: off + int32(len(write)),
		})
	}

	// NOTE: We allocate more than we actually need
	// in order to implement AsBytes and FromBytes
	// efficiently.
	backing := make([]byte, Size+Meta)
	return &Page{
		Data:    backing[:Size],
		Extents: extents,
	}
}

func FromBytes(data []byte) (*Page, error) {
	// TODO: Implement.
	// 		 We need to avoid extra allocations here!
	return nil, nil
}

// OccludesStream will tell you if the page's cached contents
// fully occlude the underlying stream. Or in other words:
// If true, we do not need to read from the underlying stream.
func (p *Page) OccludesStream(pageOff int32, length int32) bool {
	if len(p.Extents) == 0 {
		return true
	}

	l := length
	for _, extent := range p.Extents {
		l -= int32(extent.OffHi - extent.OffLo)
	}

	if l < 0 {
		log.Warnf("bug: extents of inode sum up < 0")
	}

	return l == 0
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
			// TODO: What to do in this case?
			// allocate more memory or error out?
			break
		}

		binary.LittleEndian.PutUint32(pmeta[off:off+4], uint32(extent.OffLo))
		binary.LittleEndian.PutUint32(pmeta[off:off+4], uint32(extent.OffHi))
	}

	return pdata
}

// AddExtent adds newly written data in `write` to the page
// at `off` (relative to the page start!). off + len(write) may not
// exceed the page size! This is a programmer error.
//
// Internally, the data is copied to the page buffer and we keep
// note of the new data in an extent, possibly merging with existing
// ones. This is a relatively fast operation.
func (p *Page) AddExtent(off int32, write []byte) {
	offPlusWrite := off + int32(len(write))

	if offPlusWrite > int32(len(p.Data)) {
		// TODO: panic here? This would mean that at least a part
		//       of the data is written over the page bound
		//       and would be lost, indicating a bug elsehwere.
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

	minExIdx := sort.Search(len(p.Extents), func(i int) bool {
		return off <= p.Extents[i].OffHi
	})

	maxExIdx := minExIdx + sort.Search(len(p.Extents[minExIdx:]), func(i int) bool {
		return offPlusWrite <= p.Extents[i].OffLo
	})

	if minExIdx == maxExIdx {
		// The write happens inside a single extent.
		// Borders do not need to be adjusted.
		return
	}

	if minExIdx > len(p.Extents) {
		// This means that no extent was affected because we wrote beyond any
		// existing extent. Append a new extent to the end of the list.

		// TODO: Case to append new extent at end.
		// Possibly join last index if adjacent?

		p.Extents = append(p.Extents, Extent{
			OffLo: off,
			OffHi: offPlusWrite,
		})
		return
	}

	// Join all affected in the range to one single extent,
	// and move rest of extents further and cut to new size:
	p.Extents[minExIdx].OffHi = p.Extents[maxExIdx].OffHi
	copy(p.Extents[minExIdx+1:], p.Extents[maxExIdx:])
	p.Extents = p.Extents[:len(p.Extents)-(maxExIdx-minExIdx)]
}
