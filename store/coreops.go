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
	"github.com/disorganizer/brig/store/compress"
	"github.com/disorganizer/brig/store/wire"
	"github.com/disorganizer/brig/util"
	"github.com/disorganizer/brig/util/ipfsutil"
	"github.com/gogo/protobuf/proto"
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
func (st *Store) Mkdir(repoPath string) (*Directory, error) {
	return st.mkdir(repoPath, false)
}

// MkdirAll is like Mkdir but creates intermediate directories conviniently.
func (st *Store) MkdirAll(repoPath string) (*Directory, error) {
	return st.mkdir(repoPath, true)
}

func (st *Store) mkdirParents(path string) (*Directory, error) {
	elems := strings.Split(path, "/")

	for idx := 1; idx < len(elems)-1; idx++ {
		dirname := strings.Join(elems[:idx+1], "/")
		dir, err := st.mkdir(dirname, false)

		if err != nil {
			log.Warningf("store-add: failed to create intermediate dir `%s`: %v", dirname, err)
			return nil, err
		}

		// Return it, if it's the last path component:
		if idx+1 == len(elems)-1 {
			return dir, nil
		}
	}

	return nil, fmt.Errorf("Empty path")
}

func (st *Store) mkdir(repoPath string, createParents bool) (*Directory, error) {
	dirname, basename := path.Split(repoPath)

	// Check if the parent exists.
	// (would result in weird undefined intermediates otherwise)
	parent, err := st.fs.LookupDirectory(dirname)
	if err != nil {
		return nil, err
	}

	// If it's nil, we might need to create it:
	if parent == nil {
		if !createParents {
			return nil, NoSuchFile(dirname)
		}

		parent, err = st.mkdirParents(repoPath)
		if err != nil {
			return nil, err
		}
	}

	child, err := parent.Child(basename)
	if err != nil {
		return nil, err
	}

	if child.GetType() != NodeTypeDirectory {
		return nil, fmt.Errorf("`%s` exists and is not a directory", repoPath)
	} else {
		// Notthing to do really. Return the old child.
		return child.(*Directory), nil
	}

	// Create it then!
	dir, err := newEmptyDirectory(st.fs, parent, basename)
	if err != nil {
		return nil, err
	}

	if err := st.fs.StageNode(dir); err != nil {
		return nil, err
	}

	return dir, nil
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
			log.Warningf("Recursive add: Ignoring weird file type: %v")
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
func (st *Store) AddFromReader(repoPath string, r io.Reader) error {
	repoPath = prefixSlash(repoPath)
	initialAdd := false

	st.mu.Lock()
	defer st.mu.Unlock()

	// Check if the file was already added:
	file, err := st.fs.LookupFile(repoPath)
	if err != nil {
		return err
	}

	if file != nil {
		// We know this file already.
		log.WithFields(log.Fields{
			"file": repoPath,
		}).Info("File exists; modifying.")
	} else {
		parent, err := st.mkdirParents(repoPath)
		if err != nil {
			return err
		}

		// Create a new file at specified path:
		newFile, err := newEmptyFile(st.fs, path.Base(repoPath))
		if err != nil {
			return err
		}

		if err := parent.Add(parent); err != nil {
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
		file, repoPath, &Hash{mhash}, initialAdd, sizeAcc.Size(),
	)
}

func (st *Store) insertMetadata(file *File, repoPath string, newHash *Hash, initialAdd bool, size uint64) error {
	log.Infof(
		"store-add: %s (hash: %s, key: %x)",
		repoPath,
		newHash.B58String(),
		file.Key()[10:], // TODO: Make omit() a util func
	)

	// Update metadata that might have changed:
	if file.hash.Equal(newHash) {
		log.Debugf("Refusing update.")
		return ErrNoChange
	}

	oldHash := file.Hash().Clone()
	if initialAdd {
		oldHash = nil
	}

	file.SetSize(size)
	file.SetModTime(time.Now())
	file.SetHash(newHash)

	// Create a checkpoint in the version history.
	if err := st.makeCheckpoint(file.ID(), oldHash, newHash, repoPath, repoPath); err != nil {
		return err
	}

	return st.fs.StageNode(file)
}

func (st *Store) pinOp(repoPath string, doUnpin bool) error {
	node, err := st.fs.LookupNode(repoPath)
	if err != nil {
		return err
	}

	if node == nil {
		return NoSuchFile(repoPath)
	}

	var pinMe []Node

	err = Walk(node, true, func(child Node) error {
		if child.GetType() == NodeTypeFile {
			pinMe = append(pinMe, child)
		}

		return nil
	})

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
	node, err := st.fs.LookupDirectory(repoPath)
	if err != nil {
		return false, err
	}

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

func (st *Store) makeCheckpoint(ID uint64, oldHash, newHash *Hash, oldPath, newPath string) error {
	owner, err := st.Owner()
	if err != nil {
		return err
	}

	ckp, err := st.fs.LastCheckpoint(ID)
	if err != nil {
		return err
	}

	newCkp, err := ckp.Fork(owner.ID(), oldHash, newHash, oldPath, newPath)
	if err != nil {
		return err
	}

	return st.fs.StageCheckpoint(newCkp)
}

// Stream returns the stream of the file at `path`.
func (st *Store) Stream(path string) (ipfsutil.Reader, error) {
	st.mu.Lock()
	defer st.mu.Unlock()

	file, err := st.fs.LookupFile(prefixSlash(path))
	if err != nil {
		return nil, err
	}

	if file == nil {
		return nil, NoSuchFile(path)
	}

	return file.Stream(st.IPFS)
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

func (st *Store) Lookup(repoPath string) (Node, error) {
	st.mu.Lock()
	defer st.mu.Unlock()

	return st.fs.LookupNode(repoPath)
}

func (st *Store) Root() (*Directory, error) {
	st.mu.Lock()
	defer st.mu.Unlock()

	return st.fs.Root()
}

// Remove will purge a file locally on this node.
// If `recursive` is true and if `path` is a directory, all files
// in it will be removed. If `recursive` is false, ErrNotEmpty will
// be returned upon non-empty directories.
func (st *Store) Remove(repoPath string, recursive bool) error {
	st.mu.Lock()
	defer st.mu.Unlock()

	repoPath = prefixSlash(repoPath)
	node, err := st.fs.LookupNode(repoPath)
	if err != nil {
		return err
	}

	if node == nil {
		return NoSuchFile(repoPath)
	}

	if node.GetType() == NodeTypeDirectory && node.NChildren() > 0 && !recursive {
		return ErrNotEmpty
	}

	toBeRemoved := []Node{}

	err = Walk(node, true, func(child Node) error {
		toBeRemoved = append(toBeRemoved, child)
		return nil
	})

	if err != nil {
		return err
	}

	errs := util.Errors{}
	for _, child := range toBeRemoved {
		childPath := NodePath(child)
		if err = st.makeCheckpoint(child.ID(), child.Hash(), nil, childPath, childPath); err != nil {
			return err
		}

		if child.GetType() == NodeTypeFile {
			if err := st.IPFS.Unpin(child.Hash().Multihash); err != nil {
				errs = append(errs, err)
			}
		}

		parentNode, err := child.Parent()
		if err != nil {
			return err
		}

		parent, ok := parentNode.(*Directory)
		if !ok {
			return ErrBadNode
		}

		if err := parent.RemoveChild(child); err != nil {
			return err
		}
	}

	return errs.ToErr()
}

// List exports a directory listing of `root` up to `depth` levels down.
func (st *Store) List(root string, depth int) ([]Node, error) {
	root = prefixSlash(root)
	entries := []Node{}

	st.mu.Lock()
	defer st.mu.Unlock()

	node, err := st.fs.LookupDirectory(root)
	if err != nil {
		return nil, err
	}

	if node == nil {
		return nil, NoSuchFile(root)
	}

	if depth < 0 {
		depth = math.MaxInt32
	}

	err = Walk(node, false, func(child Node) error {
		if NodeDepth(child) > depth {
			return nil
		}

		entries = append(entries, child)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return entries, err
}

// The results are marshaled into a wire.Dirlist message and written to `w`.
// `depth` may be negative for unlimited recursion.
// TODO: Is this really a core-op?
func (st *Store) ListProtoNodes(root string, depth int) (*wire.Nodes, error) {
	entries, err := st.List(root, depth)
	if err != nil {
		return nil, err
	}

	// No locking required; only some in-memory conversion follows.

	nodes := &wire.Nodes{}
	for _, node := range entries {
		pnode, err := node.ToProto()
		if err != nil {
			return nil, err
		}

		// Fill in the path for the sake of exporting:
		pnode.Path = proto.String(NodePath(node))
		nodes.Nodes = append(nodes.Nodes, pnode)
	}

	return nodes, nil
}

func (st *Store) Move(oldPath, newPath string, force bool) error {
	st.mu.Lock()
	defer st.mu.Unlock()

	return st.move(oldPath, newPath, force)
}

func (st *Store) move(oldPath, newPath string, force bool) error {
	oldPath = prefixSlash(path.Clean(oldPath))
	newPath = prefixSlash(path.Clean(newPath))

	node, err := st.fs.LookupNode(oldPath)
	if err != nil {
		return err
	}

	if node == nil {
		return NoSuchFile(oldPath)
	}

	newNode, err := st.fs.LookupNode(newPath)
	if err != nil {
		return err
	}

	if newNode != nil && newNode.GetType() != NodeTypeDirectory && !force {
		return ErrExists
	}

	newPaths := make(map[string]*File)

	// Work recursively for directories:
	err = Walk(node, true, func(child Node) error {
		if child.GetType() != NodeTypeFile {
			return nil
		}

		oldChildPath := NodePath(child)
		newChildPath := path.Join(newPath, oldChildPath[len(oldPath):])
		newPaths[newChildPath] = child.(*File)

		hash := child.Hash()
		if err = st.makeCheckpoint(newNode.ID(), hash, hash, oldChildPath, newChildPath); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	// If the node at newPath was a file, we need to remove it.
	if newNode != nil && newNode.GetType() == NodeTypeFile {
		if err := nodeRemove(newNode); err != nil {
			return err
		}
	}

	// NOTE: No pinning information needs to change,
	//       Move() does not influence the Hash() of the file.
	for newPath, file := range newPaths {
		dest, err := st.mkdirParents(newPath)
		if err != nil {
			return err
		}

		if dest == nil {
			return NoSuchFile(newPath)
		}

		// Basename might have changed:
		file.SetName(path.Base(newPath))

		// Add to new parent:
		if err := dest.Add(file); err != nil {
			return err
		}

		// Remove from old Parent:
		if err := nodeRemove(file); err != nil {
			return err
		}
	}

	return nil
}

// Status shows how a Commit would look like if Commit() would be called.
func (st *Store) Status() (*Commit, error) {
	st.mu.Lock()
	defer st.mu.Unlock()

	return st.fs.Status()
}

func (st *Store) History(repoPath string) (History, error) {
	st.mu.Lock()
	defer st.mu.Unlock()

	return st.fs.HistoryByPath(repoPath)
}

// Commit saves a commit in the store history.
func (st *Store) MakeCommit(msg string) error {
	st.mu.Lock()
	defer st.mu.Unlock()

	owner, err := st.Owner()
	if err != nil {
		return err
	}

	return st.fs.MakeCommit(owner, msg)
}

// TODO: respect from/to ranges
func (st *Store) Log() ([]Node, error) {
	st.mu.Lock()
	defer st.mu.Unlock()

	var cmts []Node

	head, err := st.fs.Head()
	if err != nil {
		return nil, err
	}

	var curr Node = head
	for curr != nil {
		cmts = append(cmts, curr)

		parNode, err := curr.Parent()
		if err != nil {
			return nil, err
		}

		if parNode == nil {
			break
		}

		curr = parNode
	}

	return cmts, nil
}
