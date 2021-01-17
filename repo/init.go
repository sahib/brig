package repo

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	e "github.com/pkg/errors"
	"github.com/sahib/brig/defaults"
	"github.com/sahib/config"
	log "github.com/sirupsen/logrus"
)

func touch(path string) error {
	fd, err := os.OpenFile(path, os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	return fd.Close()
}

// InitOptions sum up the option that we can pass to Init()
type InitOptions struct {
	// BaseFolder is where the repository is located.
	BaseFolder string

	// Owner is the owner id of the repository.
	Owner string

	// BackendName says what backend we should use.
	BackendName string

	// DaemonURL is the URL that will be used for the brig daemon.
	DaemonURL string
}

// IsValidBackendName tells you if `name` is a valid backend name.
func IsValidBackendName(name string) bool {
	switch name {
	case "mock", "httpipfs":
		return true
	default:
		return false
	}
}

// Validate checks if the options are valid.
func (opts InitOptions) Validate() error {
	if !IsValidBackendName(opts.BackendName) {
		return fmt.Errorf("invalid backend name: %v", opts.BackendName)
	}

	if len(opts.Owner) == 0 {
		return fmt.Errorf("owner may not be empty")
	}

	return nil
}

// Init will create a new repository on disk at `baseFolder`.
// `owner` will be the new owner and should be something like user@domain/resource.
// `backendName` is the name of the backend, either "ipfs" or "mock".
// `daemonPort` is the port of the local daemon.
func Init(opts InitOptions) error {
	if err := opts.Validate(); err != nil {
		return err
	}

	info, err := os.Stat(opts.BaseFolder)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(opts.BaseFolder, 0700); err != nil {
			return err
		}
	} else if info.Mode().IsDir() {
		children, err := ioutil.ReadDir(opts.BaseFolder)
		if err != nil {
			return err
		}

		if len(children) > 0 {
			log.Warningf("`%s` is a directory and exists", opts.BaseFolder)
		}
	} else {
		return fmt.Errorf("`%s` is not a directory", opts.BaseFolder)
	}

	// Create (empty) folders:
	for _, emptyFolder := range []string{"metadata", "keyring"} {
		absFolder := filepath.Join(opts.BaseFolder, emptyFolder)
		if err := os.Mkdir(absFolder, 0700); err != nil {
			return e.Wrapf(err, "failed to create dir: %v (repo exists?)", absFolder)
		}
	}

	if err := touch(filepath.Join(opts.BaseFolder, "remotes.yml")); err != nil {
		return e.Wrapf(err, "failed to touch remotes.yml")
	}

	err = ioutil.WriteFile(
		filepath.Join(opts.BaseFolder, "README.md"),
		[]byte(repoReadmeTxt),
		0600,
	)
	if err != nil {
		return e.Wrap(err, "failed to write README.md")
	}

	immutables, err := config.Open(nil, immutableDefaultsV0, config.StrictnessPanic)
	if err != nil {
		return err
	}

	if err := immutables.SetString("owner", opts.Owner); err != nil {
		return err
	}

	if err := immutables.SetString("backend", opts.BackendName); err != nil {
		return err
	}

	immutablePath := filepath.Join(opts.BaseFolder, "immutable.yml")
	if err := config.ToYamlFile(immutablePath, immutables); err != nil {
		return e.Wrap(err, "failed to setup immutables config")
	}

	// Create a default config, only with the default keys applied:
	cfg, err := config.Open(nil, defaults.Defaults, config.StrictnessPanic)
	if err != nil {
		return err
	}

	if err := cfg.SetString("daemon.url", opts.DaemonURL); err != nil {
		return err
	}

	configPath := filepath.Join(opts.BaseFolder, "config.yml")
	if err := config.ToYamlFile(configPath, cfg); err != nil {
		return e.Wrap(err, "failed to setup default config")
	}

	// Create initial key pair:
	keyringFolder := filepath.Join(opts.BaseFolder, "keyring")
	if err := createKeyPair(opts.Owner, keyringFolder, 2048); err != nil {
		return e.Wrap(err, "failed to setup gpg keys")
	}

	return nil
}
