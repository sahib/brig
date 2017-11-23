package repo

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/catfs"
	e "github.com/pkg/errors"
	"github.com/spf13/viper"
)

var (
	// Do not encrypt "data" (already contains encrypred streams) and
	// also do not encrypt meta.yml (contains e.g. owner info for startup)
	excludedFromLock = []string{"meta.yml", "data"}
)

var (
	ErrBadPassword = errors.New("Failed to open repository. Probably wrong password")
)

// Repository provides access to the file structure of a single repository.
//
// Informal: This file structure currently looks like this:
// config.yml
// meta.yml
// remotes.yml
// data/
//    <backend_name>
//        (data-backend specific)
// metadata/
//    <name_1>
//        (fs-backend specific)
//    <name_2>
//        (fs-backend specific)
type Repository struct {
	mu sync.Mutex

	// Map between owner and related filesystem.
	fsMap map[string]*catfs.FS

	// Absolute path to the repository root
	BaseFolder string

	// Name of the owner of this repository
	Owner string

	// Config interface
	Config *viper.Viper
	meta   *viper.Viper

	// Remotes gives access to all known remotes
	Remotes *RemoteList
}

func touch(path string) error {
	fd, err := os.OpenFile(path, os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	return fd.Close()
}

func Init(baseFolder, owner, password, backendName string) error {
	// The basefolder has to exist:
	info, err := os.Stat(baseFolder)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(baseFolder, 0700); err != nil {
			return err
		}
	} else if info.Mode().IsDir() {
		log.Warningf("`%s` is a directory and exists", baseFolder)
	} else {
		return fmt.Errorf("`%s` is a file (should be a directory)", baseFolder)
	}

	// Create (empty) folders:
	folders := []string{"metadata", "data"}
	for _, folder := range folders {
		absFolder := filepath.Join(baseFolder, folder)
		if err := os.Mkdir(absFolder, 0700); err != nil {
			return e.Wrapf(err, "Failed to create dir: %v", absFolder)
		}
	}

	if err := touch(filepath.Join(baseFolder, "remotes.yml")); err != nil {
		return e.Wrapf(err, "Failed touch remotes.yml")
	}

	metaPath := filepath.Join(baseFolder, "meta.yml")
	metaDefault := buildMetaDefault(backendName, owner)
	if err := ioutil.WriteFile(metaPath, metaDefault, 0644); err != nil {
		return err
	}

	cfgPath := filepath.Join(baseFolder, "config.yml")
	cfgDefaults := buildConfigDefault()
	if err := ioutil.WriteFile(cfgPath, cfgDefaults, 0644); err != nil {
		return err
	}

	dataFolder := filepath.Join(baseFolder, "data", backendName)
	if err := os.MkdirAll(dataFolder, 0700); err != nil {
		return e.Wrap(err, "Failed to setup dirs for backend")
	}

	// Create initial key pair.
	if err := createKeyPair(owner, baseFolder, 2048); err != nil {
		return e.Wrap(err, "Failed to setup pgp keys")
	}

	passwdFile := filepath.Join(baseFolder, "passwd")
	passwdData := fmt.Sprintf("%s", owner)
	if err := ioutil.WriteFile(passwdFile, []byte(passwdData), 0644); err != nil {
		return err
	}

	// passwd is used to verify the user password,
	// so it needs to be locked only once on init and
	// kept out otherwise from the locking machinery.
	if err := lockFile(passwdFile, keyFromPassword(owner, password)); err != nil {
		return err
	}

	return LockRepo(baseFolder, owner, password, excludedFromLock)
}

func CheckPassword(baseFolder, password string) error {
	passwdFile := filepath.Join(baseFolder, "passwd.locked")

	// If the file does not exist yet, it probably means
	// that the repo was not initialized yet.
	// Act like the password is okay and wait for the init.
	if _, err := os.Stat(passwdFile); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	// Try to get the owner of the repo.
	// Needed for the key derivation function.
	metaPath := filepath.Join(baseFolder, "meta.yml")
	meta := viper.New()
	meta.SetConfigFile(metaPath)
	if err := meta.ReadInConfig(); err != nil {
		return err
	}

	owner := meta.GetString("repo.owner")
	key := keyFromPassword(owner, password)
	if err := checkUnlockability(passwdFile, key); err != nil {
		log.Warningf("Failed to unlock passwd file. Wrong password entered?")
		return ErrBadPassword
	}

	return nil
}

func Open(baseFolder, password string) (*Repository, error) {
	// This is only a sanity check here. If the wrong password
	// was supplied, we won't be able to unlock the repo anyways.
	// But try to bail out here with an meaningful error message.
	if err := CheckPassword(baseFolder, password); err != nil {
		return nil, err
	}

	metaPath := filepath.Join(baseFolder, "meta.yml")
	meta := viper.New()
	meta.SetConfigFile(metaPath)
	if err := meta.ReadInConfig(); err != nil {
		return nil, err
	}

	owner := meta.GetString("repo.owner")
	if err := UnlockRepo(baseFolder, owner, password, excludedFromLock); err != nil {
		return nil, err
	}

	// Make sure to load the config:
	config := viper.New()
	config.AddConfigPath(baseFolder)
	setConfigDefaults(config)

	if err := config.ReadInConfig(); err != nil {
		return nil, err
	}

	// Load the remote list:
	remotePath := filepath.Join(baseFolder, "remotes.yml")
	remotes, err := NewRemotes(remotePath)
	if err != nil {
		return nil, err
	}

	return &Repository{
		BaseFolder: baseFolder,
		meta:       meta,
		Config:     config,
		Remotes:    remotes,
		Owner:      owner,
		fsMap:      make(map[string]*catfs.FS),
	}, nil
}

func (rp *Repository) Close(password string) error {
	return LockRepo(rp.BaseFolder, rp.Owner, password, excludedFromLock)
}

func (rp *Repository) BackendName() string {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	return rp.meta.GetString("data.backend")
}

// FS returns a filesystem for `owner`. If there is none yet,
// it will create own associated to the respective owner.
func (rp *Repository) FS(owner string, bk catfs.FsBackend) (*catfs.FS, error) {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	if fs, ok := rp.fsMap[owner]; ok {
		return fs, nil
	}

	// No fs was created yet for this owner. Create it.
	// Read the fs config from the main config:
	fsCfg := &catfs.Config{}
	fsCfg.IO.CompressAlgo = rp.Config.GetString(
		"data.compress.algo",
	)
	fsCfg.Sync.ConflictStrategy = rp.Config.GetString(
		"sync.conflict_strategy",
	)
	fsCfg.Sync.IgnoreRemoved = rp.Config.GetBool(
		"sync.ignore_removed",
	)

	fsDbPath := filepath.Join(rp.BaseFolder, "metadata", owner)
	fs, err := catfs.NewFilesystem(bk, fsDbPath, owner, fsCfg)
	if err != nil {
		return nil, err
	}

	// Store for next call:
	rp.fsMap[owner] = fs
	return fs, nil
}

// OwnFS returns the filesystem for the owner.
func (rp *Repository) OwnFS(bk catfs.FsBackend) (*catfs.FS, error) {
	return rp.FS(rp.Owner, bk)
}

func (rp *Repository) Keyring() *Keyring {
	return newKeyringHandle(rp.BaseFolder)
}
