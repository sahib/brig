package repo

import (
	"crypto/rand"
	"io"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/repo/config"
	"github.com/disorganizer/brig/repo/global"
	"github.com/disorganizer/brig/util"
	logutil "github.com/disorganizer/brig/util/log"
	ipfsconfig "github.com/ipfs/go-ipfs/repo/config"
	"github.com/ipfs/go-ipfs/repo/fsrepo"
	yamlConfig "github.com/olebedev/config"
	"github.com/wayn3h0/go-uuid"
)

func initIntoGlobal(folder string, cfg *yamlConfig.Config) error {
	globalRepo, err := global.New()
	if err != nil {
		return err
	}

	ipfsAPIPort, err := globalRepo.NextIPFSAPIPort()
	if err != nil {
		return err
	}

	ipfsSwarmPort, err := globalRepo.NextIPFSSwarmPort()
	if err != nil {
		return err
	}

	daemonPort, err := globalRepo.NextDaemonPort()
	if err != nil {
		return err
	}

	cfg.Set("ipfs.apiport", ipfsAPIPort)
	cfg.Set("ipfs.swarmport", ipfsSwarmPort)
	cfg.Set("daemon.port", daemonPort)

	log.Debugf(
		"Using ports: %d (ipfs-api) %d (ipfs-swarm) %d (brigd)",
		ipfsAPIPort, ipfsSwarmPort, daemonPort,
	)

	uuid, err := cfg.String("repository.uuid")
	if err != nil {
		return err
	}

	err = globalRepo.AddRepo(global.RepoListEntry{
		UniqueID:      uuid,
		RepoPath:      folder,
		DaemonPort:    daemonPort,
		IpfsAPIPort:   ipfsAPIPort,
		IpfsSwarmPort: ipfsSwarmPort,
	})

	if err != nil {
		return err
	}

	return nil
}

// NewRepository creates a new repository at filesystem level
// and returns a Repository interface
func NewRepository(ID, pwd, folder string) (*Repository, error) {
	absFolderPath, err := filepath.Abs(folder)
	if err != nil {
		return nil, err
	}

	if _, err = os.Stat(absFolderPath); os.IsExist(err) {
		return nil, err
	}

	if err := createRepositoryTree(absFolderPath); err != nil {
		return nil, err
	}

	brigPath := filepath.Join(absFolderPath, ".brig")
	if err := createShadowFile(brigPath, ID, pwd); err != nil {
		return nil, err
	}

	cfg := config.CreateDefaultConfig()

	repoUUID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	configDefaults := map[string]interface{}{
		"repository.ID":   ID,
		"repository.uuid": repoUUID.String(),
		"ipfs.path":       filepath.Join(brigPath, "ipfs"),
	}

	for key, value := range configDefaults {
		if err = cfg.Set(key, value); err != nil {
			return nil, err
		}
	}

	if err := initIntoGlobal(brigPath, cfg); err != nil {
		return nil, err
	}

	configPath := filepath.Join(brigPath, "config")
	if _, err := config.SaveConfig(configPath, cfg); err != nil {
		return nil, err
	}

	return loadRepository(pwd, absFolderPath)
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

	empties := []string{"otr.key", "otr.buddies", "shadow"}
	for _, empty := range empties {
		fullPath := filepath.Join(brigPath, empty)
		if err := util.Touch(fullPath); err != nil {
			return err
		}
	}

	// Make the key larger than needed:
	if err := createMasterKey(brigPath, 1024); err != nil {
		return err
	}

	return CreateIpfsRepo(ipfsPath)
}

func createMasterKey(brigPath string, keySize int) error {
	keyPath := filepath.Join(brigPath, "master.key")
	fd, err := os.OpenFile(keyPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}

	defer util.Closer(fd)

	if _, err := io.CopyN(fd, rand.Reader, int64(keySize/8)); err != nil {
		return err
	}

	return nil
}

// CreateIpfsRepo initializes an empty .ipfs directory at `ipfsRootPath`.
// ipfsRootPath should contain the ".ipfs" at the end.
func CreateIpfsRepo(ipfsRootPath string) error {
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
