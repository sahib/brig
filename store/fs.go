package store

// Layout of the bolt database:
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

import (
	"encoding/binary"
	"fmt"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/disorganizer/brig/store/wire"
	"github.com/disorganizer/brig/util/trie"
	"github.com/gogo/protobuf/proto"
	"github.com/jbenet/go-multihash"
)

/////////////// ERRORS ///////////////

// FS implements the logic of brig's data model.
// It uses an underlying key/value database to
// storea a Merkle-DAG with versioned metadata,
// similar to what git does internally.
type FS struct {
	kv KV

	root *Directory

	// Path lookup trie
	ptrie *trie.Node

	// B58Hash to node
	index map[string]Node

	// UID to node
	uidIndex map[uint64]Node
}

func marshalNode(nd Node) ([]byte, error) {
	pnd, err := nd.ToProto()
	if err != nil {
		return nil, err
	}

	return proto.Marshal(pnd)
}

func unmarshalNode(fs *FS, data []byte) (Node, error) {
	pnd := &wire.Node{}
	if err := proto.Unmarshal(data, pnd); err != nil {
		return nil, err
	}

	var node Node

	switch typ := pnd.Type; typ {
	case wire.NodeType_FILE:
		node = &File{fs: fs}
	case wire.NodeType_DIRECTORY:
		node = &Directory{fs: fs}
	case wire.NodeType_COMMIT:
		node = &Commit{fs: fs}
	default:
		return nil, ErrBadNodeType(typ)
	}

	if err := node.FromProto(pnd); err != nil {
		return nil, err
	}

	return node, nil
}

// NewFilesystem returns a new FS, ready to use. It assumes the key value store
// is working and does no check on this.
func NewFilesystem(kv KV) *FS {
	fs := &FS{kv: kv}
	fs.MemIndexClear()
	return fs
}

//  MemIndexAdd adds `nd` to the in memory index.
func (fs *FS) MemIndexAdd(nd Node) {
	fs.index[nd.Hash().B58String()] = nd
	fs.uidIndex[nd.ID()] = nd
	fs.ptrie.InsertWithData(nd.Path(), nd)
}

// MemIndexSwap updates an entry of the in memory index, by deleting
// the old entry referenced by oldHash (may be nil). This is necessary
// to ensure that old hashes do not resolve to the new, updated instance.
// If the old instance is needed, it will be loaded as new instance.
// You should not need to call this function, except when implementing own Nodes.
func (fs *FS) MemIndexSwap(nd Node, oldHash *Hash) {
	if oldHash != nil {
		delete(fs.index, oldHash.B58String())
	}

	fs.index[nd.Hash().B58String()] = nd
	fs.uidIndex[nd.ID()] = nd
	fs.ptrie.InsertWithData(nd.Path(), nd)
}

// MemIndexPurge removes `nd` from the memory index.
func (fs *FS) MemIndexPurge(nd Node) {
	delete(fs.uidIndex, nd.ID())
	delete(fs.index, nd.Hash().B58String())
	fs.ptrie.Lookup(nd.Path()).Remove()
}

// MemIndexClear resets the memory index to zero.
func (fs *FS) MemIndexClear() {
	fs.ptrie = trie.NewNode()
	fs.index = make(map[string]Node)
	fs.uidIndex = make(map[uint64]Node)
}

//////////////////////////
// COMMON NODE HANDLING //
//////////////////////////

// NextID() returns a unique identifier, used to identify a single node. You
// should not need to call this function, except when implementing own nodes.
func (fs *FS) NextID() (uint64, error) {
	bkt, err := fs.kv.Bucket([]string{"stats"})
	if err != nil {
		return 0, err
	}

	nodeCount, err := bkt.Get("node-count")
	if err != nil {
		return 0, err
	}

	// nodeCount might be nil on startup:
	cnt := uint64(1)
	if nodeCount != nil {
		cnt = binary.BigEndian.Uint64(nodeCount) + 1
	}

	cntBuf := make([]byte, 8)
	binary.BigEndian.PutUint64(cntBuf, cnt)

	if err := bkt.Put("node-count", cntBuf); err != nil {
		return 0, nil
	}

	return cnt, nil
}

// loadNode loads an individual object by its hash from the object store. It
// will return nil if the hash is not existant.
func (fs *FS) loadNode(hash *Hash) (Node, error) {
	var data []byte
	var err error

	b58hash := hash.B58String()

	loadableBuckets := [][]string{
		[]string{"stage", "objects"},
		[]string{"objects"},
	}
	for _, bucketPath := range loadableBuckets {
		var bkt Bucket
		bkt, err = fs.kv.Bucket(bucketPath)
		if err != nil {
			return nil, err
		}

		data, err = bkt.Get(b58hash)
		if err != nil {
			return nil, err
		}

		if data != nil {
			break
		}
	}

	// Damn, no hash found:
	if data == nil {
		return nil, nil
	}

	return unmarshalNode(fs, data)
}

// NodeByHash returns the node identified by hash.
// If no such hash could be found, nil is returned.
func (fs *FS) NodeByHash(hash *Hash) (Node, error) {
	// Check if we have this this node in the cache already:
	b58Hash := hash.B58String()
	if cachedNode, ok := fs.index[b58Hash]; ok {
		return cachedNode, nil
	}

	// Node was not in the cache, load directly from bolt.
	nd, err := fs.loadNode(hash)
	if err != nil {
		return nil, err
	}

	if nd == nil {
		// log.Warningf("Could not load hash `%s`", hash.B58String())
		return nil, nil
	}

	fs.MemIndexSwap(nd, nil)
	return nd, nil
}

func appendDot(path string) string {
	// path.Join() calls path.Clean() which in turn
	// removes the '.' at the end when trying to join that.
	if strings.HasSuffix(path, "/") {
		return path + "."
	}

	return path + "/."
}

// Same as path.Join, but does not remove the last '.' needed for directories.
func joinButLeaveLastDot(elems ...string) string {
	if len(elems) == 0 {
		return ""
	}

	if strings.HasSuffix(elems[len(elems)-1], "/.") {
		return appendDot(path.Join(elems...))
	}

	return path.Join(elems...)
}

// ResolveNode resolves a path to a hash and resolves the corresponding node by
// calling NodeByHash(). If no node could be resolved, nil is returned.
func (fs *FS) ResolveNode(nodePath string) (Node, error) {
	// Check if it's cached already:
	trieNode := fs.ptrie.Lookup(nodePath)
	if trieNode != nil && trieNode.Data != nil {
		return trieNode.Data.(Node), nil
	}

	var hash []byte
	var err error

	prefixes := []string{"stage/tree/", "tree/"}
	for _, prefix := range prefixes {
		// getPath() does a hierarchical lookup:
		joinedPath := joinButLeaveLastDot(prefix, nodePath)
		hash, err = getPath(fs.kv, joinedPath)

		if err != nil {
			return nil, err
		}

		if hash != nil {
			break
		}
	}

	// Return both nil if nothing found:
	if hash == nil {
		return nil, nil
	}

	// Delegate the actual directory loading to Directory()
	return fs.NodeByHash(&Hash{hash})
}

// StageNode inserts a modified node to the staging area, making sure the
// modification is persistent and part of the staging commit. All parent
// directories of the node in question will be staged automatically. If there
// was no modification it will be a (quite expensive) NOOP.
func (fs *FS) StageNode(nd Node) error {
	if err := fs.stageNodeRecursive(nd); err != nil {
		return err
	}

	// Update the staging commit's root hash:
	status, err := fs.Status()
	if err != nil {
		return err
	}

	root, err := fs.Root()
	if err != nil {
		return err
	}

	if err := status.SetRoot(root.Hash()); err != nil {
		return err
	}

	return fs.saveStatus(status)
}

// NodeByUID resolves a node by it's unique ID.
// It will return nil if no corresponding node was found.
func (fs *FS) NodeByUID(uid uint64) (Node, error) {
	uidKey := strconv.FormatUint(uid, 16)
	hash, err := getPath(fs.kv, "uid/"+uidKey)
	if err != nil {
		return nil, err
	}

	mh, err := multihash.Cast(hash)
	if err != nil {
		return nil, err
	}

	return fs.NodeByHash(&Hash{mh})
}

func (fs *FS) stageNodeRecursive(nd Node) error {
	if nd.GetType() == NodeTypeCommit {
		return fmt.Errorf("BUG: Commits cannot be staged; Use MakeCommit()")
	}

	object, err := nd.ToProto()
	if err != nil {
		return err
	}

	data, err := proto.Marshal(object)
	if err != nil {
		return err
	}

	b58Hash := nd.Hash().B58String()
	if err := putPath(fs.kv, "stage/objects/"+b58Hash, data); err != nil {
		return err
	}

	uidKey := strconv.FormatUint(nd.ID(), 16)
	if err := putPath(fs.kv, "uid/"+uidKey, nd.Hash().Bytes()); err != nil {
		return err
	}

	// The key is the path of the
	nodePath := nd.Path()

	hashPath := path.Join("stage/tree", nodePath)
	switch nd.GetType() {
	case NodeTypeDirectory:
		hashPath = appendDot(hashPath)
	}

	if err := putPath(fs.kv, hashPath, nd.Hash().Bytes()); err != nil {
		return err
	}

	// Remember/Update this node in the cache if it's not yet there:
	fs.MemIndexAdd(nd)

	// We need to save parent directories too, in case the hash changed:
	// TODO: This creates many pointless roots in the stage. Maybe remember
	// some in a kill-list & do a bit of garbage collect from time to time.
	par, err := nd.Parent()
	if err != nil {
		return err
	}

	if par != nil {
		if err := fs.StageNode(par); err != nil {
			return err
		}
	}

	return nil
}

/////////////////////////
// CHECKPOINT HANDLING //
/////////////////////////

// LastCheckpoint returns the last known checkpoint for the UID referenced by `IDLink`.
func (fs *FS) LastCheckpoint(IDLink uint64) (*Checkpoint, error) {
	key := strconv.FormatUint(IDLink, 16)
	bkt, err := fs.kv.Bucket([]string{"checkpoints", key})
	if err != nil {
		return nil, err
	}

	data, err := bkt.Last()
	if err != nil {
		return nil, err
	}

	if data == nil {
		return nil, nil
	}

	ckp := &Checkpoint{}
	if err := ckp.Unmarshal(data); err != nil {
		return nil, err
	}

	return ckp, nil
}

// CheckpointAt returns the checkpoint of the file identified by the node id
// `IDLink` at `index`. If it could not be retrieved, an Error is returned.
func (fs *FS) CheckpointAt(IDLink, index uint64) (*Checkpoint, error) {
	key := strconv.FormatUint(IDLink, 16)

	bkt, err := fs.kv.Bucket([]string{"checkpoints", key})
	if err != nil {
		return nil, err
	}

	subKey := strconv.FormatUint(index, 16)
	data, err := bkt.Get(subKey)
	if err != nil {
		return nil, err
	}

	ckp := &Checkpoint{}
	if err := ckp.Unmarshal(data); err != nil {
		return nil, err
	}

	return ckp, nil
}

// History returns all checkpoints for the file referenced by the UID `IDLink`.
func (fs *FS) History(IDLink uint64) (History, error) {
	key := strconv.FormatUint(IDLink, 16)
	history := History{}

	bkt, err := fs.kv.Bucket([]string{"checkpoints", key})
	if err != nil {
		return nil, err
	}

	err = bkt.Foreach(func(key string, value []byte) error {
		ckp := &Checkpoint{}
		if err := ckp.Unmarshal(value); err != nil {
			return err
		}

		history = append(history, ckp)
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort the history by the checkpoint indices.
	// This is likely not needed, just to be sure...
	sort.Sort(&history)
	return history, nil
}

// HistoryByPath is a convinience function. It resolves `nodePath` to a Node,
// and calls History() with the node's UID.
func (fs *FS) HistoryByPath(nodePath string) (History, error) {
	nd, err := fs.ResolveNode(nodePath)
	if err != nil {
		return nil, err
	}

	if nd == nil {
		return nil, NoSuchFile(nodePath)
	}

	return fs.History(nd.ID())
}

// StageCheckpoint remembers a checkpoint the staging commit and saves it to
// the file's history.
func (fs *FS) StageCheckpoint(ckp *Checkpoint) error {
	pckp, err := ckp.ToProto()
	if err != nil {
		return err
	}

	data, err := proto.Marshal(pckp)
	if err != nil {
		return err
	}

	key := strconv.FormatUint(pckp.IdLink, 16)
	bkt, err := fs.kv.Bucket([]string{"checkpoints", key})
	if err != nil {
		return err
	}

	idx := strconv.FormatUint(pckp.Index, 16)
	if err := bkt.Put(idx, data); err != nil {
		return err
	}

	status, err := fs.Status()
	if err != nil {
		return err
	}

	status.AddCheckpointLink(ckp.MakeLink())
	return fs.saveStatus(status)
}

/////////////////////
// COMMIT HANDLING //
/////////////////////

// MakeCommit creates a new full commit in the version history.
// The current staging commit is finalized with `author` and `message`
// and gets saved. A new, identical staging commit is created pointing
// to the root of the now new HEAD.
func (fs *FS) MakeCommit(author *Author, message string) error {
	head, err := fs.Head()
	if err != nil && !IsErrNoSuchRef(err) {
		return err
	}

	status, err := fs.Status()
	if err != nil {
		return err
	}

	// Only compare with previous if we have a HEAD yet.
	if head != nil {
		if status.Root().Equal(head.Root()) {
			return ErrNoChange
		}
	}

	rootDir, err := fs.Root()
	if err != nil {
		return err
	}

	objBkt, err := fs.kv.Bucket([]string{"objects"})
	if err != nil {
		return err
	}

	treeBkt, err := fs.kv.Bucket([]string{"tree"})
	if err != nil {
		return err
	}

	err = Walk(rootDir, true, func(child Node) error {
		data, err := marshalNode(child)
		if err != nil {
			return err
		}

		b58Hash := child.Hash().B58String()
		if err := objBkt.Put(b58Hash, data); err != nil {
			return err
		}

		path := child.Path()
		if err := treeBkt.Put(path, []byte(b58Hash)); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	// NOTE: `head` may be nil, if it couldn't be resolved.
	if err := status.Finalize(author, message, head); err != nil {
		return err
	}

	statusData, err := marshalNode(status)
	if err != nil {
		return err
	}

	statusB58Hash := status.Hash().B58String()
	if err := objBkt.Put(statusB58Hash, statusData); err != nil {
		return err
	}

	if err := fs.SaveRef("HEAD", status); err != nil {
		return err
	}

	toClear := [][]string{
		[]string{"stage", "objects"},
		[]string{"stage", "tree"},
	}

	for _, key := range toClear {
		clearBkt, err := fs.kv.Bucket(key)
		if err != nil {
			return err
		}

		if err := clearBkt.Clear(); err != nil {
			return err
		}
	}

	newStatus, err := newEmptyCommit(fs)
	if err != nil {
		return err
	}

	newStatus.SetParent(status)
	newStatus.SetRoot(status.Root())
	return fs.saveStatus(newStatus)
}

///////////////////////
// METADATA HANDLING //
///////////////////////

// MetadataPut remembers a value persisntenly identified by `key`.
// It can be used as single-level key value store for user purposes.
func (fs *FS) MetadataPut(key string, value []byte) error {
	bkt, err := fs.kv.Bucket([]string{"metadata"})
	if err != nil {
		return err
	}

	return bkt.Put(key, []byte(value))
}

// MetadataGet retriesves a previosuly put key value pair.
// It will return nil if no such value could be retrieved.
func (fs *FS) MetadataGet(key string) ([]byte, error) {
	bkt, err := fs.kv.Bucket([]string{"metadata"})
	if err != nil {
		return nil, err
	}

	return bkt.Get(key)
}

////////////////////////
// REFERENCE HANDLING //
////////////////////////

// ResolveRef resolves the hash associated with `refname`. If the ref could not
// be resolved, ErrNoSuchRef is returned. Typically, Node will be a Commit. But
// there are no technical restrictions on which node typ to use.
func (fs *FS) ResolveRef(refname string) (Node, error) {
	refname = strings.ToLower(refname)
	bkt, err := fs.kv.Bucket([]string{"refs"})
	if err != nil {
		return nil, err
	}

	hash, err := bkt.Get(refname)
	if err != nil {
		return nil, err
	}

	if len(hash) == 0 {
		return nil, ErrNoSuchRef(refname)
	}

	mh, err := multihash.Cast(hash)
	if err != nil {
		return nil, err
	}

	return fs.NodeByHash(&Hash{mh})
}

// SaveRef stores a reference to `nd` persistently. The caller is responsbiel
// to ensure that the node is already in the blockstore, otherwise it won't be
// resolvable.
func (fs *FS) SaveRef(refname string, nd Node) error {
	refname = strings.ToLower(refname)
	bkt, err := fs.kv.Bucket([]string{"refs"})
	if err != nil {
		return err
	}

	return bkt.Put(refname, nd.Hash().Bytes())
}

// Head is just a shortcut for ResolveRef("HEAD").
func (fs *FS) Head() (*Commit, error) {
	nd, err := fs.ResolveRef("HEAD")
	if err != nil {
		return nil, err
	}

	cmt, ok := nd.(*Commit)
	if !ok {
		return nil, fmt.Errorf("oh-oh, HEAD is not a Commit... %v", nd == nil)
	}

	return cmt, nil
}

// SetMemRoot sets the current root, but does not store it yet. It's supposed
// to be called after in-memory modifications. Only implementors of new Nodes
// might need to call this function.
func (fs *FS) SetMemRoot(root *Directory) {
	fs.root = root
}

// Root returns the current root directory of CURR.
// It is never nil when err is nil.
func (fs *FS) Root() (*Directory, error) {
	if fs.root != nil {
		return fs.root, nil
	}

	status, err := fs.Status()
	if err != nil {
		return nil, err
	}

	return fs.DirectoryByHash(status.Root())
}

// Status returns the current staging commit.
// It is never nil, unless err is nil.
func (fs *FS) Status() (*Commit, error) {
	cmt, err := fs.loadStatus()
	if err != nil {
		return nil, err
	}

	if cmt != nil {
		return cmt, nil
	}

	cmt, err = newEmptyCommit(fs)
	if err != nil {
		return nil, err
	}

	// Setup a new commit and set root from last HEAD or new one.
	head, err := fs.Head()
	if err != nil && !IsErrNoSuchRef(err) {
		return nil, err
	}

	var rootHash *Hash

	if IsErrNoSuchRef(err) {
		// There probably wasn't a HEAD yet.
		// No root directory then. Create a shiny new one and stage it.
		newRoot, err := newEmptyDirectory(fs, nil, "/")
		if err != nil {
			return nil, err
		}

		// Can't call StageNode(), since that would call Status(),
		// causing and endless loop of grief and doom.
		if err := fs.stageNodeRecursive(newRoot); err != nil {
			return nil, err
		}

		rootHash = newRoot.Hash()
	} else {
		if err := cmt.SetParent(head); err != nil {
			return nil, err
		}

		rootHash = head.Root()
	}

	if err := cmt.SetRoot(rootHash); err != nil {
		return nil, err
	}

	if err := fs.saveStatus(cmt); err != nil {
		return nil, err
	}

	return cmt, nil
}

func (fs *FS) loadStatus() (*Commit, error) {
	bkt, err := fs.kv.Bucket([]string{"stage"})
	if err != nil {
		return nil, err
	}

	data, err := bkt.Get("STATUS")
	if err != nil {
		return nil, err
	}

	if data == nil {
		return nil, nil
	}

	cmt, err := newEmptyCommit(fs)
	if err != nil {
		return nil, err
	}

	// It's there already. Just unmarshal it.
	pnode := &wire.Node{}
	if err := proto.Unmarshal(data, pnode); err != nil {
		return nil, err
	}

	if err := cmt.FromProto(pnode); err != nil {
		return nil, err
	}

	return cmt, nil
}

// saveStatus copies cmt to stage/STATUS.
func (fs *FS) saveStatus(cmt *Commit) error {
	if err := cmt.Finalize(StageAuthor(), "", nil); err != nil {
		return err
	}

	bkt, err := fs.kv.Bucket([]string{"stage"})
	if err != nil {
		return err
	}

	data, err := marshalNode(cmt)
	if err != nil {
		return err
	}

	if err := bkt.Put("STATUS", data); err != nil {
		return err
	}

	if err := fs.SaveRef("CURR", cmt); err != nil {
		return err
	}

	return nil
}

// ValidateRefs checks for any dead references in the merkle dag.
// Dead should currently happen on bugs. Everybody knows that
// there are no bugs here, that's why this functions is currently a NOOP.
//
// In future this needs to be called periodically and do the following:
// - Go through all commits and remember all hashes of all trees.
// - Go through all hash-buckets and delete all unreffed hashes.
func (fs *FS) ValidateRefs() error {
	return nil
}

/////////////////////////////////
// CONVINIENT ACCESS FUNCTIONS //
/////////////////////////////////

// LookupNode takes the root node and tries to resolve the path from there.
// Deleted paths are recognized in contrast to ResolveNode.
// If a path does not exist NoSuchFile is returned.
func (fs *FS) LookupNode(repoPath string) (Node, error) {
	root, err := fs.Root()
	if err != nil {
		return nil, err
	}

	return root.Lookup(repoPath)
}

func (fs *FS) LookupSettableNode(repoPath string) (SettableNode, error) {
	node, err := fs.LookupNode(repoPath)
	if err != nil {
		return nil, err
	}

	if node == nil {
		return nil, nil
	}

	snode, ok := node.(SettableNode)
	if !ok {
		return nil, ErrBadNode
	}

	return snode, nil
}

// DirectoryByHash calls NodeByHash and attempts to convert
// it to a Directory as convinience.
func (fs *FS) DirectoryByHash(hash *Hash) (*Directory, error) {
	nd, err := fs.NodeByHash(hash)
	if err != nil {
		return nil, err
	}

	if nd == nil {
		return nil, nil
	}

	dir, ok := nd.(*Directory)
	if !ok {
		return nil, ErrBadNode
	}

	return dir, nil
}

// ResolveDirectory calls ResolveNode and converts the result to a Directory.
func (fs *FS) ResolveDirectory(dirpath string) (*Directory, error) {
	nd, err := fs.ResolveNode(appendDot(path.Clean(dirpath)))
	if err != nil {
		return nil, err
	}

	if nd == nil {
		return nil, nil
	}

	dir, ok := nd.(*Directory)
	if !ok {
		return nil, ErrBadNode
	}

	return dir, nil
}

// LookupDirectory calls LookupNode and converts the result to a Directory.
func (fs *FS) LookupDirectory(repoPath string) (*Directory, error) {
	nd, err := fs.LookupNode(repoPath)
	if err != nil {
		return nil, err
	}

	if nd == nil {
		return nil, nil
	}

	dir, ok := nd.(*Directory)
	if !ok {
		return nil, ErrBadNode
	}

	return dir, nil
}

// FileByHash calls NodeByHash and converts the result to a File.
func (fs *FS) FileByHash(hash *Hash) (*File, error) {
	nd, err := fs.NodeByHash(hash)
	if err != nil {
		return nil, err
	}

	file, ok := nd.(*File)
	if !ok {
		return nil, ErrBadNode
	}

	return file, nil
}

// ResolveFile calls ResolveNode and converts the result to a file.
func (fs *FS) ResolveFile(filepath string) (*File, error) {
	nd, err := fs.ResolveNode(filepath)
	if err != nil {
		return nil, err
	}

	if nd == nil {
		return nil, nil
	}

	file, ok := nd.(*File)
	if !ok {
		return nil, ErrBadNode
	}

	return file, nil
}

// LookupFile calls LookupNode and converts the result to a file.
func (fs *FS) LookupFile(repoPath string) (*File, error) {
	nd, err := fs.LookupNode(repoPath)
	if err != nil {
		return nil, err
	}

	if nd == nil {
		return nil, nil
	}

	file, ok := nd.(*File)
	if !ok {
		return nil, ErrBadNode
	}

	return file, nil
}

// CommitByHash lookups a commit by it's hash.
// If the commit could not be found, nil is returned.
func (fs *FS) CommitByHash(hash *Hash) (*Commit, error) {
	nd, err := fs.NodeByHash(hash)
	if err != nil {
		return nil, err
	}

	if nd == nil {
		return nil, nil
	}

	cmt, ok := nd.(*Commit)
	if !ok {
		return nil, ErrBadNode
	}

	return cmt, nil
}

// Unstage resets the state of a node back to the last known commited state.
func (fs *FS) Unstage(nd Node) error {
	head, err := fs.Head()
	if err != nil {
		return err
	}

	return fs.CheckoutFile(head, nd)
}

// HaveStagedChanges returns true if there were changes in the staging area.
// If an error occurs, the first return value is undefined.
func (fs *FS) HaveStagedChanges() (bool, error) {
	head, err := fs.Head()
	if err != nil && !IsErrNoSuchRef(err) {
		return false, err
	}

	if !IsErrNoSuchRef(err) {
		// There is no HEAD yet.
		// Assume we have changes.
		return true, nil
	}

	status, err := fs.Status()
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
func (fs *FS) CheckoutCommit(cmt *Commit, force bool) error {
	// Check if the staging area is empty if no force given:
	if !force {
		haveStaged, err := fs.HaveStagedChanges()
		if err != nil {
			return err
		}

		if haveStaged {
			return ErrStageNotEmpty
		}
	}

	status, err := fs.Status()
	if err != nil {
		return err
	}

	if err := status.SetRoot(cmt.Root()); err != nil {
		return err
	}

	// Invalidate the cache, causing NodeByHash and ResolveNode to load the
	// file from the boltdb again:
	fs.MemIndexClear()
	return fs.saveStatus(status)
}

// CheckoutFile resets a certain file to the state it had in cmt. If the file
// did not exist back then, it will be deleted. `nd` is usually retrieved by
// calling ResolveNode() and sorts.
// TODO: write test for this.
func (fs *FS) CheckoutFile(cmt *Commit, nd Node) error {
	root, err := fs.DirectoryByHash(cmt.Root())
	if err != nil {
		return err
	}

	if root == nil {
		// TODO: Is this valid?
		return fmt.Errorf("No root to reset to")
	}

	// TODO: Better resolve by UID here?
	//       Would need to find the commit with the last modification though.
	oldNode, err := root.Lookup(nd.Path())
	if err != nil && !IsNoSuchFileError(err) {
		return err
	}

	// Invalidate the respective index entry, so the instance gets reloaded:
	err = Walk(nd, true, func(child Node) error {
		fs.MemIndexPurge(child)
		return nil
	})

	if err != nil {
		return err
	}

	par, err := nodeParentDir(nd)
	if err != nil {
		return err
	}

	// nd might be root itself, so par may be nil.
	if par != nil {
		if err := par.RemoveChild(nd); err != nil {
			return err
		}
	}

	if err := fs.StageNode(par); err != nil {
		return err
	}

	if oldNode == nil {
		// oldNode did not exist back then.
		// Just keep the file deleted and create a remove checkpoint.
		return makeCheckpoint(
			fs, cmt.author.ID(),
			nd.ID(),
			nd.Hash(), nil,
			nd.Path(), "",
		)
	}

	if err := par.Add(oldNode); err != nil {
		return err
	}

	// TODO: create modify checkpoint.
	// Stage the old node
	err = makeCheckpoint(
		fs, cmt.author.ID(),
		nd.ID(),
		oldNode.Hash(), nd.Hash(),
		oldNode.Path(), nd.Path(),
	)

	if err != nil {
		return err
	}

	return fs.StageNode(oldNode)
}
