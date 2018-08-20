package repo

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/sahib/config"
	"github.com/satori/go.uuid"
)

const (
	currentVersion = 0
)

// Defaults is the default validation for brig
var defaultsV0 = config.DefaultMapping{
	"repos": config.DefaultMapping{
		"__many__": config.DefaultMapping{
			"owner": config.DefaultEntry{
				Default:      "",
				NeedsRestart: true,
				Docs:         "",
			},
			"path": config.DefaultEntry{
				Default:      "",
				NeedsRestart: true,
				Docs:         "",
			},
			"password": config.DefaultEntry{
				Default:      "",
				NeedsRestart: true,
				Docs:         "",
			},
		},
	},
}

type Registry struct {
	mu  sync.Mutex
	cfg *config.Config
}

type RegistryEntry struct {
	Path     string
	Owner    string
	Password string
}

var (
	RegistryPaths = []string{
		"$HOME/.config/brig/registry.yml",
		"$HOME/.brig-registry.yml",
		"/etc/brig-registry.yml",
	}
	ErrRegistryEntryExists = errors.New("registry entry exists already")
)

func findRegistryPath() string {
	for _, path := range RegistryPaths {
		fullPath := os.ExpandEnv(path)
		if _, err := os.Stat(fullPath); err != nil {
			// Ignore any kind of errors, including
			// bad permissions or broken filesystems.
			continue
		}

		// This path seems okay.
		return fullPath
	}

	// Nothing suitable found. Use the most preferred one.
	return os.ExpandEnv(RegistryPaths[0])
}

func OpenRegistry() (*Registry, error) {
	registryPath := findRegistryPath()
	log.Debugf("using registry path `%s`", registryPath)
	registryFd, err := os.OpenFile(registryPath, os.O_RDONLY, 0600)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	var dec config.Decoder
	if err == nil {
		defer registryFd.Close()
		dec = config.NewYamlDecoder(registryFd)
	}

	// Add the file structure in a way that is safe for migration:
	mgr := config.NewMigrater(currentVersion, config.StrictnessPanic)
	mgr.Add(0, nil, defaultsV0)
	cfg, err := mgr.Migrate(dec)

	if err != nil {
		return nil, err
	}

	return &Registry{
		cfg: cfg,
	}, nil
}

func (reg *Registry) Add(entry *RegistryEntry) (string, error) {
	reg.mu.Lock()
	defer reg.mu.Unlock()

	entryUUID, err := uuid.NewV4()
	if err != nil {
		return "", err
	}

	entryName := entryUUID.String()

	// Check this unlikely case:
	if existingEntry, _ := reg.entry(entryName); existingEntry != nil {
		return "", ErrRegistryEntryExists
	}

	ownerKey := fmt.Sprintf("repos.%s.owner", entryName)
	if err := reg.cfg.SetString(ownerKey, entry.Owner); err != nil {
		return "", err
	}

	passwordKey := fmt.Sprintf("repos.%s.password", entryName)
	if err := reg.cfg.SetString(passwordKey, entry.Password); err != nil {
		return "", err
	}

	pathKey := fmt.Sprintf("repos.%s.path", entryName)
	if err := reg.cfg.SetString(pathKey, entry.Path); err != nil {
		return "", err
	}

	registryPath := findRegistryPath()
	if err := os.MkdirAll(filepath.Dir(registryPath), 0700); err != nil {
		return "", err
	}

	registryFd, err := os.OpenFile(registryPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return "", err
	}

	return entryName, reg.cfg.Save(config.NewYamlEncoder(registryFd))
}

func (reg *Registry) Entry(uuid string) (*RegistryEntry, error) {
	reg.mu.Lock()
	defer reg.mu.Unlock()

	return reg.entry(uuid)
}

func (reg *Registry) entry(uuid string) (*RegistryEntry, error) {
	if len(uuid) == 0 {
		return nil, fmt.Errorf("empty uuid")
	}

	if strings.Contains(uuid, ".") {
		return nil, fmt.Errorf("uuid should not contain dots: %s", uuid)
	}

	pathKey := fmt.Sprintf("repos.%s.path", uuid)
	path := reg.cfg.String(pathKey)
	if path == "" {
		// "" is the default value.
		return nil, fmt.Errorf("no entry for uuid `%s`", uuid)
	}

	ownerKey := fmt.Sprintf("repos.%s.owner", uuid)
	passwordKey := fmt.Sprintf("repos.%s.password", uuid)

	owner := reg.cfg.String(ownerKey)
	password := reg.cfg.String(passwordKey)

	return &RegistryEntry{
		Path:     path,
		Owner:    owner,
		Password: password,
	}, nil
}
