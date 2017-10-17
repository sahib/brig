package repo

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/backend"
	"github.com/disorganizer/brig/catfs"
	e "github.com/pkg/errors"
	"github.com/spf13/viper"
)

// Repository provides access to the file structure of a single repository.
//
// Informal: This file structure currently looks like this:
// config.yml
// meta.yml
// remotes.yml
// data/
//    <backend_name>
//        (backend specific)
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

func Init(baseFolder, owner, backendName string) error {
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

	realBackend := backend.FromName(backendName)
	if realBackend == nil {
		return fmt.Errorf("No such backend `%s`", backendName)
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

	dataFolder := filepath.Join(baseFolder, "data")
	if err := realBackend.Init(dataFolder); err != nil {
		return e.Wrap(err, "Failed to init data backend")
	}

	return nil
}

func Open(baseFolder, password string) (*Repository, error) {
	metaPath := filepath.Join(baseFolder, "meta.yml")
	meta := viper.New()
	meta.SetConfigFile(metaPath)
	if err := meta.ReadInConfig(); err != nil {
		return nil, err
	}

	owner := meta.GetString("repo.owner")
	if err := UnlockRepo(baseFolder, owner, password); err != nil {
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
	remoteFd, err := os.Open(remotePath)
	if err != nil {
		return nil, err
	}

	defer remoteFd.Close()

	remotes, err := NewRemotes(remoteFd)
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
	// Do not encrypt "data" (already contains encrypred streams) and
	// also do not encrypt meta.yml (contains e.g. owner info)
	return LockRepo(rp.BaseFolder, rp.Owner, password, []string{"data", "meta.yml"})
}

func (rp *Repository) LoadBackend() (backend.Backend, error) {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	backendName := rp.meta.GetString("data.backend")
	log.Infof("Loading backend `%s`", backendName)

	realBackend := backend.FromName(backendName)
	if realBackend == nil {
		msg := fmt.Sprintf("No such backend `%s`", backendName)
		log.Error(msg)
		return nil, fmt.Errorf("open failed: %s", msg)
	}

	return realBackend, nil
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

	// TODO: Does it make really sense to store the hash in fs?
	//       Maybe user management and repo management should be two things.
	person := catfs.Person{
		Name: owner,
		Hash: nil,
	}

	fsDbPath := filepath.Join(rp.BaseFolder, "data", owner)
	fs, err := catfs.NewFilesystem(bk, fsDbPath, &person, fsCfg)
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
