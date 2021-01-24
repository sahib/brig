// Package encrypt implements the encryption layer of brig.
// The file format used looks something like this:
//
// [HEADER][[BLOCKHEADER][PAYLOAD]...]
//
// HEADER is 20+16 bytes big and contains the following fields:
//    -   8 Byte: Magic number (to identify non-brig files quickly)
//    -   4 Byte: Flags (describing the stream)
//    -   2 Byte: Key length in bytes
//    -   2 Byte: Reserved for future use.
//	  -   4 Byte: Maximum size of each block (last may be less)
//    -  16 Byte: MAC protecting the header from forgery
//
// BLOCKHEADER contains the following fields:
//    - 8 Byte: Nonce: Derived from the current block number.
//                     The block number is checked to be correct on decryption.
//
// PAYLOAD contains the actual encrypted data, which includes a MAC at the end.
// The size of the MAC depends on the algorithm, for poly1305 it's 16 bytes.
//
// All header metadata is encoded in little endian.
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
	"errors"
	"fmt"
	"io"

	"github.com/sahib/brig/catfs/mio/compress"
	chacha "golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/sha3"
)

// Flags indicate with what options a stream was encoded.
// Some flags are not compatible to each other, see below.
type Flags int32

// Possible ciphers in Counter mode:
const (
	// FlagEmpty is invalid
	FlagEmpty = Flags(0)

	// FlagEncryptAES256GCM indicates the stream was encrypted with AES256 in GCM mode.
	// This should be fast on modern CPUs.
	FlagEncryptAES256GCM = Flags(1) << iota

	// FlagEncryptChaCha20 incidate that the stream was encrypted with ChaCha20.
	// This can be a good choice if your CPU does not support the AES-NI instruction set.
	FlagEncryptChaCha20

	// FlagCompressedInside indicates that the encrypted data was also compressed.
	// This can be used to decide at runtime what streaming is needed.
	FlagCompressedInside
)

// Other constants:
const (
	// Size of the header mac:
	macSize = 16

	// current file format version, increment on incompatible changes.
	version = 1

	// Size of the initial header:
	headerSize = 20 + macSize

	// Default maxBlockSize if not set
	defaultMaxBlockSize = 64 * 1024

	defaultDecBufferSize = defaultMaxBlockSize
	defaultEncBufferSize = defaultMaxBlockSize + 40
)

var (
	// MagicNumber contains the first 8 byte of every brig header.
	// For various reasons, it is the ascii string "moosecat".
	MagicNumber = []byte{
		0x6d, 0x6f, 0x6f, 0x73,
		0x65, 0x63, 0x61, 0x74,
	}
)

////////////////////
// Header Parsing //
////////////////////

// GenerateHeader creates a valid header for the format file
func GenerateHeader(key []byte, maxBlockSize int64, flags Flags) []byte {
	// This is in big endian:
	header := []byte{
		// magic number (8 Byte):
		0, 0, 0, 0, 0, 0, 0, 0,
		// Flags (4 byte):
		0, 0, 0, 0,
		// Key length (4 Byte):
		0, 0, 0, 0,
		// Block length (4 Byte):
		0, 0, 0, 0,
		// MAC Header (16 Byte):
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
	}

	// Magic number:
	copy(header[:len(MagicNumber)], MagicNumber)
	binary.LittleEndian.PutUint32(header[8:12], uint32(flags))

	// Encode key size (static at the moment):
	binary.LittleEndian.PutUint32(header[12:16], uint32(32))

	// Encode max block size:
	binary.LittleEndian.PutUint32(header[16:20], uint32(maxBlockSize))

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
	CipherBit Flags

	// KeyLen is the number of bytes in the encryption key.
	KeyLen uint32

	// BlockLen is the max. number of bytes in a block.
	// The last block may be smaller.
	BlockLen uint32

	Flags Flags
}

var (
	ErrSmallHeader        = errors.New("header is too small")
	ErrBadMagic           = errors.New("magic number missing")
	ErrBadFlags           = errors.New("inconsistent header flags")
	ErrBadHeaderMAC       = errors.New("header mac differs from expected")
	ErrIsCompressionMagic = errors.New("stream starts with compression magic number")
)

func cipherTypeBitFromFlags(flags Flags) (Flags, error) {
	var cipherBit Flags
	var bits = []Flags{
		FlagEncryptAES256GCM,
		FlagEncryptChaCha20,
	}

	for _, bit := range bits {
		if flags&bit == 0 {
			continue
		}

		if cipherBit != 0 {
			// only one bit at the same time allowed.
			return 0, ErrBadFlags
		}

		cipherBit = bit
	}

	if cipherBit == 0 {
		// no algorithm set: also error out.
		return 0, ErrBadFlags
	}

	return cipherBit, nil
}

// ParseHeader parses the header of the format file. Returns the flags, key
// and block length. If parsing fails, an error is returned.
func ParseHeader(header, key []byte) (*HeaderInfo, error) {
	// TODO: document assumption that len(MagicNumber) == len(compress.MagicNumber)
	if len(header) < len(MagicNumber) {
		return nil, ErrSmallHeader
	}

	if bytes.Compare(header[:len(MagicNumber)], MagicNumber) != 0 {
		if bytes.Compare(header[:len(compress.MagicNumber)], compress.MagicNumber) == 0 {
			return nil, ErrIsCompressionMagic
		}

		return nil, ErrBadMagic
	}

	if len(header) < headerSize {
		return nil, ErrSmallHeader
	}

	flags := Flags(binary.LittleEndian.Uint32(header[8:12]))
	keyLen := binary.LittleEndian.Uint32(header[12:16])
	blockLen := binary.LittleEndian.Uint32(header[16:20])

	cipherBit, err := cipherTypeBitFromFlags(flags)
	if err != nil {
		return nil, err
	}

	// Check the header mac:
	headerMac := hmac.New(sha3.New224, key)
	if _, err := headerMac.Write(header[:headerSize-macSize]); err != nil {
		return nil, err
	}

	storedMac := header[headerSize-macSize : headerSize]
	shortHeaderMac := headerMac.Sum(nil)[:macSize]
	if !hmac.Equal(shortHeaderMac, storedMac) {
		return nil, ErrBadHeaderMAC
	}

	return &HeaderInfo{
		Version:   version,
		CipherBit: cipherBit,
		KeyLen:    keyLen,
		BlockLen:  blockLen,
		Flags:     flags,
	}, nil
}

//////////////////////
// Common Utilities //
//////////////////////

func createAEADWorker(cipherType Flags, key []byte) (cipher.AEAD, error) {
	switch cipherType {
	case FlagEncryptAES256GCM:
		block, err := aes.NewCipher(key)
		if err != nil {
			return nil, err
		}
		return cipher.NewGCM(block)
	case FlagEncryptChaCha20:
		return chacha.New(key)
	default:
		return nil, fmt.Errorf("no such cipher type: %d", cipherType)
	}

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

	// Buffer for encrypted data (maxBlockSize + overhead)
	encBuf []byte
}

func (c *aeadCommon) initAeadCommon(key []byte, cipherBit Flags, maxBlockSize int64) error {
	aead, err := createAEADWorker(cipherBit, key)
	if err != nil {
		return err
	}

	c.encBuf = make([]byte, 0, maxBlockSize+int64(aead.Overhead()))
	c.nonce = make([]byte, aead.NonceSize())
	c.aead = aead
	c.key = key
	return nil
}

// Encrypt is a utility function which encrypts the data from source with key
// and writes the resulting encrypted data to dest.
func Encrypt(key []byte, source io.Reader, dest io.Writer, flags Flags) (int64, error) {
	layer, err := NewWriter(dest, key, flags)
	if err != nil {
		return 0, err
	}

	n, err := io.CopyBuffer(layer, source, make([]byte, defaultEncBufferSize))
	if err != nil {
		return n, err
	}

	return n, layer.Close()
}

// Decrypt is a utility function which decrypts the data from source with key
// and writes the resulting encrypted data to dest.
func Decrypt(key []byte, source io.Reader, dest io.Writer) (int64, error) {
	layer, err := NewReader(source, key)
	if err != nil {
		return 0, err
	}

	return io.CopyBuffer(dest, layer, make([]byte, defaultDecBufferSize))
}
