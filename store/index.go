package store

import (
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/boltdb/bolt"
	"github.com/disorganizer/brig/util/trie"
	multihash "github.com/jbenet/go-multihash"
)

var (
	bucketIndex = []byte("index")
)

// Store is responsible for adding & retrieving all files from ipfs,
// while managing their metadata in a boltDB.
type Store struct {
	db *bolt.DB

	// Trie models the directory tree.
	// The root node is the repository root.
	Trie trie.Trie
}

// FileSize is a large enough integer for file sizes, offering a few util methods.
type FileSize int64

// File represents a single file in the repository.
// It stores all metadata about it and links to the actual data.
type File struct {
	// Pointer for dynamic loading of bigger data:
	*trie.Node
	s *Store

	Size     FileSize
	Hash     multihash.Multihash
	IpfsHash multihash.Multihash

	// Size, modtime etc.
}

// Open loads an existing store, if it does not exist, it is created.
func Open(repoPath string) (*Store, error) {
	options := &bolt.Options{Timeout: 1 * time.Second}
	db, err := bolt.Open(filepath.Join(repoPath, "index.bolt"), 0600, options)

	if err != nil {
		return nil, err
	}

	store := &Store{
		db: db,
	}

	// Create initial buckets:
	db.Update(func(tx *bolt.Tx) error {
		for _, name := range []string{"index", "commits", "pinned"} {
			_, err := tx.CreateBucketIfNotExists([]byte(name))
			if err != nil {
				return fmt.Errorf("create bucket: %s", err)
			}
		}
		return nil
	})

	if err := store.load(); err != nil {
		return nil, err
	}

	return store, nil
}

// Add reads data from r, encrypts & compresses it while feeding it to ipfs.
// The resulting hash will be committed to the index.
func (s *Store) Add(path string, r io.Reader) error {
	// gets hash, size, modtime=now, ipfshash...
	// creates File{} and serializes it to DB using GOB
	// unsert .Node before serializing, set again after.
	return nil
}

// Close syncs all data. It is an error to use the store afterwards.
func (s *Store) Close() {
	s.db.Sync()
	s.db.Close()
}

// Load opens the
func (s *Store) load() error {
	s.Trie = trie.NewTrie()

	return s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketIndex)
		if bucket != nil {
			return nil
		}

		err := bucket.ForEach(func(k, v []byte) error {
			// k = absPath, v = File{} value
			file := &File{s: s}
			file.Node = s.Trie.Insert(string(k))
			return nil
		})

		return fmt.Errorf("store-load: %v", err)
	})
}
