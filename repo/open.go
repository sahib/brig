package repo

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/repo/config"
	"github.com/disorganizer/brig/repo/global"
	"github.com/disorganizer/brig/store"
	"github.com/disorganizer/brig/util/ipfsutil"
	"github.com/tsuibin/goxmpp2/xmpp"
)

var (
	ErrBadPassword = errors.New("Bad password.")
)

// Filenames that will be encrypted on close:
var lockPaths = []string{
	"index.bolt",
	"master.key",
	// TODO: What about those? Encrypt them?
	// "otr.buddies",
	// "otr.key",
}

func lookupJid(configPath string) (string, error) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return "", fmt.Errorf("Could not load config: %v", err)
	}

	jid, err := cfg.String("repository.jid")
	if err != nil {
		return "", fmt.Errorf("No jid in config: %v", err)
	}

	return jid, nil
}

// Open unencrypts all sensible data in the repository.
func Open(pwd, folder string) (*Repository, error) {
	absFolderPath, err := filepath.Abs(folder)
	brigPath := filepath.Join(absFolderPath, ".brig")

	// Figure out the JID from the config:
	jid, err := lookupJid(filepath.Join(brigPath, "config"))
	if err != nil {
		return nil, err
	}

	// Unlock all files:
	var absNames []string
	for _, name := range lockPaths {
		absName := filepath.Join(brigPath, name)
		if _, err := os.Stat(absName); err == nil {
			// File exists, this might happen on a crash or killed daemon.
			log.Warningf("File is already unlocked: %s", absName)
			continue
		}

		absNames = append(absNames, absName)
	}

	if err := UnlockFiles(jid, pwd, absNames); err != nil {
		return nil, err
	}

	return loadRepository(pwd, absFolderPath)
}

// Close encrypts sensible files in the repository.
// The password is taken from Repository.Password.
func (r *Repository) Close() error {
	var absNames []string
	for _, name := range lockPaths {
		absName := filepath.Join(r.InternalFolder, name)
		if _, err := os.Stat(absName); os.IsNotExist(err) {
			// File does not exist. Might be already locked.
			log.Warningf("File is already locked: %s", absName)
			continue
		}

		log.Infof("Locking file `%v`...", absName)
		absNames = append(absNames, absName)
	}

	if err := LockFiles(r.Jid, r.Password, absNames); err != nil {
		return err
	}

	return nil
}

// CheckPassword tries to decrypt a file in the repository.
// If that does not work, an error is returned.
func CheckPassword(folder, pwd string) error {
	absFolderPath, err := filepath.Abs(folder)
	brigPath := filepath.Join(absFolderPath, ".brig")

	jid, err := lookupJid(filepath.Join(brigPath, "config"))
	if err != nil {
		return err
	}

	entry, err := parseShadowFile(brigPath, jid)
	if err != nil {
		return err
	}

	attempt := hashPassword(entry.salt, pwd)
	if !bytes.Equal(attempt, entry.hash) {
		return ErrBadPassword
	}

	return nil
}

// loadRepository load a brig repository from a given folder.
func loadRepository(pwd, folder string) (*Repository, error) {
	absFolderPath, err := filepath.Abs(folder)
	if err != nil {
		return nil, err
	}

	brigPath := filepath.Join(absFolderPath, ".brig")
	cfg, err := config.LoadConfig(filepath.Join(brigPath, "config"))
	if err != nil {
		return nil, err
	}

	configValues := map[string]string{
		"repository.jid":  "",
		"repository.mid":  "",
		"repository.uuid": "",
	}

	for key := range configValues {
		configValues[key], err = cfg.String(key)
		if err != nil {
			return nil, err
		}
	}

	// Init the global repo (similar to .gitconfig)
	globalRepo, err := global.New()
	if err != nil {
		return nil, err
	}

	err = globalRepo.AddRepo(global.RepoListEntry{
		UniqueID:   configValues["repository.uuid"],
		RepoPath:   folder,
		DaemonPort: 6666,
		IpfsPort:   4001,
	})

	if err != nil {
		return nil, err
	}

	ipfsLayer := ipfsutil.New(filepath.Join(brigPath, "ipfs"))

	ownStore, err := store.Open(brigPath, ipfsLayer)
	if err != nil {
		return nil, err
	}

	jid := configValues["repository.jid"]
	allStores := make(map[xmpp.JID]*store.Store)
	allStores[xmpp.JID(jid)] = ownStore

	repo := Repository{
		Jid:            jid,
		Mid:            configValues["repository.mid"],
		Folder:         absFolderPath,
		InternalFolder: brigPath,
		UniqueID:       configValues["repository.uuid"],
		Config:         cfg,
		globalRepo:     globalRepo,
		OwnStore:       ownStore,
		Password:       pwd,
		IPFS:           ipfsLayer,
	}

	return &repo, nil
}
