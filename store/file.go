package store

import (
	"github.com/disorganizer/brig/store/proto"
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

	IsFile bool
	Size   FileSize

	Key  []byte
	Hash multihash.Multihash
}

// New returns a file inside a repo.
// Path is relative to the repo root.
func NewFile(store *Store, path string, hash multihash.Multihash, key []byte) (*File, error) {
	node := store.Trie.Insert(path)

	return &File{
		store:  store,
		Node:   node,
		Size:   FileSize(0), // TODO: Read from outside?
		Key:    key,
		Hash:   hash,
		IsFile: true,
	}, nil
}

func NewDir(store *Store, path string) (*File, error) {
	node := store.Trie.Insert(path)

	return &File{
		store: store,
		Node:  node,
	}, nil
}

func (f *File) Marshal() ([]byte, error) {
	dataFile := &proto.File{
		Path:     protobuf.String(f.Path()),
		Key:      f.Key,
		FileSize: protobuf.Int64(int64(f.Size)),
		IsFile:   protobuf.Bool(f.IsFile),
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
		store:  store,
		Node:   node,
		IsFile: dataFile.GetIsFile(),
		Size:   FileSize(dataFile.GetFileSize()),
		Key:    dataFile.GetKey(),
		Hash:   dataFile.GetHash(),
	}, nil
}
