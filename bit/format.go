// This package implements the encryption and file format layer of brig.
// The file format used looks something like this:
//
// [HEADER][[BLOCKHEADER][PAYLOAD]...]
//
// HEADER is 16 bytes big and contains the following fields:
//    - 8 Byte: Magic number (to identify non-brig files quickly)
//    - 2 Byte: Format version
//    - 2 Byte: Used cipher type (ChaCha20 or AES-GCM)
//    - 4 Byte: Key length in bytes.
//
// BLOCKHEADER contains the following fields:
//    - 4 Byte: Size of the Payload (1 MB or less)
//    - 8 Byte: Nonce/Block number
//
// PAYLOAD contains the actual encrypted data, possibly with padding.
//
// All metadata is encoded in big endian.
//
// EncryptedReader/EncryptedWriter are capable or reading/writing this format.
// Additionally, both support efficient seeking into the encrypted data,
// provided the underlying datastream supports seeking.
package main

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
)

// Possible ciphers in Counter mode:
const (
	aeadCipherChaCha = iota
	aeadCipherAES
)

// Other constants:
const (
	// TODO: size still needed?
	padPackSize = 4

	// Size of the initial header:
	headerSize = 16

	// Chacha20 appears to be twice as fast as AES-GCM on my machine
	defaultCipherType = aeadCipherChaCha

	// Maximum number of bytes a single payload may have
	MaxBlockSize = 1 * 1024

	// The recommended size of a buffer for efficienct reading
	GoodBufferSize = MaxBlockSize + 32
)

// Size of the used cipher's key in bytes
var KeySize int = chacha.KeySize

////////////////////
// Header Parsing //
////////////////////

// Generate a valid header for the format file:
func GenerateHeader() []byte {
	// This is in big endian:
	return []byte{
		// Brigs magic number (8 Byte):
		0x6d, 0x6f, 0x6f, 0x73,
		0x65, 0x63, 0x61, 0x74,
		// File format version (2 Byte):
		0x0, 0x1,
		// Cipher type (2 Byte):
		defaultCipherType >> 8,
		defaultCipherType & 0xff,
		// Key length (4 Byte):
		byte(uint32(chacha.KeySize) >> 24),
		byte(uint32(chacha.KeySize) >> 16),
		byte(uint32(chacha.KeySize) >> 8),
		byte(uint32(chacha.KeySize) & 0xff),
	}
}

// Parse the header of the format file:
// Returns the format version, cipher type, keylength.
// If parsing fails, an error is returned.
func ParseHeader(header []byte) (uint16, uint16, uint32, error) {
	expected := GenerateHeader()
	if bytes.Compare(header[:8], expected[:8]) != 0 {
		return 0, 0, 0, fmt.Errorf("Magic number differs")
	}

	format := binary.BigEndian.Uint16(header[8:10])
	cipher := binary.BigEndian.Uint16(header[10:12])
	switch cipher {
	case aeadCipherAES:
	case aeadCipherChaCha:
		// we support this!
	default:
		return 0, 0, 0, fmt.Errorf("Unknown cipher type: %d", cipher)
	}

	keylen := binary.BigEndian.Uint32(header[12:16])
	return format, cipher, keylen, nil
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

	// Short temporary buffer for encoding the packSize
	sizeBuf []byte

	// For more information, see:
	// https://en.wikipedia.org/wiki/Authenticated_encryption
	aead cipher.AEAD
}

func (c *aeadCommon) initAeadCommon(key []byte, cipherType uint16) error {
	c.hasher = blake2.NewBlake2B()

	aead, err := createAEADWorker(cipherType, key)
	if err != nil {
		return err
	}

	c.nonce = make([]byte, aead.NonceSize())
	c.sizeBuf = make([]byte, padPackSize)
	c.aead = aead
	return nil
}

////////////////////////////////////
// EncryptedReader Implementation //
////////////////////////////////////

// EncryptedReader decrypts and encrypted datastream from Reader.
type EncryptedReader struct {
	aeadCommon

	Reader io.Reader

	encBuf  []byte
	decBuf  []byte
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
	if r.backlog.Len() == 0 {
		_, err := r.readBlock()
		if err != nil {
			return 0, err
		}
	}
	n, _ := r.backlog.Read(dest)
	r.lastSeekPos += int64(n)
	return n, nil
}

// Fill internal buffer with current block
func (r *EncryptedReader) readBlock() (int, error) {
	n, err := r.Reader.Read(r.sizeBuf)
	if err != nil {
		return 0, err
	}

	packSize := binary.BigEndian.Uint32(r.sizeBuf)

	if packSize > MaxBlockSize+uint32(r.aead.Overhead()) {
		return 0, fmt.Errorf("Pack size exceeded expectations: %d > %d",
			packSize, MaxBlockSize)
	}

	if n, err = r.Reader.Read(r.nonce); err != nil {
		return 0, err
	} else if n != r.aead.NonceSize() {
		return 0, fmt.Errorf("Nonce size mismatch. Should: %d. Have: %d",
			r.aead.NonceSize(), n)
	}

	n, err = r.Reader.Read(r.encBuf[:packSize])
	if err != nil {
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
	blockHeaderSize := int64(padPackSize + r.aead.NonceSize())
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
	absOffsetEnc := (absOffsetDec / blockSize) * totalBlockSize
	absOffsetEnc += headerSize

	// Clear backlog, reading will cause it to be re-read
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

// Return the internal hasher
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

	version, ciperType, keylen, err := ParseHeader(header)
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

	reader.encBuf = make([]byte, 0, MaxBlockSize+reader.aead.Overhead())
	reader.decBuf = make([]byte, 0, MaxBlockSize)
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
	packBuf *bytes.Buffer

	encBuf []byte
}

func (w *EncryptedWriter) Write(p []byte) (int, error) {
	n, _ := w.packBuf.Write(p)
	if w.packBuf.Len() < MaxBlockSize {
		// Fake the amount of data we've written:
		return n, nil
	}

	_, err := w.flushPack(MaxBlockSize)
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (w *EncryptedWriter) flushPack(chunkSize int) (int, error) {
	// Try to update the checksum as we run:
	src := w.packBuf.Bytes()[:chunkSize]

	// Make sure to advance this many bytes
	// in case any leftovers are in the buffer.
	defer w.packBuf.Read(src[:chunkSize])

	if _, err := w.hasher.Write(src); err != nil {
		return 0, err
	}

	// Create a new Nonce for this block:
	//storeVarint(w.nonce, loadVarint(w.nonce)+1)
	blockNum := binary.BigEndian.Uint64(w.nonce)
	binary.BigEndian.PutUint64(w.nonce, blockNum+1)

	// Encrypt the text:
	w.encBuf = w.aead.Seal(w.encBuf[:0], w.nonce, src, nil)

	// Pass it to the underlying writer:
	// storeVarint(w.sizeBuf, uint64(len(encrypted)))
	binary.BigEndian.PutUint32(w.sizeBuf, uint32(len(w.encBuf)))

	written := 0
	if n, err := w.Writer.Write(w.sizeBuf); err == nil {
		written += n
	}

	if n, err := w.Writer.Write(w.nonce); err == nil {
		written += n
	}

	if n, err := w.Writer.Write(w.encBuf); err == nil {
		written += n
	}

	// len(encrypted) might be more than len(w.packBuf)
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
	for w.packBuf.Len() > 0 {
		size := w.packBuf.Len()
		if size > MaxBlockSize {
			size = MaxBlockSize
		}

		_, err := w.flushPack(size)
		if err != nil {
			return nil
		}
	}
	return nil
}

// Return the internal hasher
func (w *EncryptedWriter) Hash() hash.Hash {
	return w.hasher
}

// NewEncryptedWriter returns a new EncryptedWriter which encrypts data with a
// certain key.
func NewEncryptedWriter(w io.Writer, key []byte) (*EncryptedWriter, error) {
	writer := &EncryptedWriter{
		Writer:  w,
		packBuf: bytes.NewBuffer(make([]byte, 0, MaxBlockSize)),
	}

	if err := writer.initAeadCommon(key, defaultCipherType); err != nil {
		return nil, err
	}

	writer.encBuf = make([]byte, 0, MaxBlockSize+writer.aead.Overhead())

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
	return io.CopyBuffer(layer, source, make([]byte, GoodBufferSize))
}

// Decrypt is a utility function which decrypts the data from source with key
// and writes the resulting encrypted data to dest.
func Decrypt(key []byte, source io.Reader, dest io.Writer) (int64, error) {
	layer, err := NewEncryptedReader(source, key)
	if err != nil {
		return 0, err
	}

	defer layer.Close()
	return io.CopyBuffer(dest, layer, make([]byte, GoodBufferSize))
}
