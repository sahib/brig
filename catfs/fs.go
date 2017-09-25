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

// TODO: I hate to duplicate that struct here.
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

type StatInfo struct {
	Path  string
	Hash  h.Hash
	Size  uint64
	Inode uint64
	IsDir bool
}

type HistEntry struct {
	Path   string
	Change string
	Ref    h.Hash
}

type DiffPair struct {
	Src StatInfo
	Dst StatInfo
}

type Diff struct {
	Added   []StatInfo
	Removed []StatInfo
	Ignored []StatInfo

	Merged   []DiffPair
	Conflict []DiffPair
}

// TODO: Decide on naming: rev(ision), refname and tag.
type LogEntry struct {
	Hash h.Hash
	Msg  string
	Tags []string
	Date time.Time
}

/////////////////////
// UTILITY HELPERS //
/////////////////////

func nodeToStat(nd n.Node) *StatInfo {
	return &StatInfo{
		Path:  nd.Path(),
		Hash:  nd.Hash().Clone(),
		IsDir: nd.Type() == n.NodeTypeDirectory,
		Inode: nd.Inode(),
		Size:  nd.Size(),
	}
}

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

///////////////////////////////
// ACTUAL API IMPLEMENTATION //
///////////////////////////////

func NewFilesystem(backend FsBackend, dbPath string, owner *Person) (*FS, error) {
	kv, err := db.NewDiskDatabase(dbPath)
	if err != nil {
		return nil, err
	}

	lkr := c.NewLinker(kv)

	person := n.NewPerson(owner.Name, owner.Hash)
	if err := lkr.SetOwner(person); err != nil {
		return nil, err
	}

	fs := &FS{
		kv:       kv,
		lkr:      lkr,
		gc:       c.NewGarbageCollector(lkr, kv, nil),
		gcTicker: time.NewTicker(5 * time.Second),
		bk:       backend,
	}

	go func() {
		for range fs.gcTicker.C {
			fs.mu.Lock()

			log.Debugf("gc: running")
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

func (fs *FS) Stat(path string) (*StatInfo, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	nd, err := fs.lkr.LookupNode(path)
	if err != nil {
		return nil, err
	}

	if nd.Type() == n.NodeTypeGhost {
		return nil, ie.NoSuchFile(path)
	}

	return nodeToStat(nd), nil
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

	file, err := fs.lkr.LookupFile(path)
	if err == ie.ErrBadNode {
		return nil, ie.NoSuchFile(path)
	}

	if err != nil {
		return nil, err
	}

	rawStream, err := fs.bk.Cat(file.Content())
	if err != nil {
		return nil, err
	}

	return mio.NewOutStream(rawStream, file.Key())
}

// Open returns a file like object that can be used for modifying a file in memory.
// If you want to have seekable read-only stream, use Cat(), it has less overhead.
func (fs *FS) Open(path string) (*Handle, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	file, err := fs.lkr.LookupFile(path)
	if err != nil {
		return nil, err
	}

	return newHandle(file), nil
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

func (fs *FS) Head() (string, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	head, err := fs.lkr.Head()
	if err != nil {
		return "", err
	}

	return head.Hash().B58String(), nil
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
			Ref:    change.Head.Hash().Clone(),
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

	return vcs.Sync(remote.lkr, fs.lkr, nil)
}

func (fs *FS) MakeDiff(remote *FS) (*Diff, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	realDiff, err := vcs.MakeDiff(remote.lkr, fs.lkr, nil)
	if err != nil {
		return nil, err
	}

	// "fake" is the diff that we give to the outside.
	// Internally we have a bit more knowledge.
	fakeDiff := &Diff{}

	// Convert the simple slice parts:
	for _, nd := range realDiff.Added {
		fakeDiff.Added = append(fakeDiff.Added, *nodeToStat(nd))
	}

	for _, nd := range realDiff.Ignored {
		fakeDiff.Ignored = append(fakeDiff.Added, *nodeToStat(nd))
	}

	for _, nd := range realDiff.Removed {
		fakeDiff.Removed = append(fakeDiff.Removed, *nodeToStat(nd))
	}

	// And also convert the slightly more complex pairs:
	for _, pair := range realDiff.Merged {
		fakeDiff.Merged = append(fakeDiff.Merged, DiffPair{
			Src: *nodeToStat(pair.Src),
			Dst: *nodeToStat(pair.Dst),
		})
	}

	for _, pair := range realDiff.Conflict {
		fakeDiff.Conflict = append(fakeDiff.Conflict, DiffPair{
			Src: *nodeToStat(pair.Src),
			Dst: *nodeToStat(pair.Dst),
		})
	}

	return fakeDiff, nil
}

// Log returns a list of commits starting with the staging commit until the
// initial commit. For each commit, metadata is collected.
func (fs *FS) Log() ([]LogEntry, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	names, err := fs.lkr.ListRefs()
	if err != nil {
		return nil, err
	}

	hashToRef := make(map[string][]string)

	for _, name := range names {
		cmt, err := fs.lkr.ResolveRef(name)
		if err != nil {
			return nil, err
		}

		if cmt != nil {
			key := cmt.Hash().B58String()
			hashToRef[key] = append(hashToRef[key], name)
		}
	}

	entries := []LogEntry{}
	return entries, c.Log(fs.lkr, func(cmt *n.Commit) error {
		entries = append(entries, LogEntry{
			Hash: cmt.Hash().Clone(),
			Msg:  cmt.Message(),
			Tags: hashToRef[cmt.Hash().B58String()],
			Date: cmt.ModTime(),
		})

		return nil
	})
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

// Tag saves a human readable name for the revision pointed to by `rev`.
// There are three pre-defined tags available:
//
// - HEAD: The last full commit.
// - CURR: The current commit (== staging commit)
// - INIT: the initial commit.
//
// The tagname is case-insensitive.
// See TODO for more details on what is allowed as `rev`.
func (fs *FS) Tag(rev, name string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	cmt, err := parseRev(fs.lkr, rev)
	if err != nil {
		return err
	}

	return fs.lkr.SaveRef(name, cmt)
}
