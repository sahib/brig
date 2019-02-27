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

// Init will create a new repository on disk at `baseFolder`.
// `owner` will be the new owner and should be something like user@domain/resource.
// `backendName` is the name of the backend, either "ipfs" or "mock".
// `daemonPort` is the port of the local daemon.
func Init(baseFolder, owner, password, backendName string, daemonPort int64) error {
	// The basefolder has to exist:
	info, err := os.Stat(baseFolder)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(baseFolder, 0700); err != nil {
			return err
		}
	} else if info.Mode().IsDir() {
		children, err := ioutil.ReadDir(baseFolder)
		if err != nil {
			return err
		}

		if len(children) > 0 {
			log.Warningf("`%s` is a directory and exists", baseFolder)
		}
	} else {
		return fmt.Errorf("`%s` is a file (should be a directory)", baseFolder)
	}

	// Create (empty) folders:
	folders := []string{"metadata", "data"}
	for _, folder := range folders {
		absFolder := filepath.Join(baseFolder, folder)
		if err := os.Mkdir(absFolder, 0700); err != nil {
			return e.Wrapf(err, "Failed to create dir: %v (repo exists?)", absFolder)
		}
	}

	if err := touch(filepath.Join(baseFolder, "remotes.yml")); err != nil {
		return e.Wrapf(err, "Failed touch remotes.yml")
	}

	if err := touch(filepath.Join(baseFolder, "INIT_TAG")); err != nil {
		return e.Wrapf(err, "Failed touch INIT_TAG")
	}

	ownerPath := filepath.Join(baseFolder, "OWNER")
	if err := ioutil.WriteFile(ownerPath, []byte(owner), 0644); err != nil {
		return err
	}

	backendNamePath := filepath.Join(baseFolder, "BACKEND")
	if err := ioutil.WriteFile(backendNamePath, []byte(backendName), 0644); err != nil {
		return err
	}

	// For future use: If we ever need to migrate the repo.
	versionPath := filepath.Join(baseFolder, "VERSION")
	if err := ioutil.WriteFile(versionPath, []byte("1"), 0644); err != nil {
		return err
	}

	// Create a default config, only with the default keys applied:
	cfg, err := config.Open(nil, defaults.Defaults, config.StrictnessPanic)
	if err != nil {
		return err
	}

	if err := cfg.SetInt("daemon.port", daemonPort); err != nil {
		return err
	}

	configPath := filepath.Join(baseFolder, "config.yml")
	if err := config.ToYamlFile(configPath, cfg); err != nil {
		return e.Wrap(err, "Failed to setup default config")
	}

	dataFolder := filepath.Join(baseFolder, "data", backendName)
	if err := os.MkdirAll(dataFolder, 0700); err != nil {
		return e.Wrap(err, "Failed to setup dirs for backend")
	}

	// Create initial key pair:
	if err := createKeyPair(owner, baseFolder, 2048); err != nil {
		return e.Wrap(err, "Failed to setup gpg keys")
	}

	passwdFile := filepath.Join(baseFolder, "passwd")
	passwdData := fmt.Sprintf("%s", owner)
	if err := ioutil.WriteFile(passwdFile, []byte(passwdData), 0644); err != nil {
		return err
	}

	// passwd is used to verify the user password,
	// so it needs to be locked only once on init and
	// kept out otherwise from the locking machinery.
	if err := lockFile(passwdFile, keyFromPassword(owner, password)); err != nil {
		return e.Wrapf(err, "passwd-lock")
	}

	return nil
}

// OverwriteConfigKey allows to overwrite a single key/val pair in the config,
// without requiring a running daemon or an opened repository.
// It is not performant and should be use with care.
func OverwriteConfigKey(repoPath string, key string, val interface{}) error {
	configPath := filepath.Join(repoPath, "config.yml")
	cfg, err := defaults.OpenMigratedConfig(configPath)
	if err != nil {
		return e.Wrapf(err, "failed to set ipfs port")
	}

	if err := cfg.Set(key, val); err != nil {
		return err
	}

	fd, err := os.OpenFile(configPath, os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	defer fd.Close()

	return cfg.Save(config.NewYamlEncoder(fd))
}
