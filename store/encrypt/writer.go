package encrypt

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

var (
	ErrBadBlockSize = errors.New("Underlying reader failed to read full w.maxBlockSize")
	ErrMixedMethods = errors.New("Mixing Write() and ReadFrom() is not allowed.")
)

// Writer encrypts the data stream before writing to Writer.
type Writer struct {
	// Internal Writer we would write to.
	io.Writer

	// Common fields with Reader
	aeadCommon

	// A buffer that is max. w.maxBlockSize big.
	// Used for caching leftover data between writes.
	rbuf *bytes.Buffer

	// Index of the currently written block.
	blockCount uint64

	// True after the first write.
	headerWritten bool

	// w.maxBlockSize is the maximum number of bytes a single payload may have
	maxBlockSize int64
}

func (w *Writer) GoodDecBufferSize() int64 {
	return w.maxBlockSize
}

func (w *Writer) GoodEncBufferSize() int64 {
	return w.maxBlockSize + 40
}

func (w *Writer) emitHeaderIfNeeded() error {
	if !w.headerWritten {
		w.headerWritten = true
		header := GenerateHeader(w.key, w.maxBlockSize)

		if _, err := w.Writer.Write(header); err != nil {
			return err
		}
	}

	return nil
}

func (w *Writer) Write(p []byte) (int, error) {
	if err := w.emitHeaderIfNeeded(); err != nil {
		return 0, err
	}

	for int64(w.rbuf.Len()) >= w.maxBlockSize {
		if _, err := w.flushPack(w.rbuf.Next(int(w.maxBlockSize))); err != nil {
			return 0, err
		}
	}

	// Remember left-overs for next write:
	if _, err := w.rbuf.Write(p); err != nil {
		return 0, nil
	}

	// Fake the amount of data we've written:
	return len(p), nil
}

func (w *Writer) flushPack(pack []byte) (int, error) {
	// Create a new Nonce for this block:
	binary.LittleEndian.PutUint64(w.nonce, w.blockCount)

	// Encrypt the text:
	w.encBuf = w.aead.Seal(w.encBuf[:0], w.nonce, pack, nil)

	// Pass it to the underlying writer:
	nNonce, err := w.Writer.Write(w.nonce)
	if err != nil {
		return nNonce, err
	}

	w.blockCount++
	nBuf, err := w.Writer.Write(w.encBuf)
	return nNonce + nBuf, err
}

// Close the Writer and write any left-over blocks
// This does not close the underlying data stream.
func (w *Writer) Close() error {
	if err := w.emitHeaderIfNeeded(); err != nil {
		return err
	}

	// Flush last block of data if any:
	for w.rbuf.Len() > 0 {
		n := int64(w.rbuf.Len())
		if n > w.maxBlockSize {
			n = w.maxBlockSize
		}

		if _, err := w.flushPack(w.rbuf.Next(int(n))); err != nil {
			return err
		}
	}
	return nil
}

// ReadFrom writes all readable from `r` into `w`.
//
// It is intentend as optimized way to copy the whole stream without
// unneeded copying in between. io.Copy() will use this function automatically.
//
// It returns the number of read bytes and any encountered error (no io.EOF)
func (w *Writer) ReadFrom(r io.Reader) (int64, error) {
	if err := w.emitHeaderIfNeeded(); err != nil {
		return 0, err
	}

	n, nprev := int64(0), -1
	buf := make([]byte, defaultDecBufferSize)

	// Check if a previous Write() wrote to rbuf.
	if w.rbuf.Len() > 0 {
		return 0, ErrMixedMethods
	}

	for {
		nread, rerr := r.Read(buf)
		if rerr != nil && rerr != io.EOF {
			return n, rerr
		}

		n += int64(nread)

		// Sanity check: check if previous block was properly aligned:
		if nprev >= 0 && int64(nprev) != w.maxBlockSize && rerr != io.EOF {
			return n, ErrBadBlockSize
		}

		if nread > 0 {
			_, werr := w.flushPack(buf[:nread])
			w.rbuf.Reset()

			if werr != nil {
				return n, werr
			}
		}

		nprev = nread

		if rerr == io.EOF {
			break
		}
	}

	return n, nil
}

// NewWriter calls NewWriterWithTypeAndBlockSize with a sane default cipher type
// and a sane default max block size.
func NewWriter(w io.Writer, key []byte) (*Writer, error) {
	return NewWriterWithType(w, key, defaultCipherType)
}

// NewWriterWithType calls NewWriterWithTypeAndBlockSize with a a sane default maxblocksize.
func NewWriterWithType(w io.Writer, key []byte, cipherType uint16) (*Writer, error) {
	return NewWriterWithTypeAndBlockSize(w, key, cipherType, defaultMaxBlockSize)
}

// NewWriterWithTypeAndBlockSize returns a new Writer which encrypts data with a
// certain key. If `compressionFlag` is true, the compression
// flag in the file header will also be true. Otherwise no compression is done.
func NewWriterWithTypeAndBlockSize(w io.Writer, key []byte, cipherType uint16, maxBlockSize int64) (*Writer, error) {
	ew := &Writer{
		Writer:       w,
		rbuf:         &bytes.Buffer{},
		maxBlockSize: maxBlockSize,
	}

	if err := ew.initAeadCommon(key, cipherType, ew.maxBlockSize); err != nil {
		return nil, err
	}

	return ew, nil
}
