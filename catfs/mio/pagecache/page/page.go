package page

import (
	"errors"
	"io"

	log "github.com/sirupsen/logrus"
)

const (
	// Size is the default size for a page.
	// Last page might be smaller.
	Size = 64 * 1024
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

// OccludesStream will tell you if the page's cached contents
// fully occlude the underlying stream. Or in other words:
// If true, we do not need to read from the underlying stream.
func (p *Page) OccludesStream(length int32) bool {
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
	// TODO: Implement.
	// 		 We need to avoid extra allocations here!
	return nil
}

func FromBytes(data []byte) (*Page, error) {
	// TODO: Implement.
	// 		 We need to avoid extra allocations here!
	return nil, nil
}

func (p *Page) Reader(r io.Reader) *pageReader {
	return &pageReader{
		page:   p,
		reader: r,
	}
}

type pageReader struct {
	page       *Page
	reader     io.Reader
	seekOffset int
}

func (pr *pageReader) Read(buf []byte) (int, error) {
	// TODO: overlay the underlying stream with the specified
	//       extents and write result to buf.
	// TODO: increment seekOffset if we have to read from the underlying
	//       stream. We should probably read all of the underlying page
	//       if we have to. Otherwise next page might have to seek.
	return 0, nil
}

func (pr *pageReader) SeekOffset() int {
	return pr.seekOffset
}
