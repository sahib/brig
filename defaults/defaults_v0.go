package defaults

import (
	"github.com/sahib/config"
)

// Defaults is the default validation for brig
var DefaultsV0 = config.DefaultMapping{
	"daemon": config.DefaultMapping{
		"port": config.DefaultEntry{
			Default:      6666,
			NeedsRestart: true,
			Docs:         "Port of the daemon process",
			Validator:    config.IntRangeValidator(1, 655356),
		},
	},
	"fs": config.DefaultMapping{
		"sync": config.DefaultMapping{
			"ignore_removed": config.DefaultEntry{
				Default:      false,
				NeedsRestart: false,
				Docs:         "Do not remove what the remote removed",
			},
			"ignore_moved": config.DefaultEntry{
				Default:      false,
				NeedsRestart: false,
				Docs:         "Do not move what the remote moved",
			},
			"conflict_strategy": config.DefaultEntry{
				Default:      "marker",
				NeedsRestart: false,
				Validator: config.EnumValidator(
					"marker", "ignore",
				),
			},
		},
		"compress": config.DefaultMapping{
			"default_algo": config.DefaultEntry{
				Default:      "snappy",
				NeedsRestart: false,
				Docs:         "What compression algorithm to use by default",
				Validator: config.EnumValidator(
					"snappy", "lz4", "none",
				),
			},
		},
		"autocommit": config.DefaultMapping{
			"enabled": config.DefaultEntry{
				Default:      true,
				NeedsRestart: false,
				Docs:         "Enable the autocommit logic",
			},
			"interval": config.DefaultEntry{
				Default:      "5m",
				NeedsRestart: false,
				Docs:         "In what interval to make automatic commits",
				Validator: config.DurationValidator(),
			},
		},
	},
	"repo": config.DefaultMapping{
		"current_user": config.DefaultEntry{
			Default:      "",
			NeedsRestart: true,
			Docs:         "The repository owner that is published to the outside",
		},
	},
	"data": config.DefaultMapping{
		"ipfs": config.DefaultMapping{
			"path": config.DefaultEntry{
				Default:      "",
				NeedsRestart: true,
				Docs:         "Root directory of the ipfs repository",
			},
		},
	},
}
