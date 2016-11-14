package store

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"math"
	"os"
	"path"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/store/compress"
	"github.com/disorganizer/brig/util"
	"github.com/disorganizer/brig/util/ipfsutil"
)

// Mkdir creates a new, empty directory. It's a NOOP if the directory already exists.
// TODO: Do not return directory - it's not locked.
func (st *Store) Mkdir(repoPath string) error {
	_, err := mkdir(st.fs, repoPath, false)
	return err
}

// MkdirAll is like Mkdir but creates intermediate directories conviniently.
func (st *Store) MkdirAll(repoPath string) error {
	_, err := mkdir(st.fs, repoPath, true)
	return err
}

// Stage reads the data at the physical path `filePath` and adds it to the store
// at `repoPath` by hashing, compressing and encrypting the file.
// Directories will be added recursively.
func (st *Store) Stage(filePath, repoPath string) error {
	return st.StageDir(filePath, prefixSlash(repoPath))
}

// StageDir traverses all files in a directory and calls StageFromReader on them.
func (st *Store) StageDir(filePath, repoPath string) error {
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

			compressAlgo, chooseErr := compress.ChooseCompressAlgo(path, fd)
			if err != nil {
				return chooseErr
			}

			err = st.StageFromReader(currPath, fd, compressAlgo)
		case mode.IsDir():
			err = st.Mkdir(currPath)
		default:
			log.Warningf("Recursive add: Ignoring weird file type: %v")
			return nil
		}

		if err != nil {
			log.WithFields(log.Fields{
				"file_path": filePath,
				"repo_path": repoPath,
				"curr_path": currPath,
			}).Warningf("StageDir: %v", err)
		}

		return nil
	})

	return walkErr
}

// StageFromReader reads data from r, encrypts & compresses it while feeding it to ipfs.
// The resulting hash will be committed to the index.
func (st *Store) StageFromReader(repoPath string, r io.Reader, compressAlgo compress.AlgorithmType) error {
	repoPath = prefixSlash(repoPath)

	st.mu.Lock()
	defer st.mu.Unlock()

	// Control how many bytes are written to the encryption layer:
	sizeAcc := &util.SizeAccumulator{}
	teeR := io.TeeReader(r, sizeAcc)

	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return err
	}

	stream, err := NewFileReader(key, teeR, compressAlgo)
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

	owner, err := st.Owner()
	if err != nil {
		return err
	}

	if _, err := stageFile(st.fs, repoPath, &Hash{mhash}, sizeAcc.Size(), owner.ID(), key); err != nil {
		return err
	}

	return nil
}

func (st *Store) pinOp(repoPath string, doUnpin bool) error {
	node, err := st.fs.LookupNode(repoPath)
	if err != nil {
		return err
	}

	var pinMe []Node

	err = Walk(node, true, func(child Node) error {
		if child.GetType() == NodeTypeFile {
			pinMe = append(pinMe, child)
		}

		return nil
	})

	if err != nil {
		return err
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
	node, err := st.fs.LookupDirectory(repoPath)
	if err != nil {
		return false, err
	}

	return st.IPFS.IsPinned(node.Hash().Multihash)
}

// Touch creates a new empty file.
// It is provided as convenience wrapper around StageFromReader.
func (st *Store) Touch(repoPath string) error {
	return st.StageFromReader(prefixSlash(repoPath), bytes.NewReader([]byte{}), compress.AlgoNone)
}

func (st *Store) makeCheckpointByOwner(ID uint64, oldHash, newHash *Hash, oldPath, newPath string) error {
	owner, err := st.Owner()
	if err != nil {
		return err
	}

	return makeCheckpoint(st.fs, owner.ID(), ID, oldHash, newHash, oldPath, newPath)
}

// Stream returns the stream of the file at `path`.
func (st *Store) Stream(path string) (ipfsutil.Reader, error) {
	st.mu.Lock()
	defer st.mu.Unlock()

	file, err := st.fs.LookupFile(prefixSlash(path))
	if err != nil {
		return nil, err
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

	// Only kill the link of the node to it's parent. If `node` is a directory
	// it already contains the hashes of it's children.
	parentNode, err := node.Parent()
	if err != nil {
		return err
	}

	parent, ok := parentNode.(*Directory)
	if !ok {
		return ErrBadNode
	}

	if err := parent.RemoveChild(node); err != nil {
		return err
	}

	if err := st.fs.StageNode(parent); err != nil {
		return err
	}

	errs := util.Errors{}
	for _, child := range toBeRemoved {
		childPath := NodePath(child)
		if err = st.makeCheckpointByOwner(child.ID(), child.Hash(), nil, childPath, childPath); err != nil {
			return err
		}

		if child.GetType() == NodeTypeFile {
			if err := st.IPFS.Unpin(child.Hash().Multihash); err != nil {
				errs = append(errs, err)
			}
		}
	}

	return errs.ToErr()
}

// List exports a directory listing of `root` up to `depth` levels down.
// TODO: This should use a locked closure.
func (st *Store) List(root string, depth int, visit func(Node) error) error {
	root = prefixSlash(root)

	st.mu.Lock()
	defer st.mu.Unlock()

	node, err := st.fs.LookupDirectory(root)
	if err != nil {
		return err
	}

	if depth < 0 {
		depth = math.MaxInt32
	}

	err = Walk(node, false, func(child Node) error {
		if NodeDepth(child) > depth {
			return nil
		}

		return visit(child)
	})

	if err != nil {
		return err
	}

	return err
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

	newNode, err := st.fs.LookupNode(newPath)
	if err != nil && !IsNoSuchFileError(err) {
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
		return st.makeCheckpointByOwner(
			node.ID(),
			hash,
			hash,
			oldChildPath,
			newChildPath,
		)
	})

	if err != nil {
		return err
	}

	// If the node at newPath was a file, we need to remove it.
	if newNode != nil && newNode.GetType() == NodeTypeFile {
		// TODO: use store.Remove() here (for checkpoints)
		if err := nodeRemove(newNode); err != nil {
			return err
		}
	}

	// NOTE: No pinning information needs to change,
	//       Move() does not influence the Hash() of the file.
	for newPath, file := range newPaths {
		dest, err := mkdirParents(st.fs, newPath)
		if err != nil {
			return err
		}

		if dest == nil {
			return NoSuchFile(newPath)
		}

		// Remove from old Parent:
		if err := nodeRemove(file); err != nil {
			return err
		}

		// Basename might have changed:
		file.SetName(path.Base(newPath))

		// Add to new parent:
		if err := dest.Add(file); err != nil {
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
// TODO: This should use a locked closure.
// TODO: return []Commit?
func (st *Store) Log(visit func(*Commit) error) error {
	st.mu.Lock()
	defer st.mu.Unlock()

	head, err := st.fs.Head()
	if err != nil {
		return err
	}

	var curr Node = head
	for curr != nil {
		currCmt, ok := curr.(*Commit)
		if !ok {
			return ErrBadNode
		}

		if err := visit(currCmt); err != nil {
			return err
		}

		parNode, err := curr.Parent()
		if err != nil {
			return err
		}

		if parNode == nil {
			break
		}

		curr = parNode
	}

	return nil
}

// Reset resets `repoPath` to the state pointed to as `commitRef`.
// If the file did not exist back then, it will be deleted.
// A Reset("/path", "HEAD") equals an unstage operation.
// A Reset("/", "QmXYZ") resets the working tree to QmXYZ.
// Note that uncommitted changes will be lost!
func (st *Store) Reset(repoPath, commitRef string) error {
	st.mu.Lock()
	defer st.mu.Unlock()

	repoPath = prefixSlash(repoPath)
	node, err := st.fs.LookupSettableNode(repoPath)
	if err != nil {
		return err
	}

	cmt, err := resolveCommitRef(st.fs, commitRef)
	if err != nil {
		return err
	}

	switch typ := node.GetType(); typ {
	case NodeTypeFile, NodeTypeDirectory:
		return resetNode(st.fs, node, cmt)
	case NodeTypeCommit:
		return fmt.Errorf("Can't reset a commit (use checkout?)")
	default:
		return fmt.Errorf("BUG: Unhandled node type in reset: %d", typ)
	}
}
