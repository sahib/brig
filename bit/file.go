package bit

import (
	"io"
	"os"
	"time"

	multihash "github.com/jbenet/go-multihash"
)

// Size of the file as large enough integer.
// Use this type to ensure you don't make mistakes on 32bit.
type FileSize int64

// Returns the size of the file when it's encrypted
// Encrpyted file are slightly larger than the source files.
func (s FileSize) EncrpytedSize() int64 {
	// TODO
	return int64(s)
}

// File represents a File that is managed by brig.
type File interface {
	// Path relative to the repo root
	Path() string

	// File size of the file in bytes
	Size() FileSize

	// Does this file represent a directory?
	IsDir() bool

	// Modification timestamp (with timezone)
	ModTime() time.Time

	// Hash of the unencrypted file
	Hash() multihash.Multihash

	// Hash of the encrypted file from IPFS
	IpfsHash() multihash.Multihash

	// Reader returns a valid EncryptedReader or an error.
	Reader(from io.Reader) (*EncryptedReader, error)

	// Writer returns a valid EncryptedReader or an error.
	Writer(to io.Writer) (*EncryptedWriter, error)
}

type fileBuf struct {
	key      []byte // TODO: pass to NewSourceFile?
	path     string
	size     FileSize
	modTime  time.Time
	isDir    bool
	hash     multihash.Multihash
	ipfsHash multihash.Multihash
}

func NewSourceFile(path string, key []byte) (File, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	return &fileBuf{
		path:    path,
		key:     key,
		modTime: info.ModTime(),
		size:    FileSize(info.Size()),
		isDir:   info.IsDir(),
	}, nil
}

func (f *fileBuf) Path() string {
	// TODO: Possible memory optimisation:
	//       Store paths in pathtricia trie.
	return f.path
}

func (f *fileBuf) Size() FileSize {
	return f.size
}

func (f *fileBuf) IsDir() bool {
	return f.isDir
}

func (f *fileBuf) ModTime() time.Time {
	return f.modTime
}

func (f *fileBuf) Hash() multihash.Multihash {
	return f.hash
}

func (f *fileBuf) IpfsHash() multihash.Multihash {
	return f.ipfsHash
}

func (f *fileBuf) Writer(to io.Writer) (*EncryptedWriter, error) {
	return NewEncryptedWriter(to, f.key)
}

func (f *fileBuf) Reader(from io.Reader) (*EncryptedReader, error) {
	return NewEncryptedReader(from, f.key)
}
