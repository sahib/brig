package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/disorganizer/brig/util"
	"github.com/disorganizer/brig/util/ipfsutil"
	"github.com/tsuibin/goxmpp2/xmpp"
)

var (
	// ErrNoSuchFile is returned whenever a path could not be resolved to a file.
	ErrNoSuchFile = fmt.Errorf("No such file or directory")
)

// Store is responsible for adding & retrieving all files from ipfs,
// while managing their metadata in a boltDB.
type Store struct {
	db *bolt.DB

	// Root models the directory tree, aka Trie.
	// The root node is the repository root.
	Root *File

	// Internal path of the repository.
	repoPath string

	// IpfsNode holds information how and where to reach
	// the ipfs daemon process.
	ipfsNode *ipfsutil.Node

	// xmpp connection wrapper
	XMPP *Connector
}

func (s *Store) IpfsNode() (*ipfsutil.Node, error) {
	if s.ipfsNode != nil {
		return s.ipfsNode, nil
	}

	// Start an offline node until we're fully connected.
	// Local operations will work with an offline mode too.
	ipfsNode, err := ipfsutil.StartNode(filepath.Join(s.repoPath, "ipfs"), false)
	if err != nil {
		return nil, err
	}

	s.ipfsNode = ipfsNode
	return ipfsNode, nil
}

// Open loads an existing store, if it does not exist, it is created.
// For full function, Connect() should be called afterwards.
func Open(repoPath string) (*Store, error) {
	options := &bolt.Options{Timeout: 1 * time.Second}
	db, err := bolt.Open(filepath.Join(repoPath, "index.bolt"), 0600, options)

	if err != nil {
		return nil, err
	}

	store := &Store{
		db:       db,
		repoPath: repoPath,
		XMPP:     NewConnector(repoPath),
	}

	// Create initial buckets:
	err = db.Update(func(tx *bolt.Tx) error {
		for _, name := range []string{"index", "commits", "checkpoints"} {
			if _, berr := tx.CreateBucketIfNotExists([]byte(name)); berr != nil {
				return fmt.Errorf("create bucket: %s", berr)
			}
		}
		return nil
	})

	if err != nil {
		log.Warningf("store-create failed: %v", err)
	}

	// Load all paths from the database into the trie:
	rootDir, err := newDirUnlocked(store, "/")
	if err != nil {
		return nil, err
	}

	store.Root = rootDir

	err = store.viewWithBucket("index", func(tx *bolt.Tx, bucket *bolt.Bucket) error {
		return bucket.ForEach(func(k []byte, v []byte) error {
			if _, loadErr := Unmarshal(store, v); loadErr != nil {
				log.Warningf("store-unmarshal: fail on `%s`: %v", k, err)
				return loadErr
			}

			return nil
		})
	})

	return store, err
}

// Mkdir creates a new, empty directory.
// If the directory already exists, this is a NOOP.
func (s *Store) Mkdir(repoPath string) (*File, error) {
	dir, err := NewDir(s, repoPath)
	if err != nil {
		return nil, err
	}

	dir.sync()
	return dir, err
}

// Add reads the data at the physical path `filePath` and adds it to the store
// at `repoPath` by hashing, compressing and encrypting the file.
// Directories will be added recursively.
func (s *Store) Add(filePath, repoPath string) error {
	return s.AddDir(filePath, repoPath)
}

// AddDir traverses all files in a directory and calls AddFromReader on them.
func (s *Store) AddDir(filePath, repoPath string) error {
	walkErr := filepath.Walk(filePath, func(path string, info os.FileInfo, err error) error {
		// Simply skip errorneous files:
		if err != nil {
			log.Warningf("Walk: %v", err)
			return err
		}

		// Map the file path relative to repoPath:
		currPath := filepath.Join(repoPath, path[len(filePath):])

		switch mode := info.Mode(); {
		case mode.IsRegular():
			fd, openErr := os.Open(path)
			if openErr != nil {
				return openErr
			}
			defer util.Closer(fd)

			err = s.AddFromReader(currPath, fd, info.Size())
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

	return walkErr
}

// AddFromReader reads data from r, encrypts & compresses it while feeding it to ipfs.
// The resulting hash will be committed to the index.
func (s *Store) AddFromReader(repoPath string, r io.Reader, size int64) error {
	ipfsNode, err := s.IpfsNode()
	if err != nil {
		return err
	}

	// Check if the file was already added:
	file := s.Root.Lookup(repoPath)
	initialAdd := false

	log.Debugf("bolt lookup: %v", file != nil)

	if file != nil {
		// We know this file already.
		log.WithFields(log.Fields{
			"file": repoPath,
		}).Info("Updating file.")
	} else {
		// Create intermediate directories:
		elems := strings.Split(repoPath, string(filepath.Separator))
		if len(elems) > 1 {
			for idx := range elems[1 : len(elems)-1] {
				dir := strings.Join(elems[idx:len(elems)-1], string(filepath.Separator))

				if _, err := s.Mkdir(dir); err != nil {
					log.Warningf("store-add: failed to create intermediate dir %s: %v", dir, err)
				}
			}
		}

		newFile, err := NewFile(s, repoPath)
		if err != nil {
			return err
		}

		file, initialAdd = newFile, true
	}

	// Control how many bytes are written to the encryption layer:
	sizeAcc := &util.SizeAccumulator{}
	teeR := io.TeeReader(r, sizeAcc)

	stream, err := NewFileReader(file.Key(), teeR, size)
	if err != nil {
		return err
	}

	mhash, err := ipfsutil.Add(ipfsNode, stream)
	if err != nil {
		return err
	}

	log.Infof("store-add: %s (hash: %s, key: %x)", repoPath, mhash.B58String(), file.Key()[10:])

	// Update metadata that might have changed:
	file.Lock()
	defer file.Unlock()

	if file.hash.Equal(&Hash{mhash}) {
		log.Debugf("Refusing update.")
		return ErrNoChange
	}

	oldMeta := file.Metadata
	if initialAdd {
		oldMeta = nil
	}

	file.Metadata = &Metadata{
		size:    int64(sizeAcc.Size()),
		modTime: time.Now(),
		hash:    &Hash{mhash},
		key:     file.Metadata.key,
	}

	// Create a checkpoint in the version history.
	// TODO: Move is not yet supported, probably use own function for this.
	//       (store.Move() or something)
	err = s.MakeCheckpoint(oldMeta, file.Metadata, repoPath, repoPath)
	if err != nil {
		return err
	}

	// If all went well, save it to bolt.
	// This will also sync intermediate directories.
	file.sync()
	return nil
}

// Touch creates a new empty file.
func (s *Store) Touch(repoPath string) error {
	return s.AddFromReader(repoPath, bytes.NewReader([]byte{}), 0)
}

// marshalFile converts a file to a protobuf and
func (s *Store) marshalFile(file *File, repoPath string) error {
	return s.updateWithBucket("index", func(tx *bolt.Tx, bucket *bolt.Bucket) error {
		data, err := file.Marshal()
		if err != nil {
			return err
		}

		if err := bucket.Put([]byte(repoPath), data); err != nil {
			return err
		}

		return nil
	})
}

// Stream returns the stream of the file at `path`.
func (s *Store) Stream(path string) (ipfsutil.Reader, error) {
	file := s.Root.Lookup(path)
	if file == nil {
		return nil, ErrNoSuchFile
	}

	return file.Stream()
}

// Cat will write the contents of the brig file `path` into `w`.
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

// Connect tries to connect the xmpp client and the ipfs daemon to the outside world.
func (s *Store) Connect(jid xmpp.JID, password string) error {
	if s.IsOnline() {
		return nil
	}

	if err := s.XMPP.Connect(jid, password); err != nil {
		log.Warningf("Unable to connect xmpp client: %v", err)
		return err
	}

	// Check if a previous offline mode was there:
	if s.ipfsNode != nil && !s.ipfsNode.IsOnline() {
		node := s.ipfsNode
		s.ipfsNode = nil

		if err := node.Close(); err != nil {
			return err
		}
	}

	// Try to register a fresh online mode:
	if s.ipfsNode == nil {
		ipfsNode, err := ipfsutil.StartNode(filepath.Join(s.repoPath, "ipfs"), true)
		if err != nil {
			// Try to mantain a consistent state by disconnecting on error:
			s.XMPP.Disconnect()
			return err
		}

		s.ipfsNode = ipfsNode
	}

	return nil
}

// Disconnect shuts down all store services that need an connection
// to the outside.
// TODO: Make this work with xmpp etc.
func (s *Store) Disconnect() (err error) {
	if !s.IsOnline() {
		return nil
	}

	log.Debugf("Disconnecting ipfs daemon.")

	if s.ipfsNode != nil {
		node := s.ipfsNode
		s.ipfsNode = nil

		if err = node.Close(); err != nil {
			log.Warningf("Unable to close ipfs node: %v", err)
		}
	}

	log.Debugf("Disconnecting xmpp client.")

	// Try to close xmpp, even if ipfs is still running:
	if err = s.XMPP.Disconnect(); err != nil {
		log.Warningf("Unable to disconnect xmpp client: %v", err)
	}

	return err
}

// IsOnline checks if both xmpp and ipfs is up and running.
func (s *Store) IsOnline() bool {
	return s.XMPP.IsOnline() && (s.ipfsNode != nil && s.ipfsNode.IsOnline())
}

// Close syncs all data. It is an error to use the store afterwards.
func (s *Store) Close() error {
	if err := s.Disconnect(); err != nil {
		return err
	}

	if err := s.db.Sync(); err != nil {
		log.Warningf("store-sync: %v", err)
		return err
	}

	if err := s.db.Close(); err != nil {
		log.Warningf("store-close: %v", err)
		return err
	}

	return nil
}

// Rm will purge a file locally on this node.
func (s *Store) Rm(path string) error {
	node := s.Root.Lookup(path)

	if node == nil {
		return ErrNoSuchFile
	}

	// TODO: Implement dir walk...
	if node.Kind() != FileTypeRegular {
		return fmt.Errorf("TODO: rm does not work on directories yet")
	}

	// Remove from trie & remove from bolt db.
	err := s.updateWithBucket("index", func(tx *bolt.Tx, bckt *bolt.Bucket) error {
		return bckt.Delete([]byte(path))
	})

	if err != nil {
		return err
	}

	log.Debugf("rm: Making checkpoint: %v", node.Metadata)
	if err := s.MakeCheckpoint(node.Metadata, nil, path, path); err != nil {
		return err
	}

	node.Remove()
	return nil
}

// Export marshals all relevant inside the database, so a cloned
// repository may import them again.
// The exported data includes:
//  - All files (including their history and keys)
//  - All commits.
//  - Pinning information.
//
// TODO: Describe json stream format.
//
// w is not closed after Export.
func (s *Store) Export(w io.Writer) (n int, err error) {
	// TODO: Export commits (not implemented)
	// TODO: Export pinning information.
	s.Root.Walk(true, func(child *File) bool {
		data, errJSON := child.MarshalJSON()
		if err != nil {
			err = errJSON
			return false
		}

		ndata, errWrite := w.Write(data)
		if errWrite != nil {
			err = errWrite
			return false
		}

		n += ndata
		return true
	})

	if err != nil {
		return
	}

	return
}

// Import unmarshals the data written by export.
// If succesful, a new store with the data is created.
func (s *Store) Import(repoPath string, r io.Reader) error {
	dec := json.NewDecoder(r)
	for {
		var m map[string]interface{}
		if err := dec.Decode(&m); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
	}

	return nil
}
