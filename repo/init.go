package repo

import (
	"crypto/rand"
	"io"
	"os"
	"path/filepath"

	"code.google.com/p/go-uuid/uuid"
	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/repo/config"
	"github.com/disorganizer/brig/repo/global"
	"github.com/disorganizer/brig/store"
	logutil "github.com/disorganizer/brig/util/log"
	ipfsconfig "github.com/ipfs/go-ipfs/repo/config"
	"github.com/ipfs/go-ipfs/repo/fsrepo"
)

// NewRepository creates a new repository at filesystem level
// and returns a Repository interface
func NewRepository(jid, pwd, folder string) (*Repository, error) {
	absFolderPath, err := filepath.Abs(folder)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(absFolderPath); os.IsExist(err) {
		return nil, err
	}

	if err := createRepositoryTree(absFolderPath); err != nil {
		return nil, err
	}

	cfg := config.CreateDefaultConfig()
	minilockID, err := GenerateMinilockID(jid, pwd)
	if err != nil {
		return nil, err
	}

	configDefaults := map[string]interface{}{
		"repository.jid":  jid,
		"repository.uuid": uuid.NewRandom().String(),
		"repository.mid":  minilockID,
		"ipfs.path":       filepath.Join(absFolderPath, ".brig", "ipfs"),
	}

	for key, value := range configDefaults {
		if err = cfg.Set(key, value); err != nil {
			return nil, err
		}
	}

	configPath := filepath.Join(absFolderPath, ".brig", "config")
	if _, err := config.SaveConfig(configPath, cfg); err != nil {
		return nil, err
	}

	return LoadRepository(pwd, absFolderPath)
}

// CloneRepository clones a brig repository in a git like way
func CloneRepository() *Repository {
	return nil
}

// LoadRepository load a brig repository from a given folder.
func LoadRepository(pwd, folder string) (*Repository, error) {
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
		"repository.jid":  "",
		"repository.mid":  "",
		"repository.uuid": "",
	}

	for key := range configValues {
		configValues[key], err = cfg.String(key)
		if err != nil {
			return nil, err
		}
	}

	// Init the global repo (similar to .gitconfig)
	globalRepo, err := global.New()
	if err != nil {
		return nil, err
	}

	globalRepo.AddRepo(global.RepoListEntry{
		UniqueID:   configValues["repository.uuid"],
		RepoPath:   folder,
		DaemonPort: 6666,
		IpfsPort:   4001,
	})

	store, err := store.Open(brigPath)
	if err != nil {
		return nil, err
	}

	repo := Repository{
		Jid:            configValues["repository.jid"],
		Mid:            configValues["repository.mid"],
		Folder:         absFolderPath,
		InternalFolder: brigPath,
		UniqueID:       configValues["repository.uuid"],
		Config:         cfg,
		globalRepo:     globalRepo,
		Store:          store,
		Password:       pwd,
	}

	return &repo, nil
}

// touch works like the unix touch(1)
func touch(path string) error {
	fd, err := os.Create(path)
	if err != nil {
		return err
	}

	return fd.Close()
}

func createRepositoryTree(absFolderPath string) error {
	if err := os.Mkdir(absFolderPath, 0755); err != nil {
		return err
	}

	brigPath := filepath.Join(absFolderPath, ".brig")
	if err := os.Mkdir(brigPath, 0755); err != nil {
		return err
	}

	ipfsPath := filepath.Join(brigPath, "ipfs")
	if err := os.Mkdir(ipfsPath, 0755); err != nil {
		return err
	}

	boltDbPath := filepath.Join(brigPath, "index.bolt")
	if err := touch(boltDbPath); err != nil {
		return err
	}

	// Make the key larger than needed:
	if err := createMasterKey(brigPath, 1024); err != nil {
		return err
	}

	return createIPFS(ipfsPath)
}

func createMasterKey(brigPath string, keySize int) error {
	keyPath := filepath.Join(brigPath, "master.key")
	fd, err := os.OpenFile(keyPath, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}

	defer fd.Close()

	if _, err := io.CopyN(fd, rand.Reader, int64(keySize/8)); err != nil {
		return err
	}

	return nil
}

func createIPFS(ipfsRootPath string) error {
	logger := &logutil.Writer{Level: log.InfoLevel}
	cfg, err := ipfsconfig.Init(logger, 2048)
	if err != nil {
		return err
	}

	if err := fsrepo.Init(ipfsRootPath, cfg); err != nil {
		return err
	}

	return nil
}
