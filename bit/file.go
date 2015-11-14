// package bit
package main

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/binary"
	"flag"
	"fmt"
	"hash"
	"io"
	"log"
	"os"
	"time"

	chacha "github.com/codahale/chacha20poly1305"
	multihash "github.com/jbenet/go-multihash"
)

type File interface {
	// Path relative to the repo root
	Path() string

	// File size of the file in bytes
	Size() int

	// Modification timestamp (with timezone)
	Mtime() time.Time

	// Hash of the unencrypted file
	Hash() multihash.Multihash

	// Hash of the encrypted file from IPFS
	IpfsHash() multihash.Multihash
}

func NewFile(path string) (*File, error) {
	// TODO:
	return nil, nil
}

// Possible ciphers in Counter mode:
const (
	aeadCipherChaCha = iota
	aeadCipherAES
)

//
const (
	maxPackSize       = 1 * 1024 * 1024
	goodBufSize       = maxPackSize + 32
	padPackSize       = 4
	defaultCipherType = aeadCipherChaCha
)

func createAEADWorker(cipherType int, key []byte) (cipher.AEAD, error) {
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
	Hasher hash.Hash

	// Nonce that form the first aead.NonceSize() bytes
	// of the output
	nonce []byte

	// Short temporary buffer for encoding the packSize
	sizeBuf []byte

	// For more information, see:
	// https://en.wikipedia.org/wiki/Authenticated_encryption
	aead cipher.AEAD
}

func (c *aeadCommon) initAeadCommon(key []byte) error {
	c.Hasher = sha1.New()

	aead, err := createAEADWorker(defaultCipherType, key)
	if err != nil {
		return err
	}

	c.nonce = make([]byte, aead.NonceSize())
	c.sizeBuf = make([]byte, padPackSize)
	c.aead = aead
	return nil
}

type AEADReader struct {
	aeadCommon

	Reader io.ReadSeeker

	packBuf []byte
	backlog *bytes.Buffer
}

// Read from source and decrypt + hash it.
//
// This method always decrypts one block to optimize for continous reads. If
// dest is too small to hold the block, the decrypted text is cached for the
// next read.
func (r *AEADReader) Read(dest []byte) (int, error) {
	if r.backlog.Len() == 0 {
		_, err := r.readBlock()
		if err != nil {
			return 0, err
		}
	}

	return r.backlog.Read(dest)
}

// Fill internal buffer with current block
func (r *AEADReader) readBlock() (int, error) {
	n, err := r.Reader.Read(r.sizeBuf)
	if err != nil {
		return 0, err
	}

	packSize := binary.BigEndian.Uint32(r.sizeBuf)
	if packSize > maxPackSize+uint32(r.aead.Overhead()) {
		return 0, fmt.Errorf("Pack size exceeded expectations: %d > %d",
			packSize, maxPackSize)
	}

	if n, err = r.Reader.Read(r.nonce); err != nil {
		return 0, err
	} else if n != r.aead.NonceSize() {
		return 0, fmt.Errorf("Nonce size mismatch. Should: %d. Have: %d",
			r.aead.NonceSize(), n)
	}

	n, err = r.Reader.Read(r.packBuf[:packSize])
	if err != nil {
		return 0, err
	}

	decrypted, err := r.aead.Open(nil, r.nonce, r.packBuf[:n], nil)
	if err != nil {
		return 0, err
	}

	if _, err = r.Hasher.Write(decrypted); err != nil {
		return 0, err
	}

	return r.backlog.Write(decrypted)
}

func convertEncToDec(n, blockSize, totalBlockSize, headerSize int64) int64 {
	encBlocks := n / totalBlockSize
	encOffset := n % totalBlockSize
	return (encBlocks * blockSize) + encOffset - headerSize
}

// Seek into the encrypted stream.
//
// Note that the seek offset is relative to the decrypted data,
// not to the underlying, encrypted stream.
//
// Mixing SEEK_CUR and SEEK_SET might not a good idea,
// since a seek might involve reading a whole encrypted block.
// Therefore relative seek offset
func (r *AEADReader) Seek(offset int64, whence int) (int64, error) {
	// Constants and assumption on the stream below:
	blockSize := int64(maxPackSize)
	headerSize := int64(padPackSize + r.aead.NonceSize())
	totalBlockSize := headerSize + blockSize + int64(r.aead.Overhead())

	// absolute Offset in the decrypted stream
	absOffsetDec := int64(0)

	// Find out where we are currently:
	absSeekOffsetEnc, err := r.Reader.Seek(0, os.SEEK_CUR)
	if err != nil {
		return 0, err
	}

	// Convert the encrypted offset to the decrypted offset:
	absSeekOffsetDec := convertEncToDec(absSeekOffsetEnc,
		blockSize, totalBlockSize, headerSize)

	// Convert possibly relative offset to absolute offset:
	switch whence {
	case os.SEEK_CUR:
		absOffsetDec = absSeekOffsetDec + offset
	case os.SEEK_SET:
		absOffsetDec = offset
	case os.SEEK_END:
		return 0, fmt.Errorf("There is no definite end, can't use SEEK_END")
	}

	// Caller wanted to know only the current stream pos:
	if absOffsetDec == absSeekOffsetDec {
		return absOffsetDec, nil
	}

	// Convert decrypted offset to encrypted offset
	absOffsetEnc := (absOffsetDec / blockSize) * totalBlockSize

	// Clear backlog, reading will cause it to be re-read
	r.backlog.Reset()

	// Seek to the beginning of the encrypted block:
	_, err = r.Reader.Seek(absOffsetEnc, os.SEEK_SET)
	if err != nil {
		return 0, err
	}

	// TODO: make this dummy buf persist during seeks
	dummy := make([]byte, absOffsetDec%blockSize)
	if _, err = r.Read(dummy); err != nil {
		return 0, err
	}

	return absOffsetDec, nil
}

func (r *AEADReader) Close() error {
	return nil
}

func NewAEADReader(r io.ReadSeeker, key []byte) (*AEADReader, error) {
	reader := &AEADReader{
		Reader:  r,
		backlog: new(bytes.Buffer),
	}

	if err := reader.initAeadCommon(key); err != nil {
		return nil, err
	}

	reader.packBuf = make([]byte, 0, maxPackSize+reader.aead.Overhead())
	return reader, nil
}

////////////

type AEADWriter struct {
	// Common fields with AEADReader
	aeadCommon

	// Internal Writer we would write to.
	Writer io.Writer

	// A buffer that is maxPackSize big.
	// Used for caching blocks
	packBuf *bytes.Buffer
}

func inc(arr []byte) {
	i := uint64(0)
	for idx, val := range arr {
		i += uint64(val) << (8 * uint64(idx))
	}
	i++
	for idx, _ := range arr {
		arr[idx] = byte(i % 256)
		i = i >> 8
		if i == 0 {
			break
		}
	}
}

func (w *AEADWriter) Write(p []byte) (int, error) {
	w.packBuf.Write(p)
	if w.packBuf.Len() < maxPackSize {
		return 0, nil
	}

	return w.flushPack(maxPackSize)
}

func (w *AEADWriter) flushPack(chunkSize int) (int, error) {
	// Try to update the checksum as we run:
	src := w.packBuf.Bytes()[:chunkSize]

	// Make sure to advance this many bytes
	// in case any leftovers are in the buffer.
	defer w.packBuf.Read(src[:chunkSize])

	if _, err := w.Hasher.Write(src); err != nil {
		return 0, err
	}

	// Create a new Nonce for this block:
	inc(w.nonce)

	// Encrypt the text:
	encrypted := w.aead.Seal(nil, w.nonce, src, nil)

	// Pass it to the underlying writer:
	binary.BigEndian.PutUint32(w.sizeBuf, uint32(len(encrypted)))

	w.Writer.Write(w.sizeBuf)
	w.Writer.Write(w.nonce)
	w.Writer.Write(encrypted)

	// len(encrypted) might be more than len(w.packBuf)
	return len(encrypted) + len(w.nonce) + len(w.sizeBuf), nil
}

func (w *AEADWriter) Close() error {
	_, err := w.flushPack(w.packBuf.Len())
	return err
}

func NewAEADWriter(w io.Writer, key []byte) (*AEADWriter, error) {
	writer := &AEADWriter{
		Writer:  w,
		packBuf: bytes.NewBuffer(make([]byte, 0, maxPackSize)),
	}

	if err := writer.initAeadCommon(key); err != nil {
		return nil, err
	}

	return writer, nil
}

func securedTransfer(reader io.Reader, writer io.Writer, encrypt bool) error {
	r := bufio.NewReader(reader)
	buf := make([]byte, 0, goodBufSize)

	for {
		n, err := r.Read(buf[:cap(buf)])
		buf = buf[:n]
		if n == 0 {
			if err == nil {
				continue
			}
			if err == io.EOF {
				break
			}

			return err
		}

		if err != nil && err != io.EOF {
			return err
		}

		_, err = writer.Write(buf)
		if err != nil {
			return err
		}
	}

	return nil
}

func Encrypt(key []byte, source io.ReadSeeker, dest io.Writer) error {
	layer, err := NewAEADWriter(dest, key)
	if err != nil {
		return err
	}

	defer layer.Close()
	return securedTransfer(source, layer, true)
}

func Decrypt(key []byte, source io.ReadSeeker, dest io.Writer) error {
	layer, err := NewAEADReader(source, key)
	if err != nil {
		return err
	}

	defer layer.Close()
	return securedTransfer(layer, dest, true)
}

func main() {
	decryptMode := flag.Bool("d", false, "Decrypt.")
	flag.Parse()

	key := []byte("01234567890ABCDE01234567890ABCDE")

	var err error
	if *decryptMode == false {
		err = Encrypt(key, os.Stdin, os.Stdout)
	} else {
		source, _ := os.Open("/tmp/dump")

		err = Decrypt(key, source, os.Stdout)
	}

	if err != nil {
		log.Fatal(err)
		return
	}
}
