package core

// Layout of the key/value store:
//
// objects/<NODE_HASH>                   => NODE_METADATA
// tree/<FULL_NODE_PATH>                 => NODE_HASH
// inode/<NODE_HASH>                     => NODE_HASH
//
// stage/objects/<NODE_HASH>             => NODE_METADATA
// stage/tree/<FULL_NODE_PATH>           => NODE_HASH
// stage/STATUS                          => COMMIT_METADATA
//
// stats/node-count/<COUNT>              => UINT64
// refs/<REFNAME>                        => NODE_HASH
// metadata/                             => BYTES (Caller defined data)
//
// Defined by caller:
// metadata/id      => USER_ID
// metadata/hash    => USER_HASH
// metadata/version => DB_FORMAT_VERSION_NUMBER
//
// NODE is either a Commit, a Directory or a File.
// FULL_NODE_PATH may contain slashes and in case of directories,
// it will contain a trailing slash.
//
// The following refs are defined by the system:
// HEAD -> Points to the latest finished commit, or nil.
// CURR -> Points to the staging commit.
//
// In git terminology, this file implements the following commands:
// - git add:    StageNode(): Create and Update Nodes.
// - git reset:  UnstageNode(): Reset to last known state.
// - git status: Status()
// - git commit: MakeCommit()

import (
	"encoding/binary"
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/sahib/brig/catfs/db"
	ie "github.com/sahib/brig/catfs/errors"
	n "github.com/sahib/brig/catfs/nodes"
	h "github.com/sahib/brig/util/hashlib"
	"github.com/sahib/brig/util/trie"
	capnp "zombiezen.com/go/capnproto2"
)

// Linker implements the basic logic of brig's data model
// It uses an underlying key/value database to
// storea a Merkle-DAG with versioned metadata,
// similar to what git does internally.
type Linker struct {
	kv db.Database

	// root of the filesystem
	root *n.Directory

	// Path lookup trie
	ptrie *trie.Node

	// B58Hash to node
	index map[string]n.Node

	// UID to node
	inodeIndex map[uint64]n.Node

	// user name to user id (cached)
	userIndex map[string]int32

	// Cache for the linker owner.
	owner string
}

// NewFilesystem returns a new lkr, ready to use. It assumes the key value store
// is working and does no check on this.
func NewLinker(kv db.Database) *Linker {
	lkr := &Linker{kv: kv}
	lkr.MemIndexClear()
	return lkr
}

//  MemIndexAdd adds `nd` to the in memory index.
func (lkr *Linker) MemIndexAdd(nd n.Node) {
	lkr.index[nd.Hash().B58String()] = nd
	lkr.inodeIndex[nd.Inode()] = nd
	lkr.ptrie.InsertWithData(nd.Path(), nd)
}

// MemIndexSwap updates an entry of the in memory index, by deleting
// the old entry referenced by oldHash (may be nil). This is necessary
// to ensure that old hashes do not resolve to the new, updated instance.
// If the old instance is needed, it will be loaded as new instance.
// You should not need to call this function, except when implementing own Nodes.
func (lkr *Linker) MemIndexSwap(nd n.Node, oldHash h.Hash) {
	if oldHash != nil {
		delete(lkr.index, oldHash.B58String())
	}

	lkr.MemIndexAdd(nd)
}

// MemIndexPurge removes `nd` from the memory index.
func (lkr *Linker) MemIndexPurge(nd n.Node) {
	delete(lkr.inodeIndex, nd.Inode())
	delete(lkr.index, nd.Hash().B58String())
	lkr.ptrie.Lookup(nd.Path()).Remove()
}

// MemIndexClear resets the memory index to zero.
func (lkr *Linker) MemIndexClear() {
	lkr.ptrie = trie.NewNode()
	lkr.index = make(map[string]n.Node)
	lkr.inodeIndex = make(map[uint64]n.Node)
}

//////////////////////////
// COMMON NODE HANDLING //
//////////////////////////

// NextInode() returns a unique identifier, used to identify a single node. You
// should not need to call this function, except when implementing own nodes.
func (lkr *Linker) NextInode() uint64 {
	nodeCount, err := lkr.kv.Get("stats", "node-count")
	if err != nil && err != db.ErrNoSuchKey {
		return 0
	}

	// nodeCount might be nil on startup:
	cnt := uint64(1)
	if nodeCount != nil {
		cnt = binary.BigEndian.Uint64(nodeCount) + 1
	}

	cntBuf := make([]byte, 8)
	binary.BigEndian.PutUint64(cntBuf, cnt)

	batch := lkr.kv.Batch()
	batch.Put(cntBuf, "stats", "node-count")
	if err := batch.Flush(); err != nil {
		return 0
	}

	return cnt
}

func (lkr *Linker) FilesByContents(contents []h.Hash) (map[string]*n.File, error) {
	result := make(map[string]*n.File)

	err := lkr.kv.Keys(func(key []string) error {
		// Filter non-node storage:
		fullKey := strings.Join(key, "/")
		if !strings.HasPrefix(fullKey, "/objects") &&
			!strings.HasPrefix(fullKey, "/stage/objects") {
			return nil
		}

		data, err := lkr.kv.Get(key...)
		if err != nil {
			return err
		}

		nd, err := n.UnmarshalNode(data)
		if err != nil {
			return err
		}

		if nd.Type() != n.NodeTypeFile {
			return nil
		}

		file, ok := nd.(*n.File)
		if !ok {
			return ie.ErrBadNode
		}

		for _, content := range contents {
			if content.Equal(file.Content()) {
				result[content.B58String()] = file
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// loadNode loads an individual object by its hash from the object store. It
// will return nil if the hash is not there.
func (lkr *Linker) loadNode(hash h.Hash) (n.Node, error) {
	var data []byte
	var err error

	b58hash := hash.B58String()

	// First look in the stage:
	loadableBuckets := [][]string{
		[]string{"stage", "objects", b58hash},
		[]string{"objects", b58hash},
	}

	for _, bucketPath := range loadableBuckets {
		data, err = lkr.kv.Get(bucketPath...)
		if err != nil && err != db.ErrNoSuchKey {
			return nil, err
		}

		if data != nil {
			return n.UnmarshalNode(data)
		}
	}

	// Damn, no hash found:
	return nil, nil
}

// NodeByHash returns the node identified by hash.
// If no such hash could be found, nil is returned.
func (lkr *Linker) NodeByHash(hash h.Hash) (n.Node, error) {
	// Check if we have this this node in the memory cache already:
	b58Hash := hash.B58String()
	if cachedNode, ok := lkr.index[b58Hash]; ok {
		return cachedNode, nil
	}

	// Node was not in the cache, load directly from kv.
	nd, err := lkr.loadNode(hash)
	if err != nil {
		return nil, err
	}

	if nd == nil {
		// log.Warningf("Could not load hash `%s`", hash.B58String())
		return nil, nil
	}

	lkr.MemIndexSwap(nd, nil)
	return nd, nil
}

func appendDot(path string) string {
	// path.Join() calls path.Clean() which in turn
	// removes the '.' at the end when trying to join that.
	// But since we use the dot to mark directories we shouldn't do that.
	if strings.HasSuffix(path, "/") {
		return path + "."
	}

	return path + "/."
}

// ResolveNode resolves a path to a hash and resolves the corresponding node by
// calling NodeByHash(). If no node could be resolved, nil is returned.
// It does not matter if the node was deleted in the meantime. If so,
// a Ghost node is returned which stores the last known state.
func (lkr *Linker) ResolveNode(nodePath string) (n.Node, error) {
	// Check if it's cached already:
	trieNode := lkr.ptrie.Lookup(nodePath)
	if trieNode != nil && trieNode.Data != nil {
		return trieNode.Data.(n.Node), nil
	}

	fullPaths := [][]string{
		[]string{"stage", "tree", nodePath},
		[]string{"tree", nodePath},
	}

	for _, fullPath := range fullPaths {
		b58Hash, err := lkr.kv.Get(fullPath...)
		if err != nil && err != db.ErrNoSuchKey {
			return nil, err
		}

		bhash, err := h.FromB58String(string(b58Hash))
		if err != nil {
			return nil, err
		}

		if bhash != nil {
			return lkr.NodeByHash(h.Hash(bhash))
		}
	}

	// Return nil if nothing found:
	return nil, nil
}

// StageNode inserts a modified node to the staging area, making sure the
// modification is persistent and part of the staging commit. All parent
// directories of the node in question will be staged automatically. If there
// was no modification it will be a (quite expensive) NOOP.
func (lkr *Linker) StageNode(nd n.Node) (err error) {
	batch := lkr.kv.Batch()
	defer func() {
		if err != nil {
			batch.Rollback()
		} else {
			err = batch.Flush()
		}
	}()

	if err := lkr.stageNodeRecursive(batch, nd); err != nil {
		return err
	}

	// Update the staging commit's root hash:
	status, err := lkr.Status()
	if err != nil {
		return fmt.Errorf("Failed to retrieve status: %v", err)
	}

	root, err := lkr.Root()
	if err != nil {
		return err
	}

	status.SetRoot(root.Hash())
	return lkr.saveStatus(status)
}

// NodeByInode resolves a node by it's unique ID.
// It will return nil if no corresponding node was found.
func (lkr *Linker) NodeByInode(uid uint64) (n.Node, error) {
	b58Hash, err := lkr.kv.Get("inode", strconv.FormatUint(uid, 10))
	if err != nil && err != db.ErrNoSuchKey {
		return nil, err
	}

	hash, err := h.FromB58String(string(b58Hash))
	if err != nil {
		return nil, err
	}

	return lkr.NodeByHash(hash)
}

func (lkr *Linker) stageNodeRecursive(batch db.Batch, nd n.Node) error {
	if nd.Type() == n.NodeTypeCommit {
		return fmt.Errorf("BUG: Commits cannot be staged; Use MakeCommit()")
	}

	data, err := n.MarshalNode(nd)
	if err != nil {
		return err
	}

	b58Hash := nd.Hash().B58String()
	batch.Put(data, "stage", "objects", b58Hash)

	uidKey := strconv.FormatUint(nd.Inode(), 10)
	batch.Put([]byte(nd.Hash().B58String()), "inode", uidKey)

	hashPath := []string{"stage", "tree", nd.Path()}
	if nd.Type() == n.NodeTypeDirectory {
		hashPath = append(hashPath, ".")
	}

	batch.Put([]byte(b58Hash), hashPath...)

	// Remember/Update this node in the cache if it's not yet there:
	lkr.MemIndexAdd(nd)

	// We need to save parent directories too, in case the hash changed:
	// Note that this will create many pointless directories in staging.
	// That's okay since we garbage collect it every few seconds
	// on a higher layer.
	par, err := nd.Parent(lkr)
	if err != nil {
		return err
	}

	if par != nil {
		if err := lkr.stageNodeRecursive(batch, par); err != nil {
			return err
		}
	}

	return nil
}

/////////////////////
// COMMIT HANDLING //
/////////////////////

// SetMergeMarker sets the current status to be a merge commit.
// Note that this function only will have a result when MakeCommit() is called afterwards.
// Otherwise, the changes will not be written to disk.
func (lkr *Linker) SetMergeMarker(with string, remoteHead h.Hash) error {
	status, err := lkr.Status()
	if err != nil {
		return err
	}

	status.SetMergeMarker(with, remoteHead)
	return lkr.saveStatus(status)
}

// MakeCommit creates a new full commit in the version history.
// The current staging commit is finalized with `author` and `message`
// and gets saved. A new, identical staging commit is created pointing
// to the root of the now new HEAD.
func (lkr *Linker) MakeCommit(author string, message string) (err error) {
	batch := lkr.kv.Batch()
	defer func() {
		if err != nil {
			batch.Rollback()
		} else {
			err = batch.Flush()
		}
	}()

	head, err := lkr.Head()
	if err != nil && !ie.IsErrNoSuchRef(err) {
		return err
	}

	status, err := lkr.Status()
	if err != nil {
		return err
	}

	// Only compare with previous if we have a HEAD yet.
	if head != nil {
		if status.Root().Equal(head.Root()) {
			return ie.ErrNoChange
		}
	}

	rootDir, err := lkr.Root()
	if err != nil {
		return err
	}

	// Go over all files/directories and save them in tree & objects.
	// Note that this will only move nodes that are reachable from the current
	// commit root. Intermediate nodes will not be copied.
	exportedInodes := make(map[uint64]bool)
	err = n.Walk(lkr, rootDir, true, func(child n.Node) error {
		data, err := n.MarshalNode(child)
		if err != nil {
			return err
		}

		b58Hash := child.Hash().B58String()
		batch.Put(data, "objects", b58Hash)
		exportedInodes[child.Inode()] = true

		childPath := child.Path()
		if child.Type() == n.NodeTypeDirectory {
			childPath = appendDot(childPath)
		}

		batch.Put([]byte(b58Hash), "tree", childPath)
		return nil
	})

	if err != nil {
		return err
	}

	if head != nil {
		if err := status.SetParent(lkr, head); err != nil {
			return err
		}
	}

	// NOTE: `head` may be nil, if it couldn't be resolved,
	//        or (maybe more likely) if this is the first commit.
	if err := status.BoxCommit(author, message); err != nil {
		return err
	}

	statusData, err := n.MarshalNode(status)
	if err != nil {
		return err
	}

	statusB58Hash := status.Hash().B58String()
	batch.Put(statusData, "objects", statusB58Hash)

	if err := lkr.SaveRef("HEAD", status); err != nil {
		return err
	}

	// Check if we have already tagged the initial commit.
	if _, err := lkr.ResolveRef("INIT"); err != nil {
		if !ie.IsErrNoSuchRef(err) {
			// Some other error happened.
			return err
		}

		// This is probably the first commit. Tag it.
		if err := lkr.SaveRef("INIT", status); err != nil {
			return err
		}
	}

	// Fixate the moved paths in the stage:
	if err := lkr.commitMoveMapping(status, exportedInodes); err != nil {
		return err
	}

	// Clear the staging area.
	toClear := [][]string{
		[]string{"stage", "objects"},
		[]string{"stage", "tree"},
		[]string{"stage", "moves"},
	}

	for _, key := range toClear {
		batch.Clear(key...)
	}

	newStatus, err := n.NewEmptyCommit(lkr.NextInode())
	if err != nil {
		return err
	}

	newStatus.SetParent(lkr, status)
	newStatus.SetRoot(status.Root())
	return lkr.saveStatus(newStatus)
}

///////////////////////
// METADATA HANDLING //
///////////////////////

// MetadataPut remembers a value persisntenly identified by `key`.
// It can be used as single-level key value store for user purposes.
func (lkr *Linker) MetadataPut(key string, value []byte) error {
	batch := lkr.kv.Batch()
	batch.Put([]byte(value), "metadata", key)
	return batch.Flush()
}

// MetadataGet retriesves a previosuly put key value pair.
// It will return nil if no such value could be retrieved.
func (lkr *Linker) MetadataGet(key string) ([]byte, error) {
	return lkr.kv.Get("metadata", key)
}

////////////////////////
// OWNERSHIP HANDLING //
////////////////////////

func (lkr *Linker) Owner() (string, error) {
	if lkr.owner != "" {
		return lkr.owner, nil
	}

	data, err := lkr.MetadataGet("owner")
	if err != nil {
		return "", err
	}

	// Cache owner, we don't want to reload it again and again.
	// It will usually not change during runtime, except SetOwner
	// is called (which is invalidating the cache anyways)
	lkr.owner = string(data)
	return lkr.owner, nil
}

func (lkr *Linker) SetOwner(owner string) error {
	lkr.owner = owner
	return lkr.MetadataPut("owner", []byte(owner))
}

////////////////////////
// REFERENCE HANDLING //
////////////////////////

// ResolveRef resolves the hash associated with `refname`. If the ref could not
// be resolved, ErrNoSuchRef is returned. Typically, Node will be a Commit. But
// there are no technical restrictions on which node typ to use.
func (lkr *Linker) ResolveRef(refname string) (n.Node, error) {
	refname = strings.ToLower(refname)

	nUps := 0
	for idx := len(refname) - 1; idx >= 0; idx-- {
		if refname[idx] == '^' {
			nUps++
		} else {
			break
		}
	}

	// Strip the ^s:
	refname = refname[:len(refname)-nUps]

	// Special case: the status commit is not part of the normal object store.
	// Still make it able to resolve it by it's refname "curr".
	if refname == "curr" || refname == "status" {
		return lkr.Status()
	}

	b58Hash, err := lkr.kv.Get("refs", refname)
	if err != nil && err != db.ErrNoSuchKey {
		return nil, err
	}

	if len(b58Hash) == 0 {
		return nil, ie.ErrNoSuchRef(refname)
	}

	hash, err := h.FromB58String(string(b58Hash))
	if err != nil {
		return nil, err
	}

	nd, err := lkr.NodeByHash(h.Hash(hash))
	if err != nil {
		return nil, err
	}

	// Possibly advance a few commits until we hit the one
	// the user required.
	cmt, ok := nd.(*n.Commit)
	if ok {
		for i := 0; i < nUps; i++ {
			parentNd, err := cmt.Parent(lkr)
			if err != nil {
				return nil, err
			}

			if parentNd == nil {
				// TODO: log a warning here?
				break
			}

			parentCmt, ok := parentNd.(*n.Commit)
			if !ok {
				break
			}

			cmt = parentCmt
		}

		nd = cmt
	}

	return nd, nil
}

// SaveRef stores a reference to `nd` persistently. The caller is responsbiel
// to ensure that the node is already in the blockstore, otherwise it won't be
// resolvable.
func (lkr *Linker) SaveRef(refname string, nd n.Node) error {
	batch := lkr.kv.Batch()
	refname = strings.ToLower(refname)
	batch.Put([]byte(nd.Hash().B58String()), "refs", refname)
	return batch.Flush()
}

// ListRefs lists all currently known refs.
func (lkr *Linker) ListRefs() ([]string, error) {
	refs := []string{}
	walker := func(key []string) error {
		if len(key) <= 2 {
			return nil
		}

		refs = append(refs, key[2])
		return nil
	}

	if err := lkr.kv.Keys(walker, "refs"); err != nil {
		return nil, err
	}

	return refs, nil
}

func (lkr *Linker) RemoveRef(refname string) error {
	batch := lkr.kv.Batch()
	batch.Erase("refs", refname)
	return batch.Flush()
}

// Head is just a shortcut for ResolveRef("HEAD").
func (lkr *Linker) Head() (*n.Commit, error) {
	nd, err := lkr.ResolveRef("HEAD")
	if err != nil {
		return nil, err
	}

	cmt, ok := nd.(*n.Commit)
	if !ok {
		return nil, fmt.Errorf("oh-oh, HEAD is not a Commit... %v", nd)
	}

	return cmt, nil
}

// MemSetRoot sets the current root, but does not store it yet. It's supposed
// to be called after in-memory modifications. Only implementors of new Nodes
// might need to call this function.
func (lkr *Linker) MemSetRoot(root *n.Directory) {
	lkr.root = root
}

// Root returns the current root directory of CURR.
// It is never nil when err is nil.
func (lkr *Linker) Root() (*n.Directory, error) {
	if lkr.root != nil {
		return lkr.root, nil
	}

	status, err := lkr.Status()
	if err != nil {
		return nil, err
	}

	return lkr.DirectoryByHash(status.Root())
}

// Status returns the current staging commit.
// It is never nil, unless err is nil.
func (lkr *Linker) Status() (cmt *n.Commit, err error) {
	cmt, err = lkr.loadStatus()
	if err != nil {
		return nil, err
	}

	if cmt != nil {
		return cmt, nil
	}

	batch := lkr.kv.Batch()
	defer func() {
		if err != nil {
			batch.Rollback()
		} else {
			err = batch.Flush()
		}
	}()

	// Shoot, no commit exists yet.
	// We need to create an initial one.
	cmt, err = n.NewEmptyCommit(lkr.NextInode())
	if err != nil {
		return nil, err
	}

	// Setup a new commit and set root from last HEAD or new one.
	head, err := lkr.Head()
	if err != nil && !ie.IsErrNoSuchRef(err) {
		return nil, err
	}

	var rootHash h.Hash

	if ie.IsErrNoSuchRef(err) {
		// There probably wasn't a HEAD yet.
		// TODO: Replace ResolveDirectory -> Resolve* can be removed.
		if root, err := lkr.ResolveDirectory("/"); err == nil {
			rootHash = root.Hash()
		} else {
			// No root directory then. Create a shiny new one and stage it.
			inode := lkr.NextInode()
			newRoot, err := n.NewEmptyDirectory(lkr, nil, "/", lkr.owner, inode)
			if err != nil {
				return nil, err
			}

			// Can't call StageNode(), since that would call Status(),
			// causing and endless loop of grief and doom.
			if err := lkr.stageNodeRecursive(batch, newRoot); err != nil {
				return nil, err
			}

			rootHash = newRoot.Hash()
		}
	} else {
		if err := cmt.SetParent(lkr, head); err != nil {
			return nil, err
		}

		rootHash = head.Root()
	}

	cmt.SetRoot(rootHash)

	if err := lkr.saveStatus(cmt); err != nil {
		return nil, err
	}

	return cmt, nil
}

func (lkr *Linker) loadStatus() (*n.Commit, error) {
	data, err := lkr.kv.Get("stage", "STATUS")
	if err != nil && err != db.ErrNoSuchKey {
		return nil, err
	}

	if data == nil {
		return nil, nil
	}

	msg, err := capnp.Unmarshal(data)
	if err != nil {
		return nil, err
	}

	// It's there already. Just unmarshal it.
	cmt := &n.Commit{}
	if err := cmt.FromCapnp(msg); err != nil {
		return nil, err
	}

	return cmt, nil
}

// saveStatus copies cmt to stage/STATUS.
func (lkr *Linker) saveStatus(cmt *n.Commit) (err error) {
	batch := lkr.kv.Batch()
	defer func() {
		if err != nil {
			batch.Rollback()
		} else {
			err = batch.Flush()
		}
	}()

	head, err := lkr.Head()
	if err != nil && !ie.IsErrNoSuchRef(err) {
		return err
	}

	if head != nil {
		if err := cmt.SetParent(lkr, head); err != nil {
			return err
		}
	}

	if err := cmt.BoxCommit(n.AuthorOfStage, ""); err != nil {
		return err
	}

	data, err := n.MarshalNode(cmt)
	if err != nil {
		return err
	}

	inode := strconv.FormatUint(cmt.Inode(), 10)
	batch.Put(data, "stage", "STATUS")
	batch.Put([]byte(cmt.Hash().B58String()), "inode", inode)

	if err := lkr.SaveRef("CURR", cmt); err != nil {
		return err
	}

	return nil
}

/////////////////////////////////
// CONVINIENT ACCESS FUNCTIONS //
/////////////////////////////////

// LookupNode takes the root node and tries to resolve the path from there.
// Deleted paths are recognized in contrast to ResolveNode.
// If a path does not exist NoSuchFile is returned.
func (lkr *Linker) LookupNode(repoPath string) (n.Node, error) {
	root, err := lkr.Root()
	if err != nil {
		return nil, err
	}

	return root.Lookup(lkr, repoPath)
}

// TODO: Write tests for the At() variants.
func (lkr *Linker) LookupNodeAt(cmt *n.Commit, repoPath string) (n.Node, error) {
	root, err := lkr.DirectoryByHash(cmt.Root())
	if err != nil {
		return nil, err
	}

	if root == nil {
		return nil, nil
	}

	return root.Lookup(lkr, repoPath)
}

func (lkr *Linker) LookupModNode(repoPath string) (n.ModNode, error) {
	node, err := lkr.LookupNode(repoPath)
	if err != nil {
		return nil, err
	}

	if node == nil {
		return nil, nil
	}

	snode, ok := node.(n.ModNode)
	if !ok {
		return nil, ie.ErrBadNode
	}

	return snode, nil
}

func (lkr *Linker) LookupModNodeAt(cmt *n.Commit, repoPath string) (n.ModNode, error) {
	node, err := lkr.LookupNodeAt(cmt, repoPath)
	if err != nil {
		return nil, err
	}

	if node == nil {
		return nil, nil
	}

	snode, ok := node.(n.ModNode)
	if !ok {
		return nil, ie.ErrBadNode
	}

	return snode, nil
}

// DirectoryByHash calls NodeByHash and attempts to convert
// it to a Directory as convinience.
func (lkr *Linker) DirectoryByHash(hash h.Hash) (*n.Directory, error) {
	nd, err := lkr.NodeByHash(hash)
	if err != nil {
		return nil, err
	}

	if nd == nil {
		return nil, nil
	}

	dir, ok := nd.(*n.Directory)
	if !ok {
		return nil, ie.ErrBadNode
	}

	return dir, nil
}

// ResolveDirectory calls ResolveNode and converts the result to a Directory.
func (lkr *Linker) ResolveDirectory(dirpath string) (*n.Directory, error) {
	nd, err := lkr.ResolveNode(appendDot(path.Clean(dirpath)))
	if err != nil {
		return nil, err
	}

	if nd == nil {
		return nil, nil
	}

	dir, ok := nd.(*n.Directory)
	if !ok {
		return nil, ie.ErrBadNode
	}

	return dir, nil
}

// LookupDirectory calls LookupNode and converts the result to a Directory.
// TODO: Now that we have ghosts - does it make sense to do what Resolve()
//       does? i.e. just lookup the dir in the kv and use that.
//       This woild be likely more efficient and would save some code.
func (lkr *Linker) LookupDirectory(repoPath string) (*n.Directory, error) {
	nd, err := lkr.LookupNode(repoPath)
	if err != nil {
		return nil, err
	}

	if nd == nil {
		return nil, nil
	}

	dir, ok := nd.(*n.Directory)
	if !ok {
		return nil, ie.ErrBadNode
	}

	return dir, nil
}

// FileByHash calls NodeByHash and converts the result to a File.
func (lkr *Linker) FileByHash(hash h.Hash) (*n.File, error) {
	nd, err := lkr.NodeByHash(hash)
	if err != nil {
		return nil, err
	}

	file, ok := nd.(*n.File)
	if !ok {
		return nil, ie.ErrBadNode
	}

	return file, nil
}

// ResolveFile calls ResolveNode and converts the result to a file.
func (lkr *Linker) ResolveFile(filepath string) (*n.File, error) {
	nd, err := lkr.ResolveNode(filepath)
	if err != nil {
		return nil, err
	}

	if nd == nil {
		return nil, nil
	}

	file, ok := nd.(*n.File)
	if !ok {
		return nil, ie.ErrBadNode
	}

	return file, nil
}

// LookupFile calls LookupNode and converts the result to a file.
func (lkr *Linker) LookupFile(repoPath string) (*n.File, error) {
	nd, err := lkr.LookupNode(repoPath)
	if err != nil {
		return nil, err
	}

	if nd == nil {
		return nil, nil
	}

	file, ok := nd.(*n.File)
	if !ok {
		return nil, ie.ErrBadNode
	}

	return file, nil
}

// LookupGhost calls LookupNode and converts the result to a ghost.
func (lkr *Linker) LookupGhost(repoPath string) (*n.Ghost, error) {
	nd, err := lkr.LookupNode(repoPath)
	if err != nil {
		return nil, err
	}

	if nd == nil {
		return nil, nil
	}

	ghost, ok := nd.(*n.Ghost)
	if !ok {
		return nil, ie.ErrBadNode
	}

	return ghost, nil
}

// CommitByHash lookups a commit by it's hash.
// If the commit could not be found, nil is returned.
func (lkr *Linker) CommitByHash(hash h.Hash) (*n.Commit, error) {
	nd, err := lkr.NodeByHash(hash)
	if err != nil {
		return nil, err
	}

	if nd == nil {
		return nil, nil
	}

	cmt, ok := nd.(*n.Commit)
	if !ok {
		return nil, ie.ErrBadNode
	}

	return cmt, nil
}

// HaveStagedChanges returns true if there were changes in the staging area.
// If an error occurs, the first return value is undefined.
func (lkr *Linker) HaveStagedChanges() (bool, error) {
	head, err := lkr.Head()
	if err != nil && !ie.IsErrNoSuchRef(err) {
		return false, err
	}

	if ie.IsErrNoSuchRef(err) {
		// There is no HEAD yet. Assume we have changes.
		return true, nil
	}

	status, err := lkr.Status()
	if err != nil {
		return false, err
	}

	// Check if the root hashes of CURR and HEAD differ.
	return !status.Root().Equal(head.Root()), nil
}

// CheckoutCommit resets the current staging commit back to the commit
// referenced by cmt. If force is false, it will check if there any staged errors in
// the staging area and return ErrStageNotEmpty if there are any. If force is
// true, all changes will be overwritten.
// TODO: write test for this.
func (lkr *Linker) CheckoutCommit(cmt *n.Commit, force bool) (err error) {
	// Check if the staging area is empty if no force given:
	batch := lkr.kv.Batch()
	defer func() {
		if err != nil {
			batch.Rollback()
		} else {
			err = batch.Flush()
		}
	}()

	if !force {
		haveStaged, err := lkr.HaveStagedChanges()
		if err != nil {
			return err
		}

		if haveStaged {
			return ie.ErrStageNotEmpty
		}
	}

	status, err := lkr.Status()
	if err != nil {
		return err
	}

	root, err := lkr.DirectoryByHash(cmt.Root())
	if err != nil {
		return err
	}

	// Set the current virtual in-memory cached root
	lkr.MemSetRoot(root)
	status.SetRoot(cmt.Root())

	// Invalidate the cache, causing NodeByHash and ResolveNode to load the
	// file from the boltdb again:
	lkr.MemIndexClear()
	return lkr.saveStatus(status)
}

// CheckoutFile resets a certain file to the state it had in cmt. If the file
// did not exist back then, it will be deleted. `nd` is usually retrieved by
// calling ResolveNode() and sorts.
func (lkr *Linker) CheckoutFile(cmt *n.Commit, ndPath string) (err error) {
	root, err := lkr.DirectoryByHash(cmt.Root())
	if err != nil {
		return err
	}

	if root == nil {
		return fmt.Errorf("no root to reset to")
	}

	currNode, err := lkr.LookupModNode(ndPath)
	if err != nil && !ie.IsNoSuchFileError(err) {
		return err
	}

	oldNode, err := root.Lookup(lkr, ndPath)
	if err != nil && !ie.IsNoSuchFileError(err) {
		return err
	}

	// Invalidate the respective index entry, so the instances gets reloaded:
	if currNode != nil {
		err = n.Walk(lkr, currNode, true, func(child n.Node) error {
			lkr.MemIndexPurge(child)
			return nil
		})
	}

	if err != nil {
		return err
	}

	var par *n.Directory
	if ndPath != "/" {
		par, err = lkr.LookupDirectory(path.Dir(ndPath))
		if err != nil {
			return err
		}
	}

	if par == nil {
		return fmt.Errorf("checkout by commit if you want to checkout previous roots")
	}

	// Make sure the actual checkout will land as one batch on disk:
	batch := lkr.kv.Batch()
	defer func() {
		if err != nil {
			batch.Rollback()
		} else {
			err = batch.Flush()
		}
	}()

	// Remove old node, if needed.
	if currNode != nil {
		if err := par.RemoveChild(lkr, currNode); err != nil {
			return err
		}

		lkr.MemIndexPurge(currNode)

		if err := lkr.StageNode(par); err != nil {
			return err
		}
	}

	// old Node might not have yet existed back then.
	// If so, simply do not re-add it.
	if oldNode != nil {
		if err := par.Add(lkr, oldNode); err != nil {
			return err
		}

		return lkr.StageNode(oldNode)
	}

	return nil
}

// AddMoveMapping takes note that the the node `from` has been moved to `to`
// in the staging commit.
func (lkr *Linker) AddMoveMapping(from, to n.Node) (err error) {
	// Make sure the actual checkout will land as one batch on disk:

	srcInode := strconv.FormatUint(from.Inode(), 10)
	srcToDstKey := []string{"stage", "moves", srcInode}

	dstInode := strconv.FormatUint(to.Inode(), 10)
	dstToSrcKey := []string{"stage", "moves", dstInode}

	batch := lkr.kv.Batch()
	if _, err = lkr.kv.Get(srcToDstKey...); err == db.ErrNoSuchKey {
		line := []byte(fmt.Sprintf("> inode %d", to.Inode()))
		batch.Put(line, srcToDstKey...)
		batch.Put(line, "stage", "moves", "overlay", srcInode)
	}

	// Also remember the move in the other direction.
	// This might come in handy for the
	if _, err = lkr.kv.Get(dstToSrcKey...); err == db.ErrNoSuchKey {
		line := []byte(fmt.Sprintf("< inode %d", from.Inode()))
		batch.Put(line, dstToSrcKey...)
		batch.Put(line, "stage", "moves", "overlay", dstInode)
	}

	return batch.Flush()
}

func (lkr *Linker) parseMoveMappingLine(line string) (n.Node, MoveDir, error) {
	splitLine := strings.SplitN(line, " ", 3)
	if len(splitLine) < 3 {
		return nil, 0, fmt.Errorf("Malformed stage move line: `%s`", line)
	}

	dir := moveDirFromString(splitLine[0])
	if dir == MoveDirUnknown {
		return nil, 0, fmt.Errorf("Unrecognized move direction `%s`", splitLine[0])
	}

	switch splitLine[1] {
	case "inode":
		inode, err := strconv.ParseUint(splitLine[2], 10, 64)
		if err != nil {
			return nil, 0, err
		}

		node, err := lkr.NodeByInode(inode)
		if err != nil {
			return nil, 0, err
		}

		return node, dir, nil
	case "hash":
		hash, err := h.FromB58String(splitLine[2])
		if err != nil {
			return nil, 0, err
		}

		node, err := lkr.NodeByHash(hash)
		if err != nil {
			return nil, 0, err
		}

		return node, dir, nil
	default:
		return nil, 0, fmt.Errorf("Unsupported move map type: %s", splitLine[1])
	}
}

func (lkr *Linker) commitMoveMapping(status *n.Commit, exported map[uint64]bool) error {
	batch := lkr.kv.Batch()
	walker := func(key []string) error {
		inode, err := strconv.ParseUint(key[len(key)-1], 10, 64)
		if err != nil {
			return err
		}

		// Only export move mapping that relate to nodes that were actually
		// exported from staging. We do not want to export intermediate moves.
		if _, ok := exported[inode]; !ok {
			return nil
		}

		data, err := lkr.kv.Get(key...)
		if err != nil {
			return err
		}

		dstNode, moveDirection, err := lkr.parseMoveMappingLine(string(data))
		if err != nil {
			return err
		}

		if moveDirection == MoveDirDstToSrc {
			return nil
		}

		if dstNode == nil {
			return fmt.Errorf("Failed to find dest node for commit map: %v", string(data))
		}

		srcNode, err := lkr.NodeByInode(inode)
		if err != nil {
			return err
		}

		if srcNode == nil {
			return fmt.Errorf("Failed to find source node for commit map: %d", inode)
		}

		// Write a bidirectional mapping for this node:
		dstB58 := dstNode.Hash().B58String()
		srcB58 := srcNode.Hash().B58String()

		forwardLine := fmt.Sprintf("%s hash %s", moveDirection, dstB58)
		batch.Put(
			[]byte(forwardLine),
			"moves", status.Hash().B58String(), srcB58,
		)

		batch.Put(
			[]byte(forwardLine),
			"moves", "overlay", srcB58,
		)

		reverseLine := fmt.Sprintf(
			"%s hash %s",
			moveDirection.Invert(),
			srcB58,
		)

		batch.Put(
			[]byte(reverseLine),
			"moves", status.Hash().B58String(), dstB58,
		)

		batch.Put(
			[]byte(reverseLine),
			"moves", "overlay", dstB58,
		)

		// We need to verify that all ghosts will be copied out from staging.
		// In some special cases, not all used ghosts are reachable in
		// MakeCommit.
		//
		// Consider for example this case:
		//
		// $ touch x
		// $ commit
		// $ move x y
		// $ touch x
		// $ commit
		//
		// => In the last commit the ghost from the move (x) is overwritten by
		// a new file and thus will not be reachable anymore. In order to store
		// the full history of the file we need to also keep this ghost.
		for _, checkHash := range []string{dstB58, srcB58} {
			srcKey := []string{"stage", "objects", checkHash}
			dstKey := []string{"objects", checkHash}

			_, err = lkr.kv.Get(dstKey...)
			if err == db.ErrNoSuchKey {
				err = nil

				// This part of the move was not reachable, we need to copy it
				// to the object store additionally.
				if err := db.CopyKey(lkr.kv, srcKey, dstKey); err != nil {
					return err
				}
			}

			if err != nil {
				return err
			}
		}

		// We already have a bidir mapping for this node, no need to mention
		// them further.  (would not hurt, but would be duplicated work)
		delete(exported, srcNode.Inode())
		delete(exported, dstNode.Inode())

		return nil
	}

	if err := lkr.kv.Keys(walker, "stage", "moves"); err != nil {
		batch.Rollback()
		return err
	}

	return batch.Flush()
}

const (
	MoveDirUnknown = iota
	MoveDirSrcToDst
	MoveDirDstToSrc
	MoveDirNone
)

type MoveDir int

func (md MoveDir) String() string {
	switch md {
	case MoveDirSrcToDst:
		return ">"
	case MoveDirDstToSrc:
		return "<"
	case MoveDirNone:
		return "*"
	default:
		return ""
	}
}

func (md MoveDir) Invert() MoveDir {
	switch md {
	case MoveDirSrcToDst:
		return MoveDirDstToSrc
	case MoveDirDstToSrc:
		return MoveDirSrcToDst
	default:
		return md
	}
}

func moveDirFromString(spec string) MoveDir {
	switch spec {
	case ">":
		return MoveDirSrcToDst
	case "<":
		return MoveDirDstToSrc
	case "*":
		return MoveDirNone
	default:
		return MoveDirUnknown
	}
}

func (lkr *Linker) MoveEntryPoint(nd n.Node) (n.Node, MoveDir, error) {
	moveData, err := lkr.kv.Get(
		"stage", "moves", "overlay",
		strconv.FormatUint(nd.Inode(), 10),
	)

	if err != nil && err != db.ErrNoSuchKey {
		return nil, MoveDirUnknown, err
	}

	if moveData == nil {
		moveData, err = lkr.kv.Get("moves", "overlay", nd.Hash().B58String())
		if err != nil && err != db.ErrNoSuchKey {
			return nil, MoveDirUnknown, err
		}

		if moveData == nil {
			return nil, MoveDirNone, nil
		}
	}

	node, moveDir, err := lkr.parseMoveMappingLine(string(moveData))
	if err != nil {
		return nil, MoveDirUnknown, err
	}

	if node == nil {
		// No move mapping found for this node.
		// Note that this not an error.
		return nil, MoveDirNone, nil
	}

	return node, moveDir, err
}

// MoveMapping will lookup if the node pointed to by `nd` was part of a moving
// operation and if so, to what node it was moved and if it was the source or
// the dest node.
func (lkr *Linker) MoveMapping(cmt *n.Commit, nd n.Node) (n.Node, MoveDir, error) {
	// Stage and committed space use a different format to store move mappings.
	// This is because in staging nodes can still be modified, so the "dest"
	// part of the mapping is a moving target. Therefore we store the destination
	// not as hash or path (which also might be moved), but as inode reference.
	// Inodes always resolve to the latest version of a node.
	// When committing, the mappings will be "fixed" by converting the inode to
	// a hash value, to make sure we link to a specific version.
	status, err := lkr.Status()
	if err != nil {
		return nil, MoveDirUnknown, err
	}

	// Only look into staging if we are actually in the STATUS commit.
	// The lookups in the stage level are on an inode base. This would
	// cause jumping around in the history for older commits.
	if cmt == nil || cmt.Hash().Equal(status.Hash()) {
		inodeKey := strconv.FormatUint(nd.Inode(), 10)
		moveData, err := lkr.kv.Get("stage", "moves", inodeKey)
		if err != nil && err != db.ErrNoSuchKey {
			return nil, MoveDirUnknown, err
		}

		if err != db.ErrNoSuchKey {
			node, moveDir, err := lkr.parseMoveMappingLine(string(moveData))
			if err != nil {
				return nil, MoveDirUnknown, err
			}

			if node != nil {
				return node, moveDir, err
			}
		}
	}

	if cmt == nil {
		return nil, MoveDirNone, nil
	}

	moveData, err := lkr.kv.Get("moves", cmt.Hash().B58String(), nd.Hash().B58String())
	if err != nil && err != db.ErrNoSuchKey {
		return nil, MoveDirUnknown, err
	}

	if moveData == nil {
		return nil, MoveDirNone, nil
	}

	node, moveDir, err := lkr.parseMoveMappingLine(string(moveData))
	if err != nil {
		return nil, MoveDirUnknown, err
	}

	if node == nil {
		// No move mapping found for this node.
		// Note that this not an error.
		return nil, MoveDirNone, nil
	}

	return node, moveDir, err
}

// ExpandAbbrev tries to find an object reference that stats with `abbrev`.
// If so, it will return the respective hash for it.
// If none is found, it is considered as an error.
// If more than one was found ie.ErrAmbigious is returned.
func (lkr *Linker) ExpandAbbrev(abbrev string) (h.Hash, error) {
	prefixes := [][]string{
		{"stage", "objects"},
		{"objects"},
	}

	for _, prefix := range prefixes {
		matches, err := lkr.kv.Glob(append(prefix, abbrev))
		if err != nil {
			return nil, err
		}

		if len(matches) > 1 {
			return nil, ie.ErrAmbigiousRev
		}

		if len(matches) == 0 {
			continue
		}

		match := matches[0]
		return h.FromB58String(match[len(match)-1])
	}

	return nil, fmt.Errorf("No such abbrev: %v", abbrev)
}
