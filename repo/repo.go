package repo

import (
	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/id"
	"github.com/disorganizer/brig/repo/global"
	"github.com/disorganizer/brig/store"
	"github.com/disorganizer/brig/util/ipfsutil"
	yamlConfig "github.com/olebedev/config"
)

// TODO: Make ipfs keypair export/import-able to enable rescue-mode.

// Repository represents a handle to one physical brig repository.
// It groups the APIs to all useful files in it.
type Repository struct {
	// Repository is identified by a brig account
	ID id.ID

	// Folder of repository
	Folder         string
	InternalFolder string

	// TODO: still required?
	// User supplied password:
	Password string

	Config *yamlConfig.Config

	// Remotes stores the metadata of all communication partners
	Remotes RemoteStore

	allStores map[id.ID]*store.Store

	// OwnStore is the store.Store used to save our own files in.
	// This is guaranteed to be non-nil.
	OwnStore *store.Store

	// IPFS management layer.
	IPFS *ipfsutil.Node

	// globalRepo is a common file of all brig instances
	// on a machine. It's purpose to make the brig repos
	// easily findable and to avoid collisions on network ports.
	globalRepo *global.Repository
}

func (rp *Repository) Peer() id.Peer {
	hash, err := rp.IPFS.Identity()
	if err != nil {
		log.Warningf("Cannot retrieve ipfs id: %v", err)
		hash = ""
	}
	return id.NewPeer(rp.ID, hash)

}

func (rp *Repository) AddStore(ID id.ID, st *store.Store) {
	rp.allStores[ID] = st
}

func (rp *Repository) RmStore(ID id.ID) {
	delete(rp.allStores, ID)
}

func (rp *Repository) Store(ID id.ID) *store.Store {
	return rp.allStores[ID]
}
