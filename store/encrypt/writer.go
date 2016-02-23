package encrypt

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

var (
	ErrBadBlockSize = errors.New("Underlying reader failed to read full MaxBlockSize")
	ErrMixedMethods = errors.New("Mixing Write() and ReadFrom() is not allowed.")
)

// Writer encrypts the data stream before writing to Writer.
type Writer struct {
	// Internal Writer we would write to.
	io.Writer

	// Common fields with Reader
	aeadCommon

	// A buffer that is max. MaxBlockSize big.
	// Used for caching leftover data between writes.
	rbuf *bytes.Buffer

	// Index of the currently written block.
	blockCount uint64

	// True after the first write.
	headerWritten bool

	// length is the total number of bytes passed to the writer.
	// This is needed to make SEEK_END on the reader side work.
	length int64
}

func (w *Writer) emitHeaderIfNeeded() error {
	if !w.headerWritten {
		w.headerWritten = true
		header := GenerateHeader(w.key, w.length)

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

	for w.rbuf.Len() >= MaxBlockSize {
		if _, err := w.flushPack(w.rbuf.Next(MaxBlockSize)); err != nil {
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
		n := w.rbuf.Len()
		if n > MaxBlockSize {
			n = MaxBlockSize
		}

		if _, err := w.flushPack(w.rbuf.Next(n)); err != nil {
			return err
		}
	}
	return nil
}

// R
func (w *Writer) ReadFrom(r io.Reader) (int64, error) {
	if err := w.emitHeaderIfNeeded(); err != nil {
		return 0, err
	}

	n, nprev := int64(0), -1
	buf := make([]byte, GoodDecBufferSize)

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
		if nprev >= 0 && nprev != MaxBlockSize && rerr != io.EOF {
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

// NewWriter returns a new Writer which encrypts data with a
// certain key. If `compressionFlag` is true, the compression
// flag in the file header will also be true. Otherwise no compression is done.
func NewWriter(w io.Writer, key []byte, length int64) (*Writer, error) {
	writer := &Writer{
		Writer: w,
		rbuf:   &bytes.Buffer{},
		length: length,
	}

	if err := writer.initAeadCommon(key, defaultCipherType); err != nil {
		return nil, err
	}

	return writer, nil
}
