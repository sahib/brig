package repo

import (
	"github.com/disorganizer/brig/repo/global"
	"github.com/disorganizer/brig/store"
	"github.com/disorganizer/brig/util/ipfsutil"
	yamlConfig "github.com/olebedev/config"
	"github.com/tsuibin/goxmpp2/xmpp"
)

// Repository represents a handle to one physical brig repository.
// It groups the APIs to all useful files in it.
type Repository struct {
	// Repository is identified by a XMPP Account: name@domain.tld/ressource
	Jid string

	// Minilock ID
	Mid string

	// Folder of repository
	Folder         string
	InternalFolder string

	// UUID which represents a unique repository
	UniqueID string

	// User supplied password:
	Password string

	Config *yamlConfig.Config

	allStores map[xmpp.JID]*store.Store

	// OwnStore is the store.Store used to save our own files in.
	// This is guaranteed to be non-nil.
	OwnStore *store.Store

	// IPFS management layer.
	IPFS *ipfsutil.Node

	// TODO: document...
	globalRepo *global.Repository
}

func (rp *Repository) AddStore(jid xmpp.JID, st *store.Store) {
	rp.allStores[jid] = st
}

func (rp *Repository) RmStore(jid xmpp.JID) {
	delete(rp.allStores, jid)
}

func (rp *Repository) Store(jid xmpp.JID) *store.Store {
	return rp.allStores[jid]
}
