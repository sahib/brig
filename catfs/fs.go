package catfs

import (
	"bytes"
	"crypto/rand"
	"io"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"

	c "github.com/disorganizer/brig/catfs/core"
	"github.com/disorganizer/brig/catfs/db"
	ie "github.com/disorganizer/brig/catfs/errors"
	"github.com/disorganizer/brig/catfs/mio"
	"github.com/disorganizer/brig/catfs/mio/compress"
	n "github.com/disorganizer/brig/catfs/nodes"
	"github.com/disorganizer/brig/catfs/vcs"
	"github.com/disorganizer/brig/util"
	h "github.com/disorganizer/brig/util/hashlib"
)

type Person struct {
	Name string
	Hash h.Hash
}

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

	// underlying key/value store
	kv db.Database

	// linker (holds all nodes together)
	lkr *c.Linker

	// garbage collector for dead metadata links
	gc *c.GarbageCollector

	// ticker that drives the gc background routine
	gcTicker *time.Ticker

	// Actual storage backend (e.g. ipfs or memory)
	bk FsBackend
}

func NewFilesystem(backend FsBackend, dbPath string, owner *Person) (*FS, error) {
	kv, err := db.NewDiskDatabase(dbPath)
	if err != nil {
		return nil, err
	}

	lkr := c.NewLinker(kv)

	if err := lkr.SetOwner(n.NewPerson(owner.Name, owner.Hash)); err != nil {
		return nil, err
	}

	// Make sure the garbage collector is run in constant time intervals.

	fs := &FS{
		kv:       kv,
		lkr:      lkr,
		gc:       c.NewGarbageCollector(lkr, kv, nil),
		gcTicker: time.NewTicker(5 * time.Second),
		bk:       backend,
	}

	go func() {
		for timestamp := range fs.gcTicker.C {
			fs.mu.Lock()

			log.Debugf("gc: running at %v", timestamp)
			if err := fs.gc.Run(true); err != nil {
				log.Warnf("failed to run GC: %v", err)
			}

			fs.mu.Unlock()
		}
	}()

	return fs, nil
}

func (fs *FS) Close() error {
	// Stop the GC loop
	fs.gcTicker.Stop()
	return fs.kv.Close()
}

func (fs *FS) Export(w io.Writer) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	return fs.kv.Export(w)
}

func (fs *FS) Import(r io.Reader) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if err := fs.kv.Import(r); err != nil {
		return err
	}

	// disk (probably) changed, delete memcache:
	fs.lkr.MemIndexClear()
	return nil
}

/////////////////////
// CORE OPERATIONS //
/////////////////////

func lookupFileOrDir(lkr *c.Linker, path string) (n.ModNode, error) {
	nd, err := lkr.LookupNode(path)
	if err != nil {
		return nil, err
	}

	if nd == nil || nd.Type() == n.NodeTypeGhost {
		return nil, ie.NoSuchFile(path)
	}

	modNd, ok := nd.(n.ModNode)
	if !ok {
		return nil, ie.ErrBadNode
	}

	return modNd, nil
}

func (fs *FS) Move(src, dst string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	srcNd, err := lookupFileOrDir(fs.lkr, src)
	if err != nil {
		return err
	}

	return c.Move(fs.lkr, srcNd, dst)
}

func (fs *FS) Mkdir(path string, createParents bool) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	_, err := c.Mkdir(fs.lkr, path, createParents)
	return err
}

func (fs *FS) Remove(path string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	nd, err := lookupFileOrDir(fs.lkr, path)
	if err != nil {
		return err
	}

	_, _, err = c.Remove(fs.lkr, nd, true, true)
	return err
}

type NodeInfo struct {
	Path  string
	Size  uint64
	Inode uint64
	IsDir bool
}

func (fs *FS) Stat(path string) (*NodeInfo, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	nd, err := fs.lkr.LookupNode(path)
	if err != nil {
		return nil, err
	}

	if nd.Type() == n.NodeTypeGhost {
		return nil, ie.NoSuchFile(path)
	}

	return &NodeInfo{
		Path:  nd.Path(),
		IsDir: nd.Type() == n.NodeTypeDirectory,
		Inode: nd.Inode(),
		Size:  nd.Size(),
	}, nil
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
	fs.mu.Lock()
	defer fs.mu.Unlock()

	return fs.pin(path, fs.bk.Pin)
}

func (fs *FS) Unpin(path string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	return fs.pin(path, fs.bk.Unpin)
}

func (fs *FS) IsPinned(path string) (bool, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// TODO: What happens for directories?
	return false, nil
}

////////////////////////
// STAGING OPERATIONS //
////////////////////////

func (fs *FS) Touch(path string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	return fs.Stage(prefixSlash(path), bytes.NewReader([]byte{}))
}

func prefixSlash(s string) string {
	if !strings.HasPrefix(s, "/") {
		return "/" + s
	}

	return s
}

func (fs *FS) Stage(path string, r io.Reader) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

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

	nu := c.NodeUpdate{
		Hash:   hash,
		Key:    key,
		Author: owner.String(),
		Size:   sizeAcc.Size(),
	}

	_, err = c.Stage(fs.lkr, path, &nu)
	return err
}

////////////////////
// I/O OPERATIONS //
////////////////////

// Cat will open a file read-only and expose it's underlying data as stream.
// If no such path is known or it was deleted, nil is returned as stream.
func (fs *FS) Cat(path string) (mio.Stream, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	return nil, nil
}

// Open returns a file like object that can be used for modifying a file in memory.
// If you want to have seekable read-only stream, use Cat(), it has less overhead.
func (fs *FS) Open(path string) (*Handle, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	return nil, nil
}

////////////////////
// VCS OPERATIONS //
////////////////////

// MakeCommit bundles all staged changes into one commit described by `msg`.
// If no changes were made since the last call to MakeCommit() ErrNoConflict
// is returned (TODO: move to errors package)
func (fs *FS) MakeCommit(msg string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	owner, err := fs.lkr.Owner()
	if err != nil {
		return err
	}

	return fs.lkr.MakeCommit(owner, msg)
}

type HistEntry struct {
	Path   string
	Change string
	Ref    string
}

// History returns all modifications of a node with one entry per commit.
func (fs *FS) History(path string) ([]HistEntry, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	nd, err := fs.lkr.LookupModNode(path)
	if err != nil {
		return nil, err
	}

	head, err := fs.lkr.Head()
	if err != nil {
		return nil, err
	}

	hist, err := vcs.History(fs.lkr, nd, head, nil)
	if err != nil {
		return nil, err
	}

	entries := []HistEntry{}
	for _, change := range hist {
		entries = append(entries, HistEntry{
			Path:   change.Curr.Path(),
			Change: change.Mask.String(),
			Ref:    change.Head.String(),
		})
	}

	return entries, nil
}

// Sync will synchronize the state of two filesystems.
// If one of filesystems have unstaged changes, they will be committted first.
// If our filesystem was changed by Sync(), a new merge commit will also be created.
//
// TODO: Provide way to configure sync config.
func (fs *FS) Sync(remote *FS) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	return vcs.Sync(fs.lkr, remote.lkr, nil)
}

type Diff struct {
	Ignored  map[string]*NodeInfo
	Removed  map[string]*NodeInfo
	Added    map[string]*NodeInfo
	Merged   map[string]*NodeInfo
	Conflict map[string]*NodeInfo
}

func (fs *FS) Diff(remote *FS) (*Diff, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	return nil, nil
}

type LogEntry struct {
	Ref  string
	Msg  string
	Date time.Time
}

func (fs *FS) Log() ([]LogEntry, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	entries := []LogEntry{}
	err := c.Log(fs.lkr, func(cmt *n.Commit) error {
		entries = append(entries, LogEntry{
			Ref:  cmt.Hash().B58String(),
			Msg:  cmt.Message(),
			Date: cmt.ModTime(),
		})
		return nil
	})

	if err != nil {
		return nil, err
	}

	return entries, nil
}

func (fs *FS) Reset(path, rev string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	nd, err := fs.lkr.LookupNode(path)
	if err != nil {
		return err
	}

	cmt, err := parseRev(fs.lkr, rev)
	if err != nil {
		return err
	}

	return fs.lkr.CheckoutFile(cmt, nd)
}

func (fs *FS) Checkout(rev string, force bool) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	cmt, err := parseRev(fs.lkr, rev)
	if err != nil {
		return err
	}

	return fs.lkr.CheckoutCommit(cmt, force)
}

func (fs *FS) Tag(rev, name string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	cmt, err := parseRev(fs.lkr, rev)
	if err != nil {
		return err
	}

	return fs.lkr.SaveRef(rev, cmt)
}
