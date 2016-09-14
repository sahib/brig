package store

// Layout of the bolt database:
//
// objects/<NODE_HASH>                   => NODE_METADATA
// tree/<FULL_NODE_PATH>                 => NODE_HASH
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
	"errors"
	"fmt"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/disorganizer/brig/id"
	"github.com/disorganizer/brig/store/wire"
	"github.com/disorganizer/brig/util/trie"
	"github.com/gogo/protobuf/proto"
	"github.com/jbenet/go-multihash"
)

// TODO: Clear cache when invalid?

/////////////// ERRORS ///////////////

var (
	ErrBadNode = errors.New("Cannot convert to concrete type. Broken input data?")
)

type ErrBadNodeType int

func (e ErrBadNodeType) Error() string {
	return fmt.Sprintf("Bad node type in db: %d", int(e))
}

type ErrNoHashFound struct {
	b58hash string
	where   string
}

func (e ErrNoHashFound) Error() string {
	return fmt.Sprintf("No such hash in `%s`: '%s'", e.where, e.b58hash)
}

type FS struct {
	kv KV

	// Path lookup trie
	root *trie.Node

	// B58Hash to node
	index map[string]*trie.Node
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

	switch typ := pnd.GetType(); typ {
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

func NewFilesystem(kv KV) *FS {
	return &FS{
		kv:    kv,
		root:  trie.NewNode(),
		index: make(map[string]*trie.Node),
	}
}

//////////////////////////
// COMMON NODE HANDLING //
//////////////////////////

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

// LoadObject loads an individual object by its hash from the object store.
func (fs *FS) loadNode(hash *Hash) (Node, error) {
	var data []byte
	var err error

	b58hash := hash.B58String()

	loadableBuckets := []string{"objects", "stage/objects"}
	for _, bucketPath := range loadableBuckets {
		var bkt Bucket
		bkt, err = fs.kv.Bucket([]string{bucketPath})
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

	fmt.Println("lookupNode", data, b58hash)

	// Damn, no hash found:
	if data == nil {
		return nil, ErrNoHashFound{
			b58hash,
			strings.Join(loadableBuckets, " and "),
		}
	}

	return unmarshalNode(fs, data)
}

func (fs *FS) NodeByHash(hash *Hash) (Node, error) {
	// Check if we have this this node in the cache already:
	b58Hash := hash.B58String()
	if trieNode, ok := fs.index[b58Hash]; ok && trieNode.Data != nil {
		return trieNode.Data.(Node), nil
	}

	// Node was not in the cache, load directly from bolt.
	nd, err := fs.loadNode(hash)
	if err != nil {
		return nil, err
	}

	// NOTE: This will indirectly load parent directories (by calling
	//       Parent(), if not done yet!  We might be stuck in an endless loop if we
	//       have cycles in our DAG.
	fs.index[b58Hash] = fs.root.InsertWithData(NodePath(nd), nd)
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

func joinButLeaveLastDot(elems ...string) string {
	if len(elems) == 0 {
		return ""
	}

	if strings.HasSuffix(elems[len(elems)-1], "/.") {
		return appendDot(path.Join(elems...))
	}

	return path.Join(elems...)
}

func (fs *FS) ResolveNode(nodePath string) (Node, error) {
	// Check if it's cached already:
	trieNode := fs.root.Lookup(nodePath)
	if trieNode != nil && trieNode.Data != nil {
		return trieNode.Data.(Node), nil
	}

	var hash []byte
	var err error

	prefixes := []string{"tree/", "stage/tree/"}
	for _, prefix := range prefixes {
		// getPath() does a hierarchical lookup:
		joinedPath := joinButLeaveLastDot(prefix, nodePath)
		fmt.Println("looking up path:", joinedPath, nodePath)
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

func (fs *FS) StageNode(nd Node) error {
	if nd.GetType() == NodeTypeCommit {
		return fmt.Errorf("Commits cannot be staged")
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

	// The key is the path of the
	nodePath := NodePath(nd)
	hashPath := path.Join("stage/tree", nodePath)
	switch nd.GetType() {
	case NodeTypeDirectory:
		hashPath = appendDot(hashPath)
	}

	fmt.Println("Stage:", hashPath)

	if err := putPath(fs.kv, hashPath, nd.Hash().Bytes()); err != nil {
		return err
	}

	// Remember/Update this node in the cache if it's not yet there:
	fs.index[b58Hash] = fs.root.InsertWithData(nodePath, nd)

	// We need to save parent directories too, in case the hash changed:
	// TODO: Is that a good idea? Many stages will cause many pointless
	//       root dirs with different checksums.
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
		return nil, fmt.Errorf("No last checkpoint")
	}

	ckp := newEmptyCheckpoint()
	if err := ckp.Unmarshal(data); err != nil {
		return nil, err
	}

	return ckp, nil
}

func (fs *FS) History(IDLink uint64) (History, error) {
	key := strconv.FormatUint(IDLink, 16)
	history := History{}

	bkt, err := fs.kv.Bucket([]string{"checkpoints", key})
	if err != nil {
		return nil, err
	}

	err = bkt.Foreach(func(key string, value []byte) error {
		ckp := newEmptyCheckpoint()
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

func (fs *FS) HistoryByPath(nodePath string) (History, error) {
	nd, err := fs.ResolveNode(nodePath)
	if err != nil {
		return nil, err
	}

	return fs.History(nd.ID())
}

func (fs *FS) StageCheckpoint(ckp *Checkpoint) error {
	pckp, err := ckp.ToProto()
	if err != nil {
		return err
	}

	data, err := proto.Marshal(pckp)
	if err != nil {
		return err
	}

	bkt, err := fs.kv.Bucket([]string{"checkpoints"})
	if err != nil {
		return err
	}

	key := strconv.FormatUint(pckp.GetIdLink(), 16)
	if err := bkt.Put(key, data); err != nil {
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

func (fs *FS) MakeCommit(author id.Peer, message string) error {
	head, err := fs.Head()
	if err != nil {
		return err
	}

	status, err := fs.Status()
	if err != nil {
		return err
	}

	if status.Root().Equal(head.Root()) {
		return ErrNoChange
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

		path := NodePath(child)
		if err := treeBkt.Put(path, []byte(b58Hash)); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

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

	stageBkt, err := fs.kv.Bucket([]string{"stage"})
	if err != nil {
		return err
	}

	return stageBkt.Put("STATUS", nil)
}

///////////////////////
// METADATA HANDLING //
///////////////////////

func (fs *FS) MetadataPut(key string, value []byte) error {
	bkt, err := fs.kv.Bucket([]string{"metadata"})
	if err != nil {
		return err
	}

	return bkt.Put(key, []byte(value))
}

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

	mh, err := multihash.Cast(hash)
	if err != nil {
		return nil, err
	}

	return fs.NodeByHash(&Hash{mh})
}

func (fs *FS) SaveRef(refname string, nd Node) error {
	refname = strings.ToLower(refname)
	bkt, err := fs.kv.Bucket([]string{"refs"})
	if err != nil {
		return err
	}

	return bkt.Put(refname, nd.Hash().Bytes())
}

// Basically a shortcut for fs.ResolveRef("HEAD").(*Commit)
func (fs *FS) Head() (*Commit, error) {
	nd, err := fs.ResolveRef("HEAD")
	if err != nil {
		return nil, err
	}

	cmt, ok := nd.(*Commit)
	if !ok {
		return nil, fmt.Errorf("oh-oh, HEAD is not a Commit...")
	}

	return cmt, nil
}

func (fs *FS) Root() (*Directory, error) {
	// NOTE: Status() might call Root() and not other way round.
	//       We're on duty of creating a new root if necessary.
	dir, err := fs.ResolveDirectory("/")
	if err != nil {
		return nil, err
	}

	if dir != nil {
		return dir, nil
	}

	// No root directory yet? Create a shiny new one.
	return newEmptyDirectory(fs, nil, "/")
}

// Guarantee it's not nil when err == nil
func (fs *FS) Status() (*Commit, error) {
	bkt, err := fs.kv.Bucket([]string{"stage"})
	if err != nil {
		return nil, err
	}

	data, err := bkt.Get("STATUS")
	if err != nil {
		return nil, err
	}

	cmt, err := newEmptyCommit(fs)
	if err != nil {
		return nil, err
	}

	if data == nil {
		// Setup a new commit and set root.
		head, err := fs.Head()
		if err != nil {
			return nil, err
		}

		if err := cmt.SetParent(head); err != nil {
			return nil, err
		}

		root, err := fs.Root()
		if err != nil {
			return nil, err
		}

		if err := cmt.SetRoot(root.Hash()); err != nil {
			return nil, err
		}
	} else {
		pnode := &wire.Node{}
		if err := proto.Unmarshal(data, pnode); err != nil {
			return nil, err
		}

		if err := cmt.FromProto(pnode); err != nil {
			return nil, err
		}
	}

	return cmt, nil
}

func (fs *FS) saveStatus(cmt *Commit) error {
	bkt, err := fs.kv.Bucket([]string{"stage"})
	if err != nil {
		return err
	}

	data, err := marshalNode(cmt)
	if err != nil {
		return err
	}

	return bkt.Put("STATUS", data)
}

func (fs *FS) RemoveUnreffedNodes() error {
	// TODO: This is a NO-OP currently.
	// In future this needs to be called periodically and do the following:
	// - Go through all commits and remember all hashes of all trees.
	// - Go through all hash-buckets and delete all unreffed hashes.
	// - Also delete checkpoints of removed files.
	return nil
}

/////////////////////////////////
// CONVINIENT ACCESS FUNCTIONS //
/////////////////////////////////

func (fs *FS) LookupNode(repoPath string) (Node, error) {
	root, err := fs.Root()
	if err != nil {
		return nil, err
	}

	return root.Lookup(repoPath)
}

func (fs *FS) DirectoryByHash(hash *Hash) (*Directory, error) {
	nd, err := fs.NodeByHash(hash)
	if err != nil {
		return nil, err
	}

	dir, ok := nd.(*Directory)
	if !ok {
		return nil, ErrBadNode
	}

	return dir, nil
}

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

func (fs *FS) CommitByHash(hash *Hash) (*Commit, error) {
	nd, err := fs.NodeByHash(hash)
	if err != nil {
		return nil, err
	}

	cmt, ok := nd.(*Commit)
	if !ok {
		return nil, ErrBadNode
	}

	return cmt, nil
}
