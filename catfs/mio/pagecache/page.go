package overlay

import (
	"io"

	log "github.com/sirupsen/logrus"
)

// Extent marks a single write or
// several writes that were joined to one.
type Extent struct {
	offLo, offHi int32
}

type Page struct {
	Data    []byte
	Length  int32
	Extents []Extent
}

func (p *Page) OccludesStream() bool {
	l := p.Length

	for _, extent := range p.Extents {
		l -= int32(extent.offHi - extent.offLo)
	}

	if l < 0 {
		log.Warnf("bug: extents of inode sum up < 0")
	}

	return l == 0
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
