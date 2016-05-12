package store

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/disorganizer/brig/id"
	"github.com/disorganizer/brig/store/compress"
	"github.com/disorganizer/brig/store/wire"
	"github.com/disorganizer/brig/util"
	"github.com/disorganizer/brig/util/ipfsutil"
)

var (
	// ErrNoSuchFile is returned whenever a path could not be resolved to a file.
	ErrNoSuchFile = fmt.Errorf("No such file or directory")
	ErrExists     = fmt.Errorf("File exists")
	ErrNotEmpty   = fmt.Errorf("Cannot remove: Directory is not empty")
	ErrEmptyStage = fmt.Errorf("Nothing staged. No commit done.")
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
	ID id.ID

	// IPFS manager layer (from daemon.Server)
	IPFS *ipfsutil.Node
}

func prefixSlash(s string) string {
	if !strings.HasPrefix(s, "/") {
		return "/" + s
	}

	return s
}

func (st *Store) loadIndex() error {
	return st.viewWithBucket("index", func(tx *bolt.Tx, bkt *bolt.Bucket) error {
		// Check if the root directory already exists:
		if bkt.Get([]byte("/")) == nil {
			rootDir, err := newDirUnlocked(st, "/")
			if err != nil {
				return err
			}

			st.Root = rootDir
		}

		return bkt.ForEach(func(k []byte, v []byte) error {
			file := emptyFile(st)
			if err := file.Unmarshal(st, v); err != nil {
				log.Warningf("store-unmarshal: fail on `%s`: %v", k, err)
				return err
			}

			return nil
		})
	})
}

func (st *Store) createInitialCommit() error {
	needsInit := false

	err := st.viewWithBucket("refs", func(tx *bolt.Tx, bkt *bolt.Bucket) error {
		needsInit = (bkt.Get([]byte("HEAD")) == nil)
		return nil
	})

	if err != nil {
		return err
	}

	if !needsInit {
		return nil
	}

	// No commit yet, create initial commit.
	rootCommit := NewEmptyCommit(st, st.ID)
	rootCommit.Message = "Initial commit"
	rootCommit.Hash = st.Root.Hash().Clone()

	data, err := rootCommit.MarshalProto()
	if err != nil {
		return err
	}

	// Insert initial commit to `commits` bucket:
	err = st.updateWithBucket("commits", func(tx *bolt.Tx, bkt *bolt.Bucket) error {
		return bkt.Put(rootCommit.Hash.Bytes(), data)
	})

	return st.updateHEAD(rootCommit)
}

func (st *Store) updateHEAD(cmt *Commit) error {
	return st.updateWithBucket("refs", func(tx *bolt.Tx, bkt *bolt.Bucket) error {
		data, err := cmt.MarshalProto()
		if err != nil {
			return err
		}

		return bkt.Put([]byte("HEAD"), data)
	})
}

// Open loads an existing store at `brigPath/$ID/index.bolt`, if it does not
// exist, it is created.  For full function, Connect() should be called
// afterwards.
func Open(brigPath string, ID id.ID, IPFS *ipfsutil.Node) (*Store, error) {
	options := &bolt.Options{Timeout: 1 * time.Second}
	dbDir := filepath.Join(
		brigPath,
		"bolt."+strings.Replace(string(ID), "/", "-", -1),
	)

	if err := os.MkdirAll(dbDir, 0777); err != nil {
		return nil, err
	}

	db, err := bolt.Open(filepath.Join(dbDir, "index.bolt"), 0600, options)

	if err != nil {
		return nil, err
	}

	st := &Store{
		db:       db,
		ID:       ID,
		repoPath: brigPath,
		IPFS:     IPFS,
	}

	// Create initial buckets:
	err = db.Update(func(tx *bolt.Tx) error {
		buckets := []string{
			"index",       // File-Path to file protobuf.
			"stage",       // Staged files (path to current checkpoint)
			"commits",     // Commit-Hash to commit protobuf.
			"checkpoints", // File-Path to History (== mod_time to checkpoint)
			"refs",        // Special names for certain commits (e.g. HEAD)
		}

		for _, name := range buckets {
			if _, berr := tx.CreateBucketIfNotExists([]byte(name)); berr != nil {
				return fmt.Errorf("create bucket: %s", berr)
			}
		}
		return nil
	})

	if err != nil {
		log.Warningf("store-create failed: %v", err)
	}

	// Load all paths from the database into the trie.
	// This also creates a root node if none exists yet.
	if err := st.loadIndex(); err != nil {
		return nil, err
	}

	if err := st.createInitialCommit(); err != nil {
		return nil, err
	}

	return st, err
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

			err = s.AddFromReader(currPath, fd)
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
func (s *Store) AddFromReader(repoPath string, r io.Reader) error {
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

	// TODO: Make algo configurable/add heuristic too choose
	//       a suitable algorithm
	stream, err := NewFileReader(file.Key(), teeR, compress.AlgoSnappy)
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
	} else {
		// Remove the current hash from the merkle tree
		// before setting the new one:
		file.purgeHash()
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
	return s.AddFromReader(prefixSlash(repoPath), bytes.NewReader([]byte{}))
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
		log.Warningf("Could not create stream: %v", err)
		return err
	}

	if _, err := io.Copy(w, cleanStream); err != nil {
		log.Warningf("Could not copy stream: %v", err)
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
		err = s.updateWithBucket("index", func(tx *bolt.Tx, bkt *bolt.Bucket) error {
			return bkt.Delete([]byte(childPath))
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
		child.Lock()
		child.Metadata.size *= -1
		child.updateParents()
		child.Unlock()
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
func (st *Store) ListProto(root string, depth int) (*wire.Dirlist, error) {
	entries, err := st.List(root, depth)
	if err != nil {
		return nil, err
	}

	dirlist := &wire.Dirlist{}
	for _, entry := range entries {
		protoFile, err := entry.ToProto()
		if err != nil {
			return nil, err
		}

		// Be sure to mask out key and hash.
		dirlist.Entries = append(dirlist.Entries, &wire.Dirent{
			Path:     protoFile.Path,
			FileSize: protoFile.FileSize,
			Kind:     protoFile.Kind,
			ModTime:  protoFile.ModTime,
		})
	}

	return dirlist, nil
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
		err = st.updateWithBucket("index", func(tx *bolt.Tx, bkt *bolt.Bucket) error {
			return bkt.Delete([]byte(oldChildPath))
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
func (s *Store) Export() (*wire.Store, error) {
	// TODO: Export commits (not implemented)
	// TODO: Export pinning information.
	protoStore := &wire.Store{}

	var err error

	s.Root.Walk(true, func(child *File) bool {
		// Note: Walk() already calls Lock()
		protoFile, errPbf := child.ToProto()
		if err != nil {
			err = errPbf
			return false
		}

		history, errHist := s.History(child.node.Path())
		if errHist != nil {
			err = errHist
			return false
		}

		protoHist, errPbh := history.ToProto()
		if err != nil {
			err = errPbh
			return false
		}

		protoPack := &wire.Pack{
			File:    protoFile,
			History: protoHist,
		}

		protoStore.Packs = append(protoStore.Packs, protoPack)
		return true
	})

	if err != nil {
		return nil, err
	}

	return protoStore, nil
}

// Import unmarshals the data written by export.
// If succesful, a new store with the data is created.
func (s *Store) Import(protoStore *wire.Store) error {
	for _, pack := range protoStore.Packs {
		file := emptyFile(s)
		if err := file.Import(pack.GetFile()); err != nil {
			return err
		}

		// TODO: Restore history.
		log.Debugf("Imported: %v", file.Path())
		file.Sync()
		file.updateParents()
	}

	return nil
}

// Head returns the most recent commit.
// Commit will be always non-nil if error is nil,
// the initial commit has no changes.
func (st *Store) Head() (*Commit, error) {
	cmt := NewEmptyCommit(st, st.ID)

	err := st.viewWithBucket("refs", func(tx *bolt.Tx, bkt *bolt.Bucket) error {
		data := bkt.Get([]byte("HEAD"))
		if data == nil {
			return fmt.Errorf("No HEAD in database")
		}

		return cmt.UnmarshalProto(data)
	})

	if err != nil {
		return nil, err
	}

	return cmt, nil
}

// Status shows how a Commit would look like if Commit() would be called.
func (st *Store) Status() (*Commit, error) {
	head, err := st.Head()
	if err != nil {
		return nil, err
	}

	cmt := NewEmptyCommit(st, st.ID)
	cmt.Parent = head
	cmt.Hash = st.Root.Hash().Clone()
	cmt.Message = "Uncommitted changes"

	err = st.viewWithBucket("stage", func(tx *bolt.Tx, bkt *bolt.Bucket) error {
		return bkt.ForEach(func(bpath, bckpnt []byte) error {
			file := st.Root.Lookup(string(bpath))
			if file == nil {
				return ErrNoSuchFile
			}

			// TODO: unmarshal
			checkpoint := &Checkpoint{}
			if err := checkpoint.Unmarshal(bckpnt); err != nil {
				return err
			}

			cmt.Changes[file] = checkpoint
			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return cmt, nil
}

// TODO: Global store lock, calling MakeCommit more than once is racy!

// Commit saves a commit in the store history.
func (st *Store) MakeCommit(msg string) error {
	cmt, err := st.Status()
	if err != nil {
		return err
	}

	return st.db.Update(func(tx *bolt.Tx) error {
		// Check if the stage area contains something:
		stage := tx.Bucket([]byte("stage"))
		if stage == nil {
			return ErrNoSuchBucket{"stage"}
		}

		if stage.Stats().KeyN == 0 {
			return ErrEmptyStage
		}

		// Flush the staging area:
		if err := tx.DeleteBucket([]byte("stage")); err != nil {
			return err
		}

		if _, err := tx.CreateBucket([]byte("stage")); err != nil {
			return err
		}

		cmts := tx.Bucket([]byte("commits"))
		if cmts == nil {
			return ErrNoSuchBucket{"commits"}
		}

		// Put the new commit in the commits bucket:
		cmt.Message = msg
		data, err := cmt.MarshalProto()
		if err != nil {
			return err
		}

		if err := cmts.Put(cmt.Hash.Bytes(), data); err != nil {
			return err
		}

		// Update HEAD:
		refs := tx.Bucket([]byte("refs"))
		if refs == nil {
			return ErrNoSuchBucket{"refs"}
		}

		return refs.Put([]byte("HEAD"), data)
	})
}

// TODO: respect from/to ranges
func (st *Store) Log() (Commits, error) {
	var cmts Commits

	err := st.viewWithBucket("commits", func(tx *bolt.Tx, bkt *bolt.Bucket) error {
		return bkt.ForEach(func(k, v []byte) error {
			cmt := NewEmptyCommit(st, st.ID)
			if err := cmt.UnmarshalProto(v); err != nil {
				return err
			}

			cmts = append(cmts, cmt)
			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	sort.Sort(cmts)
	return cmts, nil
}
