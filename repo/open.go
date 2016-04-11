package repo

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/id"
	"github.com/disorganizer/brig/repo/config"
	"github.com/disorganizer/brig/store"
	"github.com/disorganizer/brig/util/ipfsutil"
)

var (
	ErrBadPassword = errors.New("Bad password.")
)

// Filenames that will be encrypted on close
// and decrypted upon opening the repository.
func absLockPaths(brigPath string) []string {
	lockPaths := []string{
		filepath.Join(brigPath, "master.key"),
		filepath.Join(brigPath, "remotes.yml"),
		filepath.Join(brigPath, "otr.key"),
		filepath.Join(brigPath, "otr.buddies"),
	}

	matches, err := filepath.Glob(filepath.Join(brigPath, "bolt.*"))
	if err != nil {
		panic(fmt.Sprintf("Bad pattern in glob: %s", err))
	}

	for _, match := range matches {
		lockPaths = append(lockPaths, filepath.Join(match, "index.bolt"))
	}

	return lockPaths
}

func lookupID(configPath string) (id.ID, error) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return "", fmt.Errorf("Could not load config: %v", err)
	}

	idString, err := cfg.String("repository.id")
	if err != nil {
		return "", fmt.Errorf("No ID in config: %v", err)
	}

	ID, err := id.Cast(idString)
	if err != nil {
		return "", err
	}

	return ID, nil
}

// Open decrypts all sensible data in the repository.
func Open(pwd, folder string) (*Repository, error) {
	absFolderPath, err := filepath.Abs(folder)
	brigPath := filepath.Join(absFolderPath, ".brig")

	// Figure out the JID from the config:
	ID, err := lookupID(filepath.Join(brigPath, "config"))
	if err != nil {
		return nil, err
	}

	// Unlock all files:
	var absNames []string
	for _, absName := range absLockPaths(brigPath) {
		if info, err := os.Stat(absName); err == nil {
			// File exists, this might happen on a crash or killed daemon.
			if info.Size() != 0 {
				log.Warningf("File is already unlocked: %s", absName)
			}
			continue
		}

		absNames = append(absNames, absName)
	}

	if err := UnlockFiles(string(ID), pwd, absNames); err != nil {
		return nil, err
	}

	return loadRepository(pwd, absFolderPath)
}

// Close encrypts sensible files in the repository.
// The password is taken from Repository.Password.
func (r *Repository) Close() error {
	var absNames []string
	for _, absName := range absLockPaths(r.InternalFolder) {
		info, err := os.Stat(absName)
		if os.IsNotExist(err) {
			// File does not exist. Might be already locked.
			log.Warningf("File is already locked: %s", absName)
			continue
		}

		// Work around minilock refusing to encrypt empty files.
		// (leave them as they are)
		if info.Size() == 0 {
			continue
		}

		log.Infof("Locking file `%v`...", absName)
		absNames = append(absNames, absName)
	}

	if err := LockFiles(string(r.ID), r.Password, absNames); err != nil {
		return err
	}

	return nil
}

// CheckPassword tries to decrypt a file in the repository.
// If that does not work, an error is returned.
func CheckPassword(folder, pwd string) error {
	absFolderPath, err := filepath.Abs(folder)
	brigPath := filepath.Join(absFolderPath, ".brig")

	ID, err := lookupID(filepath.Join(brigPath, "config"))
	if err != nil {
		return err
	}

	entry, err := parseShadowFile(brigPath, string(ID))
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
		"repository.id":   "",
		"repository.mid":  "",
		"repository.uuid": "",
	}

	for key := range configValues {
		configValues[key], err = cfg.String(key)
		if err != nil {
			return nil, err
		}
	}

	idString, err := cfg.String("repository.id")
	if err != nil {
		return nil, err
	}

	ID, err := id.Cast(idString)
	if err != nil {
		return nil, err
	}

	fd, err := os.Open(filepath.Join(brigPath, "remotes.yml"))
	if err != nil {
		return nil, err
	}

	remoteStore, err := NewYAMLRemotes(fd)
	if err != nil {
		return nil, err
	}

	// TODO: remove?
	ipfsAPIPort, err := cfg.Int("ipfs.apiport")
	if err != nil {
		return nil, err
	}

	ipfsSwarmPort, err := cfg.Int("ipfs.swarmport")
	if err != nil {
		return nil, err
	}

	ipfsLayer := ipfsutil.NewWithPorts(
		filepath.Join(brigPath, "ipfs"),
		ipfsAPIPort,
		ipfsSwarmPort,
	)

	ownStore, err := store.Open(brigPath, ID, ipfsLayer)
	if err != nil {
		return nil, err
	}

	mid, err := cfg.String("repository.mid")
	if err != nil {
		return nil, err
	}

	uuid, err := cfg.String("repository.uuid")
	if err != nil {
		return nil, err
	}

	allStores := make(map[id.ID]*store.Store)
	allStores[ID] = ownStore

	repo := Repository{
		ID:             ID,
		Mid:            mid,
		Folder:         absFolderPath,
		Remotes:        remoteStore,
		InternalFolder: brigPath,
		UniqueID:       uuid,
		Config:         cfg,
		OwnStore:       ownStore,
		Password:       pwd,
		IPFS:           ipfsLayer,
	}

	return &repo, nil
}
