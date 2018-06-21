package defaults

import (
	"os"

	e "github.com/pkg/errors"
	"github.com/sahib/brig/config"
)

// CurrentVersion is the current version of brig's config
const CurrentVersion = 0

// Defaults is the default validation for brig
var Defaults = DefaultsV0

// OpenMigratedConfig takes the config.yml at path and loads it.
// If required, it also migrates the config structure to the newest
// version - brig can always rely on the latest config keys to be present.
func OpenMigratedConfig(path string) (*config.Config, error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, e.Wrap(err, "failed to open config")
	}

	defer fd.Close()

	// Add there any migrations with mgr.Add if needed.
	mgr := config.NewMigrater(CurrentVersion)
	mgr.Add(0, nil, DefaultsV0)

	return mgr.Migrate(config.NewYamlDecoder(fd))
}
