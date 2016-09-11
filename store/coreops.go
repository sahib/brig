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
	"github.com/disorganizer/brig/store/compress"
	"github.com/disorganizer/brig/store/wire"
	"github.com/disorganizer/brig/util"
	"github.com/disorganizer/brig/util/ipfsutil"
)

var (
	ErrExists   = fmt.Errorf("File exists")
	ErrNotEmpty = fmt.Errorf("Cannot remove: Directory is not empty")
)

type errNoSuchFile struct {
	path string
}

func (e *errNoSuchFile) Error() string {
	return "No such file or directory: " + e.path
}

// NoSuchFile creates a new error that reports `path` as missing
// TODO: move to errors.go?
func NoSuchFile(path string) error {
	return &errNoSuchFile{path}
}

// IsNoSuchFileError asserts that `err` means that the file could not be found
func IsNoSuchFileError(err error) bool {
	_, ok := err.(*errNoSuchFile)
	return ok
}

// Mkdir creates a new, empty directory. It's a NOOP if the directory already exists.
func (st *Store) Mkdir(repoPath string) (*File, error) {
	return st.mkdir(repoPath, false)
}

// MkdirAll is like Mkdir but creates intermediate directories conviniently.
func (st *Store) MkdirAll(repoPath string) (*File, error) {
	return st.mkdir(repoPath, true)
}

func (st *Store) mkdir(repoPath string, createParents bool) (*File, error) {
	if createParents {
		if err := st.mkdirParents(repoPath); err != nil {
			return nil, err
		}
	} else {
		// Check if the parent exists.
		// (would result in weird undefined intermediates otherwise)
		parentPath := path.Dir(repoPath)
		if parent := st.Root.Lookup(parentPath); parent == nil {
			return nil, NoSuchFile(parentPath)
		}
	}

	dir, err := NewDir(st, prefixSlash(repoPath))
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
func (st *Store) AddDir(filePath, repoPath string) error {
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

			err = st.AddFromReader(currPath, fd)
		case mode.IsDir():
			_, err = st.Mkdir(currPath)
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

func (st *Store) mkdirParents(path string) error {
	elems := strings.Split(path, "/")

	for idx := 1; idx < len(elems)-1; idx++ {
		dir := strings.Join(elems[:idx+1], "/")
		if _, err := st.Mkdir(dir); err != nil {
			log.Warningf("store-add: failed to create intermediate dir `%s`: %v", dir, err)
			return err
		}
	}

	return nil
}

// AddFromReader reads data from r, encrypts & compresses it while feeding it to ipfs.
// The resulting hash will be committed to the index.
func (st *Store) AddFromReader(repoPath string, r io.Reader) error {
	repoPath = prefixSlash(repoPath)
	initialAdd := false

	st.mu.Lock()
	defer st.mu.Unlock()

	// Check if the file was already added:
	file := st.Root.Lookup(repoPath)

	if file != nil {
		// We know this file already.
		log.WithFields(log.Fields{
			"file": repoPath,
		}).Info("File exists; modifying.")
	} else {
		if err := st.mkdirParents(repoPath); err != nil {
			return err
		}

		// Create a new file at specified path:
		newFile, err := NewFile(st, repoPath)
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

	mhash, err := ipfsutil.Add(st.IPFS, stream)
	if err != nil {
		return err
	}

	if err := st.IPFS.Pin(mhash); err != nil {
		return err
	}

	return st.insertMetadata(
		file, repoPath, &Hash{mhash}, initialAdd, int64(sizeAcc.Size()),
	)
}

func (st *Store) insertMetadata(file *File, repoPath string, newHash *Hash, initialAdd bool, size int64) error {
	log.Infof(
		"store-add: %s (hash: %s, key: %x)",
		repoPath,
		newHash.B58String(),
		file.Key()[10:], // TODO: Make omit() a util func
	)

	file.Lock()
	defer file.Unlock()

	// Update metadata that might have changed:
	if file.hash.Equal(newHash) {
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
		size:    size,
		modTime: time.Now(),
		hash:    newHash,
		key:     file.Metadata.key,
		kind:    FileTypeRegular,
	}

	file.updateParents()

	// Create a checkpoint in the version history.
	if err := st.MakeCheckpoint(oldMeta, file.Metadata, repoPath, repoPath); err != nil {
		return err
	}

	// If all went well, save it to bolt.
	// This will also sync intermediate directories.
	file.sync()
	return nil
}

func (st *Store) pinOp(repoPath string, doUnpin bool) error {
	node := st.Root.Lookup(repoPath)
	if node == nil {
		return NoSuchFile(repoPath)
	}

	var pinMe []*File

	switch kind := node.Kind(); kind {
	case FileTypeDir:
		node.Walk(true, func(child *File) bool {
			if child.kind == FileTypeRegular {
				pinMe = append(pinMe, child)
			}

			return true
		})
	case FileTypeRegular:
		pinMe = append(pinMe, node)
	default:
		return fmt.Errorf("Bad node kind: %d", kind)
	}

	fn := st.IPFS.Pin
	if doUnpin {
		fn = st.IPFS.Unpin
	}

	var errs util.Errors
	for _, toPin := range pinMe {
		if err := fn(toPin.Hash().Multihash); err != nil {
			errs = append(errs, err)
		}
	}

	return errs.ToErr()
}

func (st *Store) Pin(repoPath string) error {
	st.mu.Lock()
	defer st.mu.Unlock()

	return st.pinOp(repoPath, false)
}

func (st *Store) Unpin(repoPath string) error {
	st.mu.Lock()
	defer st.mu.Unlock()

	return st.pinOp(repoPath, false)
}

func (st *Store) IsPinned(repoPath string) (bool, error) {
	node := st.Root.Lookup(repoPath)
	if node == nil {
		return false, NoSuchFile(repoPath)
	}

	return st.IPFS.IsPinned(node.Hash().Multihash)
}

// Touch creates a new empty file.
// It is provided as convenience wrapper around AddFromReader.
func (st *Store) Touch(repoPath string) error {
	return st.AddFromReader(prefixSlash(repoPath), bytes.NewReader([]byte{}))
}

// Stream returns the stream of the file at `path`.
func (st *Store) Stream(path string) (ipfsutil.Reader, error) {
	st.mu.Lock()
	defer st.mu.Unlock()

	file := st.Root.Lookup(prefixSlash(path))
	if file == nil {
		return nil, NoSuchFile(path)
	}

	return file.Stream()
}

// Cat will write the contents of the brig file `path` into `w`.
func (st *Store) Cat(path string, w io.Writer) error {
	cleanStream, err := st.Stream(path)
	if err != nil {
		log.Warningf("Could not create stream: %v", err)
		return err
	}

	// No locking required, data comes from ipfs.

	if _, err := io.Copy(w, cleanStream); err != nil {
		log.Warningf("Could not copy stream: %v", err)
		return err
	}

	return nil
}

// Remove will purge a file locally on this node.
// If `recursive` is true and if `path` is a directory, all files
// in it will be removed. If `recursive` is false, ErrNotEmpty will
// be returned upon non-empty directories.
func (st *Store) Remove(path string, recursive bool) error {
	st.mu.Lock()
	defer st.mu.Unlock()

	return st.remove(path, recursive)
}

func (st *Store) remove(path string, recursive bool) (err error) {
	path = prefixSlash(path)

	node := st.Root.Lookup(path)
	if node == nil {
		return NoSuchFile(path)
	}

	if node.Kind() == FileTypeDir && node.NChildren() > 0 && !recursive {
		return ErrNotEmpty
	}

	toBeRemoved := []*File{}

	node.Walk(true, func(child *File) bool {
		childPath := child.Path()

		// Remove from trie & remove from bolt db.
		err = st.updateWithBucket("index", func(tx *bolt.Tx, bkt *bolt.Bucket) error {
			return bkt.Delete([]byte(childPath))
		})

		if err != nil {
			return false
		}

		if err = st.MakeCheckpoint(node.Metadata, nil, childPath, childPath); err != nil {
			return false
		}

		toBeRemoved = append(toBeRemoved, child)
		return true
	})

	errs := util.Errors{}

	for _, child := range toBeRemoved {
		if child.Kind() == FileTypeRegular {
			if err := st.IPFS.Unpin(child.Hash().Multihash); err != nil {
				errs = append(errs, err)
			}
		}

		child.Lock()
		child.Metadata.size *= -1
		child.updateParents()
		child.Unlock()
		child.Remove()

	}

	return errs.ToErr()
}

// List exports a directory listing of `root` up to `depth` levels down.
func (st *Store) List(root string, depth int) (entries []*File, err error) {
	root = prefixSlash(root)
	entries = []*File{}

	st.mu.Lock()
	defer st.mu.Unlock()

	node := st.Root.Lookup(root)
	if node == nil {
		return nil, NoSuchFile(root)
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

	// No locking required; only some in-memory conversion follows.

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

func (st *Store) Move(oldPath, newPath string) error {
	st.mu.Lock()
	defer st.mu.Unlock()

	return st.move(oldPath, newPath)
}

func (st *Store) move(oldPath, newPath string) (err error) {
	oldPath, newPath = prefixSlash(oldPath), prefixSlash(newPath)

	node := st.Root.Lookup(oldPath)
	if node == nil {
		return NoSuchFile(oldPath)
	}

	if newNode := st.Root.Lookup(newPath); newNode != nil {
		return ErrExists
	}

	newPaths := make(map[string]*File)

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

		newPaths[newChildPath] = child

		md := node.Metadata
		if err = st.MakeCheckpoint(md, md, oldChildPath, newChildPath); err != nil {
			return false
		}

		return true
	})

	if err != nil {
		return err
	}

	// Note: No pinning information needs to change,
	//       Move() does not influence the Hash() of the file.
	for newPath, node := range newPaths {
		node.Remove()
		node.insert(st.Root, newPath)
	}

	return nil
}

// Status shows how a Commit would look like if Commit() would be called.
func (st *Store) Status() (*Commit, error) {
	return st.status()
}

// Unlocked version of Status()
func (st *Store) status() (*Commit, error) {
	head, err := st.head()
	if err != nil {
		return nil, err
	}

	cmt := NewEmptyCommit(st, st.ID)
	cmt.Parent = head
	cmt.Message = "Uncommitted changes"
	cmt.TreeHash = st.Root.Hash().Clone()

	hash, err := st.makeCommitHash(cmt, head)
	if err != nil {
		return nil, err
	}

	cmt.Hash = hash

	err = st.viewWithBucket("stage", func(tx *bolt.Tx, bkt *bolt.Bucket) error {
		return bkt.ForEach(func(bpath, bckpnt []byte) error {
			checkpoint := &Checkpoint{}
			if err := checkpoint.Unmarshal(bckpnt); err != nil {
				return err
			}

			cmt.Checkpoints = append(cmt.Checkpoints, checkpoint)
			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return cmt, nil
}

// Commit saves a commit in the store history.
func (st *Store) MakeCommit(msg string) error {
	st.mu.Lock()
	defer st.mu.Unlock()

	if msg == "" {
		return ErrEmptyCommitMessage
	}

	cmt, err := st.status()
	if err != nil {
		return err
	}
}

// TODO: respect from/to ranges
func (fs *FS) Log() (*Commits, error) {
	var cmts Commits

	head, err := fs.Head()
	if err != nil {
		return nil, err
	}

	for curr := head; curr != nil; curr = curr.ParentCommit() {
		cmts = append(cmts, curr)
	}

	sort.Sort(&cmts)
	return &cmts, nil
}
