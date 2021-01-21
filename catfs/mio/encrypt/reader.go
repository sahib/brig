package encrypt

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// Reader decrypts and encrypted stream from Reader.
type Reader struct {
	// Underlying reader
	io.Reader

	aeadCommon

	// Caches leftovers from unread blocks
	backlog *bytes.Reader

	// Last index of the byte the user visited.
	// (Used to avoid re-reads in Seek())
	// This does *not* equal the seek offset of the underlying stream.
	lastDecSeekPos int64

	// lastEncSeekPos saves the current position of the underlying stream.
	// it is used mostly for ensuring SEEK_END works.
	lastEncSeekPos int64

	// Parsed header info
	info *HeaderInfo

	// true once readHeader() was called
	parsedHeader bool

	// Buffer for decrypted data (MaxBlockSize big)
	decBuf []byte

	// true as long readBlock was not successful
	isInitialRead bool

	// Total size of the underlying stream in bytes.
	// This is only set when SEEK_END was used.
	endOffsetEnc int64
}

func (r *Reader) readHeaderIfNotDone() error {
	if r.parsedHeader {
		return nil
	}

	r.parsedHeader = true

	header := make([]byte, headerSize)
	n, err := io.ReadFull(r.Reader, header)
	if err != nil {
		return err
	}

	if n != headerSize {
		return fmt.Errorf("No valid header found, damaged file?")
	}

	info, err := ParseHeader(header, r.key)
	if err != nil {
		return err
	}

	if info.Version != 1 {
		return fmt.Errorf("this implementation does not support versions != 1")
	}

	if uint32(len(r.key)) != info.KeyLen {
		return fmt.Errorf(
			"key length differs: file=%d, user=%d",
			info.KeyLen,
			len(r.key),
		)
	}

	r.info = info
	if err := r.initAeadCommon(
		r.key,
		info.CipherBit,
		int64(r.info.BlockLen),
	); err != nil {
		return err
	}

	r.lastEncSeekPos += headerSize
	r.decBuf = make([]byte, 0, r.info.BlockLen)
	return nil
}

// Flags will return the flags stored in the header of the encrypted stream.
// If the header was not read yet, it will attempt to read it.
func (r *Reader) Flags() (Flags, error) {
	// Make sure we have the info needed to parse the header:
	if err := r.readHeaderIfNotDone(); err != nil {
		return 0, err
	}

	return r.info.Flags, nil
}

// Read from source and decrypt.
//
// This method always decrypts one block to optimize for continuous reads. If
// dest is too small to hold the block, the decrypted text is cached for the
// next read.
func (r *Reader) Read(dest []byte) (int, error) {
	// Make sure we have the info needed to parse the header:
	if err := r.readHeaderIfNotDone(); err != nil {
		return 0, err
	}

	readBytes := 0

	// Try our best to fill len(dest)
	for readBytes < len(dest) {
		if r.backlog.Len() == 0 {
			if _, rerr := r.readBlock(); rerr != nil && rerr != io.EOF {
				return readBytes, rerr
			}
		}

		n, berr := r.backlog.Read(dest[readBytes:])
		r.lastDecSeekPos += int64(n)
		readBytes += n

		if berr == io.EOF {
			return readBytes, io.EOF
		}
	}

	return readBytes, nil
}

// Fill internal buffer with current block
func (r *Reader) readBlock() (int, error) {
	if r.info == nil {
		return 0, fmt.Errorf("Invalid header data")
	}

	// Read nonce:
	if n, err := io.ReadFull(r.Reader, r.nonce); err != nil {
		return 0, err
	} else if n != r.aead.NonceSize() {
		return 0, fmt.Errorf(
			"nonce size mismatch; should: %d - have: %d (err: %v)",
			r.aead.NonceSize(),
			n,
			err,
		)
	}

	// Convert to block number:
	readBlockNum := binary.LittleEndian.Uint64(r.nonce)

	// Check the block number:
	currBlockNum := uint64(r.lastDecSeekPos / int64(r.info.BlockLen))
	if currBlockNum != readBlockNum {
		return 0, fmt.Errorf(
			"bad block number; as %d, should be %d", readBlockNum, currBlockNum,
		)
	}

	// Read the *whole* block from the raw stream
	N := int(r.info.BlockLen) + r.aead.Overhead()
	n, err := io.ReadAtLeast(r.Reader, r.encBuf[:N], N)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return 0, err
	}

	r.lastEncSeekPos += int64(n) + int64(len(r.nonce))

	r.decBuf, err = r.aead.Open(r.decBuf[:0], r.nonce, r.encBuf[:n], nil)
	if err != nil {
		return 0, err
	}

	r.backlog = bytes.NewReader(r.decBuf)
	r.isInitialRead = false

	return len(r.decBuf), nil
}

// Seek into the encrypted stream.
//
// Note that the seek offset is relative to the decrypted data,
// not to the underlying, encrypted stream.
func (r *Reader) Seek(offset int64, whence int) (int64, error) {
	// Check if seeking is supported:
	seeker, ok := r.Reader.(io.Seeker)
	if !ok {
		return 0, fmt.Errorf("seek is not supported by underlying stream")
	}

	if err := r.readHeaderIfNotDone(); err != nil {
		return 0, err
	}

	// set to true when an actual call to seeker.Seek() was made.
	wasMoved := false

	// Constants and assumption on the stream below:
	blockHeaderSize := int64(r.aead.NonceSize())
	blockOverhead := blockHeaderSize + int64(r.aead.Overhead())
	totalBlockSize := blockOverhead + int64(r.info.BlockLen)

	// absolute Offset in the decrypted stream
	absOffsetDec := int64(0)

	// Convert possibly relative offset to absolute offset:
	switch whence {
	case io.SeekCurrent:
		absOffsetDec = r.lastDecSeekPos + offset
	case io.SeekStart:
		absOffsetDec = offset
	case io.SeekEnd:
		// Try to figure out the end of the stream.
		// This might be inefficient for some underlying readers,
		// but is probably okay for ipfs.
		//
		// NOTE: We cache this not only for performance, but also
		//       as a workaround for a bug in ipfs.
		//       See: https://github.com/ipfs/go-ipfs/issues/2567
		if r.endOffsetEnc < 0 {
			endOffsetEnc, err := seeker.Seek(0, io.SeekEnd)
			if err != nil && err != io.EOF {
				return 0, err
			}

			r.endOffsetEnc = endOffsetEnc
		}

		// This computation is verbose on purporse,
		// since the details might be confusing.
		encLen := (r.endOffsetEnc - headerSize)
		encRest := encLen % totalBlockSize
		decBlocks := encLen / totalBlockSize

		endOffsetDec := decBlocks * int64(r.info.BlockLen)
		if encRest > 0 {
			endOffsetDec += encRest - blockOverhead
		}
		absOffsetDec = endOffsetDec + offset

		if absOffsetDec < 0 {
			// That's the wrong end of file...
			return 0, io.EOF
		}

		// For SEEK_END we need to make sure that we move the seek pointer
		// back to a sensible position when we decide that no actual move
		// is necessary further down this function.
		defer func() {
			if !wasMoved {
				seeker.Seek(r.lastEncSeekPos, io.SeekStart)
			}
		}()
	}

	if absOffsetDec < 0 {
		return 0, fmt.Errorf("Negative seek index: %d", absOffsetDec)
	}

	// Caller wanted to know only the current stream pos:
	if absOffsetDec == r.lastDecSeekPos {
		return absOffsetDec, nil
	}

	// Convert decrypted offset to encrypted offset
	absOffsetEnc := headerSize + ((absOffsetDec / int64(r.info.BlockLen)) * totalBlockSize)

	// Check if we're still in the same block as last time:
	blockNum := absOffsetEnc / totalBlockSize
	lastBlockNum := r.lastDecSeekPos / int64(r.info.BlockLen)

	r.lastDecSeekPos = absOffsetDec

	if lastBlockNum != blockNum || r.isInitialRead || whence == io.SeekEnd {
		r.lastEncSeekPos = absOffsetEnc

		// Seek to the beginning of the encrypted block:
		wasMoved = true
		if _, err := seeker.Seek(absOffsetEnc, io.SeekStart); err != nil {
			return 0, err
		}

		// Make read consume the current block:
		if _, err := r.readBlock(); err != nil {
			return 0, err
		}
	}

	// reslice the backlog, so Read() does not return skipped data.
	if _, err := r.backlog.Seek(
		absOffsetDec%int64(r.info.BlockLen),
		io.SeekStart,
	); err != nil {
		return 0, err
	}

	return absOffsetDec, nil
}

// WriteTo copies all data from `r` to `w`.
//
// It is intended to avoid unneeded copying by choosing a suitable buffer size
// and by directly reading block after block. io.Copy will use it automatically.
//
// It returns the number of written bytes and possible errors (but no io.EOF)
func (r *Reader) WriteTo(w io.Writer) (int64, error) {
	// Make sure we have the info needed to parse the header:
	if err := r.readHeaderIfNotDone(); err != nil {
		return 0, err
	}

	n := int64(0)

	// Backlog might be still filled if Read() or Seek() was done before:
	if r.backlog.Len() > 0 {
		bn, err := r.backlog.WriteTo(w)
		if err != nil {
			return bn, err
		}

		n += bn
		r.lastDecSeekPos += bn
	}

	for {
		nread, rerr := r.readBlock()
		if rerr != nil && rerr != io.EOF {
			return n, rerr
		}

		r.lastDecSeekPos += int64(nread)

		nwrite, werr := w.Write(r.decBuf[:nread])
		if werr != nil {
			return n, werr
		}

		n += int64(nwrite)

		if nwrite != nread {
			return n, io.ErrShortWrite
		}

		if rerr == io.EOF {
			break
		}
	}

	return n, nil
}

// NewReader creates a new encrypted reader and validates the file header.
// The key is required to be KeySize bytes long.
func NewReader(r io.Reader, key []byte) (*Reader, error) {
	reader := &Reader{
		Reader:        r,
		backlog:       bytes.NewReader([]byte{}),
		parsedHeader:  false,
		isInitialRead: true,
		endOffsetEnc:  -1,
		aeadCommon: aeadCommon{
			key: key,
		},
	}

	return reader, nil
}
