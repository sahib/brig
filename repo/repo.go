package repo

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/disorganizer/brig/catfs"
	e "github.com/pkg/errors"
	"github.com/spf13/viper"
)

// Repository provides access to the file structure of a single repository.
//
// Informal: This file structure currently looks like this:
// config.yml
// remotes.yml
// data/
//    (backend specific)
// metadata/
//    <name_1>
//        (backend specific)
//    <name_2>
//        (backend specific)
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

	// Remotes gives access to all known remotes
	Remotes *RemoteList

	FSBackend catfs.FsBackend
}

func touch(path string) error {
	fd, err := os.OpenFile(path, os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	return fd.Close()
}

func isEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}

func Init(baseFolder, owner string, backend RepoBackend) error {
	// The basefolder has to exist:
	info, err := os.Stat(baseFolder)
	if os.IsNotExist(err) {
		return err
	}

	repoDirIsEmpty, err := isEmpty(baseFolder)
	if err != nil {
		return err
	}

	if !info.IsDir() || !repoDirIsEmpty {
		return fmt.Errorf(
			"`%s` is not a directory or it's not empty",
			baseFolder,
		)
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

	if err := touch(filepath.Join(baseFolder, "config.yml")); err != nil {
		return e.Wrapf(err, "Failed touch config.yml")
	}

	whoamiPath := filepath.Join(baseFolder, "whoami")
	if err := ioutil.WriteFile(whoamiPath, []byte(owner), 0644); err != nil {
		return err
	}

	dataFolder := filepath.Join(baseFolder, "data")
	if err := backend.Init(dataFolder); err != nil {
		return e.Wrap(err, "Failed to init data backend")
	}

	return nil
}

func Open(baseFolder string, backend catfs.FsBackend) (*Repository, error) {
	// Make sure to load the config:
	config := viper.New()
	config.AddConfigPath(baseFolder)
	setConfigDefaults(config)

	if err := config.ReadInConfig(); err != nil {
		return nil, err
	}

	// Load the remote list:
	remotePath := filepath.Join(baseFolder, "remotes.yml")
	fd, err := os.Open(remotePath)
	if err != nil {
		return nil, err
	}

	defer fd.Close()

	remotes, err := NewRemotes(fd)
	if err != nil {
		return nil, err
	}

	whoamiPath := filepath.Join(baseFolder, "whoami")
	owner, err := ioutil.ReadFile(whoamiPath)
	if err != nil {
		return nil, err
	}

	return &Repository{
		BaseFolder: baseFolder,
		Config:     config,
		Remotes:    remotes,
		Owner:      string(owner),
		FSBackend:  backend,
		fsMap:      make(map[string]*catfs.FS),
	}, nil
}

func (rp *Repository) Close() error {
	// TODO: Close() currently does nothing, but it should encrypt config/remotes
	//       so they can get decrypted again on startup.
	return nil
}

// FS returns a filesystem for `owner`. If there is none yet,
// it will create own associated to the respective owner.
func (rp *Repository) FS(owner string) (*catfs.FS, error) {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	if fs, ok := rp.fsMap[owner]; ok {
		return fs, nil
	}

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

	// TODO: Does it make really sense to store the hash in fs?
	//       Maybe user management and repo management should be two things.
	person := catfs.Person{
		Name: owner,
		Hash: nil,
	}

	fsDbPath := filepath.Join(rp.BaseFolder, "data", owner)
	fs, err := catfs.NewFilesystem(rp.FSBackend, fsDbPath, &person, fsCfg)
	if err != nil {
		return nil, err
	}

	// Store for next call:
	rp.fsMap[owner] = fs
	return fs, nil
}

// OwnFS returns the filesystem for the owner.
func (rp *Repository) OwnFS() (*catfs.FS, error) {
	return rp.FS(rp.Owner)
}
