package repo

import (
	"os"
	"path/filepath"

	e "github.com/pkg/errors"
	"github.com/sahib/brig/defaults"
	"github.com/sahib/config"
)

// Config utilities.

// OverwriteConfigKey allows to overwrite a single key/val pair in the config,
// without requiring a running daemon or an opened repository. It is not fast
// and should be only used for one-off commands. For long running commands
// you should open the repository.
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
