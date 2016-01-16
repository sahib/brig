package store

import (
	"bytes"
	"crypto/sha512"
	"encoding/binary"
	"io"
	"os"

	"github.com/disorganizer/brig/store/proto"
	"github.com/disorganizer/brig/util/security"
	"github.com/disorganizer/brig/util/trie"
	protobuf "github.com/gogo/protobuf/proto"
	"github.com/jbenet/go-multihash"
)

// File represents a single file in the repository.
// It stores all metadata about it and links to the actual data.
type File struct {
	// Pointer for dynamic loading of bigger data:
	*trie.Node
	store *Store

	Key  []byte
	Size FileSize
	Hash multihash.Multihash
}

// New returns a file inside a repo.
// Path is relative to the repo root.
func NewFile(store *Store, path string, hash multihash.Multihash) (*File, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	key, err := Fingerprint(path)
	if err != nil {
		return nil, err
	}

	node := store.Trie.Insert(path)

	return &File{
		store: store,
		Node:  node,
		Size:  FileSize(info.Size()),
		Key:   key,
		Hash:  hash,
	}, nil
}

func (f *File) Marshal() ([]byte, error) {
	dataFile := &proto.File{
		Path:     protobuf.String(f.Path()),
		Key:      f.Key,
		FileSize: protobuf.Int64(int64(f.Size)),
		Hash:     f.Hash,
	}

	data, err := protobuf.Marshal(dataFile)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func Unmarshal(store *Store, buf []byte) (*File, error) {
	dataFile := &proto.File{}
	if err := protobuf.Unmarshal(buf, dataFile); err != nil {
		return nil, err
	}

	node := store.Trie.Insert(dataFile.GetPath())

	return &File{
		store: store,
		Node:  node,
		Size:  FileSize(dataFile.GetFileSize()),
		Key:   dataFile.GetKey(),
		Hash:  dataFile.GetHash(),
	}, nil
}

// Fingerprint calculates an AES-Key from a file.
func Fingerprint(path string) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 8192)
	if _, err := io.Copy(bytes.NewBuffer(buf), fd); err != nil {
		return nil, err
	}

	sizeBuf := make([]byte, binary.MaxVarintLen64)
	binary.PutVarint(sizeBuf, info.Size())

	cksum := sha512.Sum512(buf)
	key := security.Scrypt(cksum[:], sizeBuf, 32)
	return key, nil
}
