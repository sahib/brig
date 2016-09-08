package store

// Layout of the bolt database:
//
// objects/<NODE_HASH>            => NODE_METADATA
// tree/<FULL_NODE_PATH>          => NODE_HASH
// refs/<REFNAME>                 => NODE_HASH
// checkpoints/<NODE_HASH>/<IDX>  => CHECKPOINT_DATA
// stage/objects/<NODE_HASH>      => NODE_METADATA
// stage/tree/<FULL_NODE_PATH>    => NODE_HASH
//
// NODE is either a Commit, a Directory or a File.
// FULL_NODE_PATH may contain slashes and in case of directories,
// it will contain a trailing slash.

import (
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/disorganizer/brig/store/wire"
	"github.com/disorganizer/brig/util/trie"
	"github.com/gogo/protobuf/proto"
)

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

type ErrNoPathFound struct {
	path  string
	where string
}

func (e ErrNoPathFound) Error() string {
	return fmt.Sprintf("No such path in `%s`: '%s'", e.where, e.path)
}

type FS struct {
	kv KV

	// Path lookup trie
	root *trie.Node

	// B58Hash to node
	index map[string]*trie.Node
}

func NewFilesystem(kv KV) *FS {
	return &FS{
		kv:    kv,
		root:  trie.NewNode(),
		index: make(map[string]*trie.Node),
	}
}

var (
	loadableBuckets = []string{"objects", "stage/objects"}
)

// LoadObject loads an individual object by its hash from the object store.
func (fs *FS) loadNode(hash *Hash) (Node, error) {
	var data []byte
	var err error

	b58hash := hash.B58String()

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
			strings.Join(loadableBuckets, " + "),
		}
	}

	node := &wire.Node{}
	if err := proto.Unmarshal(data, node); err != nil {
		return nil, err
	}
	fmt.Println("lookupNode unmarshal done")

	typ := node.GetType()
	switch typ {
	case wire.NodeType_FILE:
		// TODO
	case wire.NodeType_DIRECTORY:
		dir := &Directory{fs: fs}
		if err := dir.FromProto(node); err != nil {
			return nil, err
		}

		return dir, nil
	case wire.NodeType_COMMIT:
		// TODO
	}

	return nil, ErrBadNodeType(typ)
}

// TODO: Root() should read HEAD and return the referenced directory in there.

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
	fs.index[b58Hash] = fs.root.InsertWithData(nodePath(nd), nd)
	return nd, nil
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
		hash, err = getPath(fs.kv, prefix+nodePath)
		fmt.Println("looking up path:", prefix+nodePath)

		if err != nil {
			return nil, err
		}

		if hash != nil {
			break
		}
	}

	if hash == nil {
		return nil, ErrNoPathFound{
			nodePath,
			strings.Join(prefixes, " and "),
		}
	}

	x := &Hash{hash}
	fmt.Println("Resolved to hash:", string(hash), string(x.Bytes()))

	// Delegate the actual directory loading to Directory()
	return fs.NodeByHash(&Hash{hash})
}

func (fs *FS) StageNode(nd Node) error {
	bkt, err := fs.kv.Bucket([]string{"stage/objects"})
	if err != nil {
		return err
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
	if err := bkt.Put(b58Hash, data); err != nil {
		return err
	}

	// Clean() will also remove trailing slashes:
	hashPath := path.Join("stage/tree", path.Clean(nodePath(nd)))
	switch nd.GetType() {
	case NodeTypeDirectory:
		hashPath += "/."
	}

	fmt.Println(hashPath, nodePath(nd), "insert", b58Hash)

	if err := putPath(fs.kv, hashPath, nd.Hash().Bytes()); err != nil {
		return err
	}

	// TODO: Insert to fs.index and fs.root.

	// We need to save parent directories too, in case the hash changed:
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

func (fs *FS) MakeCommit() (*Commit, error) {
	// TODO: Copy everything in stage/objects and stage/tree to
	//       objects/ and tree/. Also make a commit of all current checkpoints.
	return nil, nil
}

func ResolveRef(refname string) (Node, error) {
	// TODO: Resolve refs/<refname> to objects/$HASH
	return nil, nil
}

func (fs *FS) SaveRef(refname string, nd Node) error {
	// TODO: Place refname and nd.Hash() in refs/<refname>
	return nil
}

func (fs *FS) RemoveUnreffedNodes() error {
	// TODO: This is a NO-OP currently.
	// In future this needs to be called periodically and do the following:
	// - Go through all commits and remember all hashes of all trees.
	// - Go through all hash-buckets and delete all unreffed hashes.
	return nil
}

////////////////////////////

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
	nd, err := fs.ResolveNode(path.Clean(dirpath) + "/.")
	if err != nil {
		return nil, err
	}

	dir, ok := nd.(*Directory)
	if !ok {
		return nil, ErrBadNode
	}

	return dir, nil
}
