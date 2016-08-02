package cmdline

import (
	"fmt"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/disorganizer/brig"
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

func setLogPath(path string) error {
	switch path {
	case "stdout":
		log.SetOutput(os.Stdout)
	case "stderr":
		log.SetOutput(os.Stderr)
	default:
		fd, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}

		log.SetOutput(fd)
	}

	return nil
}

////////////////////////////
// Commandline definition //
////////////////////////////

// RunCmdline starts a brig commandline tool.
func RunCmdline(args []string) int {
	app := cli.NewApp()
	app.Name = "brig"
	app.Usage = "Secure and dezentralized file synchronization"
	app.Version = fmt.Sprintf(
		"%s [buildtime: %s]",
		brig.VersionString(),
		brig.BuildTime,
	)

	// Groups:
	repoGroup := formatGroup("repository")
	wdirGroup := formatGroup("working")
	vcscGroup := formatGroup("version control")
	advnGroup := formatGroup("advanced")
	miscGroup := formatGroup("misc")

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "nodaemon,n",
			Usage: "Don't run the daemon",
		},
		cli.StringFlag{
			Name:  "password,x",
			Usage: "Supply user password",
			Value: "",
		},
		cli.StringFlag{
			Name:   "path",
			Usage:  "Path of the repository",
			Value:  ".",
			EnvVar: "BRIG_PATH",
		},
		cli.StringFlag{
			Name:   "log-path,l",
			Usage:  "Where to output the log. May be 'stderr' (default) or 'stdout'",
			Value:  "stderr",
			EnvVar: "BRIG_LOG",
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
			ArgsUsage:   "[--password|-x]",
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
			Name:        "pin",
			Category:    repoGroup,
			Usage:       "Pin a file locally to this machine",
			Action:      withArgCheck(needAtLeast(1), withDaemon(handlePin, true)),
			ArgsUsage:   "<file>",
			Description: "Ensure that <file> is physically stored on this machine.",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "unpin,u",
					Usage: "Remove a pin again",
				},
				cli.BoolFlag{
					Name:  "is-pinned, i",
					Usage: "Check if <file> is pinned",
				},
			},
		},
		cli.Command{
			Name:        "net",
			Category:    repoGroup,
			Usage:       "Query and modify network status",
			ArgsUsage:   "[offline|online|status]",
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
			Name:        "status",
			Category:    vcscGroup,
			Usage:       "Print which file are in the staging area",
			Description: "Show all changed files since the last commit and what a new commit would contain",
			Action:      withDaemon(handleStatus, true),
		},
		cli.Command{
			Name:        "diff",
			Category:    vcscGroup,
			Usage:       "Show what changed between two commits",
			ArgsUsage:   "[--from <hash> | --to <hash>]",
			Description: "Show the difference between two points in the history",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "from,f",
					Value: "",
					Usage: "Lower range limit; initial commit if omitted",
				},
				cli.StringFlag{
					Name:  "to,t",
					Value: "",
					Usage: "Upper range limit; HEAD if omitted",
				},
			},
			Action: withDaemon(handleDiff, true),
		},
		cli.Command{
			Name:        "log",
			Category:    vcscGroup,
			Usage:       "Show all commits in a certain range",
			ArgsUsage:   "[--from <hash> | --to <hash>]",
			Description: "List a short summary of all commits in a range or all of them",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "from,f",
					Value: "",
					Usage: "Lower range limit; initial commit if omitted",
				},
				cli.StringFlag{
					Name:  "to,t",
					Value: "",
					Usage: "Upper range limit; HEAD if omitted",
				},
			},
			Action: withDaemon(handleLog, true),
		},
		cli.Command{
			Name:        "commit",
			Category:    vcscGroup,
			Usage:       "Print which file are in the staging area",
			Description: "Show all changed files since the last commit and what a new commit would contain",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "message,m",
					Value: "",
					Usage: "Provide a meaningful commit message",
				},
			},
			Action: withDaemon(handleCommit, true),
		},
		cli.Command{
			Name:        "remote",
			Category:    repoGroup,
			Usage:       "Remote management.",
			ArgsUsage:   "[add|remove|list|locate|self]",
			Description: "Add, remove, list, locate remotes and print own identity",
			Subcommands: []cli.Command{
				cli.Command{
					Name:        "add",
					Usage:       "Add a specific remote",
					ArgsUsage:   "<brig-id> <ipfs-hash>",
					Description: "Adds a specific user (brig-remote-id) with a specific identity (ipfs-hash) to remotes",
					Action:      withArgCheck(needAtLeast(2), withDaemon(handleRemoteAdd, true)),
				},
				cli.Command{
					Name:        "remove",
					Usage:       "Remove a specifc remote",
					ArgsUsage:   "<brig-remote-id>",
					Description: "Removes a specific remote from remotes.",
					Action:      withArgCheck(needAtLeast(1), withDaemon(handleRemoteRemove, true)),
				},
				cli.Command{
					Name:        "list",
					Usage:       "List status of known remotes",
					Description: "Lists all known remotes and their status",
					Action:      withDaemon(handleRemoteList, true),
				},
				cli.Command{
					Name:        "locate",
					Usage:       "Search a specific remote",
					ArgsUsage:   "<brig-remote-id>",
					Description: "Locates all remotes with the given brig-remote-id ",
					Action:      withArgCheck(needAtLeast(1), withDaemon(handleRemoteLocate, true)),
				},
				cli.Command{
					Name:        "self",
					Usage:       "Print identity",
					Description: "Prints the users identity and online status",
					Action:      withDaemon(handleRemoteSelf, true),
				},
			},
		},
		cli.Command{
			Name:        "tree",
			Usage:       "List files in a tree",
			ArgsUsage:   "[/brig-path] [--depth|-d]",
			Description: "Lists all files of a specific brig path in a tree like-manner",
			Category:    wdirGroup,
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
			Name:        "ls",
			Usage:       "List files",
			ArgsUsage:   "[/brig-path] [--depth|-d] [--recursive|-r]",
			Description: "Lists all files of a specific brig path in a ls-like manner",
			Category:    wdirGroup,
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
			Name:        "mkdir",
			Category:    wdirGroup,
			Usage:       "Create an empty directory",
			ArgsUsage:   "</dirname>",
			Description: "Create a empty directory",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "parents, p",
					Usage: "Create parent directories as needed",
				},
			},
			Action: withArgCheck(needAtLeast(1), withDaemon(handleMkdir, true)),
		},
		cli.Command{
			Name:        "add",
			Category:    wdirGroup,
			Usage:       "Transer file into brig's control",
			ArgsUsage:   "</file>",
			Description: "Add a specific file to the brig repository",
			Action:      withArgCheck(needAtLeast(1), withDaemon(handleAdd, true)),
		},
		cli.Command{
			Name:        "rm",
			Category:    wdirGroup,
			Usage:       "Remove the file and optionally old versions of it",
			ArgsUsage:   "</file> [--recursive|-r]",
			Description: "Remove a spcific file or directory",
			Action:      withArgCheck(needAtLeast(1), withDaemon(handleRm, true)),
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "recursive,r",
					Usage: "Remove directories recursively",
				},
			},
		},
		cli.Command{
			Name:        "mv",
			Category:    wdirGroup,
			Usage:       "Move a specific file",
			ArgsUsage:   "</sourcefile> </destinationfile>",
			Description: "Move a file from SOURCE to DEST",
			Action:      withArgCheck(needAtLeast(2), withDaemon(handleMv, true)),
		},
		cli.Command{
			Name:        "cat",
			Category:    wdirGroup,
			Usage:       "Concatenates a file",
			ArgsUsage:   "</file>",
			Description: "Concatenates files and print them on stdout",
			Action:      withArgCheck(needAtLeast(1), withDaemon(handleCat, true)),
		},
		cli.Command{
			Name:     "daemon",
			Category: advnGroup,
			Usage:    "Manually run the daemon process",
			Subcommands: []cli.Command{
				cli.Command{
					Name:        "launch",
					Category:    advnGroup,
					Usage:       "Start the daemon process",
					Description: "Start the brig daemon process, unlock the repository and go online",
					Action:      withExit(handleDaemon),
				},
				cli.Command{
					Name:        "quit",
					Category:    advnGroup,
					Usage:       "Manually kill the daemon process",
					Description: "Disconnect from ipfs network, shutdown the daemon and lock the repository",
					Action:      withDaemon(handleDaemonQuit, false),
				},
				cli.Command{
					Name:        "ping",
					Category:    advnGroup,
					Usage:       "See if the daemon responds in a timely fashion",
					Description: "Checks if deamon is running and reports the response time",
					Action:      withDaemon(handleDaemonPing, false),
				},
				cli.Command{
					Name:        "wait",
					Category:    advnGroup,
					Usage:       "Block until the daemon is available",
					Description: "Wait blocks until the daemon is available",
					Action:      withExit(handleDaemonWait),
				},
			},
		},
		cli.Command{
			Name:     "debug",
			Category: miscGroup,
			Usage:    "Access to several debugging facilities",
			Subcommands: []cli.Command{
				cli.Command{
					Name:        "export",
					Usage:       "Export the currently loaded, own store to stdout",
					Description: "This is mainly useful for analyzing the current store contents",
					Action:      withDaemon(handleDebugExport, true),
				},
				cli.Command{
					Name:   "import",
					Usage:  "Import a previosuly exported store from stdin",
					Action: withDaemon(handleDebugImport, true),
				},
			},
		},
		cli.Command{
			Name:     "config",
			Category: miscGroup,
			Usage:    "Access, list and modify configuration values",
			Subcommands: []cli.Command{
				cli.Command{
					Name:        "list",
					Usage:       "Show current config values",
					Description: "Show the current brig configuration",
					Action:      withExit(withConfig(handleConfigList)),
				},
				cli.Command{
					Name:        "get",
					Usage:       "Get a specific config value",
					Description: "Get a specific config value and print it to stdout",
					ArgsUsage:   "<configkey>",
					Action:      withArgCheck(needAtLeast(1), withExit(withConfig(handleConfigGet))),
				},
				cli.Command{
					Name:        "set",
					Usage:       "Set a specific config value",
					Description: "Set a given config option to the given value",
					ArgsUsage:   "<configkey> <value>",
					Action:      withArgCheck(needAtLeast(2), withExit(withConfig(handleConfigSet))),
				},
			},
		},
		cli.Command{
			Name:        "mount",
			Category:    miscGroup,
			Usage:       "Mount a brig repository",
			ArgsUsage:   "[--umount|-u] <mountpath>",
			Description: "Mount a brig repository as FUSE filesystem",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "umount,u",
					Usage: "Unmount the specified directory",
				},
			},
			Action: withArgCheck(needAtLeast(1), withDaemon(handleMount, true)),
		},
	}

	app.Before = func(ctx *cli.Context) error {
		return setLogPath(ctx.String("log-path"))
	}

	app.Run(args)
	return 0
}
