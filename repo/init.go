package repo

import (
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"code.google.com/p/go-uuid/uuid"
	"github.com/disorganizer/brig/repo/config"
	"github.com/disorganizer/brig/repo/global"
	"github.com/disorganizer/brig/store"
	ipfsconfig "github.com/ipfs/go-ipfs/repo/config"
	"github.com/ipfs/go-ipfs/repo/fsrepo"
	yamlConfig "github.com/olebedev/config"
)

// FsRepository represents data a brig repository consists of
type FsRepository struct {

	// Repository is identified by a XMPP Account: name@domain.tld/ressource
	Jid string

	// Minilock ID
	Mid string

	// Folder of repository
	Folder         string
	InternalFolder string

	// UUID which represents a unique repository
	UniqueID string

	// TODO: Just for prototype testing, should be deleted in final version
	Password string

	Config *yamlConfig.Config

	globalRepo *global.GlobalRepository

	Store *store.Store
}

// Interface methods

// Open a encrypted repository
func (r *FsRepository) Lock() error {
	fmt.Println("Opening repository.")
	return nil
}

// Close a open repository
func (r *FsRepository) Unlock() error {
	fmt.Println("Closing repository.")
	return nil
}

// NewFsRepository creates a new repository at filesystem level
// and returns a Repository interface
func NewFsRepository(jid, pass, folder string) (*FsRepository, error) {
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
	minilockID, err := GenerateMinilockID(jid, pass)
	if err != nil {
		return nil, err
	}

	configDefaults := map[string]interface{}{
		"repository.jid":      jid,
		"repository.password": pass,
		"repository.uuid":     uuid.NewRandom().String(),
		"repository.mid":      minilockID,
		"ipfs.path":           filepath.Join(absFolderPath, ".brig", "ipfs"),
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

	return LoadFsRepository(absFolderPath)
}

// CloneFsRepository clones a brig repository in a git like way
func CloneFsRepository() *FsRepository {
	return nil
}

// LoadFsRepository load a brig repository from a given folder.
func LoadFsRepository(folder string) (*FsRepository, error) {
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
		"repository.jid":      "",
		"repository.mid":      "",
		"repository.uuid":     "",
		"repository.password": "",
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

	repo := FsRepository{
		Jid:            configValues["repository.jid"],
		Mid:            configValues["repository.mid"],
		Password:       configValues["repository.password"],
		Folder:         absFolderPath,
		InternalFolder: brigPath,
		UniqueID:       configValues["repository.uuid"],
		Config:         cfg,
		globalRepo:     globalRepo,
		Store:          store,
	}

	return &repo, nil
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

	// TODO: touch() util
	boltDbPath := filepath.Join(brigPath, "index.bolt")
	if fd, err := os.Create(boltDbPath); err != nil {
		return err
	} else {
		fd.Write([]byte(""))
		fd.Close()
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
	// TODO: write to log, not stdout
	cfg, err := ipfsconfig.Init(os.Stdout, 2048)
	if err != nil {
		return err
	}

	if err := fsrepo.Init(ipfsRootPath, cfg); err != nil {
		return err
	}

	return nil
}
