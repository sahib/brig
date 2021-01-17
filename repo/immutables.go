package repo

import (
	"os"

	e "github.com/pkg/errors"
	"github.com/sahib/config"
)

const (
	currentImmutablesVersion = 0
)

var immutableDefaultsV0 = config.DefaultMapping{
	"backend": config.DefaultEntry{
		Default:      "httpipfs",
		NeedsRestart: true,
		Docs:         "What backend type this repository uses",
		Validator: config.EnumValidator(
			"httpipfs",
			"mock",
		),
	},
	"owner": config.DefaultEntry{
		Default:      "",
		NeedsRestart: true,
		Docs:         "The owner of this repository passed at init",
	},
	"init_tag": config.DefaultEntry{
		Default:      "",
		NeedsRestart: true,
		Docs:         "Hash of the first commit",
	},
	"version": config.DefaultEntry{
		Default:      currentImmutablesVersion,
		NeedsRestart: true,
		Docs:         "Layout version of this repository",
	},
}

func openMigratedImmutables(path string) (*config.Config, error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, e.Wrapf(err, "failed to open config path %s", path)
	}

	defer fd.Close()

	mgr := config.NewMigrater(currentImmutablesVersion, config.StrictnessPanic)
	mgr.Add(0, nil, immutableDefaultsV0)

	cfg, err := mgr.Migrate(config.NewYamlDecoder(fd))
	if err != nil {
		return nil, e.Wrap(err, "failed to migrate or open")
	}

	return cfg, nil
}

// Immutables gives access to different values that can not be changed
// by the user and were determined during the init of the repository.
type Immutables struct {
	cfg *config.Config
}

// NewImmutables loads the immutable.yml at `path`
func NewImmutables(path string) (*Immutables, error) {
	cfg, err := openMigratedImmutables(path)
	if err != nil {
		return nil, err
	}

	return &Immutables{cfg: cfg}, nil
}

// Owner returns the owner of the repository.
func (i *Immutables) Owner() string {
	return i.cfg.String("owner")
}

// Backend returns the chosen backend of the repository.
func (i *Immutables) Backend() string {
	return i.cfg.String("backend")
}
