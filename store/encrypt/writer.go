package encrypt

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
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
		if _, err := w.flushPack(MaxBlockSize); err != nil {
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

func (w *Writer) flushPack(chunkSize int) (int, error) {
	// Create a new Nonce for this block:
	if _, err := rand.Read(w.nonce); err != nil {
		return 0, err
	}

	// Encrypt the text:
	w.encBuf = w.aead.Seal(w.encBuf[:0], w.nonce, w.rbuf.Next(chunkSize), nil)

	// Pass it to the underlying writer:
	nNonce, err := w.Writer.Write(w.nonce)
	if err != nil {
		return nNonce, err
	}

	binary.BigEndian.PutUint64(w.blocknum, w.blockCount)
	nBlockNum, err := w.Writer.Write(w.blocknum)
	if err != nil {
		return nNonce + nBlockNum, err
	}

	w.blockCount++
	nBuf, err := w.Writer.Write(w.encBuf)
	return nNonce + nBlockNum + nBuf, err
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
