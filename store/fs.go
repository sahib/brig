package store

import (
	"errors"
	"fmt"
	"path"

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
	return fmt.Sprintf("No such hash in `%s`: %s", e.where, e.b58hash)
}

type FS struct {
	kv KV

	// Path lookup trie
	root *trie.Node

	// B58Hash to node
	index map[string]*trie.Node
}

// LoadObject loads an individual object by its hash from the object store.
func (fs *FS) loadNode(hash *Hash) (Node, error) {
	bkt, err := fs.kv.Bucket([]string{"objects"})
	if err != nil {
		return nil, err
	}

	data, err := bkt.Get(hash.B58String())
	if err != nil {
		return nil, err
	}

	node := &wire.Node{}
	if err := proto.Unmarshal(data, node); err != nil {
		return nil, err
	}

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
	if trieNode != nil {
		return trieNode.Data.(Node), nil
	}

	// The actual hash of a directory is contained in the "." field:
	lookupPath := path.Join("tree", path.Clean(nodePath))

	// getPath() does a hierarchical lookup:
	hash, err := getPath(fs.kv, lookupPath)
	if err != nil {
		return nil, err
	}

	// Delegate the actual directory loading to Directory()
	return fs.NodeByHash(&Hash{hash})
}

func (fs *FS) SaveNode(nd Node) error {
	bkt, err := fs.kv.Bucket([]string{"staged"})
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
	hashPath := path.Join("objects", path.Clean(nodePath(nd)))
	switch nd.GetType() {
	case NodeTypeDirectory:
		hashPath += "/"
	}

	if err := putPath(fs.kv, hashPath, []byte(b58Hash)); err != nil {
		return err
	}

	// We need to save parent directories too, in case the hash changed:
	par, err := nd.Parent()
	if err != nil {
		return err
	}

	if nd != nil {
		if err := fs.SaveNode(par); err != nil {
			return err
		}
	}

	return nil
}

func (fs *FS) MakeCommit() (*Commit, error) {
	// TODO
	return nil, nil
}

func ResolveRef(refname string) (Node, error) {
	// TODO
	return nil, nil
}

func (fs *FS) SaveRef(refname string, nd Node) error {
	// TODO
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
	nd, err := fs.ResolveNode(path.Join(dirpath, "/"))
	if err != nil {
		return nil, err
	}

	dir, ok := nd.(*Directory)
	if !ok {
		return nil, ErrBadNode
	}

	return dir, nil
}
