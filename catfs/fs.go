package catfs

import (
	"bytes"
	"crypto/rand"
	"io"
	"strings"
	"sync"

	"github.com/disorganizer/brig/catfs/db"
	"github.com/disorganizer/brig/catfs/mio"
	"github.com/disorganizer/brig/catfs/mio/compress"
	n "github.com/disorganizer/brig/catfs/nodes"
	"github.com/disorganizer/brig/util"
	h "github.com/disorganizer/brig/util/hashlib"
)

// FS (short for Filesystem) is the central API entry for everything related to
// paths.  It exposes a POSIX-like interface where path are mapped to the
// actual underlying hashes and the associated metadata.
//
// Additionally it supports version control commands like MakeCommit(),
// Checkout() etc.  The API is file-centric, i.e. directories are created on
// the fly for some operations like Stage(). Empty directories can be created
// via Mkdir() though.
type FS struct {
	mu sync.Mutex

	kv  db.Database
	lkr *Linker
	bk  FsBackend
}

func NewFilesystem(dbPath, owner string) (*FS, error) {
	return &FS{}, nil
}

func (fs *FS) Close() error {
	return nil
}

func (fs *FS) Export(w io.Writer) error {
	return nil
}

func (fs *FS) Import(r io.Reader) error {
	return nil
}

/////////////////////
// CORE OPERATIONS //
/////////////////////

func (fs *FS) Move(src, dst string) error {
	return nil
}

func (fs *FS) Mkdir(path string, createParents bool) error {
	return nil
}

func (fs *FS) Remove(path string) error {
	return nil
}

type NodeInfo struct {
	Path  string
	Type  int
	Size  uint64
	Inode uint64
}

func (fs *FS) Stat(path string) (*NodeInfo, error) {
	return nil, nil
}

////////////////////////
// PINNING OPERATIONS //
////////////////////////

func (fs *FS) pin(path string, op func(hash h.Hash) error) error {
	nd, err := fs.lkr.LookupNode(path)
	if err != nil {
		return err
	}

	return n.Walk(fs.lkr, nd, true, func(child n.Node) error {
		if child.Type() == n.NodeTypeFile {
			if err := op(nd.Hash()); err != nil {
				return err
			}
		}

		return nil
	})
}

func (fs *FS) Pin(path string) error {
	return fs.pin(path, fs.bk.Pin)
}

func (fs *FS) Unpin(path string) error {
	return fs.pin(path, fs.bk.Unpin)
}

func (fs *FS) IsPinned(path string) (bool, error) {
	// TODO: What happens for directories?
	return false, nil
}

////////////////////////
// STAGING OPERATIONS //
////////////////////////

func (fs *FS) Touch(path string) error {
	return fs.Stage(prefixSlash(path), bytes.NewReader([]byte{}))
}

func prefixSlash(s string) string {
	if !strings.HasPrefix(s, "/") {
		return "/" + s
	}

	return s
}

func (fs *FS) Stage(path string, r io.Reader) error {
	return nil
	path = prefixSlash(path)

	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Control how many bytes are written to the encryption layer:
	sizeAcc := &util.SizeAccumulator{}
	teeR := io.TeeReader(r, sizeAcc)

	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return err
	}

	stream, err := mio.NewInStream(teeR, key, compress.AlgoSnappy)
	if err != nil {
		return err
	}

	hash, err := fs.bk.Add(stream)
	if err != nil {
		return err
	}

	if err := fs.bk.Pin(hash); err != nil {
		return err
	}

	owner, err := fs.lkr.Owner()
	if err != nil {
		return err
	}

	nu := NodeUpdate{
		Hash:   hash,
		Key:    key,
		Author: owner.String(),
		Size:   sizeAcc.Size(),
	}

	_, err = stage(fs.lkr, path, &nu)
	return err
}

////////////////////
// I/O OPERATIONS //
////////////////////

// Cat will open a file read-only and expose it's underlying data as stream.
// If no such path is known or it was deleted, nil is returned as stream.
func (fs *FS) Cat(path string) (mio.Stream, error) {
	return nil, nil
}

// Open returns a file like object that can be used for modifying a file in memory.
// If you want to have seekable read-only stream, use Cat(), it has less overhead.
func (fs *FS) Open(path string) (*Handle, error) {
	return nil, nil
}

////////////////////
// VCS OPERATIONS //
////////////////////

// MakeCommit bundles all staged changes into one commit described by `msg`.
// If no changes were made since the last call to MakeCommit() ErrNoConflict
// is returned (TODO: move to errors package)
func (fs *FS) MakeCommit(msg string) error {
	return nil
}

// History returns all modifications of a node with one entry per commit.
func (fs *FS) History(path string) error {
	return nil
}

// Sync will synchronize the state of two filesystems.
// If one of filesystems have unstaged changes, they will be committted first.
// If our filesystem was changed by Sync(), a new merge commit will also be created.
func (fs *FS) Sync(remote *FS) error {
	return nil
}

type Diff struct {
	Ignored  map[string]*NodeInfo
	Removed  map[string]*NodeInfo
	Added    map[string]*NodeInfo
	Merged   map[string]*NodeInfo
	Conflict map[string]*NodeInfo
}

func (fs *FS) Diff(remote *FS) (*Diff, error) {
	return nil, nil
}

type LogEntry struct {
}

func (fs *FS) Log() ([]LogEntry, error) {
	return nil, nil
}

func (fs *FS) Reset(path, rev string) error {
	return nil
}
