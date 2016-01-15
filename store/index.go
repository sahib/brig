package store

import (
	"fmt"
	"io"
	"path/filepath"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/disorganizer/brig/util/ipfsutil"
	"github.com/disorganizer/brig/util/trie"
	"github.com/jbenet/go-multihash"
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

	IpfsCtx *ipfsutil.Context
}

// Load opens the
func (s *Store) loadTrie() error {
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

		if err == nil {
			return nil
		}

		return fmt.Errorf("store-load: %v", err)
	})
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
		IpfsCtx: &ipfsutil.Context{
			Path: filepath.Join(repoPath, "ipfs"),
		},
	}

	// Create initial buckets:
	err = db.Update(func(tx *bolt.Tx) error {
		for _, name := range []string{"index", "commits", "pinned"} {
			if _, berr := tx.CreateBucketIfNotExists([]byte(name)); berr != nil {
				return fmt.Errorf("create bucket: %s", berr)
			}
		}
		return nil
	})

	if err != nil {
		log.Warningf("store-create-table failed: %v", err)
	}

	if err := store.loadTrie(); err != nil {
		return nil, err
	}

	return store, nil
}

var TestKey = []byte("01234567890ABCDE01234567890ABCDE")

// Add reads data from r, encrypts & compresses it while feeding it to ipfs.
// The resulting hash will be committed to the index.
func (s *Store) Add(path string, r io.Reader) (multihash.Multihash, error) {
	// TODO
	// gets hash, size, modtime=now, ipfshash...
	// creates File{} and serializes it to DB using GOB
	// insert .Node before serializing, set again after.
	stream, err := NewFileReader(TestKey, r)
	if err != nil {
		return nil, err
	}

	hash, err := ipfsutil.Add(s.IpfsCtx, stream)
	if err != nil {
		return nil, err
	}

	err = s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketIndex)
		if bucket == nil {
			return fmt.Errorf("Add: No index bucket")
		}

		if err := bucket.Put([]byte(path), hash); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return hash, nil
}

func (s *Store) Cat(path string, w io.Writer) error {
	hash, err := s.PathToHash(path)
	if err != nil {
		return err
	}

	fmt.Println("HASH", hash.B58String())

	ipfsStream, err := ipfsutil.Cat(s.IpfsCtx, hash)
	if err != nil {
		return err
	}
	defer ipfsStream.Close()

	cleanStream, err := NewIpfsReader(TestKey, ipfsStream)
	if err != nil {
		return err
	}

	if _, err := io.Copy(w, cleanStream); err != nil {
		return err
	}

	return nil
}

func (s *Store) PathToHash(path string) (multihash.Multihash, error) {
	var hash multihash.Multihash

	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketIndex)
		if bucket == nil {
			return fmt.Errorf("PathToHash: No index bucket")
		}

		foundHash := bucket.Get([]byte(path))
		if foundHash == nil {
			return fmt.Errorf("cat: no hash to path `%s`", path)
		}

		hash = make([]byte, len(foundHash))
		copy(hash, foundHash)
		return nil
	})

	return hash, err
}

// Close syncs all data. It is an error to use the store afterwards.
func (s *Store) Close() {
	if err := s.db.Sync(); err != nil {
		log.Warningf("store-sync: %v", err)
	}

	if err := s.db.Close(); err != nil {
		log.Warningf("store-close: %v", err)
	}
}

// Remove will purge a file locally on this node.
// The file might still be available somewhere else.
func (s *Store) Remove(path string) error {
	return nil
}
