package catfs

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/sahib/config"
	capnp "zombiezen.com/go/capnproto2"

	e "github.com/pkg/errors"
	c "github.com/sahib/brig/catfs/core"
	"github.com/sahib/brig/catfs/db"
	ie "github.com/sahib/brig/catfs/errors"
	"github.com/sahib/brig/catfs/mio"
	"github.com/sahib/brig/catfs/mio/compress"
	n "github.com/sahib/brig/catfs/nodes"
	"github.com/sahib/brig/catfs/vcs"
	"github.com/sahib/brig/util"
	h "github.com/sahib/brig/util/hashlib"
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

	// underlying key/value store
	kv db.Database

	// linker (holds all nodes together)
	lkr *c.Linker

	// garbage collector for dead metadata links
	gc *c.GarbageCollector

	// channel to schedule gc runs and quit the gc loop
	gcControl chan bool

	// channel to schedule auto commits and quit the loop
	autoCommitControl chan bool

	// Actual storage backend (e.g. ipfs or memory)
	bk FsBackend

	// internal config
	cfg *config.Config

	// cache for the isPinned operation
	pinner *Pinner

	// wether this fs is read only and cannot be changed.
	// It can be change by applying patches though.
	readOnly bool
}

var ErrReadOnly = errors.New("fs is read only")

// StatInfo describes the metadata of a single node.
// The concept is comparable to the POSIX stat() call.
type StatInfo struct {
	// Path is the full path to the file
	Path string

	// TreeHash is the hash of the node in the DAG
	TreeHash h.Hash
	// ContentHash is the actual hash of the content
	// (used to test for content equality)
	ContentHash h.Hash
	// BackendHash is the hash under which the file is reachable
	// in the backend.
	BackendHash h.Hash

	// User is the name of the user that modified this node last.
	User string
	// Size in bytes
	Size uint64
	// Inode is a unique number specific to this node
	Inode uint64
	// Depth is the hierarchy level inside of this node (root has 0)
	Depth int
	// ModTime is the last modification timestamp
	ModTime time.Time

	// IsDir tells you if this node is a dir
	IsDir bool
	// IsPinned tells you if this node is pinned (either implicit or explicit)
	IsPinned bool
	// IsExplicit is true when the user pinned this node on purpose
	IsExplicit bool
}

// DiffPair is a pair of nodes.
// It is returned by MakeDiff(), where the source
// is a node on the remote side and the dst node is
// a node on our side.
type DiffPair struct {
	Src StatInfo
	Dst StatInfo
}

// Diff is a list of things that changed between to commits
type Diff struct {
	// Added is a list of nodes that were added newly
	Added []StatInfo

	// Removed is a list of nodes that were removed
	Removed []StatInfo

	// Ignored is a list of nodes that were not considered
	Ignored []StatInfo

	// Missing is a list of nodes that the remoe side is missing
	Missing []StatInfo

	// Moved is a list of nodes that changed path
	Moved []DiffPair

	// Merged is a list of nodes that can be merged automatically
	Merged []DiffPair

	// Conflict is a list of nodes that cannot be merged automatically
	Conflict []DiffPair
}

// Commit gives information about a single commit.
type Commit struct {
	// Hash is the id of this commit
	Hash h.Hash
	// Msg describes the committed contents
	Msg string
	// Tags is a user defined list of tags
	// (tags like HEAD, CURR and INIT are assigned dynamically as exception)
	Tags []string
	// Date is the time when the commit was made
	Date time.Time
}

// Change describes a single change to a node between to version
type Change struct {
	// Path is the node that was changed
	Path string

	// Change describes what was changed
	Change string

	// MovedTo indicates that the node at this Path was moved to
	// another location and that there is no node at this location now.
	MovedTo string

	// WasPreviouslyAt is filled when the node was moved
	// and was previously at another location.
	WasPreviouslyAt string

	// Head is the commit after the change
	Head *Commit

	// Next is the commit before the change
	Next *Commit
}

// ExplicitPin is a pair of path and commit id.
type ExplicitPin struct {
	Path   string
	Commit string
}

/////////////////////
// UTILITY HELPERS //
/////////////////////

func (fs *FS) nodeToStat(nd n.Node) *StatInfo {
	isPinned, isExplicit, err := fs.pinner.IsNodePinned(nd)
	if err != nil {
		log.Warningf("stat: failed to acquire pin state: %v", err)
	}

	var content h.Hash
	if file, ok := nd.(*n.File); ok {
		content = file.ContentHash()
	}

	return &StatInfo{
		Path:        nd.Path(),
		TreeHash:    nd.TreeHash().Clone(),
		User:        nd.User(),
		ModTime:     nd.ModTime(),
		IsDir:       nd.Type() == n.NodeTypeDirectory,
		Inode:       nd.Inode(),
		Size:        nd.Size(),
		Depth:       n.Depth(nd),
		IsPinned:    isPinned,
		IsExplicit:  isExplicit,
		ContentHash: content,
		BackendHash: nd.BackendHash().Clone(),
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

func (fs *FS) handleGcEvent(nd n.Node) bool {
	if nd.Type() != n.NodeTypeFile {
		return true
	}

	file, ok := nd.(*n.File)
	if !ok {
		return true
	}

	content := file.BackendHash()
	log.Infof("unpinning gc'd node %v", content.B58String())

	// This node will not be reachable anymore by brig.
	// Make sure it is also unpinned to save space.
	if err := fs.pinner.Unpin(file.Inode(), file.BackendHash(), true); err != nil {
		log.Warningf("unpinning attempt failed: %v", err)
	}

	// Still return true, no need to stop the GC
	return true
}

///////////////////////////////
// ACTUAL API IMPLEMENTATION //
///////////////////////////////

func (fs *FS) doGcRun() {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	owner, err := fs.lkr.Owner()
	if err != nil {
		log.Warningf("gc: failed to get owner: %v", err)
		return
	}

	log.Debugf("filesystem GC (for %s): running", owner)
	if err := fs.gc.Run(true); err != nil {
		log.Warnf("failed to run GC: %v", err)
	}
}

func NewFilesystem(backend FsBackend, dbPath string, owner string, readOnly bool, fsCfg *config.Config) (*FS, error) {
	kv, err := db.NewBadgerDatabase(dbPath)
	if err != nil {
		return nil, err
	}

	lkr := c.NewLinker(kv)

	if err := lkr.SetOwner(owner); err != nil {
		return nil, err
	}

	pinCache, err := NewPinner(lkr, backend)
	if err != nil {
		return nil, err
	}

	// NOTE: We do not need to validate fsCfg here.
	// This is already done on the side of our config module.
	// (we just need to convert a few keys to the vcs.SyncOptions enum later).

	fs := &FS{
		kv:                kv,
		lkr:               lkr,
		bk:                backend,
		cfg:               fsCfg,
		readOnly:          readOnly,
		gcControl:         make(chan bool),
		autoCommitControl: make(chan bool),
		pinner:            pinCache,
	}

	// Start the garbage collection background task.
	// It will run locked every few seconds and removes unreachable
	// objects from the staging area.
	fs.gc = c.NewGarbageCollector(lkr, kv, fs.handleGcEvent)

	go fs.gcLoop()
	go fs.autoCommitLoop()

	return fs, nil
}

func (fs *FS) gcLoop() {
	gcTicker := time.NewTicker(120 * time.Second)
	defer gcTicker.Stop()
	for {
		select {
		case state := <-fs.gcControl:
			if state {
				fs.doGcRun()
			} else {
				// Quit the gc loop:
				log.Debugf("Quitting the GC loop")
				return
			}
		case <-gcTicker.C:
			fs.doGcRun()
		}
	}
}

func (fs *FS) autoCommitLoop() {
	lastCheck := time.Now()
	checkTicker := time.NewTicker(1 * time.Second)
	defer checkTicker.Stop()

	for {
		select {
		case <-fs.autoCommitControl:
			log.Debugf("quitting the auto commit loop")
			return
		case <-checkTicker.C:
			isEnabled := fs.cfg.Bool("autocommit.enabled")
			if !isEnabled {
				continue
			}

			if time.Since(lastCheck) >= fs.cfg.Duration("autocommit.interval") {
				lastCheck = time.Now()
				msg := fmt.Sprintf("auto commit at »%s«", time.Now().Format(time.RFC822))
				if err := fs.MakeCommit(msg); err != nil && err != ie.ErrNoChange {
					log.Warningf("failed to create auto commit: %v", err)
				}
			}
		}
	}
}

func (fs *FS) Close() error {
	go func() {
		fs.gcControl <- false
	}()

	go func() {
		fs.autoCommitControl <- false
	}()

	if err := fs.pinner.Close(); err != nil {
		log.Warnf("Failed to close pin cache: %v", err)
	}

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

	if fs.readOnly {
		return ErrReadOnly
	}

	srcNd, err := lookupFileOrDir(fs.lkr, src)
	if err != nil {
		return err
	}

	return c.Move(fs.lkr, srcNd, dst)
}

func (fs *FS) Copy(src, dst string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if fs.readOnly {
		return ErrReadOnly
	}

	srcNd, err := lookupFileOrDir(fs.lkr, src)
	if err != nil {
		return err
	}

	_, err = c.Copy(fs.lkr, srcNd, dst)
	return err
}

func (fs *FS) Mkdir(dir string, createParents bool) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if fs.readOnly {
		return ErrReadOnly
	}

	// "brig mkdir ." somehow is able to overwrite everything:
	dir = strings.TrimLeft(path.Clean(dir), ".")
	_, err := c.Mkdir(fs.lkr, dir, createParents)
	return err
}

func (fs *FS) Remove(path string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if fs.readOnly {
		return ErrReadOnly
	}

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

	return fs.nodeToStat(nd), nil
}

// List returns stat info for each node below (and including) root.
// Nodes deeper than maxDepth will not be shown. If maxDepth is a
// negative number, all nodes will be shown.
func (fs *FS) List(root string, maxDepth int) ([]*StatInfo, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// NOTE: This method is highly inefficient:
	//       - iterates over all nodes even if maxDepth is >= 0
	//
	// Fix whenever it proves to be a problem.
	// I don't want to engineer something now until I know what's needed.
	rootNd, err := fs.lkr.LookupNode(root)
	if err != nil {
		return nil, err
	}

	if rootNd.Type() == n.NodeTypeGhost {
		return nil, ie.NoSuchFile(root)
	}

	// Start counting max depth relative to the root:
	if maxDepth >= 0 {
		maxDepth += n.Depth(rootNd)
	}

	result := []*StatInfo{}
	err = n.Walk(fs.lkr, rootNd, false, func(child n.Node) error {
		if maxDepth < 0 || n.Depth(child) <= maxDepth {
			if maxDepth >= 0 && child.Path() == root {
				return nil
			}

			// Ghost nodes should not be visible to the outside.
			if child.Type() == n.NodeTypeGhost {
				return nil
			}

			result = append(result, fs.nodeToStat(child))
		}

		return nil
	})

	sort.Slice(result, func(i, j int) bool {
		iDepth := result[i].Depth
		jDepth := result[j].Depth

		if iDepth == jDepth {
			return result[i].Path < result[j].Path
		}

		return iDepth < jDepth
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

////////////////////////
// PINNING OPERATIONS //
////////////////////////

// preCache makes the backend fetch the data already from the network,
// even though it might not be needed yet.
func (fs *FS) preCache(path string) error {
	stream, err := fs.Cat(path)
	if err != nil {
		return err
	}

	_, err = io.Copy(ioutil.Discard, stream)
	return err
}

func (fs *FS) preCacheInBackground(path string) {
	if !fs.cfg.Bool("pre_cache.enabled") {
		return
	}

	go func() {
		if err := fs.preCache(path); err != nil {
			log.Debugf("failed to pre-cache `%s`: %v", path, err)
		}
	}()
}

func (fs *FS) Pin(path string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	nd, err := lookupFileOrDir(fs.lkr, path)
	if err != nil {
		return err
	}

	if err := fs.pinner.PinNode(nd, true); err != nil {
		return err
	}

	// Make sure the data is available:
	// (this is some sort of `cat path > /dev/null`)
	fs.preCacheInBackground(path)
	return nil
}

func (fs *FS) Unpin(path string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	nd, err := lookupFileOrDir(fs.lkr, path)
	if err != nil {
		return err
	}

	return fs.pinner.UnpinNode(nd, true)
}

func (fs *FS) ListExplicitPins(prefix, fromRef, toRef string) ([]ExplicitPin, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	var from, to *n.Commit
	var err error

	if fromRef != "" {
		from, err = parseRev(fs.lkr, fromRef)
		if err != nil {
			return nil, e.Wrapf(err, "parse from ref")
		}
	}

	if toRef != "" {
		to, err = parseRev(fs.lkr, toRef)
		if err != nil {
			return nil, e.Wrapf(err, "parse to ref")
		}
	}

	results := []ExplicitPin{}
	err = fs.lkr.IterAll(from, to, func(nd n.ModNode, cmt *n.Commit) error {
		if nd.Type() != n.NodeTypeFile {
			return nil
		}

		if !strings.HasPrefix(nd.Path(), prefix) {
			return nil
		}

		_, isExplicit, err := fs.pinner.IsPinned(nd.Inode(), nd.BackendHash())
		if err != nil {
			return err
		}

		// isExplicit implies isPinned.
		if isExplicit {
			results = append(results, ExplicitPin{
				Path:   nd.Path(),
				Commit: cmt.TreeHash().B58String(),
			})
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (fs *FS) ClearExplicitPins(prefix, fromRef, toRef string) (int, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	return fs.setExplicitPins(false, prefix, fromRef, toRef)
}

func (fs *FS) SetExplicitPin(prefix, fromRef, toRef string) (int, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	return fs.setExplicitPins(true, prefix, fromRef, toRef)
}

func (fs *FS) setExplicitPins(doPin bool, prefix, fromRef, toRef string) (int, error) {
	var err error
	var from, to *n.Commit

	if fromRef != "" {
		from, err = parseRev(fs.lkr, fromRef)
		if err != nil {
			return 0, e.Wrapf(err, "parse from ref")
		}
	}

	if toRef != "" {
		to, err = parseRev(fs.lkr, toRef)
		if err != nil {
			return 0, e.Wrapf(err, "parse to ref")
		}
	}

	processed := 0

	return processed, fs.lkr.IterAll(from, to, func(nd n.ModNode, cmt *n.Commit) error {
		if nd.Type() != n.NodeTypeFile {
			return nil
		}

		if !strings.HasPrefix(nd.Path(), prefix) {
			return nil
		}

		if doPin {
			err = fs.pinner.PinNode(nd, true)
		} else {
			err = fs.pinner.UnpinNode(nd, true)
		}

		if err != nil {
			return err
		}

		processed++
		return nil
	})
}

// IsPinned returns true for files and directories that are pinned.
// A directory only counts as pinned if all files and directories
// in it are also pinned.
func (fs *FS) IsPinned(path string) (bool, bool, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	nd, err := lookupFileOrDir(fs.lkr, path)
	if err != nil {
		return false, false, err
	}

	return fs.pinner.IsNodePinned(nd)
}

////////////////////////
// STAGING OPERATIONS //
////////////////////////

func prefixSlash(s string) string {
	if !strings.HasPrefix(s, "/") {
		return "/" + s
	}

	return s
}

// Touch creates an empty file at `path` if it does not exist yet.
// If it exists, it's mod time is being updated to the current time.
func (fs *FS) Touch(path string) error {
	fs.mu.Lock()

	if fs.readOnly {
		fs.mu.Unlock()
		return ErrReadOnly
	}

	nd, err := fs.lkr.LookupNode(path)
	if err != nil && !ie.IsNoSuchFileError(err) {
		fs.mu.Unlock()
		return err
	}

	if nd != nil {
		modNd, ok := nd.(n.ModNode)
		if !ok {
			// Probably a ghost node.
			fs.mu.Unlock()
			return nil
		}

		modNd.SetModTime(time.Now())
		fs.mu.Unlock()
		return nil
	}

	// We may not call Stage() with a lock.
	fs.mu.Unlock()

	// Notthing there, stage an empty file.
	return fs.Stage(prefixSlash(path), bytes.NewReader([]byte{}))
}

// Truncate cuts of the output of the file at `path` to `size`.
// `size` should be between 0 and the size of the file,
// all other values will be ignored.
//
// Note that this is not implemented as an actual IO operation.
// It is possible to go back to a bigger size until the actual
// content was changed via Stage().
func (fs *FS) Truncate(path string, size uint64) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if fs.readOnly {
		return ErrReadOnly
	}

	nd, err := fs.lkr.LookupModNode(path)
	if err != nil {
		return err
	}

	if nd.Type() != n.NodeTypeFile {
		return fmt.Errorf("`%s` is not a file", path)
	}

	nd.SetSize(size)
	return fs.lkr.StageNode(nd)
}

func peekHeader(r io.Reader) ([]byte, io.Reader, error) {
	headerBuf := make([]byte, 4*1024)
	n, err := r.Read(headerBuf)
	if err != nil && err != io.EOF {
		return nil, nil, err
	}

	headerBuf = headerBuf[:n]
	return headerBuf, util.PrefixReader(headerBuf, r), nil
}

func (fs *FS) computePreconditions(path string, rs io.ReadSeeker) (h.Hash, uint64, compress.AlgorithmType, error) {
	// Save a little header of the things we read,
	// but avoid reading it twice.
	headerBuf, pr, err := peekHeader(rs)
	if err != nil {
		return nil, 0, compress.AlgoNone, err
	}

	hashWriter := h.NewHashWriter()
	hashReader := io.TeeReader(pr, hashWriter)

	sizeAcc := &util.SizeAccumulator{}
	sizeReader := io.TeeReader(hashReader, sizeAcc)

	if _, err := io.Copy(ioutil.Discard, sizeReader); err != nil {
		return nil, 0, compress.AlgoNone, err
	}

	// Go back to the beginning of the file:
	if _, err := rs.Seek(0, io.SeekStart); err != nil {
		return nil, 0, compress.AlgoNone, err
	}

	algo, err := compress.GuessAlgorithm(path, headerBuf)
	if err != nil {
		// Use the default algorithm set in the config:
		algo, err = compress.AlgoFromString(fs.cfg.String("compress.default_algo"))
		if err != nil {
			return nil, 0, compress.AlgoNone, err
		}

		log.Warningf("failed to guess suitable zip algo: %v", err)
	}

	log.Debugf("using '%s' compression for file %s", algo, path)
	contentHash := hashWriter.Finalize()
	size := sizeAcc.Size()
	return contentHash, size, algo, nil
}

func deriveKeyFromContent(content h.Hash, size uint64) []byte {
	salt := make([]byte, 8)
	binary.LittleEndian.PutUint64(salt, size)
	return util.DeriveKey(content, salt, 32)
}

func (fs *FS) renewPins(oldFile, newFile *n.File) error {
	pinExplicit := false

	if oldFile != nil {
		oldBackendHash := oldFile.BackendHash()
		if oldBackendHash.Equal(newFile.BackendHash()) {
			// Nothing changed, nothing to do...
			return nil
		}

		_, isExplicit, err := fs.pinner.IsNodePinned(oldFile)
		if err != nil {
			return err
		}

		// If the old file was pinned explicitly, we should also pin
		// the new file explicitly to carry over that info.
		pinExplicit = isExplicit

		if !isExplicit {
			if err := fs.pinner.UnpinNode(oldFile, pinExplicit); err != nil {
				return err
			}
		}
	}

	return fs.pinner.PinNode(newFile, pinExplicit)
}

// Stage reads all data from `r` and stores as content of the node at `path`.
// If `path` already exists, it will be updated.
func (fs *FS) Stage(path string, r io.ReadSeeker) error {
	fs.mu.Lock()

	if fs.readOnly {
		return ErrReadOnly
	}

	path = prefixSlash(path)

	// See if we already have such a file.
	// If not we gonna need to generate new key for it
	// based on the content hash.
	var oldFile *n.File
	oldNode, err := fs.lkr.LookupNode(path)

	// Check that we're handling the right kind of node.
	// We should be able to add on-top of ghosts, but directorie
	// are pointless as input.
	if err == nil {
		switch oldNode.Type() {
		case n.NodeTypeDirectory:
			fs.mu.Unlock()
			return fmt.Errorf("Cannot stage over directory: %v", path)
		case n.NodeTypeGhost:
			// Act like there was no such node:
			err = ie.NoSuchFile(path)
		case n.NodeTypeFile:
			var ok bool
			oldFile, ok = oldNode.(*n.File)
			if !ok {
				fs.mu.Unlock()
				return ie.ErrBadNode
			}
		}
	}

	if err != nil && !ie.IsNoSuchFileError(err) {
		fs.mu.Unlock()
		return err
	}

	// Copy self, so we do not need to fear race conditions below.
	var oldFileCopy *n.File
	if oldFile != nil {
		oldFileCopy = oldFile.Copy(oldFile.Inode()).(*n.File)
	}

	// Unlock the fs lock while adding the stream to the backend.
	// This is not required for the data integrity of the fs.
	fs.mu.Unlock()

	contentHash, size, compressAlgo, err := fs.computePreconditions(path, r)
	if err != nil {
		return err
	}

	var key []byte
	if oldFileCopy == nil {
		// only create a new key for new files.
		// The key depends on the content hash and the size.
		key = deriveKeyFromContent(contentHash, size)
	} else {
		if contentHash.Equal(oldFileCopy.ContentHash()) {
			log.Infof("content of %s did not change; not modifying", path)
			return nil
		}

		// Next generations of the same file get the same key.
		key = oldFileCopy.Key()
	}

	stream, err := mio.NewInStream(r, key, compressAlgo)
	if err != nil {
		return err
	}

	backendHash, err := fs.bk.Add(stream)
	if err != nil {
		return err
	}

	// Lock it again for the metadata staging:
	fs.mu.Lock()
	defer fs.mu.Unlock()

	newFile, err := c.Stage(fs.lkr, path, contentHash, backendHash, size, key)
	if err != nil {
		return err
	}

	return fs.renewPins(oldFileCopy, newFile)
}

////////////////////
// I/O OPERATIONS //
////////////////////

// Cat will open a file read-only and expose it's underlying data as stream.
// If no such path is known or it was deleted, nil is returned as stream.
func (fs *FS) Cat(path string) (mio.Stream, error) {
	fs.mu.Lock()

	file, err := fs.lkr.LookupFile(path)
	if err == ie.ErrBadNode {
		fs.mu.Unlock()
		return nil, ie.NoSuchFile(path)
	}

	if err != nil {
		fs.mu.Unlock()
		return nil, err
	}

	// Copy all attributes, since accessing them beyond the lock might be racy.
	size := file.Size()
	backendHash := file.BackendHash().Clone()
	key := make([]byte, len(file.Key()))
	copy(key, file.Key())

	fs.mu.Unlock()

	// NOTE: This part of the code is not locked by fs.mu!
	rawStream, err := fs.bk.Cat(backendHash)
	if err != nil {
		return nil, err
	}

	stream, err := mio.NewOutStream(rawStream, key)
	if err != nil {
		return nil, err
	}

	// Truncate stream to file size. Data stream might be bigger
	// for example when fuse decided to truncate the file, but
	// did not flush it already.
	return mio.LimitStream(stream, size), nil
}

// Open returns a file like object that can be used for modifying a file in memory.
// If you want to have seekable read-only stream, use Cat(), it has less overhead.
func (fs *FS) Open(path string) (*Handle, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	nd, err := fs.lkr.LookupNode(path)
	if err != nil {
		return nil, err
	}

	file, ok := nd.(*n.File)
	if !ok {
		return nil, fmt.Errorf("Can only open files: %v", path)
	}

	return newHandle(fs, file, fs.readOnly), nil
}

////////////////////
// VCS OPERATIONS //
////////////////////

// MakeCommit bundles all staged changes into one commit described by `msg`.
// If no changes were made since the last call to MakeCommit() ErrNoConflict
// is returned.
func (fs *FS) MakeCommit(msg string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	owner, err := fs.lkr.Owner()
	if err != nil {
		return err
	}

	if err := fs.lkr.MakeCommit(owner, msg); err != nil {
		return err
	}

	return nil
}

func (fs *FS) Head() (string, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	head, err := fs.lkr.Head()
	if err != nil {
		return "", err
	}

	return head.TreeHash().B58String(), nil
}

func (fs *FS) Curr() (string, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	status, err := fs.lkr.Status()
	if err != nil {
		return "", err
	}

	return status.TreeHash().B58String(), nil
}

// History returns all modifications of a node with one entry per commit.
func (fs *FS) History(path string) ([]Change, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	nd, err := fs.lkr.LookupModNode(path)
	if err != nil {
		return nil, err
	}

	status, err := fs.lkr.Status()
	if err != nil {
		return nil, err
	}

	hist, err := vcs.History(fs.lkr, nd, status, nil)
	if err != nil {
		return nil, err
	}

	hashToRef, err := fs.buildCommitHashToRefTable()
	if err != nil {
		return nil, err
	}

	entries := []Change{}
	for _, change := range hist {
		head := &Commit{
			Hash: change.Head.TreeHash().Clone(),
			Msg:  change.Head.Message(),
			Tags: hashToRef[change.Head.TreeHash().B58String()],
			Date: change.Head.ModTime(),
		}

		var next *Commit
		if change.Next != nil {
			next = &Commit{
				Hash: change.Next.TreeHash().Clone(),
				Msg:  change.Next.Message(),
				Tags: hashToRef[change.Next.TreeHash().B58String()],
				Date: change.Next.ModTime(),
			}
		}

		entries = append(entries, Change{
			Path:            change.Curr.Path(),
			Change:          change.Mask.String(),
			Head:            head,
			Next:            next,
			MovedTo:         change.MovedTo,
			WasPreviouslyAt: change.WasPreviouslyAt,
		})
	}

	return entries, nil
}

func (fs *FS) buildSyncCfg() (*vcs.SyncOptions, error) {
	// Helper method to easily pin depending on a condition variable
	doPinOrUnpin := func(doPin, explicit bool, nd n.ModNode) {
		file, ok := nd.(*n.File)
		if !ok {
			// Non-files are simply ignored.
			return
		}

		op := fs.pinner.UnpinNode
		opName := "unpin"
		if doPin {
			op = fs.pinner.PinNode
			opName = "pin"
		}

		if err := op(file, explicit); err != nil {
			log.Warningf("Failed to %s (hash: %v)", opName, file.BackendHash())
		}
	}

	conflictStrategy := vcs.ConflictStrategyFromString(
		fs.cfg.String("sync.conflict_strategy"),
	)

	if conflictStrategy == vcs.ConflictStragetyUnknown {
		return nil, fmt.Errorf("unknown conflict strategy: %v", conflictStrategy)
	}

	return &vcs.SyncOptions{
		ConflictStrategy: conflictStrategy,
		IgnoreDeletes:    fs.cfg.Bool("sync.ignore_removed"),
		IgnoreMoves:      fs.cfg.Bool("sync.ignore_moved"),
		OnAdd: func(newNd n.ModNode) bool {
			doPinOrUnpin(true, false, newNd)
			return true
		},
		OnRemove: func(oldNd n.ModNode) bool {
			doPinOrUnpin(false, true, oldNd)
			return true
		},
		OnMerge: func(newNd, oldNd n.ModNode) bool {
			_, isExplicit, err := fs.pinner.IsNodePinned(oldNd)
			if err != nil {
				log.Warnf(
					"failed to check pin status of old node `%s` (%v)",
					oldNd.Path(),
					oldNd.BackendHash(),
				)
			}

			// Pin new node with old pin state:
			doPinOrUnpin(true, isExplicit, newNd)
			doPinOrUnpin(false, true, oldNd)
			return true
		},
		OnConflict: func(src, dst n.ModNode) bool {
			// Don't need to do something,
			// conflict files will not get a pin by default.
			return true
		},
	}, nil
}

// Sync will synchronize the state of two filesystems.
// If one of filesystems have unstaged changes, they will be committted first.
// If our filesystem was changed by Sync(), a new merge commit will also be created.
func (fs *FS) Sync(remote *FS) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if fs.readOnly {
		return ErrReadOnly
	}

	syncCfg, err := fs.buildSyncCfg()
	if err != nil {
		return err
	}

	return vcs.Sync(remote.lkr, fs.lkr, syncCfg)
}

func (fs *FS) MakeDiff(remote *FS, headRevOwn, headRevRemote string) (*Diff, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	srcHead, err := parseRev(remote.lkr, headRevRemote)
	if err != nil {
		return nil, e.Wrapf(err, "parse remote ref")
	}

	dstHead, err := parseRev(fs.lkr, headRevOwn)
	if err != nil {
		return nil, e.Wrapf(err, "parse own ref")
	}

	syncCfg, err := fs.buildSyncCfg()
	if err != nil {
		return nil, err
	}

	realDiff, err := vcs.MakeDiff(remote.lkr, fs.lkr, srcHead, dstHead, syncCfg)
	if err != nil {
		return nil, e.Wrapf(err, "make diff")
	}

	// "fake" is the diff that we give to the outside.
	// Internally we have a bit more knowledge.
	fakeDiff := &Diff{}

	// Convert the simple slice parts:
	for _, nd := range realDiff.Added {
		fakeDiff.Added = append(fakeDiff.Added, *fs.nodeToStat(nd))
	}

	for _, nd := range realDiff.Ignored {
		fakeDiff.Ignored = append(fakeDiff.Ignored, *fs.nodeToStat(nd))
	}

	for _, nd := range realDiff.Removed {
		fakeDiff.Removed = append(fakeDiff.Removed, *fs.nodeToStat(nd))
	}

	for _, nd := range realDiff.Missing {
		fakeDiff.Missing = append(fakeDiff.Missing, *fs.nodeToStat(nd))
	}

	// And also convert the slightly more complex pairs:
	for _, pair := range realDiff.Moved {
		fakeDiff.Moved = append(fakeDiff.Moved, DiffPair{
			Src: *fs.nodeToStat(pair.Src),
			Dst: *fs.nodeToStat(pair.Dst),
		})
	}

	for _, pair := range realDiff.Merged {
		fakeDiff.Merged = append(fakeDiff.Merged, DiffPair{
			Src: *fs.nodeToStat(pair.Src),
			Dst: *fs.nodeToStat(pair.Dst),
		})
	}

	for _, pair := range realDiff.Conflict {
		fakeDiff.Conflict = append(fakeDiff.Conflict, DiffPair{
			Src: *fs.nodeToStat(pair.Src),
			Dst: *fs.nodeToStat(pair.Dst),
		})
	}

	return fakeDiff, nil
}

func (fs *FS) buildCommitHashToRefTable() (map[string][]string, error) {
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
			key := cmt.TreeHash().B58String()
			hashToRef[key] = append(hashToRef[key], name)
		}
	}

	return hashToRef, nil
}

// Log returns a list of commits starting with the staging commit until the
// initial commit. For each commit, metadata is collected.
func (fs *FS) Log() ([]Commit, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	hashToRef, err := fs.buildCommitHashToRefTable()
	if err != nil {
		return nil, err
	}

	entries := []Commit{}
	return entries, c.Log(fs.lkr, func(cmt *n.Commit) error {
		entries = append(entries, Commit{
			Hash: cmt.TreeHash().Clone(),
			Msg:  cmt.Message(),
			Tags: hashToRef[cmt.TreeHash().B58String()],
			Date: cmt.ModTime(),
		})

		return nil
	})
}

func (fs *FS) Reset(path, rev string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if fs.readOnly {
		return ErrReadOnly
	}

	if path == "/" || path == "" {
		return fs.Checkout(rev, false)
	}

	cmt, err := parseRev(fs.lkr, rev)
	if err != nil {
		return err
	}

	oldNode, err := vcs.ResetNode(fs.lkr, cmt, path)
	if err != nil {
		return err
	}

	// The old node does not necessarily exist:
	if oldNode != nil {
		if err := fs.pinner.UnpinNode(oldNode, false); err != nil {
			return err
		}
	}

	// Cannot (un)pin non-existing file anymore.
	newNode, err := fs.lkr.LookupNode(path)
	if ie.IsNoSuchFileError(err) {
		return nil
	}

	if err != nil {
		return err
	}

	return fs.pinner.PinNode(newNode, false)
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
func (fs *FS) Tag(rev, name string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	cmt, err := parseRev(fs.lkr, rev)
	if err != nil {
		return e.Wrap(err, "parse ref")
	}

	return fs.lkr.SaveRef(name, cmt)
}

// RemoveTag removes a previously created tag.
func (fs *FS) RemoveTag(name string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	return fs.lkr.RemoveRef(name)
}

func (fs *FS) FilesByContent(contents []h.Hash) (map[string]StatInfo, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	files, err := fs.lkr.FilesByContents(contents)
	if err != nil {
		return nil, err
	}

	infos := make(map[string]StatInfo)
	for content, file := range files {
		infos[content] = *fs.nodeToStat(file)
	}

	return infos, nil
}

func (fs *FS) ScheduleGCRun() {
	// Putting a value into gcControl might block,
	// so better do it in the background.
	go func() {
		fs.gcControl <- true
	}()
}

func (fs *FS) writeLastPatchIndex(index int64) error {
	fromIndexData := []byte(strconv.FormatInt(index, 10))
	return fs.lkr.MetadataPut("fs.last-merge-index", fromIndexData)
}

// MakePatch creates a binary patch with all file changes starting with
// `fromRev`. Note that commit information is not exported, only individual
// file and directory changes.
//
// The byte structured returned by this method may change at any point
// and may not be relied upon.
//
// The `remoteName` is the name of the remote we're creating the patch for.
// It's only used for display purpose in the commit message.
func (fs *FS) MakePatch(fromRev string, folders []string, remoteName string) ([]byte, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	haveStagedChanges, err := fs.lkr.HaveStagedChanges()
	if err != nil {
		return nil, err
	}

	// Commit changes if there are any.
	// This is a little unfortunate implication on how the current
	// way of sending getting patches work. Creating a patch itself
	// works with a staging commit, but the versioning does not work
	// anymore then, since the same version might have a different
	// set of changes.
	if haveStagedChanges {
		owner, err := fs.lkr.Owner()
		if err != nil {
			return nil, err
		}

		msg := fmt.Sprintf("»%s« merged with you", remoteName)
		if err := fs.lkr.MakeCommit(owner, msg); err != nil {
			return nil, err
		}
	}

	from, err := parseRev(fs.lkr, fromRev)
	if err != nil {
		return nil, err
	}

	patch, err := vcs.MakePatch(fs.lkr, from, folders)
	if err != nil {
		return nil, err
	}

	msg, err := patch.ToCapnp()
	if err != nil {
		return nil, err
	}

	return msg.Marshal()
}

// ApplyPatch reads the binary patch coming from MakePatch and tries to apply it.
func (fs *FS) ApplyPatch(data []byte) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	msg, err := capnp.Unmarshal(data)
	if err != nil {
		return err
	}

	patch := &vcs.Patch{}
	if err := patch.FromCapnp(msg); err != nil {
		return err
	}

	if err := vcs.ApplyPatch(fs.lkr, patch); err != nil {
		return err
	}

	// Remember what patch index we merged last.
	// This info can be read via LastPatchIndex() to determine
	// the next version to get from the remote.
	fromIndexData := []byte(strconv.FormatInt(patch.CurrIndex, 10))
	if err := fs.lkr.MetadataPut("fs.last-merge-index", fromIndexData); err != nil {
		return err
	}

	owner, err := fs.lkr.Owner()
	if err != nil {
		return err
	}

	cmtMsg := fmt.Sprintf("apply patch") // TODO: Better message.
	if err := fs.lkr.MakeCommit(owner, cmtMsg); err != nil {
		// An empty patch is perfectly valid:
		if err == ie.ErrNoChange {
			return nil
		}

		return err
	}

	return nil
}

// LastPatchIndex will return the current version of this filesystem
// regarding patch state.
func (fs *FS) LastPatchIndex() (int64, error) {
	fromIndexData, err := fs.lkr.MetadataGet("fs.last-merge-index")
	if err != nil && err != db.ErrNoSuchKey {
		return -1, err
	}

	// If we did not merge yet with anyone we have to
	// ask for a full fetch.
	if err == db.ErrNoSuchKey {
		return 0, nil
	}

	return strconv.ParseInt(string(fromIndexData), 10, 64)
}
