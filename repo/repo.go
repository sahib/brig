package repo

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	log "github.com/Sirupsen/logrus"
	e "github.com/pkg/errors"
	"github.com/sahib/brig/catfs"
	fserr "github.com/sahib/brig/catfs/errors"
	"github.com/sahib/brig/defaults"
	"github.com/sahib/config"
)

var (
	// Do not encrypt "data" (already contains encrypted streams) and
	excludedFromLock   = []string{"data", "OWNER", "BACKEND"}
	excludedFromUnlock = []string{"passwd.locked"}
)

var (
	ErrBadPassword = errors.New("Failed to open repository. Probably wrong password")
)

// Repository provides access to the file structure of a single repository.
//
// Informal: This file structure currently looks like this:
// config.yml
// OWNER
// BACKEND
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

	// Name of the backend in use
	backendName string

	// Absolute path to the repository root
	BaseFolder string

	// Name of the owner of this repository
	Owner string

	// Config interface
	Config *config.Config

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
			return e.Wrapf(err, "Failed to create dir: %v (repo exists?)", absFolder)
		}
	}

	if err := touch(filepath.Join(baseFolder, "remotes.yml")); err != nil {
		return e.Wrapf(err, "Failed touch remotes.yml")
	}

	ownerPath := filepath.Join(baseFolder, "OWNER")
	if err := ioutil.WriteFile(ownerPath, []byte(owner), 0644); err != nil {
		return err
	}

	backendNamePath := filepath.Join(baseFolder, "BACKEND")
	if err := ioutil.WriteFile(backendNamePath, []byte(backendName), 0644); err != nil {
		return err
	}

	// For future use: If we ever need to migrate the repo.
	versionPath := filepath.Join(baseFolder, "VERSION")
	if err := ioutil.WriteFile(versionPath, []byte{1}, 0644); err != nil {
		return err
	}

	// Create a default config, only with the default keys applied:
	cfg, err := config.Open(nil, defaults.Defaults)
	if err != nil {
		return err
	}

	configPath := filepath.Join(baseFolder, "config.yml")
	if err := config.ToYamlFile(configPath, cfg); err != nil {
		return e.Wrap(err, "Failed to setup default config")
	}

	dataFolder := filepath.Join(baseFolder, "data", backendName)
	if err := os.MkdirAll(dataFolder, 0700); err != nil {
		return e.Wrap(err, "Failed to setup dirs for backend")
	}

	// Create initial key pair:
	if err := createKeyPair(owner, baseFolder, 2048); err != nil {
		return e.Wrap(err, "Failed to setup gpg keys")
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

	logDir := filepath.Join(baseFolder, "logs")
	if err := os.MkdirAll(logDir, 0700); err != nil {
		return err
	}

	return LockRepo(
		baseFolder,
		owner,
		password,
		excludedFromLock,
		excludedFromUnlock,
	)
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
	ownerPath := filepath.Join(baseFolder, "OWNER")
	owner, err := ioutil.ReadFile(ownerPath)
	if err != nil {
		return e.Wrap(err, "failed to read OWNER")
	}

	key := keyFromPassword(string(owner), password)
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

	ownerPath := filepath.Join(baseFolder, "OWNER")
	owner, err := ioutil.ReadFile(ownerPath)
	if err != nil {
		return nil, e.Wrap(err, "failed to read OWNER")
	}

	err = UnlockRepo(
		baseFolder,
		string(owner),
		password,
		excludedFromLock,
		excludedFromUnlock,
	)

	if err != nil {
		return nil, err
	}

	cfgPath := filepath.Join(baseFolder, "config.yml")
	cfg, err := defaults.OpenMigratedConfig(cfgPath)
	if err != nil {
		return nil, err
	}

	cfg.SetString("repo.current_user", string(owner))

	// Load the remote list:
	remotePath := filepath.Join(baseFolder, "remotes.yml")
	remotes, err := NewRemotes(remotePath)
	if err != nil {
		return nil, err
	}

	backendNamePath := filepath.Join(baseFolder, "BACKEND")
	backendName, err := ioutil.ReadFile(backendNamePath)

	return &Repository{
		BaseFolder:  baseFolder,
		backendName: string(backendName),
		Config:      cfg,
		Remotes:     remotes,
		Owner:       string(owner),
		fsMap:       make(map[string]*catfs.FS),
	}, nil
}

func (rp *Repository) Close(password string) error {
	return LockRepo(
		rp.BaseFolder,
		rp.Owner,
		password,
		excludedFromLock,
		excludedFromUnlock,
	)
}

func (rp *Repository) BackendName() string {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	return rp.backendName
}

// HaveFS will return true if we have data for a certain owner.
func (rp *Repository) HaveFS(owner string) bool {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	fsDbPath := filepath.Join(rp.BaseFolder, "metadata", owner)
	if _, err := os.Stat(fsDbPath); err != nil {
		return false
	}

	return true
}

// FS returns a filesystem for `owner`. If there is none yet,
// it will create own associated to the respective owner.
func (rp *Repository) FS(owner string, bk catfs.FsBackend) (*catfs.FS, error) {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	if fs, ok := rp.fsMap[owner]; ok {
		return fs, nil
	}

	isReadOnly := rp.Owner != owner

	// No fs was created yet for this owner.
	// Create it & give it a part of the main config.
	fsCfg := rp.Config.Section("fs")
	fsDbPath := filepath.Join(rp.BaseFolder, "metadata", owner)
	fs, err := catfs.NewFilesystem(bk, fsDbPath, owner, isReadOnly, fsCfg)
	if err != nil {
		return nil, err
	}

	// Create an initial commit if there was none yet:
	if _, err := fs.Head(); fserr.IsErrNoSuchRef(err) {
		if err := fs.MakeCommit("initial commit"); err != nil {
			return nil, err
		}
	}

	// Store for next call:
	rp.fsMap[owner] = fs
	return fs, nil
}

func (rp *Repository) CurrentUser() string {
	return rp.Config.String("repo.current_user")
}

func (rp *Repository) SetCurrentUser(user string) {
	rp.Config.Set("repo.current_user", user)
}

func (rp *Repository) Keyring() *Keyring {
	return newKeyringHandle(rp.BaseFolder)
}

func (rp *Repository) BackendPath(name string) string {
	return filepath.Join(rp.BaseFolder, "data", name)
}
