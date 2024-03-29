package defaults

import (
	"errors"
	"net/url"
	"runtime"

	"github.com/sahib/config"
)

// DaemonDefaultURL returns the default URL for the current OS.
func DaemonDefaultURL() string {
	// If the platform supports unix sockets,
	// we should make use of it.
	switch runtime.GOOS {
	case "linux", "darwin":
		// See "man 7 unix" - we use an abstract unix domain socket.
		// This means there is no socket file on the file system.
		// (other tools use unix:@/path, but Go does not support that notation)
		// This also means that there are no user rights on the socket file.
		// If you need this, specify the url in the config.
		return "unix:/tmp/brig.socket?abstract=true"
	default:
		return "tcp://127.0.0.1:6666"
	}
}

func urlValidator(val interface{}) error {
	s, ok := val.(string)
	if !ok {
		return errors.New("url is not an string")
	}

	_, err := url.Parse(s)
	return err
}

// DefaultsV0 is the default config validation for brig
var DefaultsV0 = config.DefaultMapping{
	"daemon": config.DefaultMapping{
		"url": config.DefaultEntry{
			Default:      DaemonDefaultURL(),
			NeedsRestart: true,
			Docs:         "URL of the daemon process.",
			Validator:    urlValidator,
		},
		"ipfs_path_or_url": config.DefaultEntry{
			Default:      "",
			NeedsRestart: true,
			Docs:         "URL or path to the IPFS repository you want to use.",
		},
		"enable_pprof": config.DefaultEntry{
			Default:      true,
			NeedsRestart: true,
			Docs:         "Enable a ppropf profile server on startup (see »brig d p --help«)",
		},
	},
	"events": config.DefaultMapping{
		"enabled": config.DefaultEntry{
			Default:      true,
			NeedsRestart: false,
			Docs:         "Wether we should handle incoming events and publish auto update events.",
		},
		"recv_interval": config.DefaultEntry{
			Default:      "100ms",
			NeedsRestart: false,
			Docs:         "Time window in which events are buffered before handling them.",
		},
		"recv_max_events_per_second": config.DefaultEntry{
			Default:      0.5,
			NeedsRestart: false,
			Docs:         "How many incoming events per second to process at max.",
		},
		"send_interval": config.DefaultEntry{
			Default:      "200ms",
			NeedsRestart: false,
			Docs:         "Time window in which events are buffered before sending them.",
		},
		"send_max_events_per_second": config.DefaultEntry{
			Default:      5.0,
			NeedsRestart: false,
			Docs:         "How many outgoing events per second to send out at max",
		},
	},
	"gateway": config.DefaultMapping{
		"enabled": config.DefaultEntry{
			Default:      false,
			NeedsRestart: false,
			Docs:         "Wether the gateway should be running. Will start when enabled.",
		},
		"port": config.DefaultEntry{
			Default:      6001,
			NeedsRestart: false,
			Docs:         "On what port the gateway runs on.",
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
				Docs:         "Enable debug mode (load resources from filesystem).",
			},
		},
		"auth": config.DefaultMapping{
			"anon_allowed": config.DefaultEntry{
				Default:      false,
				NeedsRestart: false,
				Docs:         "Wether a login is required.",
			},
			"anon_user": config.DefaultEntry{
				Default:      "anon",
				NeedsRestart: false,
				Docs:         "What user to copy settings (folder, rights etc.) from.",
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
				Docs:         "Do not remove what the remote removed.",
			},
			"ignore_moved": config.DefaultEntry{
				Default:      false,
				NeedsRestart: false,
				Docs:         "Do not move what the remote moved",
			},
			"pin_added": config.DefaultEntry{
				Default:      false,
				NeedsRestart: false,
				Docs:         "Do not pin files which were added at the remote",
			},
			"conflict_strategy": config.DefaultEntry{
				Default:      "marker",
				NeedsRestart: false,
				Validator: config.EnumValidator(
					"marker", "ignore", "embrace",
				),
				Docs: `What strategy to apply in case of conflicts:

  * marker: Create a conflict file with the remote's version.
  * ignore: Ignore the remote version completely and keep our version.
  * embrace: Take the remote version and replace ours with it.
`,
			},
		},
		"pre_cache": config.DefaultMapping{
			"enabled": config.DefaultEntry{
				Default:      false,
				NeedsRestart: false,
				Docs:         "pre-cache files up-on pinning.",
			},
		},
		"pagecache": config.DefaultMapping{
			"max_memory": config.DefaultEntry{
				Default:      "1G",
				NeedsRestart: true,
				Docs:         "Consume at max this amount of memory for the pagecache",
			},
			"l2compress": config.DefaultEntry{
				Default:      true,
				NeedsRestart: true,
				Docs:         "Compress swapped pages over max_memory before going to disk",
			},
		},
		"repin": config.DefaultMapping{
			"enabled": config.DefaultEntry{
				Default:      true,
				NeedsRestart: false,
				Docs:         "Perform repinning to reclaim space (see »brig pin repin --help«)",
			},
			"interval": config.DefaultEntry{
				Default:      "15m",
				NeedsRestart: false,
				Docs:         "In what time interval to trigger repinning automatically.",
				Validator:    config.DurationValidator(),
			},
			"quota": config.DefaultEntry{
				Default:      "5GB",
				NeedsRestart: false,
				Docs: `Maximum stored amount of pinned files to have.

  If the quota limit is hit, old versions of a file are unpinned first on the
  next repin. Biggest file first.
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
			"pin_unpinned": config.DefaultEntry{
				Default:      false,
				NeedsRestart: false,
				Docs: `Pin unpinned files:

  * 'true'  if you want maximum permitted mirroring
  * 'false' if you want to save traffic

  If a file version »n« is such that (min_depth <= »n« < max_depth),
  then the repinner will pin such version if pin_unpinned is set to true.
  Otherwise, it will keep the file unpinned, i.e. not cached at the backend.
`,
			},
		},
		"autocommit": config.DefaultMapping{
			"enabled": config.DefaultEntry{
				Default:      true,
				NeedsRestart: false,
				Docs:         "Wether to make automatic commits in a fixed interval.",
			},
			"interval": config.DefaultEntry{
				Default:      "5m",
				NeedsRestart: false,
				Docs:         "In what interval to make automatic commits.",
				Validator:    config.DurationValidator(),
			},
		},
	},
	"repo": config.DefaultMapping{
		"current_user": config.DefaultEntry{
			Default:      "",
			NeedsRestart: false,
			Docs:         "The repository owner that is published to the outside.",
		},
		"autogc": config.DefaultMapping{
			"enabled": config.DefaultEntry{
				Default:      true,
				NeedsRestart: false,
				Docs:         "Wether to make automatic commits in a fixed interval.",
			},
			"interval": config.DefaultEntry{
				Default:      "60m",
				NeedsRestart: false,
				Docs:         "In what interval to make automatic commits.",
				Validator:    config.DurationValidator(),
			},
		},
	},
	"mounts": config.DefaultMapping{
		// This key stands for the fstab name entry:
		"__many__": config.DefaultMapping{
			"path": config.DefaultEntry{
				Default:      "",
				NeedsRestart: true,
				Docs:         "The place where the mount path can be found.",
			},
			"read_only": config.DefaultEntry{
				Default:      false,
				NeedsRestart: true,
				Docs:         "Wether this mount should be done read-only.",
			},
			"offline": config.DefaultEntry{
				Default:      false,
				NeedsRestart: true,
				Docs:         "Error out on remote files early if set true.",
			},
			"root": config.DefaultEntry{
				Default:      "/",
				NeedsRestart: true,
				Docs:         "The virtual root of the mount.",
			},
		},
	},
}
