package repo

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/sahib/config"
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
				Docs:         "Owner of the repository",
			},
			"addr": config.DefaultEntry{
				Default:      "",
				NeedsRestart: true,
				Docs:         "Backend Address of this repository",
			},
			"local_port": config.DefaultEntry{
				Default:      6666,
				NeedsRestart: true,
				Docs:         "The port of the brigd service",
			},
			"path": config.DefaultEntry{
				Default:      "",
				NeedsRestart: true,
				Docs:         "Path to the repository",
			},
			"is_default": config.DefaultEntry{
				Default:      false,
				NeedsRestart: true,
				Docs:         "Use this repo as default?",
			},
		},
	},
}

type Registry struct {
	mu  sync.Mutex
	cfg *config.Config
}

type RegistryEntry struct {
	Path      string
	Owner     string
	Addr      string
	Port      int64
	IsDefault bool
}

var (
	ErrRegistryEntryExists = errors.New("registry entry exists already")
)

func findRegistryPath() string {
	var registryPaths []string
	if path := os.Getenv("BRIG_REGISTRY_PATH"); path != "" {
		registryPaths = []string{path}
	} else {
		home := ""
		user, err := user.Current()
		if err != nil {
			home = os.Getenv("HOME")
		} else {
			home = user.HomeDir
		}

		registryPaths = []string{
			fmt.Sprintf("%s/.config/brig/registry.yml", home),
			fmt.Sprintf("%s/.brig-registry.yml", home),
			"/etc/brig-registry.yml",
		}
	}

	for _, path := range registryPaths {
		if _, err := os.Stat(path); err != nil {
			// Ignore any kind of errors, including
			// bad permissions or broken filesystems.
			continue
		}

		// This path seems okay.
		return path
	}

	// Nothing suitable found. Use the most preferred one.
	return registryPaths[0]
}

func OpenRegistry() (*Registry, error) {
	registryPath := findRegistryPath()
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

	uuidString := entryUUID.String()

	// Check this unlikely case:
	if existingEntry, _ := reg.entry(uuidString); existingEntry != nil {
		return "", ErrRegistryEntryExists
	}

	entries, err := reg.list()
	if err != nil {
		return "", err
	}

	entry.IsDefault = len(entries) == 0
	if err := reg.update(uuidString, entry); err != nil {
		return "", err
	}

	return entryUUID.String(), nil
}

func (reg *Registry) Update(uuid string, entry *RegistryEntry) error {
	reg.mu.Lock()
	defer reg.mu.Unlock()

	return reg.update(uuid, entry)
}

func (reg *Registry) update(uuid string, entry *RegistryEntry) error {
	ownerKey := fmt.Sprintf("repos.%s.owner", uuid)
	if err := reg.cfg.SetString(ownerKey, entry.Owner); err != nil {
		return err
	}

	pathKey := fmt.Sprintf("repos.%s.path", uuid)
	if err := reg.cfg.SetString(pathKey, entry.Path); err != nil {
		return err
	}

	isDefaultKey := fmt.Sprintf("repos.%s.is_default", uuid)
	if err := reg.cfg.SetBool(isDefaultKey, entry.IsDefault); err != nil {
		return err
	}

	addrKey := fmt.Sprintf("repos.%s.addr", uuid)
	if err := reg.cfg.SetString(addrKey, entry.Addr); err != nil {
		return err
	}

	if entry.Port != 0 {
		portKey := fmt.Sprintf("repos.%s.local_port", uuid)
		if err := reg.cfg.SetInt(portKey, entry.Port); err != nil {
			return err
		}
	}

	registryPath := findRegistryPath()
	if err := os.MkdirAll(filepath.Dir(registryPath), 0700); err != nil {
		return err
	}

	registryFd, err := os.OpenFile(registryPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	return reg.cfg.Save(config.NewYamlEncoder(registryFd))
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

	isDefaultKey := fmt.Sprintf("repos.%s.is_default", uuid)
	ownerKey := fmt.Sprintf("repos.%s.owner", uuid)
	addrKey := fmt.Sprintf("repos.%s.addr", uuid)
	portKey := fmt.Sprintf("repos.%s.local_port", uuid)

	isDefault := reg.cfg.Bool(isDefaultKey)
	owner := reg.cfg.String(ownerKey)
	addr := reg.cfg.String(addrKey)
	port := reg.cfg.Int(portKey)

	return &RegistryEntry{
		Path:      path,
		Owner:     owner,
		Addr:      addr,
		Port:      port,
		IsDefault: isDefault,
	}, nil
}

func (reg *Registry) List() ([]*RegistryEntry, error) {
	reg.mu.Lock()
	defer reg.mu.Unlock()

	return reg.list()
}

func (reg *Registry) list() ([]*RegistryEntry, error) {
	entries := []*RegistryEntry{}

	for _, key := range reg.cfg.Keys() {
		if !strings.HasSuffix(key, ".path") {
			continue
		}

		split := strings.Split(key, ".")
		if len(split) < 3 {
			return nil, fmt.Errorf("broken key in global registry: %s", key)
		}

		uuid := split[1]

		entry, err := reg.entry(uuid)
		if err != nil {
			return nil, err
		}

		entries = append(entries, entry)
	}

	return entries, nil
}
