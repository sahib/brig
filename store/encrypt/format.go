// Package encrypt implements the encryption layer of brig.
// The file format used looks something like this:
//
// [HEADER][[BLOCKHEADER][PAYLOAD]...]
//
// HEADER is 28 bytes big and contains the following fields:
//    - 8 Byte: Magic number (to identify non-brig files quickly)
//    - 2 Byte: Format version
//    - 1 Byte: Used cipher type (ChaCha20 or AES-GCM)
//    - 1 Byte: Compression algorithm (currently 0 for none and 1 for snappy)
//    - 4 Byte: Key length in bytes.
//	  - 4 Byte: Maximum size of each block (last may be less)
//    - 8 Byte: MAC protecting the header from forgery
//
// BLOCKHEADER contains the following fields:
//    - 8 Byte: Nonce/Block number
//
// PAYLOAD contains the actual encrypted data, possibly with padding.
//
// All metadata is encoded in big endian.
//
// Reader/Writer are capable or reading/writing this format.
// Additionally, both support efficient seeking into the encrypted data,
// provided the underlying datastream supports seeking.
package encrypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"encoding/binary"
	"fmt"
	"io"

	chacha "github.com/codahale/chacha20poly1305"
	"golang.org/x/crypto/sha3"
)

// Possible ciphers in Counter mode:
const (
	aeadCipherChaCha = iota
	aeadCipherAES
)

// Other constants:
const (
	// Size of the header mac:
	macSize = 8

	// Size of the initial header:
	headerSize = 20 + macSize

	// Chacha20 appears to be twice as fast as AES-GCM on my machine
	defaultCipherType = aeadCipherChaCha

	// MaxBlockSize is the maximum number of bytes a single payload may have
	MaxBlockSize = 64 * 1024

	// GoodEncBufferSize is the recommended size of buffers
	GoodEncBufferSize = MaxBlockSize + 32

	// GoodDecBufferSize is the recommended size of buffers
	GoodDecBufferSize = MaxBlockSize
)

var (
	// MagicNumber contains the first 8 byte of every brig header.
	// For various reasons, it is the ascii string "moosecat".
	MagicNumber = []byte{
		0x6d, 0x6f, 0x6f, 0x73,
		0x65, 0x63, 0x61, 0x74,
	}
)

// KeySize of the used cipher's key in bytes.
var KeySize = chacha.KeySize

////////////////////
// Header Parsing //
////////////////////

// Sometimes go is a bit weird:
func btoi(b bool) byte {
	if b {
		return 1
	}
	return 0
}

// GenerateHeader creates a valid header for the format file
func GenerateHeader(key []byte, compression bool) []byte {
	// This is in big endian:
	header := []byte{
		// Brigs magic number (8 Byte):
		0, 0, 0, 0, 0, 0, 0, 0,
		// File format version (2 Byte):
		0x0, 0x1,
		// Cipher type (1 Byte):
		defaultCipherType,
		// Cipher type (1 Byte):
		btoi(compression),
		// Key length (4 Byte):
		0, 0, 0, 0,
		// Block length (4 Byte):
		0, 0, 0, 0,
		// MAC Header (8 Byte):
		0, 0, 0, 0, 0, 0, 0, 0,
	}

	// Magic number:
	copy(header[:len(MagicNumber)], MagicNumber)

	// Encode key size:
	binary.BigEndian.PutUint32(header[12:16], uint32(KeySize))

	// Encode max block size:
	binary.BigEndian.PutUint32(header[16:20], uint32(MaxBlockSize))

	// This needs to be done last:
	headerMac := hmac.New(sha3.New224, key)
	if _, err := headerMac.Write(header[:20]); err != nil {
		return nil
	}

	shortHeaderMac := headerMac.Sum(nil)[:macSize]
	copy(header[20:28], shortHeaderMac)

	return header
}

// HeaderInfo represents a parsed header.
type HeaderInfo struct {
	// Version of the file format. Currently always 1.
	Version uint16
	// Cipher type used in the file.
	Cipher uint8
	// Compressed stores the compression algorithm id or 0.
	Compressed uint8
	// Keylen is the number of bytes in the encryption key.
	Keylen uint32
	// Blocklen is the max. number of bytes in a block.
	// The last block might be smaller.
	Blocklen uint32
}

// ParseHeader parses the header of the format file.
// Returns the format version, cipher type, keylength and block length. If
// parsing fails, an error is returned.
func ParseHeader(header, key []byte) (*HeaderInfo, error) {
	if bytes.Compare(header[:len(MagicNumber)], MagicNumber) != 0 {
		return nil, fmt.Errorf("Magic number in header differs")
	}

	version := binary.BigEndian.Uint16(header[8:10])
	cipher := uint8(header[10])
	switch cipher {
	case aeadCipherAES:
	case aeadCipherChaCha:
		// we support this!
	default:
		return nil, fmt.Errorf("Unknown cipher type: %d", cipher)
	}

	compressed := uint8(header[11])
	keylen := binary.BigEndian.Uint32(header[12:16])
	blocklen := binary.BigEndian.Uint32(header[16:20])

	if blocklen != MaxBlockSize {
		return nil, fmt.Errorf("Unsupported block length in header: %d", blocklen)
	}

	// Check the header mac:
	headerMac := hmac.New(sha3.New224, key)
	if _, err := headerMac.Write(header[:20]); err != nil {
		return nil, err
	}

	storedMac := header[20:28]
	shortHeaderMac := headerMac.Sum(nil)[:macSize]
	if !hmac.Equal(shortHeaderMac, storedMac) {
		fmt.Printf("%x %x\n", storedMac, shortHeaderMac)
		return nil, fmt.Errorf("Header MAC differs from expected.")
	}

	return &HeaderInfo{
		Version:    version,
		Cipher:     cipher,
		Keylen:     keylen,
		Blocklen:   blocklen,
		Compressed: compressed,
	}, nil
}

//////////////////////
// Common Utilities //
//////////////////////

func createAEADWorker(cipherType uint8, key []byte) (cipher.AEAD, error) {
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
	// Nonce that form the first aead.NonceSize() bytes
	// of the output
	nonce []byte

	// Key used for encryption/decryption
	key []byte

	// For more information, see:
	// https://en.wikipedia.org/wiki/Authenticated_encryption
	aead cipher.AEAD

	// Buffer for encrypted data (MaxBlockSize + overhead)
	encBuf []byte

	// Buffer for decrypted data (MaxBlockSize)
	decBuf []byte
}

func (c *aeadCommon) initAeadCommon(key []byte, cipherType uint8) error {
	aead, err := createAEADWorker(cipherType, key)
	if err != nil {
		return err
	}

	c.nonce = make([]byte, aead.NonceSize())
	c.aead = aead
	c.key = key

	c.encBuf = make([]byte, 0, MaxBlockSize+aead.Overhead())
	c.decBuf = make([]byte, 0, MaxBlockSize)

	return nil
}

// Encrypt is a utility function which encrypts the data from source with key
// and writes the resulting encrypted data to dest.
func Encrypt(key []byte, source io.Reader, dest io.Writer) (n int64, outErr error) {
	layer, err := NewWriter(dest, key, false)
	if err != nil {
		return 0, err
	}

	defer func() {
		if err := layer.Close(); outErr != nil && err != nil {
			outErr = err
		}
	}()

	return io.CopyBuffer(layer, source, make([]byte, GoodEncBufferSize))
}

// Decrypt is a utility function which decrypts the data from source with key
// and writes the resulting encrypted data to dest.
func Decrypt(key []byte, source io.Reader, dest io.Writer) (n int64, outErr error) {
	layer, err := NewReader(source, key)
	if err != nil {
		return 0, err
	}

	defer func() {
		if err := layer.Close(); outErr != nil && err != nil {
			outErr = err
		}
	}()

	return io.CopyBuffer(dest, layer, make([]byte, GoodDecBufferSize))
}
