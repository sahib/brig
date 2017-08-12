package cafs

// Layout of the key/value store:
//
// objects/<NODE_HASH>                   => NODE_METADATA
// tree/<FULL_NODE_PATH>                 => NODE_HASH
// uid/<NODE_HASH>                       => NODE_HASH
// checkpoints/<HEX_NODE_ID>/<IDX>       => CHECKPOINT_DATA
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

import (
	"encoding/binary"
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/disorganizer/brig/cafs/db"
	n "github.com/disorganizer/brig/cafs/nodes"
	h "github.com/disorganizer/brig/util/hashlib"
	"github.com/disorganizer/brig/util/trie"
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
	uidIndex map[uint64]n.Node
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
	lkr.uidIndex[nd.Inode()] = nd
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

	lkr.index[nd.Hash().B58String()] = nd
	lkr.uidIndex[nd.Inode()] = nd
	lkr.ptrie.InsertWithData(nd.Path(), nd)
}

// MemIndexPurge removes `nd` from the memory index.
func (lkr *Linker) MemIndexPurge(nd n.Node) {
	delete(lkr.uidIndex, nd.Inode())
	delete(lkr.index, nd.Hash().B58String())
	lkr.ptrie.Lookup(nd.Path()).Remove()
}

// MemIndexClear resets the memory index to zero.
func (lkr *Linker) MemIndexClear() {
	lkr.ptrie = trie.NewNode()
	lkr.index = make(map[string]n.Node)
	lkr.uidIndex = make(map[uint64]n.Node)
}

//////////////////////////
// COMMON NODE HANDLING //
//////////////////////////

// NextInode() returns a unique identifier, used to identify a single node. You
// should not need to call this function, except when implementing own nodes.
func (lkr *Linker) NextInode() uint64 {
	// TODO: Transactions
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
	if err := lkr.kv.Put(cntBuf, "stats", "node-count"); err != nil {
		return 0
	}

	return cnt
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

	// Node was not in the cache, load directly from bolt.
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
func (lkr *Linker) StageNode(nd n.Node) error {
	if err := lkr.stageNodeRecursive(nd); err != nil {
		return err
	}
	fmt.Println("--- status ---")

	// Update the staging commit's root hash:
	status, err := lkr.Status()
	if err != nil {
		return fmt.Errorf("Failed to retrieve status: %v", err)
	}

	fmt.Println("--- root ---")
	root, err := lkr.Root()
	if err != nil {
		return err
	}

	fmt.Println("--- save status ---", status.Inode())
	status.SetRoot(root.Hash())
	return lkr.saveStatus(status)
}

// NodeByInode resolves a node by it's unique ID.
// It will return nil if no corresponding node was found.
func (lkr *Linker) NodeByInode(uid uint64) (n.Node, error) {
	hash, err := lkr.kv.Get("inode", strconv.FormatUint(uid, 16))
	if err != nil && err != db.ErrNoSuchKey {
		return nil, err
	}

	return lkr.NodeByHash(h.Hash(hash))
}

func (lkr *Linker) stageNodeRecursive(nd n.Node) error {
	if nd.Type() == n.NodeTypeCommit {
		return fmt.Errorf("BUG: Commits cannot be staged; Use MakeCommit()")
	}

	data, err := n.MarshalNode(nd)
	if err != nil {
		return err
	}

	// TODO: Transactions?
	b58Hash := nd.Hash().B58String()
	if err := lkr.kv.Put(data, "stage", "objects", b58Hash); err != nil {
		return err
	}

	uidKey := strconv.FormatUint(nd.Inode(), 16)
	if err := lkr.kv.Put([]byte(nd.Hash().B58String()), "inode", uidKey); err != nil {
		return err
	}

	hashPath := []string{"stage", "tree", nd.Path()}
	if nd.Type() == n.NodeTypeDirectory {
		hashPath = append(hashPath, ".")
	}

	if err := lkr.kv.Put([]byte(b58Hash), hashPath...); err != nil {
		return err
	}

	// Remember/Update this node in the cache if it's not yet there:
	lkr.MemIndexAdd(nd)

	// We need to save parent directories too, in case the hash changed:
	// TODO: This creates many pointless roots in the stage. Maybe remember
	// some in a kill-list & do a bit of garbage collect from time to time.
	par, err := nd.Parent(lkr)
	if err != nil {
		return err
	}

	if par != nil {
		if err := lkr.stageNodeRecursive(par); err != nil {
			return err
		}
	}

	return nil
}

/////////////////////
// COMMIT HANDLING //
/////////////////////

// MakeCommit creates a new full commit in the version history.
// The current staging commit is finalized with `author` and `message`
// and gets saved. A new, identical staging commit is created pointing
// to the root of the now new HEAD.
func (lkr *Linker) MakeCommit(author *n.Person, message string) error {
	head, err := lkr.Head()
	if err != nil && !IsErrNoSuchRef(err) {
		return err
	}

	status, err := lkr.Status()
	if err != nil {
		return err
	}

	// Only compare with previous if we have a HEAD yet.
	if head != nil {
		if status.Root().Equal(head.Root()) {
			return ErrNoChange
		}
	}

	rootDir, err := lkr.Root()
	if err != nil {
		return err
	}

	// Go over all files/directories and save them in tree & objects.
	err = n.Walk(lkr, rootDir, true, func(child n.Node) error {
		data, err := n.MarshalNode(child)
		if err != nil {
			return err
		}

		b58Hash := child.Hash().B58String()
		if err := lkr.kv.Put(data, "objects", b58Hash); err != nil {
			return err
		}

		childPath := child.Path()
		if child.Type() == n.NodeTypeDirectory {
			childPath = appendDot(childPath)
		}

		return lkr.kv.Put([]byte(b58Hash), "tree", childPath)
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
	if err := lkr.kv.Put(statusData, "objects", statusB58Hash); err != nil {
		return err
	}

	if err := lkr.SaveRef("HEAD", status); err != nil {
		return err
	}

	// Clear the staging area.
	toClear := [][]string{
		[]string{"stage", "objects"},
		[]string{"stage", "tree"},
	}

	for _, key := range toClear {
		if err := lkr.kv.Clear(key...); err != nil {
			return err
		}
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
	return lkr.kv.Put([]byte(value), "metadata", key)
}

// MetadataGet retriesves a previosuly put key value pair.
// It will return nil if no such value could be retrieved.
func (lkr *Linker) MetadataGet(key string) ([]byte, error) {
	return lkr.kv.Get("metadata", key)
}

////////////////////////
// REFERENCE HANDLING //
////////////////////////

// ResolveRef resolves the hash associated with `refname`. If the ref could not
// be resolved, ErrNoSuchRef is returned. Typically, Node will be a Commit. But
// there are no technical restrictions on which node typ to use.
func (lkr *Linker) ResolveRef(refname string) (n.Node, error) {
	refname = strings.ToLower(refname)
	b58Hash, err := lkr.kv.Get("refs", refname)
	if err != nil && err != db.ErrNoSuchKey {
		return nil, err
	}

	if len(b58Hash) == 0 {
		return nil, ErrNoSuchRef(refname)
	}

	hash, err := h.FromB58String(string(b58Hash))
	if err != nil {
		return nil, err
	}

	return lkr.NodeByHash(h.Hash(hash))
}

// SaveRef stores a reference to `nd` persistently. The caller is responsbiel
// to ensure that the node is already in the blockstore, otherwise it won't be
// resolvable.
func (lkr *Linker) SaveRef(refname string, nd n.Node) error {
	refname = strings.ToLower(refname)
	return lkr.kv.Put([]byte(nd.Hash().B58String()), "refs", refname)
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
func (lkr *Linker) Status() (*n.Commit, error) {
	cmt, err := lkr.loadStatus()
	if err != nil {
		return nil, err
	}

	if cmt != nil {
		return cmt, nil
	}

	// Shoot, no commit exists yet.
	// We need to create an initial one.
	cmt, err = n.NewEmptyCommit(lkr.NextInode())
	if err != nil {
		return nil, err
	}

	// Setup a new commit and set root from last HEAD or new one.
	head, err := lkr.Head()
	if err != nil && !IsErrNoSuchRef(err) {
		return nil, err
	}

	var rootHash h.Hash

	if IsErrNoSuchRef(err) {
		// There probably wasn't a HEAD yet.
		if root, err := lkr.ResolveDirectory("/"); err == nil {
			rootHash = root.Hash()
		} else {
			// No root directory then. Create a shiny new one and stage it.
			inode := lkr.NextInode()
			newRoot, err := n.NewEmptyDirectory(lkr, nil, "/", inode)
			if err != nil {
				return nil, err
			}

			// Can't call StageNode(), since that would call Status(),
			// causing and endless loop of grief and doom.
			if err := lkr.stageNodeRecursive(newRoot); err != nil {
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

	fmt.Println("--- save inner ---")
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
func (lkr *Linker) saveStatus(cmt *n.Commit) error {
	head, err := lkr.Head()
	if err != nil && !IsErrNoSuchRef(err) {
		return err
	}

	if head != nil {
		if err := cmt.SetParent(lkr, head); err != nil {
			return err
		}
	}

	if err := cmt.BoxCommit(n.AuthorOfStage(), ""); err != nil {
		return err
	}

	data, err := n.MarshalNode(cmt)
	if err != nil {
		return err
	}

	// TODO: Use transactions here.
	if err := lkr.kv.Put(data, "stage", "STATUS"); err != nil {
		return err
	}

	b58Hash := cmt.Hash().B58String()
	inode := strconv.FormatUint(cmt.Inode(), 10)
	if err := lkr.kv.Put([]byte(b58Hash), "inode", inode); err != nil {
		return err
	}

	if err := lkr.SaveRef("CURR", cmt); err != nil {
		return err
	}

	return nil
}

func (lkr *Linker) ValidateRefs() error {
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

func (lkr *Linker) LookupSettableNode(repoPath string) (n.SettableNode, error) {
	node, err := lkr.LookupNode(repoPath)
	if err != nil {
		return nil, err
	}

	if node == nil {
		return nil, nil
	}

	snode, ok := node.(n.SettableNode)
	if !ok {
		return nil, n.ErrBadNode
	}

	return snode, nil
}

func (lkr *Linker) ResolveSettableNode(repoPath string) (n.SettableNode, error) {
	node, err := lkr.ResolveNode(repoPath)
	if err != nil {
		return nil, err
	}

	if node == nil {
		return nil, nil
	}

	snode, ok := node.(n.SettableNode)
	if !ok {
		return nil, n.ErrBadNode
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
		return nil, n.ErrBadNode
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
		return nil, n.ErrBadNode
	}

	return dir, nil
}

// LookupDirectory calls LookupNode and converts the result to a Directory.
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
		return nil, n.ErrBadNode
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
		return nil, n.ErrBadNode
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
		return nil, n.ErrBadNode
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
		return nil, n.ErrBadNode
	}

	return file, nil
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
		return nil, n.ErrBadNode
	}

	return cmt, nil
}

// Unstage resets the state of a node back to the last known commited state.
func (lkr *Linker) Unstage(nd n.Node) error {
	head, err := lkr.Head()
	if err != nil {
		return err
	}

	return lkr.CheckoutFile(head, nd)
}

// HaveStagedChanges returns true if there were changes in the staging area.
// If an error occurs, the first return value is undefined.
func (lkr *Linker) HaveStagedChanges() (bool, error) {
	head, err := lkr.Head()
	if err != nil && !IsErrNoSuchRef(err) {
		return false, err
	}

	if !IsErrNoSuchRef(err) {
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
// referenced by cmt. If force is false, it will check if there any stages in
// the staging area and return ErrStageNotEmpty if there are any. If force is
// true, all changes will be overwritten.
// TODO: write test for this.
func (lkr *Linker) CheckoutCommit(cmt *n.Commit, force bool) error {
	// Check if the staging area is empty if no force given:
	if !force {
		haveStaged, err := lkr.HaveStagedChanges()
		if err != nil {
			return err
		}

		if haveStaged {
			return ErrStageNotEmpty
		}
	}

	status, err := lkr.Status()
	if err != nil {
		return err
	}

	status.SetRoot(cmt.Root())

	// Invalidate the cache, causing NodeByHash and ResolveNode to load the
	// file from the boltdb again:
	lkr.MemIndexClear()
	return lkr.saveStatus(status)
}

// CheckoutFile resets a certain file to the state it had in cmt. If the file
// did not exist back then, it will be deleted. `nd` is usually retrieved by
// calling ResolveNode() and sorts.
func (lkr *Linker) CheckoutFile(cmt *n.Commit, nd n.Node) error {
	root, err := lkr.DirectoryByHash(cmt.Root())
	if err != nil {
		return err
	}

	if root == nil {
		// TODO: Is this valid?
		return fmt.Errorf("No root to reset to")
	}

	// TODO: Better resolve by UID here?
	//       Would need to find the commit with the last modification though.
	oldNode, err := root.Lookup(lkr, nd.Path())
	if err != nil && !n.IsNoSuchFileError(err) {
		return err
	}

	// Invalidate the respective index entry, so the instance gets reloaded:
	err = n.Walk(lkr, nd, true, func(child n.Node) error {
		lkr.MemIndexPurge(child)
		return nil
	})

	if err != nil {
		return err
	}

	par, err := n.ParentDirectory(lkr, nd)
	if err != nil {
		return err
	}

	// nd might be root itself, so par may be nil.
	if par != nil {
		if err := par.RemoveChild(lkr, nd); err != nil {
			return err
		}
	}

	if err := lkr.StageNode(par); err != nil {
		return err
	}

	if err := par.Add(lkr, oldNode); err != nil {
		return err
	}

	return lkr.StageNode(oldNode)
}
