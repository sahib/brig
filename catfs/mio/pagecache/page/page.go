package page

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	log "github.com/sirupsen/logrus"
)

const (
	// Size is the default size for a page.
	// Last page might be smaller.
	Size = 64 * 1024

	// Ovherhead is the number of bytes we use
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
	// TODO: IDEA: use a fixed or max amount of extents.
	Extents []Extent
	Data    []byte
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

	// TODO: allocate from pool and more than size.
	//       We can use this to make FromBytes and AsBytes efficient.
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

func (p *Page) AddExtent(off int32, write []byte) {
	// TODO: merge magic.
}

func (p *Page) Reader(r io.Reader, pageOff int32) *pageReader {
	return &pageReader{
		page:       p,
		pageOff:    pageOff,
		underlying: r,
	}
}

type pageReader struct {
	page       *Page
	pageOff    int32
	underlying io.Reader
	seekOffset int
}

func (pr *pageReader) Read(buf []byte) (int, error) {
	// TODO: overlay the underlying stream with the specified
	//       extents and write result to buf.
	// TODO: increment seekOffset if we have to read from the underlying
	//       stream. We should probably read all of the underlying page
	//       if we have to. Otherwise next page might have to seek.

	// for _, ex := range pr.page.Extents {
	// }

	return 0, nil
}

func (pr *pageReader) SeekOffset() int {
	return pr.seekOffset
}
