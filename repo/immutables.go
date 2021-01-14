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

// TODO: rename this file to immutables.go

type Immutables struct {
	cfg *config.Config
}

func NewImmutables(path string) (*Immutables, error) {
	cfg, err := openMigratedImmutables(path)
	if err != nil {
		return nil, err
	}

	return &Immutables{cfg: cfg}, nil
}

func (i *Immutables) Owner() string {
	return i.cfg.String("owner")
}

func (i *Immutables) Backend() string {
	return i.cfg.String("backend")
}
