package overlay

import (
	"io"
)

// small util to wrap a buffer we want to write to. Tells you easily how much
// data you can still write to it.
type iobuf struct {
	dst []byte
	off int
}

func (ib *iobuf) Write(src []byte) (int, error) {
	n := copy(ib.dst[ib.off:ib.off+ib.Left()], src)
	ib.off += n
	return n, nil
}

func (ib *iobuf) Len() int {
	return ib.off
}

func (ib *iobuf) Left() int {
	return len(ib.dst) - ib.off
}

// zeroPadReader wraps another reader which has data
// until `size`. If `length` > `size` than it pads the
// gap with zero reads.
type zeroPadReader struct {
	r                 io.Reader
	off, size, length int64
}

func memzero(buf []byte) {
	// TODO: Check if the for loop is faster or
	//       if we should copy() from a pre-allocated zero buf.
	for idx := range buf {
		buf[idx] = 0
	}
}

func (zpr *zeroPadReader) Read(buf []byte) (int, error) {
	if zpr.size >= zpr.length {
		// sanity check. zpr.length might be also shorter.
		// then we don't do any padding but work like
		// io.LimitReader().
		zpr.size = zpr.length
	}

	diff := zpr.length - zpr.off
	bufLen := int64(len(buf))
	if diff < bufLen {
		// clamp buf to zpr.length
		bufLen = diff
	}

	if zpr.off < zpr.size {
		// below underlying stream size:
		n, err := zpr.r.Read(buf[:bufLen])
		zpr.off += int64(n)
		return n, err
	}

	if diff > 0 {
		// above underlying stream size,
		// but below padded length.
		memzero(buf[:bufLen])
		zpr.off += bufLen
		return int(bufLen), nil
	}

	return 0, io.EOF
}

/////////

// copyNBuffer is golang's io.CopyN with added param for the buffer,
// like in io.CopyBuffer. Saves precious allocations.
func copyNBuffer(dst io.Writer, src io.Reader, n int64, buf []byte) (written int64, err error) {
	written, err = io.CopyBuffer(dst, io.LimitReader(src, n), buf)
	if written == n {
		return n, nil
	}

	if written < n && err == nil {
		// src stopped early; must have been EOF.
		err = io.EOF
	}

	return
}
