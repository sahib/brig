// Package global implements the logic behind the global config files in
// ~/.brigconfig
package global

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/user"
	"path"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/util/filelock"
	"gopkg.in/yaml.v2"
)

const (
	dirName      = ".brigconfig"
	maxPortTries = 5000
)

// Repository is the handle for the global repository.
type Repository struct {
	Folder string
}

type repoList struct {
	Repos map[string]*RepoListEntry
}

// RepoListEntry is a single entry in ~/.brigconfig/repos
type RepoListEntry struct {
	UniqueID      string
	RepoPath      string
	DaemonPort    int
	IpfsSwarmPort int
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

	// Save an empty config:
	if err := repo.save(&repoList{}); err != nil {
		return nil, err
	}

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

	// Try to load the config, so errors get reported early:
	_, err := repo.load()
	if err != nil {
		return nil, err
	}

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

func (g *Repository) load() (*repoList, error) {
	path := path.Join(g.Folder, "repos")
	data, readErr := ioutil.ReadFile(path)
	if readErr != nil {
		return nil, readErr
	}

	l := &repoList{}
	err := yaml.Unmarshal(data, &l)

	if l.Repos == nil {
		l.Repos = make(map[string]*RepoListEntry)
	}

	if err != nil {
		return nil, err
	}

	return l, nil
}

func (g *Repository) save(list *repoList) error {
	data, err := yaml.Marshal(list)
	if err != nil {
		return err
	}

	path := path.Join(g.Folder, "repos")
	if err := ioutil.WriteFile(path, data, 0644); err != nil {
		return err
	}

	return nil
}

func (g *Repository) modify(handle func(lst *repoList)) error {
	if err := g.acquireLock(); err != nil {
		return err
	}
	defer g.releaseLock()

	lst, err := g.load()
	if err != nil {
		return err
	}

	handle(lst)

	if err := g.save(lst); err != nil {
		return err
	}

	return nil
}

func (g *Repository) view(handle func(lst *repoList)) error {
	if err := g.acquireLock(); err != nil {
		return err
	}
	defer g.releaseLock()

	lst, err := g.load()
	if err != nil {
		return err
	}

	handle(lst)
	return nil
}

// AddRepo adds a new repo to ~/.brigconfig/repos
func (g *Repository) AddRepo(entry RepoListEntry) error {
	return g.modify(func(lst *repoList) {
		lst.Repos[entry.UniqueID] = &entry
	})
}

// RemoveRepo deletes an existing repo to ~/.brigconfig/repos
func (g *Repository) RemoveRepo(entry RepoListEntry) error {
	return g.modify(func(lst *repoList) {
		delete(lst.Repos, entry.UniqueID)
	})
}

// check if localhost:$port looks like it's reserved
func localPortIsReserved(port int) bool {
	addr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return true
	}

	timeout := 50 * time.Millisecond
	conn, err := net.DialTimeout("tcp", addr.String(), timeout)

	if err != nil {
		return false
	}

	conn.Close()
	return true
}

func nextFreePort(start int) (int, error) {
	// This could happen in parallel, but that wouldn't
	// get the next "logical" port. At least not faster.
	for p := start; p < start+maxPortTries; p++ {
		if !localPortIsReserved(p) {
			return p, nil
		}
	}

	return -1, fmt.Errorf("Could not find suitable port in [%d,%d]", start, start+maxPortTries)
}

func (g *Repository) findMaxPort(maxPort int) (int, error) {
	err := g.view(func(lst *repoList) {
		for _, repo := range lst.Repos {
			// There might be better heuristics than this:
			if maxPort < repo.IpfsSwarmPort {
				maxPort = repo.IpfsSwarmPort + 1
			}

			if maxPort < repo.DaemonPort {
				maxPort = repo.DaemonPort + 1
			}
		}
	})

	if err != nil {
		return -1, err
	}

	return nextFreePort(maxPort)
}

func (g *Repository) NextIPFSSwarmPort() (int, error) {
	return g.findMaxPort(4001)
}

func (g *Repository) NextDaemonPort() (int, error) {
	return g.findMaxPort(6666)
}
