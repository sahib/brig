package defaults

import (
	"github.com/sahib/config"
)

// DefaultsV0 is the default config validation for brig
var DefaultsV0 = config.DefaultMapping{
	"daemon": config.DefaultMapping{
		"port": config.DefaultEntry{
			Default:      6666,
			NeedsRestart: true,
			Docs:         "Port of the daemon process",
			Validator:    config.IntRangeValidator(1, 655356),
		},
		"ipfs_port": config.DefaultEntry{
			Default:      5001,
			NeedsRestart: true,
			Docs:         "Port of IPFS' HTTP API",
			Validator:    config.IntRangeValidator(1, 655356),
		},
		"enable_pprof": config.DefaultEntry{
			Default:      true,
			NeedsRestart: true,
			Docs:         "enable a ppropf profile server on startup",
		},
	},
	"events": config.DefaultMapping{
		"enabled": config.DefaultEntry{
			Default:      true,
			NeedsRestart: false,
			Docs:         "Wether we should handle incoming events and publish events",
		},
		"recv_interval": config.DefaultEntry{
			Default:      "100ms",
			NeedsRestart: false,
			Docs:         "Time window in which events are buffered before handling them",
		},
		"recv_max_events_per_second": config.DefaultEntry{
			Default:      0.5,
			NeedsRestart: false,
			Docs:         "How many events per second to process at max",
		},
		"send_interval": config.DefaultEntry{
			Default:      "200ms",
			NeedsRestart: false,
			Docs:         "Time window in which events are buffered before sending them",
		},
		"send_max_events_per_second": config.DefaultEntry{
			Default:      5.0,
			NeedsRestart: false,
			Docs:         "How many events per second to send out at max",
		},
	},
	"gateway": config.DefaultMapping{
		"enabled": config.DefaultEntry{
			Default:      false,
			NeedsRestart: false,
			Docs:         "Wether the gateway should be running",
		},
		"port": config.DefaultEntry{
			Default:      6001,
			NeedsRestart: false,
			Docs:         "On what port the gateway runs on",
		},
		"ui": config.DefaultMapping{
			"enabled": config.DefaultEntry{
				Default:      true,
				NeedsRestart: false,
				Docs:         "Enable the UI. This does not affect the /get endpoint.",
			},
			"debug_mode": config.DefaultEntry{
				Default:      false,
				NeedsRestart: false,
				Docs:         "Enable debug mode (load resources from filesystem)",
			},
		},
		"cert": config.DefaultMapping{
			"certfile": config.DefaultEntry{
				Default:      "",
				NeedsRestart: false,
				Docs:         "Path to an existing certificate file",
			},
			"keyfile": config.DefaultEntry{
				Default:      "",
				NeedsRestart: false,
				Docs:         "Path to an existing key file",
			},
			"domain": config.DefaultEntry{
				Default:      "",
				NeedsRestart: false,
				Docs:         "What domain to use for getting a certificate from LetsEncrypt",
			},
			"redirect": config.DefaultMapping{
				"enabled": config.DefaultEntry{
					Default:      true,
					NeedsRestart: false,
					Docs:         "Wether http request should be forwarded to https",
				},
				"http_port": config.DefaultEntry{
					Default:      5001,
					NeedsRestart: false,
					Docs:         "What port the http redirect server should run on",
				},
			},
		},
		"auth": config.DefaultMapping{
			"enabled": config.DefaultEntry{
				Default:      true,
				NeedsRestart: false,
				Docs:         "Wether the gateway should be running",
			},
			"session-encryption-key": config.DefaultEntry{
				Default:      "",
				NeedsRestart: true,
				Docs:         "Encryption key for session cookies. Generated when left empty.",
			},
			"session-authentication-key": config.DefaultEntry{
				Default:      "",
				NeedsRestart: true,
				Docs:         "Authentication key for session cookies. Generated when left empty.",
			},
			"session-csrf-key": config.DefaultEntry{
				Default:      "",
				NeedsRestart: true,
				Docs:         "Key used for CSRF protection. Generated if empty.",
			},
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
		"pre_cache": config.DefaultMapping{
			"enabled": config.DefaultEntry{
				Default:      false,
				NeedsRestart: false,
				Docs:         "pre-cache pinned files",
			},
		},
		"repin": config.DefaultMapping{
			"enabled": config.DefaultEntry{
				Default:      true,
				NeedsRestart: false,
				Docs:         "Perform repinning on old versions",
			},
			"interval": config.DefaultEntry{
				Default:      "15m",
				NeedsRestart: false,
				Docs:         "In what time interval to trigger repinning",
				Validator:    config.DurationValidator(),
			},
			"quota": config.DefaultEntry{
				Default:      "1GB",
				NeedsRestart: false,
				Docs: `Maximum stored amount of pinned files to have.

This quota is always enforced on commits. When the pinned storage exceeds the quota limit,
old versions of files are unpinned first.
`,
			},
			"min_depth": config.DefaultEntry{
				Default:      1,
				NeedsRestart: false,
				Docs:         `Keep at least »n« versions of a pinned file, even if this would exceed the quota.`,
			},
			"max_depth": config.DefaultEntry{
				Default:      10,
				NeedsRestart: false,
				Docs:         `Keep at max »n« versions of a pinned file and remove it even if it does not exceed quota.`,
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
				Validator:    config.DurationValidator(),
			},
		},
	},
	"repo": config.DefaultMapping{
		"current_user": config.DefaultEntry{
			Default:      "",
			NeedsRestart: true,
			Docs:         "The repository owner that is published to the outside",
		},
		"password_command": config.DefaultEntry{
			Default:      "",
			NeedsRestart: false,
			Docs:         "If set, the repo password is taken from stdout of this command",
		},
	},
	"mounts": config.DefaultMapping{
		"__many__": config.DefaultMapping{
			"path": config.DefaultEntry{
				Default:      "",
				NeedsRestart: true,
				Docs:         "The place where the mount path can be found",
			},
			"read_only": config.DefaultEntry{
				Default:      false,
				NeedsRestart: true,
				Docs:         "Wether this mount should be done read-only",
			},
			"root": config.DefaultEntry{
				Default:      "/",
				NeedsRestart: true,
				Docs:         "The root of the mount",
			},
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
