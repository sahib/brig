package store

import "github.com/disorganizer/brig/util/trie"
import "path"

// TODO: move to fs.go
type FS struct {
	kv KV

	// Path lookup trie
	dirRoot *trie.Node

	// B58Hash to node
	dirTable map[string]*trie.Node
}

// TODO: Root() should read HEAD and return the referenced directory in there.
// func(fs *FS) File(hash *Hash) (*Hash, error)
// NOTE: There is no remove function. Directories and files
//       just get unreffed but still exist in history.

func (fs *FS) DirectoryByHash(hash *Hash) (*Directory, error) {
	b58Hash := hash.B58String()
	if node, ok := fs.dirTable[b58Hash]; ok {
		// panic is okay, programmer must have fucked up:
		return node.Data.(*Directory), nil
	}

	// Node was not in the cache, load directly from bolt.
	bkt, err := fs.kv.Bucket([]string{"tree-hashes"})
	if err != nil {
		return nil, err
	}

	data, err := bkt.Get(b58Hash)
	if err != nil {
		return nil, err
	}

	dir := &Directory{fs: fs}
	if err := dir.Unmarshal(data); err != nil {
		return nil, err
	}

	// NOTE: This will indirectly load parent directories, if not done yet!
	//       We might be stuck in an endless loop if we have cycles in our DAG.
	fs.dirTable[b58Hash] = fs.dirRoot.InsertWithData(nodePath(dir), dir)
	return dir, nil
}

func (fs *FS) ResolveDirectory(dirpath string) (*Directory, error) {
	// Check if it's cached already:
	node := fs.dirRoot.Lookup(dirpath)
	if node != nil {
		return node.Data.(*Directory), nil
	}

	// The actual hash of a directory is contained in the "." field:
	lookupPath := path.Join("tree-paths", path.Clean(dirpath), ".")

	// This is a hierarchical lookup:
	hash, err := getPath(fs.kv, lookupPath)
	if err != nil {
		return nil, err
	}

	// Delegate the actual directory loading to Directory()
	return fs.DirectoryByHash(&Hash{hash})
}

func (fs *FS) SaveDirectory(d *Directory) error {
	bkt, err := fs.kv.Bucket([]string{"tree-hashes"})
	if err != nil {
		return err
	}

	data, err := d.Marshal()
	if err != nil {
		return err
	}

	b58Hash := d.hash.B58String()
	if err := bkt.Put(b58Hash, data); err != nil {
		return err
	}

	hashPath := path.Join("tree-paths", path.Clean(nodePath(d)), ".")
	return putPath(fs.kv, hashPath, []byte(b58Hash))
}

func (fs *FS) RemoveUnreffedNodes() error {
	// TODO: This is a NO-OP currently.
	// In future this needs to be called periodically and do the following:
	// - Go through all commits and remember all hashes of all trees.
	// - Go through all hash-buckets and delete all unreffed hashes.
	return nil
}
