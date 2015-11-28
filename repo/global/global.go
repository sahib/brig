package global

import (
	"os"
	"os/user"
	"path"

	"github.com/disorganizer/brig/repo/config"
	"github.com/disorganizer/brig/util/filelock"
	yamlConfig "github.com/olebedev/config"
)

type GlobalRepository struct {
	Folder string
	Config *yamlConfig.Config
}

type RepoListEntry struct {
	UniqueID   string
	RepoPath   string
	DaemonPort int
	IpfsPort   int
}

func (g *GlobalRepository) acquireLock() error {
	lockPath := path.Join(g.Folder, "lock")
	if err := filelock.Acquire(lockPath); err != nil {
		return err
	}

	return nil
}

func (g *GlobalRepository) releaseLock() error {
	return filelock.Release(path.Join(g.Folder, "lock"))
}

func guessGlobalFolder() string {
	curr, err := user.Current()
	if err != nil {
		return os.TempDir()
	}

	return path.Join(curr.HomeDir, ".brig")
}

func Init() (*GlobalRepository, error) {
	folder := guessGlobalFolder()
	repo := &GlobalRepository{
		Folder: folder,
	}

	if err := os.Mkdir(folder, 0755); err != nil {
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

func Load() (*GlobalRepository, error) {
	folder := guessGlobalFolder()
	repo := &GlobalRepository{
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

func New() (*GlobalRepository, error) {
	folder := guessGlobalFolder()
	if _, err := os.Stat(folder); os.IsExist(err) {
		return Load()
	}

	return Init()
}

func (g *GlobalRepository) modifyConfig(worker func(cfg *yamlConfig.Config) error) error {
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

func (g *GlobalRepository) AddRepo(entry RepoListEntry) error {
	return g.modifyConfig(func(cfg *yamlConfig.Config) error {
		repos, err := cfg.Map("repositories")
		if err != nil {
			return err
		}

		repos[entry.UniqueID] = entry
		return nil
	})
}

func (g *GlobalRepository) RemoveRepo(entry RepoListEntry) error {
	return g.modifyConfig(func(cfg *yamlConfig.Config) error {
		repos, err := cfg.Map("repositories")
		if err != nil {
			return err
		}

		delete(repos, entry.UniqueID)
		return nil
	})
}
