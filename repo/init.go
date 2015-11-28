package repo

import (
	"code.google.com/p/go-uuid/uuid"
	"fmt"
	"github.com/disorganizer/brig/repo/config"
	"github.com/disorganizer/brig/repo/global"
	ipfsconfig "github.com/ipfs/go-ipfs/repo/config"
	"github.com/ipfs/go-ipfs/repo/fsrepo"
	yamlConfig "github.com/olebedev/config"
	"os"
	"path"
	"path/filepath"
)

// Repository interface for brig repository types
type Repository interface {
	Open()
	Close()
	Lock()
	Unlock()
}

// FsRepository represents data a brig repository consists of
type FsRepository struct {

	// Repository is identified by a XMPP Account: name@domain.tld/ressource
	Jid string

	// Minilock ID
	Mid string

	// Folder of repository
	Folder string

	// UUID which represents a unique repository
	UniqueID string

	// TODO: Just for prototype testing, should be deleted in final version
	Password string

	Config *yamlConfig.Config

	globalRepo *global.GlobalRepository
}

// Interface methods

// Open a encrypted repository
func (r *FsRepository) Open() {
	fmt.Println("Opening repository.")
}

// Close a open repository
func (r *FsRepository) Close() {
	fmt.Println("Closing repository.")
}

// Lock a repository to be read only
func (r *FsRepository) Lock() {
	fmt.Println("Locking repository.")
}

// Unlock a repository to be writeable
func (r *FsRepository) Unlock() {
	fmt.Println("Unlocking repository.")
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
		"ipfs.path":           path.Join(absFolderPath, ".brig", "ipfs"),
	}

	for key, value := range configDefaults {
		if err = cfg.Set(key, value); err != nil {
			return nil, err
		}
	}

	configPath := path.Join(absFolderPath, ".brig", "config")
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

	cfg, err := config.LoadConfig(path.Join(absFolderPath, ".brig", "config"))
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

	// TODO: Use global repo
	// globalRepo, err := global.New()
	// if err != nil {
	// 	return nil, err
	// }

	repo := FsRepository{
		Jid:      configValues["repository.jid"],
		Mid:      configValues["repository.mid"],
		Password: configValues["repository.password"],
		Folder:   absFolderPath,
		UniqueID: configValues["repository.uuid"],
		Config:   cfg,
		//globalRepo: globalRepo,
	}

	return &repo, nil
}

func createRepositoryTree(absFolderPath string) error {
	if err := os.Mkdir(absFolderPath, 0755); err != nil {
		return err
	}

	brigPath := path.Join(absFolderPath, ".brig")
	if err := os.Mkdir(brigPath, 0755); err != nil {
		return err
	}

	ipfsPath := path.Join(brigPath, "ipfs")
	fmt.Println("IPFS PATH", ipfsPath, brigPath)
	if err := os.Mkdir(ipfsPath, 0755); err != nil {
		return err
	}

	return createIPFS(ipfsPath)
}

func createIPFS(ipfsRootPath string) error {
	cfg, err := ipfsconfig.Init(os.Stdout, 2048)
	if err != nil {
		return err
	}

	if err := fsrepo.Init(ipfsRootPath, cfg); err != nil {
		return err
	}

	return nil
}
