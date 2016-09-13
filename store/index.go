package store

import (
	"path/filepath"
	"strings"
	"sync"

	"github.com/disorganizer/brig/id"
	"github.com/disorganizer/brig/store/wire"
	"github.com/disorganizer/brig/util/ipfsutil"
)

// Store is responsible for adding & retrieving all files from ipfs,
// while managing their metadata in a boltDB.
type Store struct {
	// TODO: Load this.
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

	// TODO: Insert metadata here
	if err := st.storeOwner(owner); err != nil {
		return nil, err
	}

	return st, err
}

func (st *Store) storeOwner(owner id.Peer) error {
	if err := st.fs.MetadataPut("id", []byte(owner.ID())); err != nil {
		return err
	}

	if err := st.fs.MetadataPut("hash", []byte(owner.Hash())); err != nil {
		return err
	}

	return nil
}

// Owner returns the owner of the store (name + hash)
func (st *Store) Owner() (id.Peer, error) {
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

	return id.NewPeer(ident, string(bhash)), nil
}

// Close syncs all data. It is an error to use the store afterwards.
func (st *Store) Close() error {
	st.mu.Lock()
	defer st.mu.Unlock()

	return st.kv.Close()
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
	/*
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

		// TODO: Export refs, only commits are exported currently.
		// TODO: Get Head() and traverse down to root.
		//       -> History is linear?
		//       -> Merge commits have a special Merge attr?
		if err != nil {
			return nil, err
		}

		cmts := &wire.Commits{}

		err = st.viewWithBucket("commits", func(tx *bolt.Tx, bkt *bolt.Bucket) error {
			return bkt.ForEach(func(k, v []byte) error {
				cmt := &wire.Commit{}
				if err := proto.Unmarshal(v, cmt); err != nil {
					return err
				}

				cmts.Commits = append(cmts.Commits, cmt)
				return nil
			})
		})

		if err != nil {
			return nil, err
		}

		protoStore.Commits = cmts
		return protoStore, nil
	*/
	return nil, nil
}

// Import unmarshals the data written by export.
// If succesful, a new store with the data is created.
func (st *Store) Import(protoStore *wire.Store) error {
	st.mu.Lock()
	defer st.mu.Unlock()

	// TODO: re-design
	/*
		for _, pack := range protoStore.Packs {
			file := emptyFile(st)
			if err := file.Import(pack.GetFile()); err != nil {
				return err
			}

			log.Debugf("-- Imported: %v", file.Path())
			file.Sync()
			file.updateParents()

			// TODO: Only make one transaction after the for{}.
			for _, protoCheckpoint := range pack.GetHistory().GetHist() {
				err := st.updateWithBucket("refs", func(tx *bolt.Tx, bkt *bolt.Bucket) error {
					data, err := proto.Marshal(protoCheckpoint)
					if err != nil {
						return err
					}

					return bkt.Put(protoCheckpoint.GetModTime(), data)
				})

				if err != nil {
					return err
				}
			}
		}

		return st.updateWithBucket("commits", func(tx *bolt.Tx, bkt *bolt.Bucket) error {
			for _, protoCommit := range protoStore.GetCommits().GetCommits() {
				data, err := proto.Marshal(protoCommit)
				if err != nil {
					return err
				}

				if err := bkt.Put(protoCommit.GetHash(), data); err != nil {
					return err
				}
			}

			return nil
		})
	*/
	return nil
}
