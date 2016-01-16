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

	store.Trie = trie.NewTrie()
	return store, nil
}

// Add reads data from r, encrypts & compresses it while feeding it to ipfs.
// The resulting hash will be committed to the index.
func (s *Store) Add(filePath, repoPath string, r io.Reader) (multihash.Multihash, error) {
	file, err := NewFile(s, filePath)
	if err != nil {
		return nil, err
	}

	stream, err := NewFileReader(file.Key, r)
	if err != nil {
		return nil, err
	}

	hash, err := ipfsutil.Add(s.IpfsCtx, stream)
	if err != nil {
		return nil, err
	}

	file.Hash = hash

	err = s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketIndex)
		if bucket == nil {
			return fmt.Errorf("Add: No index bucket")
		}

		data, err := file.Marshal()
		if err != nil {
			return err
		}

		if err := bucket.Put([]byte(repoPath), data); err != nil {
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
	file, err := s.PathToFile(path)
	if err != nil {
		return err
	}

	fmt.Println("HASH", file.Hash.B58String())
	fmt.Printf("CAT KEY %x\n", file.Key)

	ipfsStream, err := ipfsutil.Cat(s.IpfsCtx, file.Hash)
	if err != nil {
		return err
	}
	defer ipfsStream.Close()

	cleanStream, err := NewIpfsReader(file.Key, ipfsStream)
	if err != nil {
		return err
	}

	if _, err := io.Copy(w, cleanStream); err != nil {
		return err
	}

	return nil
}

func (s *Store) PathToFile(path string) (*File, error) {
	var file *File

	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketIndex)
		if bucket == nil {
			return fmt.Errorf("PathToFile: No index bucket")
		}

		data := bucket.Get([]byte(path))
		if data == nil {
			return fmt.Errorf("cat: no file to path `%s`", path)
		}

		var err error
		if file, err = Unmarshal(s, data); err != nil {
			return err
		}

		return nil
	})

	return file, err
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
