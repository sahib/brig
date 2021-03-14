package repo

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/dustin/go-humanize"
	e "github.com/pkg/errors"
	"github.com/sahib/brig/catfs"
	fserr "github.com/sahib/brig/catfs/errors"
	"github.com/sahib/brig/catfs/mio/pagecache/mdcache"
	"github.com/sahib/brig/defaults"
	"github.com/sahib/brig/repo/hints"
	"github.com/sahib/config"
	log "github.com/sirupsen/logrus"
)

// Repository provides access to the file structure of a single repository.
//
// Informal: This file structure currently looks like this:
//
// config.yml
// immutables.yml
// remotes.yml
// keyring/
//    <remote_name>
//        key.prv
//        key.pub
// metadata/
//    <remote_name>
//        (fs-backend specific)
// gateway/
//    (gateway specific)
type Repository struct {
	mu sync.Mutex

	// Map between owner and related filesystem.
	fsMap map[string]*catfs.FS

	// Absolute path to the repository root
	BaseFolder string

	// Config interface
	Config *config.Config

	// Immutables gives access to things that do not change
	// after initializing the repository.
	Immutables *Immutables

	// Remotes gives access to all known remotes
	Remotes *RemoteList

	// Hints are streaming settings
	Hints *hints.HintManager

	// channel to control the auto gc loop
	autoGCControl chan bool
}

func loadHintManager(hintsPath string) (*hints.HintManager, error) {
	hintsFd, err := os.Open(hintsPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, e.Wrap(err, "failed to open hints.yml")
	}

	if os.IsNotExist(err) {
		// No such file yet, create it.
		return hints.NewManager(nil)
	}

	defer hintsFd.Close()

	return hints.NewManager(hintsFd)
}

// Open will open the repository at `baseFolder`
func Open(baseFolder string) (*Repository, error) {
	immutables, err := NewImmutables(filepath.Join(baseFolder, "immutable.yml"))
	if err != nil {
		return nil, e.Wrap(err, "failed to open immutable store")
	}

	cfgPath := filepath.Join(baseFolder, "config.yml")
	cfg, err := defaults.OpenMigratedConfig(cfgPath)
	if err != nil {
		return nil, err
	}

	// TODO: Why do we do this?
	cfg.SetString("repo.current_user", immutables.Owner())

	remotePath := filepath.Join(baseFolder, "remotes.yml")
	remotes, err := NewRemotes(remotePath)
	if err != nil {
		return nil, err
	}

	hintsMgr, err := loadHintManager(filepath.Join(baseFolder, "hints.yml"))
	if err != nil {
		return nil, err
	}

	return &Repository{
		BaseFolder:    baseFolder,
		Immutables:    immutables,
		Config:        cfg,
		Remotes:       remotes,
		Hints:         hintsMgr,
		fsMap:         make(map[string]*catfs.FS),
		autoGCControl: make(chan bool, 1),
	}, nil
}

// Close will lock the repository, making this instance unusable.
func (rp *Repository) Close() error {
	rp.stopAutoGCLoop()
	for owner, fs := range rp.fsMap {
		log.Infof("closing FS for %s", owner)
		fs.Close()
	}

	return nil
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
		// return cached:
		return fs, nil
	}

	isReadOnly := rp.Immutables.Owner() != owner

	// No fs was created yet for this owner.
	// Create it & give it a part of the main config.
	fsCfg := rp.Config.Section("fs")
	fsDbPath := filepath.Join(rp.BaseFolder, "metadata", owner)
	if err := os.MkdirAll(fsDbPath, 0700); err != nil && err != os.ErrExist {
		return nil, err
	}

	pageCachePath := filepath.Join(rp.BaseFolder, "pages", owner)
	if err := os.MkdirAll(pageCachePath, 0700); err != nil && err != os.ErrExist {
		return nil, err
	}

	pageCacheMaxMemorySrc := fsCfg.String("pagecache.max_memory")
	pageCacheMaxMemory, err := humanize.ParseBytes(pageCacheMaxMemorySrc)
	if err != nil {
		return nil, e.Wrapf(err, "failed to parse fs.pagecache.max_memory")
	}

	pageCache, err := mdcache.New(mdcache.Options{
		MaxMemoryUsage:    int64(pageCacheMaxMemory),
		SwapDirectory:     pageCachePath,
		L1CacheMissRefill: true,
		L2Compress:        fsCfg.Bool("pagecache.l2compress"),
	})

	if err != nil {
		return nil, err
	}

	fs, err := catfs.NewFilesystem(
		bk,
		fsDbPath,
		owner,
		isReadOnly,
		fsCfg,
		rp.Hints,
		pageCache,
	)

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
func (rp *Repository) Keyring() (*Keyring, error) {
	owner := rp.Immutables.Owner()
	path := filepath.Join(rp.BaseFolder, "keyring")
	if err := os.MkdirAll(path, 0700); err != nil {
		log.WithError(err).Warnf("failed to create keyring directory: %s", path)
		return nil, err
	}

	return newKeyringHandle(path, owner), nil
}

// SaveConfig dumps the in memory config to disk.
func (rp *Repository) SaveConfig() error {
	configPath := filepath.Join(rp.BaseFolder, "config.yml")
	return config.ToYamlFile(configPath, rp.Config)
}

// SaveHints dumps the hints settings to disk.
// You should call this whenever Hints are changed.
func (rp *Repository) SaveHints() error {
	hintsPath := filepath.Join(rp.BaseFolder, "hints.yml")
	fd, err := os.OpenFile(hintsPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	defer fd.Close()

	return rp.Hints.Save(fd)
}
