package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/disorganizer/brig/id"
	"github.com/disorganizer/brig/store/wire"
	"github.com/disorganizer/brig/util/ipfsutil"
)

// Store is responsible for adding & retrieving all files from ipfs,
// while managing their metadata in a boltDB.
type Store struct {
	db *bolt.DB

	// Root models the directory tree, aka Trie.
	// The root node is the repository root.
	Root *File

	// Internal path of the repository.
	repoPath string

	// The jabber id this store is associated to.
	ID id.ID

	// IPFS manager layer (from daemon.Server)
	IPFS *ipfsutil.Node

	// Lock for atomic operations inside the store
	// where several db operations happen in a row.
	// Access to the trie is secured by store.Root.RWMutex.
	mu sync.Mutex
}

func prefixSlash(s string) string {
	if !strings.HasPrefix(s, "/") {
		return "/" + s
	}

	return s
}

func (st *Store) loadIndex() error {
	return st.viewWithBucket("index", func(tx *bolt.Tx, bkt *bolt.Bucket) error {
		// Check if the root directory already exists:
		if bkt.Get([]byte("/")) == nil {
			rootDir, err := newDirUnlocked(st, "/")
			if err != nil {
				return err
			}

			st.Root = rootDir
		}

		return bkt.ForEach(func(k []byte, v []byte) error {
			file := emptyFile(st)
			if err := file.Unmarshal(st, v); err != nil {
				log.Warningf("store-unmarshal: fail on `%s`: %v", k, err)
				return err
			}

			return nil
		})
	})
}

func (st *Store) createInitialCommit() error {
	needsInit := false

	err := st.viewWithBucket("refs", func(tx *bolt.Tx, bkt *bolt.Bucket) error {
		needsInit = (bkt.Get([]byte("HEAD")) == nil)
		return nil
	})

	if err != nil {
		return err
	}

	if !needsInit {
		return nil
	}

	// No commit yet, create initial commit.
	rootCommit := NewEmptyCommit(st, st.ID)
	rootCommit.Message = "Initial commit"
	rootCommit.Hash = st.Root.Hash().Clone()
	rootCommit.TreeHash = st.Root.Hash().Clone()

	data, err := rootCommit.MarshalProto()
	if err != nil {
		return err
	}

	// Insert initial commit to `commits` bucket:
	err = st.updateWithBucket("commits", func(tx *bolt.Tx, bkt *bolt.Bucket) error {
		return bkt.Put(rootCommit.Hash.Bytes(), data)
	})

	return st.updateHEAD(rootCommit)
}

func (st *Store) updateHEAD(cmt *Commit) error {
	return st.updateWithBucket("refs", func(tx *bolt.Tx, bkt *bolt.Bucket) error {
		data, err := cmt.MarshalProto()
		if err != nil {
			return err
		}

		return bkt.Put([]byte("HEAD"), data)
	})
}

// Open loads an existing store at `brigPath/$ID/index.bolt`, if it does not
// exist, it is created.  For full function, Connect() should be called
// afterwards.
func Open(brigPath string, ID id.ID, IPFS *ipfsutil.Node) (*Store, error) {
	options := &bolt.Options{Timeout: 1 * time.Second}
	dbDir := filepath.Join(
		brigPath,
		"bolt."+strings.Replace(string(ID), "/", "-", -1),
	)

	if err := os.MkdirAll(dbDir, 0777); err != nil {
		return nil, err
	}

	db, err := bolt.Open(filepath.Join(dbDir, "index.bolt"), 0600, options)

	if err != nil {
		return nil, err
	}

	st := &Store{
		db:       db,
		ID:       ID,
		repoPath: brigPath,
		IPFS:     IPFS,
	}

	// Create initial buckets:
	err = db.Update(func(tx *bolt.Tx) error {
		buckets := []string{
			"index",       // File-Path to file protobuf.
			"stage",       // Staged files (path to current checkpoint)
			"commits",     // Commit-Hash to commit protobuf.
			"checkpoints", // File-Path to History (== mod_time to checkpoint)
			"refs",        // Special names for certain commits (e.g. HEAD)
		}

		for _, name := range buckets {
			if _, berr := tx.CreateBucketIfNotExists([]byte(name)); berr != nil {
				return fmt.Errorf("create bucket: %s", berr)
			}
		}
		return nil
	})

	if err != nil {
		log.Warningf("store-create failed: %v", err)
	}

	// Load all paths from the database into the trie.
	// This also creates a root node if none exists yet.
	if err := st.loadIndex(); err != nil {
		return nil, err
	}

	if err := st.createInitialCommit(); err != nil {
		return nil, err
	}

	return st, err
}

// Close syncs all data. It is an error to use the store afterwards.
func (st *Store) Close() error {
	st.mu.Lock()
	defer st.mu.Unlock()

	if err := st.db.Sync(); err != nil {
		log.Warningf("store-sync: %v", err)
		return err
	}

	if err := st.db.Close(); err != nil {
		log.Warningf("store-close: %v", err)
		return err
	}

	return nil
}

// Export marshals all relevant inside the database, so a cloned
// repository may import them again.
// The exported data includes:
//  - All files (including their history and keys)
//  - All commits.
//  - Pinning information.
//
// TODO: Describe stream format.
//
// w is not closed after Export.
func (st *Store) Export() (*wire.Store, error) {
	// TODO: Export commits (not implemented)
	// TODO: Export pinning information?
	protoStore := &wire.Store{}

	var err error

	st.mu.Lock()
	defer st.mu.Unlock()

	st.Root.Walk(true, func(child *File) bool {
		// Note: Walk() already calls Lock()
		protoFile, errPbf := child.ToProto()
		if err != nil {
			err = errPbf
			return false
		}

		if child.kind != FileTypeRegular {
			// Directories are implicit:
			return true
		}

		history, errHist := st.History(child.node.Path())
		if errHist != nil {
			err = errHist
			return false
		}

		protoHist, errPbh := history.ToProto()
		if err != nil {
			err = errPbh
			return false
		}

		protoPack := &wire.Pack{
			File:    protoFile,
			History: protoHist,
		}

		protoStore.Packs = append(protoStore.Packs, protoPack)
		return true
	})

	// TODO: Get Head() and traverse down to root.
	//       -> History is linear?
	//       -> Merge commits have a special Merge attr?

	if err != nil {
		return nil, err
	}

	return protoStore, nil
}

// Import unmarshals the data written by export.
// If succesful, a new store with the data is created.
func (st *Store) Import(protoStore *wire.Store) error {
	st.mu.Lock()
	defer st.mu.Unlock()

	for _, pack := range protoStore.Packs {
		file := emptyFile(st)
		if err := file.Import(pack.GetFile()); err != nil {
			return err
		}

		// TODO: Restore history.
		log.Debugf("Imported: %v", file.Path())
		file.Sync()
		file.updateParents()
	}

	return nil
}
