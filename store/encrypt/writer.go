package encrypt

import (
	"bytes"
	"crypto/rand"
	"io"
)

// Writer encrypts the data stream before writing to Writer.
type Writer struct {
	// Common fields with Reader
	aeadCommon

	// Internal Writer we would write to.
	Writer io.Writer

	// A buffer that is max. MaxBlockSize big.
	// Used for caching leftover data between writes.
	rbuf *bytes.Buffer

	// True after the first write.
	headerWritten bool

	// length is the total number of bytes passed to the writer.
	// This is needed to make SEEK_END on the reader side work.
	length int64

	// Compression flag for the header.
	// This module does not compression itself.
	compressionFlag bool
}

func (w *Writer) emitHeaderIfNeeded() error {
	if !w.headerWritten {
		w.headerWritten = true

		_, err := w.Writer.Write(GenerateHeader(w.key, w.length, w.compressionFlag))
		if err != nil {
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
		_, err := w.flushPack(MaxBlockSize)
		if err != nil {
			return 0, err
		}
	}

	// Remember left-overs for next write:
	_, err := w.rbuf.Write(p)
	if err != nil {
		return 0, nil
	}

	// Fake the amount of data we've written:
	return len(p), nil
}

func (w *Writer) flushPack(chunkSize int) (int, error) {
	// Create a new Nonce for this block:
	if _, err := rand.Read(w.nonce); err != nil {
		return 0, err
	}

	// Encrypt the text:
	w.encBuf = w.aead.Seal(w.encBuf[:0], w.nonce, w.rbuf.Next(chunkSize), nil)

	// Pass it to the underlying writer:
	written := 0

	n, err := w.Writer.Write(w.nonce)
	if err != nil {
		return n, err
	}

	written += n

	n, err = w.Writer.Write(w.encBuf)
	if err != nil {
		return n, err
	}

	written += n
	return written, nil
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

		if _, err := w.flushPack(n); err != nil {
			return err
		}
	}
	return nil
}

// NewWriter returns a new Writer which encrypts data with a
// certain key. If `compressionFlag` is true, the compression
// flag in the file header will also be true. Otherwise no compression is done.
func NewWriter(w io.Writer, key []byte, length int64, compressionFlag bool) (*Writer, error) {
	writer := &Writer{
		Writer:          w,
		rbuf:            &bytes.Buffer{},
		length:          length,
		compressionFlag: compressionFlag,
	}

	if err := writer.initAeadCommon(key, defaultCipherType); err != nil {
		return nil, err
	}

	return writer, nil
}
