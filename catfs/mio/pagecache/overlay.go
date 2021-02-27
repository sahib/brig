package overlay

import (
	"bytes"
	"io"
	"sync"

	"github.com/sahib/brig/catfs/mio/pagecache/page"
	"github.com/sahib/brig/util"
)

type PageLayer struct {
	// underlying stream
	rs io.ReadSeeker

	// inode is a unique identifier for the stream.
	// it is used as identifier in the page cache.
	inode int32

	// cache gives access to cached pages
	cache Cache

	// size is the number of bytes that can be read from
	// `rs` from start to end. It represents the "old" file size.
	// It's only use to decide when to stop reading from the
	// underlying stream. For deciding where EOF is, length is used.
	size int64

	// overlayOffset is the last known offset in the stream,
	// including reads from the cache.
	overlayOffset int64

	// streamOffset indicates the offset in the underlying stream `rs`.
	// It can be the same as `overlayOffset` but is not most of the time.
	// Not counted in in `streamOffset` are bytes that were read from
	// the cache exclusively, with no need to read from `rs`.
	streamOffset int64

	// length starts out same as size, but might change due to
	// calls to Truncate(). Truncate is terrible name since it
	// can be also used to extend a file's length. But that's
	// how the underlying syscall is named, so we follow that.
	length int64
}

// NOTE: assumption: `rs` is at read offset zero.
func NewPageLayer(rs io.ReadSeeker, cache Cache, inode, size int64) (*PageLayer, error) {
	if err := cache.Evict(int32(inode)); err != nil {
		return nil, err
	}

	return &PageLayer{
		rs:     rs,
		inode:  int32(inode),
		size:   size,
		length: size,
		cache:  cache,
	}, nil
}

func (l *PageLayer) ensureOffset() error {
	if l.overlayOffset == l.streamOffset {
		return nil
	}

	l.streamOffset = l.overlayOffset
	if _, err := l.rs.Seek(l.overlayOffset, io.SeekStart); err != nil {
		return err
	}

	return nil
}

func (l *PageLayer) WriteAt(buf []byte, off int64) (n int, err error) {
	// If `buf` is large enough to span over several writes then we
	// have to calculate the offset of the first page, so that new
	// data is written to the correct place.
	pageOff := off % page.Size

	// Go over all pages this write affects.
	pageLo := off / page.Size
	pageHi := (off + int64(len(buf))) / page.Size
	for pageIdx := pageLo; pageIdx <= pageHi; pageIdx++ {
		// Overlay the part of `buf` that affects this page
		// and merge with any pre-existing writes.
		if err := l.cache.Merge(
			l.inode,
			int32(pageIdx),
			int32(pageOff),
			buf,
		); err != nil {
			return -1, err
		}

		// starting from the second block the page offset will
		// be always zero. That's only relevant for len(buf) > page.Size.
		pageOff = 0
	}

	// check if this write extended the full buffer.
	// If so we need to remember the new length.
	if newOff := off + int64(len(buf)); newOff > l.length {
		l.length = newOff
	}

	// We always write the full buffer or fail in prior.
	return len(buf), nil
}

var (
	copyBufPool = &sync.Pool{
		New: func() interface{} {
			return make([]byte, page.Size)
		},
	}
)

// TODO: go docs state:
//  * ReadAt() must be allowed to call in parallel.
//    We cannot guarantee that at the moment since sometimes
//    we have to seek the underlying stream - mutex?
func (l *PageLayer) ReadAt(buf []byte, off int64) (int, error) {
	// when requesting reads beyond the size of the overlay,
	// we should immediately cancel the request.
	if off >= l.length {
		return 0, io.EOF
	}

	// small helper for copying data to buf.
	// we will never copy more than page.Size to buf.
	ib := &iobuf{dst: buf}

	// l.rs might not be as long as l.length.
	// We need to pad the rest of the stream with zeros.
	// This reader does this.
	zpr := &zeroPadReader{
		r:      l.rs,
		off:    off,
		size:   l.size,
		length: l.length,
	}

	pageOff := uint32(off % page.Size)

	// keep the copybuf around between GC runs.
	copyBuf := copyBufPool.Get().([]byte)
	defer copyBufPool.Put(copyBuf)

	// Go over all pages this read may affect.
	// We might return early due to io.EOF though.
	pageLo := off / page.Size
	pageHi := (off + int64(len(buf))) / page.Size
	for pageIdx := pageLo; pageIdx <= pageHi; pageIdx++ {
		p, err := l.cache.Lookup(l.inode, int32(pageIdx))
		switch err {
		case page.ErrCacheMiss:
			// we don't have this page cached.
			// need to read it from zpr directly.
			if err := l.ensureOffset(); err != nil {
				return ib.Len(), err
			}

			n, err := copyNBuffer(ib, zpr, int64(ib.Left()), copyBuf)
			if err != nil {
				return ib.Len(), err
			}

			l.overlayOffset += n
			l.streamOffset += n

			// NOTE: we could be clever here and cache pages that have
			//       been read often. We could even hook in things like
			//       fadvise() into this layer.
		case nil:
			// In this case we know that the page is cached.
			// We can fill `buf` with the page of the data,
			// (provided by page.Reader()), but have to watch
			// out for some special cases:
			//
			// - `buf` might be not big enough to hold all of the page.
			//   Therefore ib.Left() caps this number.
			// - This might be the last page and `buf` might be bigger
			//   than the page's contents. This is handled by making
			//   page.Reader() return io.EOF when we would read over
			//   the border.
			// - When reading from cache alone we don't need to seek,
			//   but we have to remember at what position we should
			//   be for the next read and what the current position is.
			//   For this we have l.{overlay,stream}Offset.

			// check how many bytes we can read in total:
			fullLen := util.Min64(
				l.length,
				l.overlayOffset+page.Size,
			) - l.overlayOffset

			occludesStream := p.OccludesStream(pageOff, uint32(fullLen))
			if !occludesStream {
				// only seek if we have to.
				if err := l.ensureOffset(); err != nil {
					return ib.Len(), err
				}

				pageN, err := io.ReadFull(zpr, p.Data[pageOff:])
				if err != nil {
					// TODO: eof and so on?
					return ib.Len(), err
				}

				p.AddExtent(pageOff, p.Data[pageOff:])

				l.streamOffset += int64(pageN)
			}

			r := bytes.NewReader(p.Data[pageOff:])
			n, err := copyNBuffer(ib, r, int64(ib.Left()), copyBuf)
			if err != nil {
				return ib.Len(), err
			}

			l.overlayOffset += n
		default:
			// some other error during cache lookup.
			return ib.Len(), err
		}

		// If read spans over several pages, the second
		// page has to start at zero.
		pageOff = 0
	}

	return ib.Len(), nil
}

func (l *PageLayer) Close() error {
	return l.cache.Close()
}

func (l *PageLayer) Truncate(size int64) {
	l.length = size
}

func (l *PageLayer) Length() int64 {
	return l.length
}
