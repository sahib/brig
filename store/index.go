package store

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/disorganizer/brig/store/wire"
	"github.com/disorganizer/brig/util"
	"github.com/disorganizer/brig/util/ipfsutil"
	"github.com/disorganizer/brig/util/protocol"
	"github.com/tsuibin/goxmpp2/xmpp"
)

var (
	// ErrNoSuchFile is returned whenever a path could not be resolved to a file.
	ErrNoSuchFile = fmt.Errorf("No such file or directory")
	ErrExists     = fmt.Errorf("File exists")
	ErrNotEmpty   = fmt.Errorf("Cannot remove: Directory is not empty")
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

	// The jabber id this store is associated to.
	jid xmpp.JID

	// IPFS manager layer (from daemon.Server)
	IPFS *ipfsutil.Node
}

func prefixSlash(s string) string {
	if !strings.HasPrefix(s, "/") {
		return "/" + s
	}

	return s
}

// Open loads an existing store at `brigPath/$jid/index.bolt`, if it does not
// exist, it is created.  For full function, Connect() should be called
// afterwards.
func Open(brigPath string, jid xmpp.JID, IPFS *ipfsutil.Node) (*Store, error) {
	options := &bolt.Options{Timeout: 1 * time.Second}
	dbDir := filepath.Join(
		brigPath,
		"bolt."+strings.Replace(string(jid), "/", "-", -1),
	)

	if err := os.MkdirAll(dbDir, 0777); err != nil {
		return nil, err
	}

	db, err := bolt.Open(filepath.Join(dbDir, "index.bolt"), 0600, options)

	if err != nil {
		return nil, err
	}

	store := &Store{
		db:       db,
		jid:      jid,
		repoPath: brigPath,
		IPFS:     IPFS,
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
			file := emptyFile(store)
			if loadErr := file.Unmarshal(store, v); loadErr != nil {
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
	dir, err := NewDir(s, prefixSlash(repoPath))
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
	return s.AddDir(filePath, prefixSlash(repoPath))
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
		repoPath = prefixSlash(repoPath)
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
	repoPath = prefixSlash(repoPath)

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
		sep := string(filepath.Separator)

		for idx := 1; idx < len(elems)-1; idx++ {
			dir := strings.Join(elems[:len(elems)-idx], sep)
			if _, err := s.Mkdir(dir); err != nil {
				log.Warningf("store-add: failed to create intermediate dir %s: %v", dir, err)
				return err
			}
		}

		// Create a new file at specified path:
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

	mhash, err := ipfsutil.Add(s.IPFS, stream)
	if err != nil {
		return err
	}

	log.Infof(
		"store-add: %s (hash: %s, key: %x)",
		repoPath,
		mhash.B58String(),
		file.Key()[10:],
	)

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
		kind:    FileTypeRegular,
	}

	file.updateParents()

	// Create a checkpoint in the version history.
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
// It is provided as convenience wrapper around AddFromReader.
func (s *Store) Touch(repoPath string) error {
	return s.AddFromReader(prefixSlash(repoPath), bytes.NewReader([]byte{}), 0)
}

// marshalFile converts a file to a protobuf and
func (s *Store) marshalFile(file *File, repoPath string) error {
	repoPath = prefixSlash(repoPath)

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
	file := s.Root.Lookup(prefixSlash(path))
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

// Close syncs all data. It is an error to use the store afterwards.
func (s *Store) Close() error {
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

// Remove will purge a file locally on this node.
// If `recursive` is true and if `path` is a directory, all files
// in it will be removed. If `recursive` is false, ErrNotEmpty will
// be returned upon non-empty directories.
func (s *Store) Remove(path string, recursive bool) (err error) {
	path = prefixSlash(path)
	node := s.Root.Lookup(path)
	if node == nil {
		return ErrNoSuchFile
	}

	if node.Kind() == FileTypeDir && node.NChildren() > 0 && !recursive {
		return ErrNotEmpty
	}

	toBeRemoved := []*File{}

	node.Walk(true, func(child *File) bool {
		childPath := child.Path()

		// Remove from trie & remove from bolt db.
		err = s.updateWithBucket("index", func(tx *bolt.Tx, bckt *bolt.Bucket) error {
			return bckt.Delete([]byte(childPath))
		})

		if err != nil {
			return false
		}

		if err = s.MakeCheckpoint(node.Metadata, nil, childPath, childPath); err != nil {
			return false
		}

		toBeRemoved = append(toBeRemoved, child)
		return true
	})

	for _, child := range toBeRemoved {
		child.Remove()
	}
	return nil
}

// List exports a directory listing of `root` up to `depth` levels down.
func (st *Store) List(root string, depth int) (entries []*File, err error) {
	root = prefixSlash(root)
	entries = []*File{}

	node := st.Root.Lookup(root)
	if node == nil {
		return nil, ErrNoSuchFile
	}

	if depth < 0 {
		depth = math.MaxInt32
	}

	node.Walk(false, func(child *File) bool {
		if child.Depth() > depth {
			return false
		}

		entries = append(entries, child)
		return true
	})

	return
}

// The results are marshaled into a wire.Dirlist message and written to `w`.
// `depth` may be negative for unlimited recursion.
func (st *Store) ListMarshalled(w io.Writer, root string, depth int) error {
	entries, err := st.List(root, depth)
	if err != nil {
		return err
	}

	dirlist := &wire.Dirlist{}
	for _, entry := range entries {
		protoFile, err := entry.toProtoMessage()
		if err != nil {
			return err
		}

		// Be sure to mask out key and hash.
		dirlist.Entries = append(dirlist.Entries, &wire.Dirent{
			Path:     protoFile.Path,
			FileSize: protoFile.FileSize,
			Kind:     protoFile.Kind,
			ModTime:  protoFile.ModTime,
		})
	}

	enc := protocol.NewProtocolWriter(w, true)
	if err := enc.Send(dirlist); err != nil {
		return err
	}

	return nil
}

func (st *Store) Move(oldPath, newPath string) (err error) {
	oldPath, newPath = prefixSlash(oldPath), prefixSlash(newPath)

	node := st.Root.Lookup(oldPath)
	if node == nil {
		return ErrNoSuchFile
	}

	if newNode := st.Root.Lookup(newPath); newNode != nil {
		return ErrExists
	}

	toBeRemoved := make(map[string]*File)

	// Work recursively for directories:
	node.Walk(true, func(child *File) bool {
		oldChildPath := child.Path()
		newChildPath := path.Join(newPath, oldChildPath[len(oldPath):])

		// Remove from trie & remove from bolt db.
		err = st.updateWithBucket("index", func(tx *bolt.Tx, bckt *bolt.Bucket) error {
			return bckt.Delete([]byte(oldChildPath))
		})

		if err != nil {
			return false
		}

		toBeRemoved[newChildPath] = child

		md := node.Metadata
		if err = st.MakeCheckpoint(md, md, oldChildPath, newChildPath); err != nil {
			return false
		}

		return true
	})

	if err != nil {
		return err
	}

	for newPath, node := range toBeRemoved {
		node.Remove()
		node.insert(st.Root, newPath)
	}

	return nil
}

// Export marshals all relevant inside the database, so a cloned
// repository may import them again.
// The exported data includes:
//  - All files (including their history and keys)
//  - All commits.
//  - Pinning information.
//
// TODO: Describe stream format.
//
// w is not closed after Export.
func (s *Store) Export(w io.Writer) (err error) {
	// TODO: Export commits (not implemented)
	// TODO: Export pinning information.
	enc := protocol.NewProtocolWriter(w, true)

	s.Root.Walk(true, func(child *File) bool {
		// Note: Walk() already calls Lock()
		protoFile, errPbf := child.toProtoMessage()
		if err != nil {
			err = errPbf
			return false
		}

		history, errHist := s.History(child.node.Path())
		if errHist != nil {
			err = errHist
			return false
		}

		protoHist, errPbh := history.toProtoMessage()
		if err != nil {
			err = errPbh
			return false
		}

		protoPack := &wire.Pack{
			File:    protoFile,
			History: protoHist,
		}

		if errSend := enc.Send(protoPack); err != nil {
			err = errSend
			return false
		}

		return true
	})

	return err
}

// Import unmarshals the data written by export.
// If succesful, a new store with the data is created.
func (s *Store) Import(r io.Reader) error {
	dec := protocol.NewProtocolReader(r, true)

	for {
		pack := &wire.Pack{}

		if err := dec.Recv(pack); err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		file := emptyFile(s)
		if err := file.Import(pack.GetFile()); err != nil {
			return err
		}

		// TODO: Insert history?

		log.Debugf("Imported: %v", file.Path())
		file.Sync()
		file.updateParents()
	}

	return nil
}
