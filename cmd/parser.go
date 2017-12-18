package cmd

import (
	"fmt"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	isatty "github.com/mattn/go-isatty"
	"github.com/sahib/brig/util/colors"
	formatter "github.com/sahib/brig/util/log"
	"github.com/sahib/brig/version"
	"github.com/urfave/cli"
)

func init() {
	log.SetOutput(os.Stderr)
	log.SetLevel(log.DebugLevel)

	// Only use color if we're printing to a terminal:
	if isatty.IsTerminal(os.Stdout.Fd()) {
		log.SetFormatter(&formatter.ColorfulLogFormatter{})
		colors.Enable()
	} else {
		colors.Disable()
	}
}

func formatGroup(category string) string {
	return strings.ToUpper(category) + " COMMANDS"
}

////////////////////////////
// Commandline definition //
////////////////////////////

// RunCmdline starts a brig commandline tool.
func RunCmdline(args []string) int {
	app := cli.NewApp()
	app.Name = "brig"
	app.Usage = "Secure and dezentralized file synchronization"
	app.EnableBashCompletion = true
	app.Version = fmt.Sprintf(
		"%s [buildtime: %s]",
		version.String(),
		version.BuildTime,
	)
	app.CommandNotFound = commandNotFound

	// Groups:
	repoGroup := formatGroup("repository")
	wdirGroup := formatGroup("working tree")
	vcscGroup := formatGroup("version control")
	netwGroup := formatGroup("network")

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "nodaemon,n",
			Usage: "Don't start the daemon automatically",
		},
		cli.BoolFlag{
			Name:  "no-password,x",
			Usage: "Use 'no-pass' as password",
		},
		cli.StringFlag{
			Name:  "password,p",
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
			Value:  "",
			EnvVar: "BRIG_LOG",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:        "init",
			Category:    repoGroup,
			Usage:       "Initialize an empty repository",
			ArgsUsage:   "<brig-id>",
			Description: "Creates a new brig repository folder and unlocks it.\n   The name of the folder is derivated from the given brig-id.\n   brig-id example: yourname@optionaldomain/ressource",
			Action:      withArgCheck(needAtLeast(1), withDaemon(handleInit, true)),
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "backend,b",
					Value: "mock",
					Usage: "What data backend to use for the new repo",
				},
			},
		},
		cli.Command{
			Name:        "whoami",
			Category:    netwGroup,
			Usage:       "Print information about this repository",
			Description: "Print information (like fingerprint, is-online) about this repository",
			Action:      withDaemon(handleWhoami, true),
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "fingerprint,f",
					Usage: "Only print the own fingerprint",
				},
			},
		},
		cli.Command{
			Name:        "remote",
			Category:    netwGroup,
			Usage:       "Manage what other peers can sync with us",
			ArgsUsage:   "[add|remove|list|locate|ping]",
			Description: "Add, remove, list, locate remotes and print own identity",
			Subcommands: []cli.Command{
				cli.Command{
					Name:        "add",
					Usage:       "Add a specific remote",
					ArgsUsage:   "<name> <fingerprint>",
					Description: "Adds a specific user with it's fingerprint to the remote list",
					Action:      withArgCheck(needAtLeast(2), withDaemon(handleRemoteAdd, true)),
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "folder,f",
							Value: "",
							Usage: "What folder the remote can access",
						},
					},
				},
				cli.Command{
					Name:        "remove",
					Aliases:     []string{"rm"},
					Usage:       "Remove a specifc remote",
					ArgsUsage:   "<name>",
					Description: "Removes a specific remote from remotes.",
					Action:      withArgCheck(needAtLeast(1), withDaemon(handleRemoteRemove, true)),
				},
				cli.Command{
					Name:        "list",
					Aliases:     []string{"ls"},
					Usage:       "List status of known remotes",
					Description: "Lists all known remotes and their status",
					Action:      withDaemon(handleRemoteList, true),
				},
				cli.Command{
					Name:        "edit",
					Usage:       "Edit the current remote list",
					Description: "Edit the current remote list with $EDITOR",
					Action:      withDaemon(handleRemoteEdit, true),
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "yml,y",
							Value: "",
							Usage: "Directly overwrite remote list with yml file",
						},
					},
				},
				cli.Command{
					Name:        "locate",
					Usage:       "Search a specific remote",
					ArgsUsage:   "<name>",
					Description: "Locates all remotes with the given brig-remote-id ",
					Action:      withArgCheck(needAtLeast(1), withDaemon(handleRemoteLocate, true)),
				},
				cli.Command{
					Name:        "ping",
					Usage:       "ping <remote-name>",
					Description: "Ping a remote and see if it responds",
					Action:      withArgCheck(needAtLeast(1), withDaemon(handleRemotePing, true)),
				},
			},
		},
		cli.Command{
			Name:        "pin",
			Category:    netwGroup,
			Usage:       "Pin a file to local storage",
			Action:      withArgCheck(needAtLeast(1), withDaemon(handlePin, true)),
			ArgsUsage:   "<file>",
			Description: "Ensure that <file> is physically stored on this machine.",
			Subcommands: []cli.Command{
				cli.Command{
					Name:   "add",
					Usage:  "Add a pin for a specific file or directory",
					Action: withDaemon(handlePin, true),
				},
				cli.Command{
					Name:   "rm",
					Usage:  "Remove a pin for a specific file or directory",
					Action: withDaemon(handleUnpin, true),
				},
			},
		},
		cli.Command{
			Name:        "net",
			Category:    netwGroup,
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
				cli.Command{
					Name:    "list",
					Aliases: []string{"ls"},
					Usage:   "See what other peers are online",
					Action:  withDaemon(handleOnlinePeers, true),
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
			ArgsUsage:   "[-r <name> | -v <hash>] <REMOTE-REV>",
			Description: "Show the difference between two points in the history",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "remote,r",
					Value: "",
					Usage: "Who to compare with (by default: self)",
				},
				cli.StringFlag{
					Name:  "rev,v",
					Value: "HEAD",
					Usage: "What commit to compare remote with (default: HEAD)",
				},
			},
			Action: withDaemon(handleDiff, true),
		},
		cli.Command{
			Name:        "tag",
			Category:    vcscGroup,
			Usage:       "Tag a commit with a specific name",
			ArgsUsage:   "<commit-rev> <name>",
			Description: "Give a commit an easier to remember name",
			Action:      withArgCheck(needAtLeast(1), withDaemon(handleTag, true)),
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "delete,d",
					Usage: "Delete the tag instead of creating it",
				},
			},
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
			Name:        "fetch",
			Category:    vcscGroup,
			Usage:       "Fetch the metadata from a remote",
			Description: "Fetch the metadata from a remote",
			Action:      withArgCheck(needAtLeast(1), withDaemon(handleFetch, true)),
		},
		cli.Command{
			Name:        "sync",
			Category:    vcscGroup,
			Usage:       "Sync with any partner in your remote list",
			Description: "Attempt to synchronize your files with any partner",
			Action:      withArgCheck(needAtLeast(1), withDaemon(handleSync, true)),
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "no-fetch,n",
					Usage: "Do not do a fetch before syncing",
				},
			},
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
			Name:        "reset",
			Category:    vcscGroup,
			Usage:       "Reset a file to a certain version",
			ArgsUsage:   "<file> [<commit>]",
			Description: "Reset a file to the last known state or to a certain commit",
			Action:      withArgCheck(needAtLeast(1), withDaemon(handleReset, true)),
		},
		cli.Command{
			Name:        "become",
			Category:    vcscGroup,
			Usage:       "Act as other user and view the data we synced with",
			Description: "Act as other user and view the data we synced with",
			Action:      withArgCheck(needAtLeast(1), withDaemon(handleBecome, true)),
		},
		cli.Command{
			Name:        "checkout",
			Category:    vcscGroup,
			Usage:       "Revert to a specific commit",
			ArgsUsage:   "<commit> [--force]",
			Description: "Make the staging commit equal to an old state",
			Action:      withArgCheck(needAtLeast(1), withDaemon(handleCheckout, true)),
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "force,f",
					Usage: "Remove directories recursively",
				},
			},
		},
		cli.Command{
			Name:        "history",
			Category:    vcscGroup,
			Usage:       "Show the history of the given brig file",
			Action:      withArgCheck(needAtLeast(1), withDaemon(handleHistory, true)),
			Description: "history lists all modifications of a given file",
			ArgsUsage:   "<filename>",
		},
		cli.Command{
			Name:        "stage",
			Category:    wdirGroup,
			Usage:       "Transer a file into brig's control or update an existing one",
			ArgsUsage:   "<file>",
			Description: "Stage a specific file into the brig repository",
			Action:      withArgCheck(needAtLeast(1), withDaemon(handleStage, true)),
		},
		cli.Command{
			Name:        "cat",
			Category:    wdirGroup,
			Usage:       "Output content of any file to stdout",
			ArgsUsage:   "<file>",
			Description: "Concatenates files and print them on stdout",
			Action:      withArgCheck(needAtLeast(1), withDaemon(handleCat, true)),
		},
		cli.Command{
			Name:        "info",
			Category:    wdirGroup,
			Usage:       "Lookup extended attributes of a single filesystem node",
			ArgsUsage:   "<file>",
			Description: "Stage a specific file into the brig repository",
			Action:      withArgCheck(needAtLeast(1), withDaemon(handleInfo, true)),
		},
		cli.Command{
			Name:        "rm",
			Aliases:     []string{"remove"},
			Category:    wdirGroup,
			Usage:       "Remove the file and optionally old versions of it",
			ArgsUsage:   "<file> [--recursive|-r]",
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
			Name:        "ls",
			Usage:       "List files similar to ls(1)",
			ArgsUsage:   "[/brig-path] [--depth|-d] [--recursive|-r]",
			Description: "Lists all files of a specific brig path in a ls-like manner",
			Category:    wdirGroup,
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "depth,d",
					Usage: "Max depth to traverse",
					Value: 1,
				},
				cli.BoolFlag{
					Name:  "recursive,R",
					Usage: "Allow recursive traverse",
				},
			},
			Action: withDaemon(handleList, true),
		},
		cli.Command{
			Name:        "tree",
			Usage:       "List files similar to tree(1)",
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
			Name:        "mkdir",
			Category:    wdirGroup,
			Usage:       "Create an empty directory",
			ArgsUsage:   "<dirname>",
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
			Name:        "mv",
			Category:    wdirGroup,
			Usage:       "Move a specific file",
			ArgsUsage:   "<sourcefile> <destinationfile>",
			Description: "Move a file from SOURCE to DEST",
			Action:      withArgCheck(needAtLeast(2), withDaemon(handleMv, true)),
		},
		cli.Command{
			Name:     "daemon",
			Category: repoGroup,
			Usage:    "Manually run the daemon process",
			Subcommands: []cli.Command{
				cli.Command{
					Name:        "launch",
					Usage:       "Start the daemon process",
					Description: "Start the brig daemon process, unlock the repository and go online",
					Action:      withExit(handleDaemonLaunch),
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "trace,t",
							Usage: "Create tracing output suitable for `go tool trace`",
						},
					},
				},
				cli.Command{
					Name:        "quit",
					Usage:       "Manually kill the daemon process",
					Description: "Disconnect from ipfs network, shutdown the daemon and lock the repository",
					Action:      withDaemon(handleDaemonQuit, false),
				},
				cli.Command{
					Name:        "ping",
					Usage:       "See if the daemon responds in a timely fashion",
					Description: "Checks if deamon is running and reports the response time",
					Action:      withDaemon(handleDaemonPing, false),
				},
			},
		},
		cli.Command{
			Name:     "config",
			Category: repoGroup,
			Usage:    "Access, list and modify configuration values",
			Subcommands: []cli.Command{
				cli.Command{
					Name:        "list",
					Usage:       "Show current config values",
					Description: "Show the current brig configuration",
					Action:      withDaemon(handleConfigList, true),
				},
				cli.Command{
					Name:        "get",
					Usage:       "Get a specific config value",
					Description: "Get a specific config value and print it to stdout",
					ArgsUsage:   "<configkey>",
					Action:      withArgCheck(needAtLeast(1), withDaemon(handleConfigGet, true)),
				},
				cli.Command{
					Name:        "set",
					Usage:       "Set a specific config value",
					Description: "Set a given config option to the given value",
					ArgsUsage:   "<configkey> <value>",
					Action:      withArgCheck(needAtLeast(2), withDaemon(handleConfigSet, true)),
				},
			},
		},
		cli.Command{
			Name:        "mount",
			Category:    repoGroup,
			Usage:       "Mount a brig repository",
			ArgsUsage:   "[--umount|-u] <mountpath>",
			Description: "Mount a brig repository as FUSE filesystem",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "umount,u",
					Usage: "Unmount the specified directory",
				},
			},
			Action: withDaemon(handleMount, true),
		},
		cli.Command{
			Name:        "unmount",
			Category:    repoGroup,
			Usage:       "Unmount a previosuly mounted directory",
			ArgsUsage:   "<mountpath>",
			Description: "Unmounts a FUSE filesystem",
			Action:      withDaemon(handleUnmount, true),
		},
		cli.Command{
			Name:        "gc",
			Category:    repoGroup,
			Usage:       "Trigger the ipfs garbage collector",
			ArgsUsage:   "",
			Description: "Trigger the ipfs garbage collector and print kill count",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "aggressive,a",
					Usage: "Also run the garbage collector on all filesystems immediately",
				},
			},
			Action: withDaemon(handleGc, true),
		},
	}

	app.Run(args)
	return 0
}
