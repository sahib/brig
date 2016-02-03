// Package global implements the logic behind the global config files in
// ~/.brigconfig
package global

import (
	"os"
	"os/user"
	"path"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/repo/config"
	"github.com/disorganizer/brig/util/filelock"
	yamlConfig "github.com/olebedev/config"
)

const (
	dirName = ".brigconfig"
)

// Repository is the handle for the global repository.
type Repository struct {
	Folder string
	Config *yamlConfig.Config
}

// RepoListEntry is a single entry in ~/.brigconfig/repos
type RepoListEntry struct {
	UniqueID   string
	RepoPath   string
	DaemonPort int
	IpfsPort   int
}

func (g *Repository) acquireLock() error {
	lockPath := path.Join(g.Folder, "lock")
	if err := filelock.Acquire(lockPath); err != nil {
		return err
	}

	return nil
}

func (g *Repository) releaseLock() {
	if err := filelock.Release(path.Join(g.Folder, "lock")); err != nil {
		log.Warningf("global: release-lock failed: %v", err)
	}
}

func guessGlobalFolder() string {
	curr, err := user.Current()
	if err != nil {
		return os.TempDir()
	}

	return path.Join(curr.HomeDir, dirName)
}

// Init creates a new global Repository and returns it.
func Init() (*Repository, error) {
	folder := guessGlobalFolder()
	repo := &Repository{
		Folder: folder,
	}

	if err := os.Mkdir(folder, 0755); err != nil && !os.IsExist(err) {
		return nil, err
	}

	if err := repo.acquireLock(); err != nil {
		return nil, err
	}

	defer repo.releaseLock()

	cfg := &yamlConfig.Config{
		Root: map[string]interface{}{
			"repositories": map[string]RepoListEntry{},
		},
	}

	if _, err := config.SaveConfig(path.Join(folder, "repos"), cfg); err != nil {
		return nil, err
	}

	repo.Config = cfg
	return repo, nil
}

// Load loads an existing global repository.
func Load() (*Repository, error) {
	folder := guessGlobalFolder()
	repo := &Repository{
		Folder: folder,
	}

	if err := repo.acquireLock(); err != nil {
		return nil, err
	}
	defer repo.releaseLock()

	cfg, err := config.LoadConfig(path.Join(folder, "repos"))
	if err != nil {
		return nil, err
	}

	repo.Config = cfg
	return repo, nil
}

// New loads a global repository, if it's not there, it's created.
func New() (*Repository, error) {
	folder := guessGlobalFolder()
	if _, err := os.Stat(folder); os.IsExist(err) {
		return Load()
	}

	return Init()
}

func (g *Repository) modifyConfig(worker func(cfg *yamlConfig.Config) error) error {
	if err := g.acquireLock(); err != nil {
		return err
	}
	defer g.releaseLock()

	cfg, err := config.LoadConfig(path.Join(g.Folder, "repos"))
	if err != nil {
		return err
	}

	if err := worker(cfg); err != nil {
		return err
	}

	if _, err := config.SaveConfig(path.Join(g.Folder, "repos"), cfg); err != nil {
		return err
	}

	return nil
}

// AddRepo adds a new repo to ~/.brigconfig/repos
func (g *Repository) AddRepo(entry RepoListEntry) error {
	return g.modifyConfig(func(cfg *yamlConfig.Config) error {
		repos, err := cfg.Map("repositories")
		if err != nil {
			return err
		}

		repos[entry.UniqueID] = entry
		return nil
	})
}

// RemoveRepo deletes an existing repo to ~/.brigconfig/repos
func (g *Repository) RemoveRepo(entry RepoListEntry) error {
	return g.modifyConfig(func(cfg *yamlConfig.Config) error {
		repos, err := cfg.Map("repositories")
		if err != nil {
			return err
		}

		delete(repos, entry.UniqueID)
		return nil
	})
}
