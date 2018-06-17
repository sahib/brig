package config

// Defaults is the default validation for brig
var Defaults = DefaultMapping{
	"daemon": DefaultMapping{
		"port": DefaultEntry{
			Default:      6666,
			NeedsRestart: true,
			Docs:         "Port of the daemon process",
			Validator:    IntRangeValidator(1, 655356),
		},
	},
	"fs": DefaultMapping{
		"sync": DefaultMapping{
			"ignore_removed": DefaultEntry{
				Default:      false,
				NeedsRestart: false,
				Docs:         "Do not remove what the remote removed",
			},
			"ignore_moved": DefaultEntry{
				Default:      false,
				NeedsRestart: false,
				Docs:         "Do not move what the remote moved",
			},
			"conflict_strategy": DefaultEntry{
				Default:      "marker",
				NeedsRestart: false,
				Validator: EnumValidator(
					"marker", "ignore",
				),
			},
		},
		"compress": DefaultMapping{
			"default_algo": DefaultEntry{
				Default:      "snappy",
				NeedsRestart: false,
				Docs:         "What compression algorithm to use by default",
				Validator: EnumValidator(
					"snappy", "lz4", "none",
				),
			},
		},
	},
	"repo": DefaultMapping{
		"current_user": DefaultEntry{
			Default:      "",
			NeedsRestart: true,
			Docs:         "The repository owner that is published to the outside",
		},
	},
	"data": DefaultMapping{
		"ipfs": DefaultMapping{
			"path": DefaultEntry{
				Default:      "",
				NeedsRestart: true,
				Docs:         "Root directory of the ipfs repository",
			},
		},
	},
}
