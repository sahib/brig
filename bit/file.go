// package bit
package main

import (
	"bufio"
	"crypto/cipher"
	"crypto/sha512"
	"encoding/binary"
	"flag"
	"fmt"
	"hash"
	"io"
	"log"
	"os"
	"time"

	chacha "github.com/codahale/chacha20poly1305"
	"github.com/jbenet/go-multihash"
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

const (
	maxPackSize = 16 * 1024 * 1024
	padPackSize = 4
)

type AEADReader struct {
	Reader io.Reader

	Hasher hash.Hash
	aead   cipher.AEAD

	nonce   []byte
	sizeBuf []byte
	packBuf []byte
}

func (r *AEADReader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(r.sizeBuf)
	if err != nil {
		return 0, err
	}

	packSize := binary.BigEndian.Uint32(r.sizeBuf)
	if packSize > maxPackSize {
		return 0, fmt.Errorf("Pack size exceeded expectations, ignore pack.")
	}

	n, err = r.Reader.Read(r.nonce)
	if err != nil {
		return 0, err
	}

	if n != r.aead.NonceSize() {
		return 0, fmt.Errorf("Nonce size mismatch. Should: %d. Have: %d",
			r.aead.NonceSize(), n)
	}

	// Read encrypted text, but not more than the packet indicated:
	// Since the packSize might be larger than cap(p) we use another buffer.
	if r.packBuf == nil || packSize > uint32(cap(r.packBuf)) {
		r.packBuf = make([]byte, packSize)
	}

	n, err = r.Reader.Read(r.packBuf[:packSize])
	if err != nil {
		return 0, err
	}

	// Decrypt text (store it in decrypted)
	// TODO: This might be slow since it involves two buffers?
	decrypted, err := r.aead.Open(p, r.nonce, r.packBuf[:packSize], nil)
	if err != nil {
		return 0, err
	}

	_, err = r.Hasher.Write(decrypted)
	if err != nil {
		return 0, err
	}

	return copy(p, decrypted), nil
}

func NewAEADReader(r io.Reader, key []byte) (io.Reader, error) {
	aead, err := chacha.New(key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aead.NonceSize())
	return &AEADReader{
		Hasher:  sha512.New(),
		Reader:  r,
		nonce:   nonce,
		aead:    aead,
		sizeBuf: make([]byte, padPackSize),
	}, nil
}

////////////

type AEADWriter struct {
	// Internal Writer we would write to.
	Writer io.Writer

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

func (w *AEADWriter) Write(p []byte) (int, error) {
	if maxPackSize <= padPackSize+w.aead.NonceSize()+len(p)+w.aead.Overhead() {
		log.Println("too large")
		return 0, fmt.Errorf("Maximum chunk size is 16MB")
	}

	// Try to update the checksum as we run:
	_, err := w.Hasher.Write(p)
	if err != nil {
		return 0, err
	}

	// Create a new Nonce for this block:
	// We do not want to make the nonce be random
	// so we don't skip deduplication of the encrypted data.
	hash := w.Hasher.Sum(nil)
	copy(w.nonce, hash[w.aead.NonceSize():])

	// Encrypt the text:
	encrypted := w.aead.Seal(nil, w.nonce, p, nil)

	// Pass it to the underlying writer:
	binary.BigEndian.PutUint32(w.sizeBuf, uint32(len(encrypted)))
	w.Writer.Write(w.sizeBuf)
	w.Writer.Write(w.nonce)
	w.Writer.Write(encrypted)

	// len(encrypted) might be more than len(p)
	return len(encrypted) + len(w.nonce) + len(w.sizeBuf), nil
}

func NewAEADWriter(w io.Writer, key []byte) (io.Writer, error) {
	aead, err := chacha.New(key)
	if err != nil {
		return nil, err
	}

	return &AEADWriter{
		Hasher:  sha512.New(),
		Writer:  w,
		aead:    aead,
		nonce:   make([]byte, aead.NonceSize()),
		sizeBuf: make([]byte, padPackSize),
	}, nil
}

func main() {
	var reader io.Reader
	var writer io.Writer
	var err error

	decryptMode := flag.Bool("d", false, "Decrypt.")
	flag.Parse()

	key := []byte("01234567890ABCDE01234567890ABCDE")

	if *decryptMode == false {
		writer, err = NewAEADWriter(os.Stdout, key)
		reader = os.Stdin
	} else {
		writer = os.Stdout
		reader, err = NewAEADReader(os.Stdin, key)
	}

	if err != nil {
		log.Fatal(err)
		return
	}

	r := bufio.NewReader(reader)
	buf := make([]byte, 0, 4*1024*1024)

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
			log.Fatal(err)
		}

		if err != nil && err != io.EOF {
			log.Fatal(err)
		}

		_, err = writer.Write(buf)
		if err != nil {
			log.Fatal(err)
		}
	}
}
