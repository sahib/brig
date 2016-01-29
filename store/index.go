package store

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/disorganizer/brig/util"
	"github.com/disorganizer/brig/util/ipfsutil"
)

var (
	ErrNoSuchFile = fmt.Errorf("No such file or directory")
)

// Store is responsible for adding & retrieving all files from ipfs,
// while managing their metadata in a boltDB.
type Store struct {
	db *bolt.DB

	// Root models the directory tree, aka Trie.
	// The root node is the repository root.
	Root *File

	// IpfsNode holds information how and where to reach
	// the ipfs daemon process.
	IpfsNode *ipfsutil.Node
}

// Open loads an existing store, if it does not exist, it is created.
func Open(repoPath string) (*Store, error) {
	options := &bolt.Options{Timeout: 1 * time.Second}
	db, err := bolt.Open(filepath.Join(repoPath, "index.bolt"), 0600, options)

	if err != nil {
		return nil, err
	}

	ipfsNode, err := ipfsutil.StartNode(filepath.Join(repoPath, "ipfs"))
	if err != nil {
		db.Close()
		return nil, err
	}

	store := &Store{
		db:       db,
		IpfsNode: ipfsNode,
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
		log.Warningf("store-create failed: %v", err)
	}
	fmt.Println("Done create")

	// Load all paths from the database into the trie:
	rootDir, err := newDirUnlocked(store, "/")
	if err != nil {
		return nil, err
	}

	store.Root = rootDir
	fmt.Println("Done newDir")

	err = db.View(withBucket("index", func(tx *bolt.Tx, bucket *bolt.Bucket) error {
		return bucket.ForEach(func(k []byte, v []byte) error {
			fmt.Println("Unmarshal")
			Unmarshal(store, v)
			fmt.Println("Unmarshal done")
			return nil
		})
	}))

	fmt.Println("Done open")
	return store, err
}

// Mkdir creates a new, empty directory.
// If the directory already exists, this is a NOOP.
func (s *Store) Mkdir(repoPath string) (*File, error) {
	dir, err := NewDir(s, repoPath)
	if err != nil {
		return nil, err
	}

	return dir, err
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
			defer fd.Close()

			info, err := os.Stat(path)
			if err != nil {
				return err
			}

			err = s.AddFromReader(currPath, fd, &Metadata{
				Size:    FileSize(info.Size()),
				ModTime: info.ModTime(),
			})
		case mode.IsDir():
			_, err = s.Mkdir(currPath)
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
func (s *Store) AddFromReader(repoPath string, r io.Reader, meta *Metadata) error {
	// Check if the file was already added:
	file := s.Root.Lookup(repoPath)

	log.Infof("bolt lookup: %v", file)

	if file != nil {
		// We know this file already.
		log.WithFields(log.Fields{
			"file": repoPath,
		}).Info("Updating file.")

		// TODO: Write oldFile to commit history here...
	} else {
		newFile, err := NewFile(s, repoPath, meta)
		if err != nil {
			return err
		}

		file = newFile
	}

	// Control how many bytes are written to the encryption layer:
	sizeAcc := &util.SizeAccumulator{}
	teeR := io.TeeReader(r, sizeAcc)

	stream, err := NewFileReader(file.Key, teeR)
	if err != nil {
		return err
	}

	hash, err := ipfsutil.Add(s.IpfsNode, stream)
	if err != nil {
		return err
	}

	log.Infof("ADD KEY:  %x", file.Key)
	log.Infof("ADD HASH: %s", hash.B58String())

	// Update metadata that might have changed:
	file.Lock()
	{
		file.Size = FileSize(sizeAcc.Size())
		file.ModTime = time.Now()
		file.Hash = hash
	}
	file.Unlock()

	if err := s.marshalFile(file, file.Path()); err != nil {
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
			return fmt.Errorf("index: No bucket named `%s`", name)
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

func (s *Store) Stream(path string) (ipfsutil.Reader, error) {
	file := s.Root.Lookup(path)
	if file == nil {
		return nil, ErrNoSuchFile
	}

	return file.Stream()
}

func (s *Store) Cat(path string, w io.Writer) error {
	cleanStream, err := s.Stream(path)
	if err != nil {
		return err
	}

	if _, err := io.Copy(w, cleanStream); err != nil {
		return err
	}

	return nil
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
	s.Root.Lock()
	defer s.Root.Unlock()

	node := s.Root.Lookup(path)
	if node == nil {
		log.Errorf("Could not remove `%s` from trie.", path)
	} else {
		node.Remove()
	}

	// Remove from trie & remove from bolt db.
	return s.db.Update(withBucket("index", func(tx *bolt.Tx, bucket *bolt.Bucket) error {
		return bucket.Delete([]byte(path))
	}))
}
