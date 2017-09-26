package catfs

import (
	"fmt"

	"github.com/disorganizer/brig/catfs/mio/compress"
	"github.com/disorganizer/brig/catfs/vcs"
)

// Config can be used to control specific bevhaviours of the filesystem implementation.
// It's designed to be a human readable configuration, that mwill be parsed when
// instancing the filesystem.
type Config struct {
	IO struct {
		CompressAlgo string
	}

	// Special options for the sync algorithm:
	Sync struct {
		IgnoreRemoved    bool
		ConflictStrategy string
	}
}

// DefaultConfig is a Config with sane default values
var DefaultConfig = &Config{}

func init() {
	DefaultConfig.IO.CompressAlgo = "snappy"
}

type config struct {
	sync         vcs.SyncConfig
	compressAlgo compress.AlgorithmType
}

func (cfg *Config) parseConfig() (*config, error) {
	if cfg == nil {
		cfg = DefaultConfig
	}

	vfg := &config{}

	// TODO: From String is a bad name.
	algo, err := compress.AlgoFromString(cfg.IO.CompressAlgo)
	if err != nil {
		return nil, err
	}

	cs := vcs.ConflictStrategyFromString(cfg.Sync.ConflictStrategy)
	if cs == vcs.ConflictStragetyUnknown {
		return nil, fmt.Errorf("Bad conflic strategy: %v", cfg.Sync.ConflictStrategy)
	}

	vfg.compressAlgo = algo
	vfg.sync.IgnoreDeletes = cfg.Sync.IgnoreRemoved
	vfg.sync.ConflictStrategy = cs
	return vfg, nil
}
