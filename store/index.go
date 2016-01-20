package store

import (
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/disorganizer/brig/util/ipfsutil"
	"github.com/disorganizer/brig/util/trie"
)

var (
	ErrNoSuchFile = fmt.Errorf("No such file or directory")
)

// Store is responsible for adding & retrieving all files from ipfs,
// while managing their metadata in a boltDB.
type Store struct {
	db *bolt.DB

	// Trie models the directory tree.
	// The root node is the repository root.
	Trie trie.Trie

	// IpfsCtx holds information how and where to reach
	// the ipfs daemon process.
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

// Mkdir creates a new, empty directory.
// If the directory already exists, this is a NOOP.
func (s *Store) Mkdir(repoPath string) error {
	dir, err := NewDir(s, repoPath)
	if err != nil {
		return err
	}

	return s.marshalFile(dir, repoPath)
}

func (s *Store) Add(filePath, repoPath string) error {
	// TODO: Explain this "trick"
	return s.AddDir(filePath, repoPath)
}

func (s *Store) AddDir(filePath, repoPath string) error {
	err := filepath.Walk(filePath, func(path string, info os.FileInfo, err error) error {
		// Simply skip errorneous files:
		if err != nil {
			log.Warningf("Walk: %v", err)
			return err
		}

		// Map the file path relative to repoPath:
		currPath := filepath.Join(repoPath, path[len(filePath):])

		switch mode := info.Mode(); {
		case mode.IsRegular():
			fd, err := os.Open(path)
			if err != nil {
				return err
			}

			err = s.AddFromReader(currPath, fd)
		case mode.IsDir():
			err = s.Mkdir(currPath)
		default:
			log.Warningf("Recursive add: Ignoring weird file: %v")
			return nil
		}

		if err != nil {
			log.WithFields(log.Fields{
				"file_path": filePath,
				"repo_path": repoPath,
				"curr_path": currPath,
			}).Warningf("AddDir: %v", err)
		}

		return nil
	})

	return err
}

// Add reads data from r, encrypts & compresses it while feeding it to ipfs.
// The resulting hash will be committed to the index.
func (s *Store) AddFromReader(repoPath string, r io.Reader) error {
	// Check if the file was already added:
	_, err := s.PathToFile(repoPath)

	log.Infof("bolt lookup: %v", err)

	if err != nil {
		if err != ErrNoSuchFile {
			return err
		}
	} else {
		// We know this file already.
		log.WithFields(log.Fields{
			"file": repoPath,
		}).Info("Updating file.")

		// TODO: Write oldFile to commit history here,.
	}

	key := make([]byte, 32)
	n, err := rand.Reader.Read(key)
	if err != nil {
		return err
	}

	if n != 32 {
		return fmt.Errorf("Read less than desired key size: %v", n)
	}

	stream, err := NewFileReader(key, r)
	if err != nil {
		return err
	}

	hash, err := ipfsutil.Add(s.IpfsCtx, stream)
	if err != nil {
		return err
	}

	file, err := NewFile(s, repoPath, hash, key)
	if err != nil {
		return err
	}

	if err := s.marshalFile(file, repoPath); err != nil {
		return err
	}

	return nil
}

type BucketHandler func(tx *bolt.Tx, b *bolt.Bucket) error

// withBucket wraps a bolt handler closure and passes a named bucket
// as extra parameter. Error handling is done universally for convinience.
func withBucket(name string, handler BucketHandler) func(tx *bolt.Tx) error {
	return func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(name))
		if bucket == nil {
			return fmt.Errorf("Add: No bucket named `%s`", name)
		}

		return handler(tx, bucket)
	}
}

// marshalFile converts a file to a protobuf and
func (s *Store) marshalFile(file *File, repoPath string) error {
	return s.db.Update(withBucket("index", func(tx *bolt.Tx, bucket *bolt.Bucket) error {
		data, err := file.Marshal()
		if err != nil {
			return err
		}

		if err := bucket.Put([]byte(repoPath), data); err != nil {
			return err
		}

		return nil
	}))
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

	err := s.db.View(withBucket("index", func(tx *bolt.Tx, bucket *bolt.Bucket) error {
		data := bucket.Get([]byte(path))
		if data == nil {
			return ErrNoSuchFile
		}

		var err error
		if file, err = Unmarshal(s, data); err != nil {
			return err
		}

		return nil
	}))

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

// Rm will purge a file locally on this node.
func (s *Store) Rm(path string) error {
	node := s.Trie.Lookup(path)
	if node == nil {
		log.Errorf("Could not remove `%s` from trie.", path)
	} else {
		node.Remove()
	}

	// Remove from trie, remove from bolt db.
	return s.db.Update(withBucket("index", func(tx *bolt.Tx, bucket *bolt.Bucket) error {
		return bucket.Delete([]byte(path))
	}))
}
