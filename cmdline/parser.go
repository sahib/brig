package cmdline

import (
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	colorlog "github.com/disorganizer/brig/util/log"
)

func init() {
	log.SetOutput(os.Stderr)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)

	// Log pretty text
	log.SetFormatter(&colorlog.ColorfulLogFormatter{})
}

func formatGroup(category string) string {
	return strings.ToUpper(category) + " COMMANDS"
}

////////////////////////////
// Commandline definition //
////////////////////////////

// RunCmdline starts a brig commandline tool.
func RunCmdline() int {

	app := cli.NewApp()
	app.Name = "brig"
	app.Usage = "Secure and dezentralized file synchronization"

	//groups
	repoGroup := formatGroup("repository")
	idntGroup := formatGroup("id helper")
	wdirGroup := formatGroup("working")
	advnGroup := formatGroup("advanced")
	miscGroup := formatGroup("misc")

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "nodaemon,n",
			Usage: "Don't run the daemon",
		},
		cli.StringFlag{
			Name:  "password, x",
			Usage: "Supply user password",
			Value: "",
		},
		cli.StringFlag{
			Name:   "path",
			Usage:  "Path of the repository",
			Value:  ".",
			EnvVar: "BRIG_PATH",
		},
	}

	// Commands.
	app.Commands = []cli.Command{
		{
			Name:        "init",
			Category:    repoGroup,
			Usage:       "Initialize an empty repository",
			ArgsUsage:   "<brig-id>",
			Description: "Creates a new brig repository folder and unlocks it.\n   The name of the folder is derivated from the given brig-id.\n   brig-id example: yourname@optionaldomain/ressource",
			Action:      withArgCheck(needAtLeast(1), withExit(handleInit)),
		},
		cli.Command{
			Name:        "open",
			Category:    repoGroup,
			Usage:       "Open an encrypted repository",
			ArgsUsage:   "[--password]",
			Description: "Open a closed (encrypted) brig repository by providing a password",
			Action:      withDaemon(handleOpen, true),
		},
		cli.Command{
			Name:        "close",
			Category:    repoGroup,
			Usage:       "Close an encrypted repository",
			Description: "Encrypt all metadata in the repository and go offline",
			Action:      withDaemon(handleClose, false),
		},
		cli.Command{
			Name:        "history",
			Category:    repoGroup,
			Usage:       "Show the history of the given brig file",
			Action:      withArgCheck(needAtLeast(1), withDaemon(handleHistory, true)),
			Description: "history lists all modifications of a given file",
			ArgsUsage:   "<filename>",
		},
		cli.Command{
			Name:        "net",
			Category:    repoGroup,
			Usage:       "Query and modify network status",
			ArgsUsage:   "[offline | online | status]",
			Description: "Query and modify the connection state to the ipfs network",
			Subcommands: []cli.Command{
				cli.Command{
					Name:   "offline",
					Usage:  "Disconnect from the outside world. The daemon will continue running",
					Action: withDaemon(handleOffline, true),
				},
				cli.Command{
					Name:   "online",
					Usage:  "Connect the daemon to the outside world",
					Action: withDaemon(handleOnline, true),
				},
				cli.Command{
					Name:   "status",
					Usage:  "Check if the daemon is online",
					Action: withDaemon(handleIsOnline, true),
				},
			},
		},

		cli.Command{
			Name:     "remote",
			Category: idntGroup,
			Usage:    "Manage remotes",
			Action:   withDaemon(handleRemoteList, true),
			Subcommands: []cli.Command{
				cli.Command{
					Name:   "add",
					Usage:  "Add a specific remote",
					Action: withArgCheck(needAtLeast(2), withDaemon(handleRemoteAdd, true)),
				},
				cli.Command{
					Name:   "remove",
					Usage:  "Remove a specifc remote",
					Action: withArgCheck(needAtLeast(1), withDaemon(handleRemoteRemove, true)),
				},
				cli.Command{
					Name:   "list",
					Usage:  "List remote status",
					Action: withDaemon(handleRemoteList, true),
				},
				cli.Command{
					Name:   "locate",
					Usage:  "Search a specific remote",
					Action: withArgCheck(needAtLeast(1), withDaemon(handleRemoteLocate, true)),
				},
				cli.Command{
					Name:   "self",
					Usage:  "Print own identity hash",
					Action: withDaemon(handleRemoteSelf, true),
				},
			},
		},
		cli.Command{
			Name:     "tree",
			Usage:    "List files in a tree",
			Category: wdirGroup,
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "depth, d",
					Usage: "Max depth to traverse",
					Value: -1,
				},
			},
			Action: withDaemon(handleTree, true),
		},
		cli.Command{
			Name:     "ls",
			Usage:    "List files",
			Category: wdirGroup,
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "depth, d",
					Usage: "Max depth to traverse",
					Value: -1,
				},
				cli.BoolFlag{
					Name:  "recursive,r",
					Usage: "Allow recursive traverse",
				},
			},
			Action: withDaemon(handleList, true),
		},
		cli.Command{
			Name:     "mkdir",
			Category: wdirGroup,
			Usage:    "Create an empty directory",
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleMkdir, true)),
		},
		cli.Command{
			Name:     "add",
			Category: wdirGroup,
			Usage:    "Transer file into brig's control",
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleAdd, true)),
		},
		cli.Command{
			Name:     "rm",
			Category: wdirGroup,
			Usage:    "Remove the file and optionally old versions of it",
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleRm, true)),
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "recursive,r",
					Usage: "Remove directories recursively",
				},
			},
		},
		cli.Command{
			Name:     "mv",
			Category: wdirGroup,
			Usage:    "Move a file from SOURCE to DEST",
			Action:   withArgCheck(needAtLeast(2), withDaemon(handleMv, true)),
		},
		cli.Command{
			Name:     "cat",
			Category: wdirGroup,
			Usage:    "Write ",
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleCat, true)),
		},
		cli.Command{
			Name:     "daemon",
			Category: advnGroup,
			Usage:    "Manually run the daemon process",
			// Flags
			Action: withExit(handleDaemon),
			Subcommands: []cli.Command{
				cli.Command{
					Name:     "quit",
					Category: advnGroup,
					Usage:    "Manually kill the daemon process",
					Action:   withDaemon(handleDaemonQuit, false),
				},
				cli.Command{
					Name:     "ping",
					Category: advnGroup,
					Usage:    "See if the daemon responds in a timely fashion",
					Action:   withDaemon(handleDaemonPing, false),
				},
				cli.Command{
					Name:     "wait",
					Category: advnGroup,
					Usage:    "Block until the daemon is available",
					Action:   withExit(handleDaemonWait),
				},
			},
		},
		cli.Command{
			Name:     "config",
			Category: miscGroup,
			Usage:    "Access, list and modify configuration values",
			Subcommands: []cli.Command{
				cli.Command{
					Name:   "list",
					Usage:  "Show current config values",
					Action: withExit(withConfig(handleConfigList)),
				},
				cli.Command{
					Name:   "get",
					Usage:  "Connect the daemon to the outside world",
					Action: withExit(withConfig(handleConfigGet)),
				},
				cli.Command{
					Name:   "set",
					Usage:  "Check if the daemon is online",
					Action: withExit(withConfig(handleConfigSet)),
				},
			},
		},
		cli.Command{
			Name:     "mount",
			Category: miscGroup,
			Usage:    "Handle FUSE mountpoints",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "unmount,u",
					Usage: "Unmount the specified directory",
				},
			},
			Action: withArgCheck(needAtLeast(1), withDaemon(handleMount, true)),
		},
	}

	app.Run(os.Args)
	return 0
}
