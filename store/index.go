package store

import (
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/boltdb/bolt"
	"github.com/disorganizer/brig/util/trie"
)

var (
	BucketIndex = []byte("index")
)

type Store struct {
	db   *bolt.DB
	Trie trie.Trie
}

type File struct {
	// Pointer for dynamic loading of bigger data:
	*trie.Node
	s *Store

	Hash     []byte
	IpfsHash []byte

	// Size, modtime etc.
}

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

	if err := store.Load(); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *Store) Add(path string, r io.Reader) error {
	// gets hash, size, modtime=now, ipfshash...
	// creates File{} and serializes it to DB using GOB
	// unsert .Node before serializing, set again after.
	return nil
}

func (s *Store) Load() error {
	s.Trie = trie.NewTrie()

	return s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(BucketIndex)
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

func (s *Store) Close() {
	// For safety:
	s.db.Sync()

	s.db.Close()
}
