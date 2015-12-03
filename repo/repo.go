package repo

import (
	"github.com/disorganizer/brig/repo/global"
	"github.com/disorganizer/brig/store"
	yamlConfig "github.com/olebedev/config"
)

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

	Config     *yamlConfig.Config
	Store      *store.Store
	globalRepo *global.GlobalRepository
}
