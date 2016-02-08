// Package encrypt implements the encryption layer of brig.
// The file format used looks something like this:
//
// [HEADER][[BLOCKHEADER][PAYLOAD]...]
//
// HEADER is 28 bytes big and contains the following fields:
//    -  8 Byte: Magic number (to identify non-brig files quickly)
//    -  2 Byte: Format version
//    -  2 Byte: Used cipher type (ChaCha20 or AES-GCM)
//    -  4 Byte: Key length in bytes.
//	  -  4 Byte: Maximum size of each block (last may be less)
//    TODO: Make that 8 byte, retardo.
//    - 10 Byte: Number of bytes passed to encryption (i.e. len of decrypted data)
//               This is needed to make SEEK_END work
//               (and also to make sure all data was decrypted)
//    -  8 Byte: MAC protecting the header from forgery
//
// BLOCKHEADER contains the following fields:
//    - 8 Byte: Nonce: Randomly generated, used as encryption seed.
//    - 8 Byte: Block Number: Needed to force block ordering.
//
// PAYLOAD contains the actual encrypted data, which includes a MAC.
// (The size of the MAC depends on the algorithm in use)
//
// All header metadata is encoded in big endian.
// TODO: Move to little endian, like rest of world?
//
// Reader/Writer are capable or reading/writing this format.  Additionally,
// Reader supports efficient seeking into the encrypted data, provided the
// underlying datastream supports seeking.  SEEK_END is only supported when the
// number of encrypted blocks is present in the header.
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
	headerSize = 30 + macSize

	// Chacha20 appears to be twice as fast as AES-GCM on my machine
	defaultCipherType = aeadCipherChaCha

	// MaxBlockSize is the maximum number of bytes a single payload may have
	MaxBlockSize = 64 * 1024

	// GoodEncBufferSize is the recommended size of buffers
	GoodEncBufferSize = MaxBlockSize + 40

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

// GenerateHeader creates a valid header for the format file
func GenerateHeader(key []byte, length int64) []byte {
	// This is in big endian:
	header := []byte{
		// Brigs magic number (8 Byte):
		0, 0, 0, 0, 0, 0, 0, 0,
		// File format version (2 Byte):
		0x0, 0x1,
		// Cipher type (2 Byte):
		0, defaultCipherType,
		// Key length (4 Byte):
		0, 0, 0, 0,
		// Block length (4 Byte):
		0, 0, 0, 0,
		// Length of input (10 Byte),
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		// MAC Header (8 Byte):
		0, 0, 0, 0, 0, 0, 0, 0,
	}

	// Magic number:
	copy(header[:len(MagicNumber)], MagicNumber)

	binary.BigEndian.PutUint16(header[10:12], uint16(defaultCipherType))

	// Encode key size:
	binary.BigEndian.PutUint32(header[12:16], uint32(KeySize))

	// Encode max block size:
	binary.BigEndian.PutUint32(header[16:20], uint32(MaxBlockSize))

	// Encode number of blocks:
	binary.PutVarint(header[20:30], length)

	// Calculate a MAC of the header; this needs to be done last:
	headerMac := hmac.New(sha3.New224, key)
	if _, err := headerMac.Write(header[:headerSize-macSize]); err != nil {
		return nil
	}

	// Copy the MAC to the output:
	shortHeaderMac := headerMac.Sum(nil)[:macSize]
	copy(header[headerSize-macSize:headerSize], shortHeaderMac)

	return header
}

// HeaderInfo represents a parsed header.
type HeaderInfo struct {
	// Version of the file format. Currently always 1.
	Version uint16
	// Cipher type used in the file.
	Cipher uint16
	// Keylen is the number of bytes in the encryption key.
	Keylen uint32
	// Blocklen is the max. number of bytes in a block.
	// The last block might be smaller.
	Blocklen uint32
	// Length is the number of bytes that were passed to the encryption.
	Length int64
}

// ParseHeader parses the header of the format file.
// Returns the format version, cipher type, keylength and block length. If
// parsing fails, an error is returned.
func ParseHeader(header, key []byte) (*HeaderInfo, error) {
	if bytes.Compare(header[:len(MagicNumber)], MagicNumber) != 0 {
		return nil, fmt.Errorf("Magic number in header differs")
	}

	version := binary.BigEndian.Uint16(header[8:10])
	cipher := binary.BigEndian.Uint16(header[10:12])
	switch cipher {
	case aeadCipherAES:
	case aeadCipherChaCha:
		// we support this!
	default:
		return nil, fmt.Errorf("Unknown cipher type: %d", cipher)
	}

	keylen := binary.BigEndian.Uint32(header[12:16])
	blocklen := binary.BigEndian.Uint32(header[16:20])

	if blocklen != MaxBlockSize {
		return nil, fmt.Errorf("Unsupported block length in header: %d", blocklen)
	}

	length, overflow := binary.Varint(header[20:30])
	if overflow <= 0 {
		return nil, fmt.Errorf("Block size is too big (overflow: %d)", overflow)
	}

	// Check the header mac:
	headerMac := hmac.New(sha3.New224, key)
	if _, err := headerMac.Write(header[:headerSize-macSize]); err != nil {
		return nil, err
	}

	storedMac := header[headerSize-macSize : headerSize]
	shortHeaderMac := headerMac.Sum(nil)[:macSize]
	if !hmac.Equal(shortHeaderMac, storedMac) {
		return nil, fmt.Errorf("Header MAC differs from expected.")
	}

	return &HeaderInfo{
		Version:  version,
		Cipher:   cipher,
		Keylen:   keylen,
		Blocklen: blocklen,
		Length:   length,
	}, nil
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
	// Nonce that form the first aead.NonceSize() bytes
	// of the output
	nonce []byte

	// Key used for encryption/decryption
	key []byte

	// Buffer to encode/decode the current blocknumber:
	blocknum []byte

	// For more information, see:
	// https://en.wikipedia.org/wiki/Authenticated_encryption
	aead cipher.AEAD

	// Buffer for encrypted data (MaxBlockSize + overhead)
	encBuf []byte
}

func (c *aeadCommon) initAeadCommon(key []byte, cipherType uint16) error {
	aead, err := createAEADWorker(cipherType, key)
	if err != nil {
		return err
	}

	c.nonce = make([]byte, aead.NonceSize())
	c.aead = aead
	c.key = key

	c.blocknum = make([]byte, 8)
	c.encBuf = make([]byte, 0, MaxBlockSize+aead.Overhead())
	return nil
}

// Encrypt is a utility function which encrypts the data from source with key
// and writes the resulting encrypted data to dest.
func Encrypt(key []byte, source io.Reader, dest io.Writer, size int64) (n int64, outErr error) {
	layer, err := NewWriter(dest, key, size)
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
