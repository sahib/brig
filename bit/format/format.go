// Package format implements the encryption and file format layer of brig.
// The file format used looks something like this:
//
// [HEADER][[BLOCKHEADER][PAYLOAD]...]
//
// HEADER is 20 bytes big and contains the following fields:
//    - 8 Byte: Magic number (to identify non-brig files quickly)
//    - 2 Byte: Format version
//    - 2 Byte: Used cipher type (ChaCha20 or AES-GCM)
//    - 4 Byte: Key length in bytes.
//	  - 4 Byte: Maximum size of each block (last may be less)
//
// BLOCKHEADER contains the following fields:
//    - 8 Byte: Nonce/Block number
//
// PAYLOAD contains the actual encrypted data, possibly with padding.
//
// All metadata is encoded in big endian.
//
// EncryptedReader/EncryptedWriter are capable or reading/writing this format.
// Additionally, both support efficient seeking into the encrypted data,
// provided the underlying datastream supports seeking.
package format

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"fmt"
	"hash"
	"io"
	"os"

	blake2 "github.com/codahale/blake2"
	chacha "github.com/codahale/chacha20poly1305"
	rbuf "github.com/glycerine/rbuf"
)

// Possible ciphers in Counter mode:
const (
	aeadCipherChaCha = iota
	aeadCipherAES
)

// Other constants:
const (
	// Size of the initial header:
	headerSize = 20

	// Chacha20 appears to be twice as fast as AES-GCM on my machine
	defaultCipherType = aeadCipherChaCha

	// MaxBlockSize is the maximum number of bytes a single payload may have
	MaxBlockSize = 1 * 1024 * 1024

	// GoodEncBufferSize is the recommended size of buffers
	GoodEncBufferSize = MaxBlockSize + 32

	// GoodDecBufferSize is the recommended size of buffers
	GoodDecBufferSize = MaxBlockSize
)

// Size of the used cipher's key in bytes
var KeySize = chacha.KeySize

////////////////////
// Header Parsing //
////////////////////

// GenerateHeader creates a valid header for the format file
func GenerateHeader() []byte {
	// This is in big endian:
	header := []byte{
		// Brigs magic number (8 Byte):
		0x6d, 0x6f, 0x6f, 0x73,
		0x65, 0x63, 0x61, 0x74,
		// File format version (2 Byte):
		0x0, 0x1,
		// Cipher type (2 Byte):
		defaultCipherType >> 8,
		defaultCipherType & 0xff,
		// Key length (4 Byte):
		0, 0, 0, 0,
		// Block length (4 Byte):
		0, 0, 0, 0,
	}

	binary.BigEndian.PutUint32(header[12:16], uint32(chacha.KeySize))
	binary.BigEndian.PutUint32(header[16:20], uint32(MaxBlockSize))
	return header
}

// ParseHeader parses the header of the format file.
// Returns the format version, cipher type, keylength and block length. If
// parsing fails, an error is returned.
func ParseHeader(header []byte) (format uint16, cipher uint16, keylen uint32, blocklen uint32, err error) {
	expected := GenerateHeader()
	if bytes.Compare(header[:8], expected[:8]) != 0 {
		err = fmt.Errorf("Magic number in header differs")
		return
	}

	format = binary.BigEndian.Uint16(header[8:10])
	cipher = binary.BigEndian.Uint16(header[10:12])
	switch cipher {
	case aeadCipherAES:
	case aeadCipherChaCha:
		// we support this!
	default:
		err = fmt.Errorf("Unknown cipher type: %d", cipher)
		return
	}

	keylen = binary.BigEndian.Uint32(header[12:16])
	blocklen = binary.BigEndian.Uint32(header[16:20])
	if blocklen != MaxBlockSize {
		err = fmt.Errorf("Unsupported block length in header: %d", blocklen)
	}

	return
}

//////////////////////
// Common Utilities //
//////////////////////

func createAEADWorker(cipherType uint16, key []byte) (cipher.AEAD, error) {
	switch cipherType {
	case aeadCipherAES:
		block, err := aes.NewCipher(key)
		if err != nil {
			return nil, err
		}
		return cipher.NewGCM(block)
	case aeadCipherChaCha:
		return chacha.New(key)
	}

	return nil, fmt.Errorf("No such cipher type.")
}

type aeadCommon struct {
	// Hashing io.Writer for in-band hashing.
	hasher hash.Hash

	// Nonce that form the first aead.NonceSize() bytes
	// of the output
	nonce []byte

	// For more information, see:
	// https://en.wikipedia.org/wiki/Authenticated_encryption
	aead cipher.AEAD

	// Buffer for encrypted data (MaxBlockSize + overhead)
	encBuf []byte

	// Buffer for decrypted data (MaxBlockSize)
	decBuf []byte
}

func (c *aeadCommon) initAeadCommon(key []byte, cipherType uint16) error {
	c.hasher = blake2.NewBlake2B()

	aead, err := createAEADWorker(cipherType, key)
	if err != nil {
		return err
	}

	c.nonce = make([]byte, aead.NonceSize())
	c.aead = aead

	c.encBuf = make([]byte, 0, MaxBlockSize+aead.Overhead())
	c.decBuf = make([]byte, 0, MaxBlockSize)

	return nil
}

////////////////////////////////////
// EncryptedReader Implementation //
////////////////////////////////////

// EncryptedReader decrypts and encrypted datastream from Reader.
type EncryptedReader struct {
	aeadCommon

	// Underlying io.Reader
	Reader io.Reader

	// Caches leftovers from unread blocks
	backlog *bytes.Reader

	// Last index of the byte the user visited.
	// (Used to avoid re-reads in Seek())
	// This does *not* equal the seek offset of the underlying stream.
	lastSeekPos int64
}

// Read from source and decrypt + hash it.
//
// This method always decrypts one block to optimize for continous reads. If
// dest is too small to hold the block, the decrypted text is cached for the
// next read.
func (r *EncryptedReader) Read(dest []byte) (int, error) {
	readBytes := 0

	// Try our best ot fill len(dest)
	for readBytes < len(dest) {
		if r.backlog.Len() == 0 {
			_, err := r.readBlock()
			if err != nil {
				return readBytes, err
			}
		}

		n, _ := r.backlog.Read(dest[readBytes:])
		readBytes += n
		r.lastSeekPos += int64(n)
	}

	return readBytes, nil
}

// Fill internal buffer with current block
func (r *EncryptedReader) readBlock() (int, error) {
	if n, err := r.Reader.Read(r.nonce); err != nil {
		return 0, err
	} else if n != r.aead.NonceSize() {
		return 0, fmt.Errorf("Nonce size mismatch. Should: %d. Have: %d",
			r.aead.NonceSize(), n)
	}

	// Read the *whole* block from the fs
	N := MaxBlockSize + r.aead.Overhead()
	n, err := io.ReadAtLeast(r.Reader, r.encBuf[:N], N)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return 0, err
	}

	r.decBuf, err = r.aead.Open(r.decBuf[:0], r.nonce, r.encBuf[:n], nil)
	if err != nil {
		return 0, err
	}

	if _, err = r.hasher.Write(r.decBuf); err != nil {
		return 0, err
	}

	r.backlog = bytes.NewReader(r.decBuf)
	return len(r.decBuf), nil
}

// Seek into the encrypted stream.
//
// Note that the seek offset is relative to the decrypted data,
// not to the underlying, encrypted stream.
//
// Mixing SEEK_CUR and SEEK_SET might not a good idea,
// since a seek might involve reading a whole encrypted block.
// Therefore relative seek offset
func (r *EncryptedReader) Seek(offset int64, whence int) (int64, error) {
	// Check if seeking is supported:
	seeker, ok := r.Reader.(io.ReadSeeker)
	if !ok {
		return 0, fmt.Errorf("Seek is not supported by underlying datastream")
	}

	// Constants and assumption on the stream below:
	blockSize := int64(MaxBlockSize)
	blockHeaderSize := int64(r.aead.NonceSize())
	totalBlockSize := blockHeaderSize + blockSize + int64(r.aead.Overhead())

	// absolute Offset in the decrypted stream
	absOffsetDec := int64(0)

	// Convert possibly relative offset to absolute offset:
	switch whence {
	case os.SEEK_CUR:
		absOffsetDec = r.lastSeekPos + offset
	case os.SEEK_SET:
		absOffsetDec = offset
	case os.SEEK_END:
		// We have no idea when the stream ends.
		return 0, fmt.Errorf("SEEK_END is not supported for encrypted data")
	}

	if absOffsetDec < 0 {
		return 0, fmt.Errorf("Negative seek index")
	}

	// Caller wanted to know only the current stream pos:
	if absOffsetDec == r.lastSeekPos {
		return absOffsetDec, nil
	}

	// Convert decrypted offset to encrypted offset
	absOffsetEnc := headerSize + ((absOffsetDec / blockSize) * totalBlockSize)

	// Check if we're still in the same block as last time:
	blockNum := absOffsetEnc / totalBlockSize
	lastBlockNum := r.lastSeekPos / blockSize
	r.lastSeekPos = absOffsetDec

	if lastBlockNum != blockNum {
		// Seek to the beginning of the encrypted block:
		if _, err := seeker.Seek(absOffsetEnc, os.SEEK_SET); err != nil {
			return 0, err
		}

		// Make read consume the current block:
		if _, err := r.readBlock(); err != nil {
			return 0, err
		}
	}

	// Reslice the backlog, so Read() does not return skipped data.
	r.backlog.Seek(absOffsetDec%blockSize, os.SEEK_SET)
	return absOffsetDec, nil
}

// Hash returns the internal hasher
func (r *EncryptedReader) Hash() hash.Hash {
	return r.hasher
}

// Close does finishing work.
// It does not close the underlying data stream.
//
// This is currently a No-Op, but you should not rely on that.
func (r *EncryptedReader) Close() error {
	return nil
}

// NewEncryptedReader creates a new encrypted reader and validates the file header.
// The key is required to be KeySize bytes long.
func NewEncryptedReader(r io.Reader, key []byte) (*EncryptedReader, error) {
	reader := &EncryptedReader{
		Reader:  r,
		backlog: bytes.NewReader([]byte{}),
	}

	header := make([]byte, headerSize)
	n, err := reader.Reader.Read(header)
	if err != nil {
		return nil, err
	}

	if n != headerSize {
		return nil, fmt.Errorf("No valid header found, damaged file?")
	}

	version, ciperType, keylen, _, err := ParseHeader(header)
	if err != nil {
		return nil, err
	}

	if version != 1 {
		return nil, fmt.Errorf("This implementation does not support versions != 1")
	}

	if uint32(len(key)) != keylen {
		return nil, fmt.Errorf("Key length differs: file=%d, user=%d", keylen, len(key))
	}

	if err := reader.initAeadCommon(key, ciperType); err != nil {
		return nil, err
	}

	return reader, nil
}

////////////////////////////////////
// EncryptedWriter Implementation //
////////////////////////////////////

// EncryptedWriter encrypts the datstream before writing to Writer.
type EncryptedWriter struct {
	// Common fields with EncryptedReader
	aeadCommon

	// Internal Writer we would write to.
	Writer io.Writer

	// A buffer that is MaxBlockSize big.
	// Used for caching blocks
	rbuf *rbuf.FixedSizeRingBuf
}

func (w *EncryptedWriter) Write(p []byte) (int, error) {
	for w.rbuf.Readable >= MaxBlockSize {
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

func (w *EncryptedWriter) flushPack(chunkSize int) (int, error) {
	n, err := w.rbuf.Read(w.decBuf[:chunkSize])
	if err != nil {
		return 0, err
	}

	// Try to update the checksum as we run:
	if _, err := w.hasher.Write(w.decBuf[:n]); err != nil {
		return 0, err
	}

	// Create a new Nonce for this block:
	blockNum := binary.BigEndian.Uint64(w.nonce)
	binary.BigEndian.PutUint64(w.nonce, blockNum+1)

	// Encrypt the text:
	w.encBuf = w.aead.Seal(w.encBuf[:0], w.nonce, w.decBuf[:n], nil)

	// Pass it to the underlying writer:
	written := 0
	if n, err := w.Writer.Write(w.nonce); err == nil {
		written += n
	}

	if n, err := w.Writer.Write(w.encBuf); err == nil {
		written += n
	}

	return written, nil
}

// Seek the write stream. This maps to a seek in the underlying datastream.
func (w *EncryptedWriter) Seek(offset int64, whence int) (int64, error) {
	if seeker, ok := w.Writer.(io.Seeker); ok {
		return seeker.Seek(offset, whence)
	}

	return 0, fmt.Errorf("Seek is not supported by underlying datastream")
}

// Close the EncryptedWriter and write any left-over blocks
// This does not close the underlying data stream.
func (w *EncryptedWriter) Close() error {
	for w.rbuf.Readable > 0 {
		n := MaxBlockSize
		if n > w.rbuf.Readable {
			n = w.rbuf.Readable
		}
		_, err := w.flushPack(n)
		if err != nil {
			return err
		}
	}
	return nil
}

// Hash returns the internal hasher
func (w *EncryptedWriter) Hash() hash.Hash {
	return w.hasher
}

// NewEncryptedWriter returns a new EncryptedWriter which encrypts data with a
// certain key.
func NewEncryptedWriter(w io.Writer, key []byte) (*EncryptedWriter, error) {
	writer := &EncryptedWriter{
		Writer: w,
		rbuf:   rbuf.NewFixedSizeRingBuf(MaxBlockSize * 2),
	}

	if err := writer.initAeadCommon(key, defaultCipherType); err != nil {
		return nil, err
	}

	_, err := w.Write(GenerateHeader())
	if err != nil {
		return nil, err
	}
	return writer, nil
}

// Encrypt is a utility function which encrypts the data from source with key
// and writes the resulting encrypted data to dest.
func Encrypt(key []byte, source io.Reader, dest io.Writer) (int64, error) {
	layer, err := NewEncryptedWriter(dest, key)
	if err != nil {
		return 0, err
	}

	defer layer.Close()
	return io.CopyBuffer(layer, source, make([]byte, GoodEncBufferSize))
}

// Decrypt is a utility function which decrypts the data from source with key
// and writes the resulting encrypted data to dest.
func Decrypt(key []byte, source io.Reader, dest io.Writer) (int64, error) {
	layer, err := NewEncryptedReader(source, key)
	if err != nil {
		return 0, err
	}

	defer layer.Close()
	return io.CopyBuffer(dest, layer, make([]byte, GoodDecBufferSize))
}
