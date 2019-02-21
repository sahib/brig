package repo

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	log "github.com/sirupsen/logrus"
	e "github.com/pkg/errors"
	"github.com/sahib/brig/catfs"
	fserr "github.com/sahib/brig/catfs/errors"
	"github.com/sahib/brig/defaults"
	"github.com/sahib/config"
)

var (
	// Do not encrypt "data" (already contains encrypted streams) and
	excludedFromLock   = []string{"data", "OWNER", "BACKEND", "REPO_ID", "config.yml"}
	excludedFromUnlock = []string{"passwd.locked"}
)

var (
	// ErrBadPassword is returned by Open() when the decyption password seems to be wrong.
	ErrBadPassword = errors.New("Failed to open repository. Probably wrong password")
)

// Repository provides access to the file structure of a single repository.
//
// Informal: This file structure currently looks like this:
// config.yml
// OWNER
// BACKEND
// REPO_ID
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

// CheckPassword will try to validate `password` by decrypting something
// in `baseFolder`.
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
	owner, err := ioutil.ReadFile(ownerPath) // #nosec
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

// Open will open the repository at `baseFolder` by using `password`.
func Open(baseFolder, password string) (*Repository, error) {
	// This is only a sanity check here. If the wrong password
	// was supplied, we won't be able to unlock the repo anyways.
	// But try to bail out here with an meaningful error message.
	if err := CheckPassword(baseFolder, password); err != nil {
		return nil, err
	}

	ownerPath := filepath.Join(baseFolder, "OWNER")
	owner, err := ioutil.ReadFile(ownerPath) // #nosec
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
	backendName, err := ioutil.ReadFile(backendNamePath) // #nosec
	if err != nil {
		return nil, err
	}

	return &Repository{
		BaseFolder:  baseFolder,
		backendName: string(backendName),
		Config:      cfg,
		Remotes:     remotes,
		Owner:       string(owner),
		fsMap:       make(map[string]*catfs.FS),
	}, nil
}

// Close will lock the repository, making this instance unusable.
func (rp *Repository) Close(password string) error {
	return LockRepo(
		rp.BaseFolder,
		rp.Owner,
		password,
		excludedFromLock,
		excludedFromUnlock,
	)
}

// BackendName returns the backend name used when constructing the repo.
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
	if err := os.MkdirAll(fsDbPath, 0700); err != nil && err != os.ErrExist {
		return nil, err
	}

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

// CurrentUser returns the current user of the repository.
// (i.e. what FS is being shown)
func (rp *Repository) CurrentUser() string {
	return rp.Config.String("repo.current_user")
}

// SetCurrentUser sets the current user of the repository.
// (i.e. called by "become" when changing the FS)
func (rp *Repository) SetCurrentUser(user string) {
	rp.Config.Set("repo.current_user", user)
}

// Keyring returns the keyring of the repository.
func (rp *Repository) Keyring() *Keyring {
	return newKeyringHandle(rp.BaseFolder)
}

// BackendPath returns the absolute path to the backend folder inside the repo.
func (rp *Repository) BackendPath(name string) string {
	return filepath.Join(rp.BaseFolder, "data", name)
}

// RepoID returns a unique ID specific to this repository.
func (rp *Repository) RepoID() (string, error) {
	data, err := ioutil.ReadFile(filepath.Join(rp.BaseFolder, "REPO_ID"))
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// SaveConfig dumps the in memory config to disk.
func (rp *Repository) SaveConfig() error {
	configPath := filepath.Join(rp.BaseFolder, "config.yml")
	return config.ToYamlFile(configPath, rp.Config)
}
