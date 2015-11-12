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
	"runtime/pprof"
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
	aeadCipherChaCha = iota
	aeadCipherAES
)

const (
	maxPackSize       = 4 * 1024 * 1024
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

type AEADReader struct {
	Reader io.ReadSeeker

	Hasher hash.Hash
	aead   cipher.AEAD

	nonce   []byte
	sizeBuf []byte
	packBuf []byte
	backlog *bytes.Buffer
}

func (r *AEADReader) Read(dest []byte) (int, error) {
	if r.backlog.Len() > 0 {
		n, _ := r.backlog.Read(dest)
		return n, nil
	}

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

	// This is the counterpart to above:
	// Log back any parts that do not fit into `dest`.
	copySize := len(dest)
	if len(dest) > len(decrypted) {
		copySize = len(decrypted)
	} else if len(dest) < len(decrypted) {
		r.backlog.Write(decrypted[copySize:])
		log.Println("CACHE")
	}

	// return copy(dest, decrypted[:copySize]), nil
	return copy(dest, decrypted), nil
}

func (r *AEADReader) Seek(offset int64, whence int) (int64, error) {
	// TODO: clear backlog?
	r.backlog.Reset()

	currPos, _ := r.Reader.Seek(0, os.SEEK_CUR)

	// find previous block
	blockSize := maxPackSize + padPackSize + r.aead.Overhead() + r.aead.NonceSize()
	blockPos := currPos / int64(blockSize)

	r.Reader.Seek(blockPos, os.SEEK_SET)

	// TODO: read currPos - blockPos
	return blockPos, nil
}

func (r *AEADReader) Close() error {
	return nil
}

func NewAEADReader(r io.ReadSeeker, key []byte) (io.ReadCloser, error) {
	aead, err := createAEADWorker(defaultCipherType, key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aead.NonceSize())
	return &AEADReader{
		Hasher:  sha1.New(),
		Reader:  r,
		nonce:   nonce,
		aead:    aead,
		backlog: new(bytes.Buffer),
		sizeBuf: make([]byte, padPackSize),
		packBuf: make([]byte, 0, maxPackSize+aead.Overhead()),
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

	// A buffer that is maxPackSize big.
	// Used for caching blocks
	packBuf *bytes.Buffer

	// For more information, see:
	// https://en.wikipedia.org/wiki/Authenticated_encryption
	aead cipher.AEAD
}

func (w *AEADWriter) Write(p []byte) (int, error) {
	w.packBuf.Write(p)
	if w.packBuf.Len() < maxPackSize {
		return 0, nil
	}

	return w.flushPack(maxPackSize)
}

func (w *AEADWriter) Close() error {
	_, err := w.flushPack(w.packBuf.Len())
	return err
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
	// We do not want to make the nonce be random
	// so we don't skip deduplication of the encrypted data.
	hash := w.Hasher.Sum(nil)
	copy(w.nonce, hash[w.aead.NonceSize():])

	// Encrypt the text:
	encrypted := w.aead.Seal(nil, w.nonce, src, nil)

	// Pass it to the underlying writer:
	binary.BigEndian.PutUint32(w.sizeBuf, uint32(len(encrypted)))

	w.Writer.Write(w.sizeBuf)
	w.Writer.Write(w.nonce)
	w.Writer.Write(encrypted)

	// log.Println("WROTE", len(encrypted), len(src))

	// len(encrypted) might be more than len(w.packBuf)
	return len(encrypted) + len(w.nonce) + len(w.sizeBuf), nil
}

func NewAEADWriter(w io.Writer, key []byte) (io.WriteCloser, error) {
	aead, err := createAEADWorker(defaultCipherType, key)
	if err != nil {
		return nil, err
	}

	return &AEADWriter{
		Hasher:  sha1.New(),
		Writer:  w,
		aead:    aead,
		nonce:   make([]byte, aead.NonceSize()),
		sizeBuf: make([]byte, padPackSize),
		packBuf: bytes.NewBuffer(make([]byte, 0, maxPackSize)),
	}, nil
}

func main() {
	var reader io.ReadCloser
	var writer io.WriteCloser
	var err error
	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	decryptMode := flag.Bool("d", false, "Decrypt.")
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	key := []byte("01234567890ABCDE01234567890ABCDE")

	if *decryptMode == false {
		writer, err = NewAEADWriter(os.Stdout, key)
		reader = os.Stdin
		defer writer.Close()
	} else {
		writer = os.Stdout
		reader, err = NewAEADReader(os.Stdin, key)
		defer reader.Close()
	}

	if err != nil {
		log.Fatal(err)
		return
	}

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
			log.Fatal(err)
		}

		// log.Println("READ ", n)

		if err != nil && err != io.EOF {
			log.Fatal(err)
		}

		_, err = writer.Write(buf)
		if err != nil {
			log.Fatal(err)
		}
	}
}
