package store

import (
	"bytes"
	"path/filepath"
	"strings"
	"sync"

	"github.com/disorganizer/brig/id"
	"github.com/disorganizer/brig/store/wire"
	"github.com/disorganizer/brig/util/ipfsutil"
	multihash "github.com/jbenet/go-multihash"
)

// Store is responsible for adding & retrieving all files from ipfs,
// while managing their metadata in a boltDB.
type Store struct {
	fs *FS
	kv KV

	// Internal path of the repository.
	repoPath string

	// IPFS manager layer (from daemon.Server)
	IPFS *ipfsutil.Node

	// Lock for atomic operations inside the store
	// where several db operations happen in a row.
	// Access to the trie is secured by store.Root.RWMutex.
	mu sync.Mutex
}

// Open loads an existing store at `brigPath/$ID/index.bolt`, if it does not
// exist, it is created.  For full function, Connect() should be called
// afterwards.
func Open(brigPath string, owner id.Peer, IPFS *ipfsutil.Node) (*Store, error) {
	dbDir := filepath.Join(
		brigPath,
		"bolt."+strings.Replace(string(owner.ID()), "/", "-", -1),
	)

	kv, err := NewBoltKV(dbDir)
	if err != nil {
		return nil, err
	}

	fs := NewFilesystem(kv)

	st := &Store{
		repoPath: brigPath,
		IPFS:     IPFS,
		fs:       fs,
		kv:       kv,
	}

	// This version does not attempt any version checking:
	if err := fs.MetadataPut("version", []byte("1")); err != nil {
		return nil, err
	}

	if err := st.setStoreOwner(owner); err != nil {
		return nil, err
	}

	return st, err
}

func (st *Store) setStoreOwner(owner id.Peer) error {
	if err := st.fs.MetadataPut("id", []byte(owner.ID())); err != nil {
		return err
	}

	if err := st.fs.MetadataPut("hash", []byte(owner.Hash())); err != nil {
		return err
	}

	return nil
}

// Owner returns the owner of the store (name + hash)
func (st *Store) Owner() (*Author, error) {
	bid, err := st.fs.MetadataGet("id")
	if err != nil {
		return nil, err
	}

	bhash, err := st.fs.MetadataGet("hash")
	if err != nil {
		return nil, err
	}

	ident, err := id.Cast(string(bid))
	if err != nil {
		return nil, err
	}

	hash, err := multihash.FromB58String(string(bhash))
	if err != nil {
		return nil, err
	}

	return &Author{ident, &Hash{hash}}, nil
}

// View provides a locked view on a node in the store.
// The node may be modified in the ViewNode method.
func (st *Store) ViewNode(path string, fn func(node Node) error) error {
	st.mu.Lock()
	defer st.mu.Unlock()

	node, err := st.fs.LookupNode(path)
	if err != nil {
		return err
	}

	return fn(node)
}

// ViewFile works like ViewNode, but provides a file to the closure.
func (st *Store) ViewFile(path string, fn func(file *File) error) error {
	st.mu.Lock()
	defer st.mu.Unlock()

	file, err := st.fs.LookupFile(path)
	if err != nil {
		return err
	}

	return fn(file)
}

// ViewDir works like ViewNode, but provides a file to the closure.
func (st *Store) ViewDir(path string, fn func(dir *Directory) error) error {
	st.mu.Lock()
	defer st.mu.Unlock()

	dir, err := st.fs.LookupDirectory(path)
	if err != nil {
		return err
	}

	return fn(dir)
}

// Close syncs all data. It is an error to use the store afterwards.
func (st *Store) Close() error {
	st.mu.Lock()
	defer st.mu.Unlock()

	return st.kv.Close()
}

// Export marshals all relevant inside the database, so a cloned repository may
// import them again. Currently it justs packs the whole boltdb into a byte
// message.
func (st *Store) Export() (*wire.Store, error) {
	b := &bytes.Buffer{}
	if err := st.kv.Export(b); err != nil {
		return nil, err
	}

	return &wire.Store{
		Boltdb: b.Bytes(),
	}, nil
}

// Import unmarshals the data written by export.
// If succesful, a new store with the data is created.
func (st *Store) Import(pstore *wire.Store) error {
	st.mu.Lock()
	defer st.mu.Unlock()

	return st.kv.Import(bytes.NewReader(pstore.Boltdb))
}
