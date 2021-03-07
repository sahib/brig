package pagecache

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	"github.com/sahib/brig/catfs/mio/pagecache/page"
	"github.com/sahib/brig/util"
)

// Layer is a layer above a read-only stream with write support.
type Layer struct {
	// underlying stream
	rs io.ReadSeeker

	// inode is a unique identifier for the stream.
	// it is used as identifier in the page cache.
	inode int64

	// cache gives access to cached pages
	cache Cache

	// size is the number of bytes that can be read from
	// `rs` from start to end. It represents the "old" file size.
	// It's only use to decide when to stop reading from the
	// underlying stream. For deciding where EOF is, length is used.
	size int64

	// length starts out same as size, but might change due to
	// calls to Truncate(). Truncate is terrible name since it
	// can be also used to extend a file's length. But that's
	// how the underlying syscall is named, so we follow that.
	length int64

	// overlayOffset is the last known offset in the stream,
	// including reads from the cache. It is the position in the
	// overlayed stream.
	overlayOffset int64

	// streamOffset indicates the offset in the underlying stream `rs`.
	// It can be the same as `overlayOffset` but is not most of the time.
	// Not counted in in `streamOffset` are bytes that were read from
	// the cache exclusively, with no need to read from `rs`.
	// It's not updated when data is purely read from the cache.
	streamOffset int64
}

// NewLayer returns a paged overlay for `rs`, reading and storing data from
// `cache`. `inode` will be used as cache identifier for this file. The only
// need is that it is unique to this file, otherwise it does not need any
// inode-like semantics. `size` must be known in advance and reflects the size
// of `rs`. This cannot be used for pure streaming. `rs` is assumed to be positioned
// at the zero offset. If not, subtract the offset from `size`.
func NewLayer(rs io.ReadSeeker, cache Cache, inode, size int64) (*Layer, error) {
	if err := cache.Evict(inode, size); err != nil {
		return nil, err
	}

	return &Layer{
		rs:     rs,
		inode:  inode,
		size:   size,
		length: size,
		cache:  cache,
	}, nil
}

func (l *Layer) ensureOffset(zpr *zeroPadReader) error {
	if l.overlayOffset == l.streamOffset {
		return nil
	}

	l.streamOffset = l.overlayOffset
	zpr.off = l.overlayOffset
	if _, err := l.rs.Seek(l.overlayOffset, io.SeekStart); err != nil {
		return err
	}

	return nil
}

// WriteAt writes `buf` to `off`. It will appear on the next
// read operation.
func (l *Layer) WriteAt(buf []byte, off int64) (n int, err error) {
	// If `buf` is large enough to span over several writes then we
	// have to calculate the offset of the first page, so that new
	// data is written to the correct place.
	pageOff := off % page.Size
	pageBuf := buf

	// Go over all pages this write affects.
	newOff := off + int64(len(buf))
	pageLo := off / page.Size
	pageHi := newOff / page.Size
	if newOff%page.Size == 0 {
		pageHi--
	}

	for pageIdx := pageLo; pageIdx <= pageHi; pageIdx++ {
		// Divide `buf` into small portions that will be copied
		// to the individual pages.

		mayWrite := page.Size - pageOff
		if mayWrite > int64(len(pageBuf)) {
			mayWrite = int64(len(pageBuf))
		}

		if mayWrite == 0 {
			break
		}

		// Overlay the part of `buf` that affects this page
		// and merge with any pre-existing writes.
		if err := l.cache.Merge(
			l.inode,
			uint32(pageIdx),
			uint32(pageOff),
			pageBuf[:mayWrite],
		); err != nil {
			return -1, err
		}

		// starting from the second block the page offset will
		// be always zero. That's only relevant for len(buf) > page.Size.
		pageOff = 0
		pageBuf = pageBuf[mayWrite:]
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

// ReadAt reads into `buf` from the position `off`.
//
// TODO: go docs state:
//  * ReadAt() must be allowed to call in parallel.
//    We cannot guarantee that at the moment since sometimes
//    we have to seek the underlying stream - mutex?
func (l *Layer) ReadAt(buf []byte, off int64) (int, error) {
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

	// keep the copy buf around between GC runs.
	copyBuf := copyBufPool.Get().([]byte)
	defer copyBufPool.Put(copyBuf)

	// Go over all pages this read may affect.
	// We might return early due to io.EOF though.
	newOff := off + int64(len(buf))
	pageLo := off / page.Size
	pageHi := newOff / page.Size
	if newOff%page.Size == 0 {
		pageHi--
	}

	for pageIdx := pageLo; pageIdx <= pageHi && ib.Left() > 0; pageIdx++ {
		p, err := l.cache.Lookup(l.inode, uint32(pageIdx))
		switch err {
		case page.ErrCacheMiss:
			// we don't have this page cached.
			// need to read it from zpr directly.
			if err := l.ensureOffset(zpr); err != nil {
				return ib.Len(), err
			}

			n, err := copyNBuffer(ib, zpr, int64(ib.Left()), copyBuf)
			l.overlayOffset += n
			l.streamOffset += n

			if err != nil {
				return ib.Len(), err
			}

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
				if err := l.ensureOffset(zpr); err != nil {
					return ib.Len(), err
				}

				// NOTE: Here we read the complete page (if possible)
				pageN, err := io.ReadFull(zpr, copyBuf[pageOff:])
				if pageN > 0 {
					p.Underlay(pageOff, copyBuf[pageOff:pageOff+uint32(pageN)])
					l.streamOffset += int64(pageN)
				}

				if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
					return ib.Len(), err
				}
			}

			pageMax := uint32(fullLen)
			if pageMax+pageOff > page.Size {
				pageMax = page.Size - pageOff
			}

			r := bytes.NewReader(p.Data[pageOff : pageOff+pageMax])
			n, err := copyNBuffer(ib, r, int64(ib.Left()), copyBuf)
			if err != nil && err != io.EOF {
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

// Truncate sets the size of the stream.
// There are three cases:
//
// - `size` is equal to Length(): Nothing happens.
// - `size` is less than Length(): The stream will return io.EOF earlier.
// - `size` is more than Length(): The stream will be padded with zeros.
//
// This matches the behavior of the equally confusingly named POSIX
// ftruncate() function. Note that Truncate() is a very fast operation.
func (l *Layer) Truncate(size int64) {
	l.length = size
}

// Length is the current truncated length of the overlay.
// When you did not call Truncate() it will be the size you
// passed to NewLayer(). Otherwise it is what you passed
// to the last call of Truncate().
func (l *Layer) Length() int64 {
	return l.length
}

/////////////////////////////////////
// FILE I/O COMPATIBILITY METHODS  //
//                                 //
// Do not use, unless you have to. //
// Prefer WriteAt() & ReadAt()     //
/////////////////////////////////////

// Read implements io.Reader by calling ReadAt()
// with the current offset.
func (l *Layer) Read(buf []byte) (int, error) {
	return l.ReadAt(buf, l.overlayOffset)
}

// Write writes `buf` at the current offset.
// It does NOT modify the read position.
func (l *Layer) Write(buf []byte) (int, error) {
	n, err := l.WriteAt(buf, l.overlayOffset)
	return n, err
}

// Seek changes the current offset for Write and Read.
// Note that in this implementation calling ReadAt.
// *does* change the seek offset. Use Seek() to make
// sure you're reading from the right spot.
func (l *Layer) Seek(off int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		l.overlayOffset = off
	case io.SeekCurrent:
		l.overlayOffset += off
	case io.SeekEnd:
		l.overlayOffset = l.length + off
	default:
		return 0, fmt.Errorf("invalid whence %d", whence)
	}

	return l.overlayOffset, nil
}

// Close will close the overlay and free up all resources,
// including pages in the cache.
func (l *Layer) Close() error {
	return l.cache.Evict(l.inode, l.length)
}

// WriteTo implements io.WriterTo
func (l *Layer) WriteTo(w io.Writer) (int64, error) {
	// NOTE: This method is mostly used in tests.
	// but can be also used by io.Copy() internally.
	// There is room for optimizations here:
	// Avoid one copy by directly writing to copyBuf.
	copyBuf := copyBufPool.Get().([]byte)
	defer copyBufPool.Put(copyBuf)

	wsum := int64(0)

	for {
		rn, rerr := l.ReadAt(copyBuf, l.overlayOffset)
		if rerr != nil && rerr != io.EOF {
			return wsum, rerr
		}

		wn, werr := w.Write(copyBuf[:rn])
		wsum += int64(wn)
		if werr != nil {
			return wsum, werr
		}

		if wn < rn {
			return wsum, io.ErrShortWrite
		}

		if rerr == io.EOF {
			return wsum, nil
		}

		if rn == 0 {
			return wsum, fmt.Errorf("nothing read, but no EOF")
		}
	}

}
