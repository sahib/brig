package cmd

import (
	"fmt"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/fatih/color"
	isatty "github.com/mattn/go-isatty"
	formatter "github.com/sahib/brig/util/log"
	"github.com/sahib/brig/version"
	"github.com/urfave/cli"
)

func init() {
	log.SetOutput(os.Stderr)
	log.SetLevel(log.DebugLevel)

	// Only use fancy logging if we print to a terminal:
	if isatty.IsTerminal(os.Stdout.Fd()) {
		log.SetFormatter(&formatter.ColorfulLogFormatter{})
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
		"%s [buildtime: %s] (client version)",
		version.String(),
		version.BuildTime,
	)
	app.CommandNotFound = commandNotFound
	app.Description = "brig can be used to easily store, version and synchronize files between many peers."

	// Set global options here:
	app.Before = func(ctx *cli.Context) error {
		if ctx.Bool("no-color") {
			color.NoColor = true
		}

		return nil
	}

	// Groups:
	repoGroup := formatGroup("repository")
	wdirGroup := formatGroup("working tree")
	vcscGroup := formatGroup("version control")
	netwGroup := formatGroup("network")

	// autocomplete all commands, but not their aliases.
	app.BashComplete = func(ctx *cli.Context) {
		for _, cmd := range app.Commands {
			fmt.Println(cmd.Name)
		}
	}

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
			Name:   "bind",
			Usage:  "To what host to bind to (default: localhost)",
			Value:  "localhost",
			EnvVar: "BRIG_BIND",
		},
		cli.IntFlag{
			Name:   "port,p",
			Usage:  "On what port the daemon listens on (default: 6666)",
			EnvVar: "BRIG_PORT",
			Value:  6666,
		},
		cli.BoolFlag{
			Name:  "no-color,",
			Usage: "Use 'no-pass' as password",
		},
		cli.StringFlag{
			Name:  "password,P",
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

	// TODO: Implement 'brig help online' (or similar) to open online docs in a browser.
	app.Commands = TranslateHelp([]cli.Command{
		{
			Name:         "init",
			Category:     repoGroup,
			Action:       withArgCheck(needAtLeast(1), withDaemon(handleInit, true)),
			BashComplete: completeArgsUsage,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "backend,b",
					Value: "ipfs",
					//Usage: "What data backend to use for the new repo",
				},
			},
		},
		cli.Command{
			Name:         "whoami",
			Category:     netwGroup,
			Action:       withDaemon(handleWhoami, true),
			BashComplete: completeArgsUsage,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "fingerprint,f",
					Usage: "Only print the own fingerprint",
				},
				cli.BoolFlag{
					Name:  "name,n",
					Usage: "Only print the own name",
				},
			},
		},
		cli.Command{
			Name:         "remote",
			Aliases:      []string{"rmt"},
			Category:     netwGroup,
			BashComplete: completeSubcommands,
			Subcommands: []cli.Command{
				cli.Command{
					Name:   "add",
					Action: withArgCheck(needAtLeast(2), withDaemon(handleRemoteAdd, true)),
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "folder,f",
							Usage: "What folder the remote can access",
						},
					},
				},
				cli.Command{
					Name:         "remove",
					Aliases:      []string{"rm"},
					Action:       withArgCheck(needAtLeast(1), withDaemon(handleRemoteRemove, true)),
					BashComplete: completeArgsUsage,
				},
				cli.Command{
					Name:         "list",
					Aliases:      []string{"ls"},
					Action:       withDaemon(handleRemoteList, true),
					BashComplete: completeArgsUsage,
				},
				cli.Command{
					Name:         "clear",
					Action:       withDaemon(handleRemoteClear, true),
					BashComplete: completeArgsUsage,
				},
				cli.Command{
					Name:   "edit",
					Action: withDaemon(handleRemoteEdit, true),
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "yml,y",
							Value: "",
							Usage: "Directly overwrite remote list with yml file",
						},
					},
					BashComplete: completeArgsUsage,
				},
				cli.Command{
					Name:         "ping",
					Action:       withArgCheck(needAtLeast(1), withDaemon(handleRemotePing, true)),
					BashComplete: completeArgsUsage,
				},
			},
		},
		cli.Command{
			Name:         "pin",
			Category:     netwGroup,
			Action:       withArgCheck(needAtLeast(1), withDaemon(handlePin, true)),
			BashComplete: completeSubcommands,
			Subcommands: []cli.Command{
				cli.Command{
					Name:         "add",
					Action:       withDaemon(handlePin, true),
					BashComplete: completeArgsUsage,
				},
				cli.Command{
					Name:         "rm",
					Action:       withDaemon(handleUnpin, true),
					BashComplete: completeArgsUsage,
				},
			},
		},
		cli.Command{
			Name:         "net",
			Category:     netwGroup,
			BashComplete: completeSubcommands,
			Subcommands: []cli.Command{
				cli.Command{
					Name:         "offline",
					Action:       withDaemon(handleOffline, true),
					BashComplete: completeArgsUsage,
				},
				cli.Command{
					Name:         "online",
					Action:       withDaemon(handleOnline, true),
					BashComplete: completeArgsUsage,
				},
				cli.Command{
					Name:         "status",
					Action:       withDaemon(handleIsOnline, true),
					BashComplete: completeArgsUsage,
				},
				// TODO: Should this go to remotes?
				cli.Command{
					Name:         "list",
					Aliases:      []string{"ls"},
					Action:       withDaemon(handleOnlinePeers, true),
					BashComplete: completeArgsUsage,
				},
				cli.Command{
					Name:         "locate",
					Action:       withArgCheck(needAtLeast(1), withDaemon(handleNetLocate, true)),
					BashComplete: completeArgsUsage,
					// TODO: Provide flag to indicate what part of the name to search.
					// TODO: Make timeout a "time duration" (i.e. 5s)
					// TODO: think of way to upload fingerprint of node more
					Flags: []cli.Flag{
						cli.IntFlag{
							Name:  "t,timeout",
							Value: 10,
							Usage: "Wait at most <n> seconds before bailing out",
						},
					},
				},
			},
		},
		cli.Command{
			Name:         "status",
			Aliases:      []string{"st"},
			Category:     vcscGroup,
			Action:       withDaemon(handleStatus, true),
			BashComplete: completeArgsUsage,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "tree,t",
					Usage: "View the status as a tree listing",
				},
			},
		},
		cli.Command{
			Name:         "diff",
			Category:     vcscGroup,
			BashComplete: completeArgsUsage,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "list,l",
					Usage: "Output the diff as simple list (like status)",
				},
			},
			Action: withDaemon(handleDiff, true),
		},
		cli.Command{
			Name:         "tag",
			Category:     vcscGroup,
			Action:       withArgCheck(needAtLeast(1), withDaemon(handleTag, true)),
			BashComplete: completeArgsUsage,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "delete,d",
					Usage: "Delete the tag instead of creating it",
				},
			},
		},
		cli.Command{
			Name:         "log",
			Category:     vcscGroup,
			BashComplete: completeArgsUsage,
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
			Name:         "fetch",
			Category:     vcscGroup,
			Usage:        "Fetch the metadata from a remote",
			Description:  "Fetch the metadata from a remote",
			Action:       withArgCheck(needAtLeast(1), withDaemon(handleFetch, true)),
			BashComplete: completeArgsUsage,
		},
		cli.Command{
			Name:         "sync",
			Category:     vcscGroup,
			Usage:        "Sync with any partner in your remote list",
			Description:  "Attempt to synchronize your files with any partner",
			Action:       withArgCheck(needAtLeast(1), withDaemon(handleSync, true)),
			BashComplete: completeArgsUsage,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "no-fetch,n",
					Usage: "Do not do a fetch before syncing",
				},
			},
		},
		cli.Command{
			Name:         "commit",
			Aliases:      []string{"cmt"},
			Category:     vcscGroup,
			Usage:        "Print which file are in the staging area",
			Description:  "Show all changed files since the last commit and what a new commit would contain",
			BashComplete: completeArgsUsage,
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
			Name:         "reset",
			Aliases:      []string{"co"},
			Category:     vcscGroup,
			Usage:        "Reset commits, file or directories to an old state",
			ArgsUsage:    "<commit> [<file>] [--force]",
			Description:  "",
			Action:       withArgCheck(needAtLeast(1), withDaemon(handleReset, true)),
			BashComplete: completeArgsUsage,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "force,f",
					Usage: "Reset even when there are changes in the staging area",
				},
			},
		},
		cli.Command{
			Name:         "become",
			Category:     vcscGroup,
			Usage:        "Act as other user and view the data we synced with",
			Description:  "Act as other user and view the data we synced with.",
			Action:       withDaemon(handleBecome, true),
			BashComplete: completeArgsUsage,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "self,s",
					Usage: "Become self (i.e. the owner of the repository)",
				},
			},
		},
		cli.Command{
			Name:         "history",
			Aliases:      []string{"hst", "hist"},
			Category:     vcscGroup,
			Usage:        "Show the history of the given brig file",
			Action:       withArgCheck(needAtLeast(1), withDaemon(handleHistory, true)),
			Description:  "history lists all modifications of a given file",
			ArgsUsage:    "<path>",
			BashComplete: completeArgsUsage,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "empty,e",
					Usage: "Also show commits where nothing happens",
				},
			},
		},
		cli.Command{
			Name:         "stage",
			Aliases:      []string{"stg", "add"},
			Category:     wdirGroup,
			Usage:        "Transer a file into brig's control or update an existing one",
			ArgsUsage:    "<file>",
			Description:  "Stage a specific file into the brig repository",
			Action:       withArgCheck(needAtLeast(1), withDaemon(handleStage, true)),
			BashComplete: completeArgsUsage,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "stdin,i",
					Usage: "Read data from stdin",
				},
			},
		},
		cli.Command{
			Name:         "touch",
			Aliases:      []string{"t"},
			Category:     wdirGroup,
			Usage:        "Create an empty file or update the timestamp of an existing",
			ArgsUsage:    "<file>",
			Description:  "Create an empty file or update the timestamp of an existing",
			Action:       withArgCheck(needAtLeast(1), withDaemon(handleTouch, true)),
			BashComplete: completeArgsUsage,
		},
		cli.Command{
			Name:         "cat",
			Category:     wdirGroup,
			Usage:        "Output content of any file to stdout",
			ArgsUsage:    "<file>",
			Description:  "Concatenates files and print them on stdout",
			Action:       withArgCheck(needAtLeast(1), withDaemon(handleCat, true)),
			BashComplete: completeArgsUsage,
		},
		cli.Command{
			Name:         "info",
			Category:     wdirGroup,
			Usage:        "Lookup extended attributes of a single filesystem node",
			ArgsUsage:    "<file>",
			Description:  "Stage a specific file into the brig repository",
			Action:       withArgCheck(needAtLeast(1), withDaemon(handleInfo, true)),
			BashComplete: completeArgsUsage,
		},
		cli.Command{
			Name:         "rm",
			Aliases:      []string{"remove"},
			Category:     wdirGroup,
			Usage:        "Remove the file and optionally old versions of it",
			ArgsUsage:    "<file>",
			Description:  "Remove a spcific file or directory",
			Action:       withArgCheck(needAtLeast(1), withDaemon(handleRm, true)),
			BashComplete: completeArgsUsage,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "recursive,r",
					Usage: "Remove directories recursively",
				},
			},
		},
		cli.Command{
			Name:         "ls",
			Usage:        "List files similar to ls(1)",
			ArgsUsage:    "/path",
			Description:  "Lists all files of a specific brig path in a ls-like manner",
			Category:     wdirGroup,
			BashComplete: completeArgsUsage,
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
			Name:         "tree",
			Usage:        "List files similar to tree(1)",
			ArgsUsage:    "[/brig-path] [--depth|-d]",
			Description:  "Lists all files of a specific brig path in a tree like-manner",
			Category:     wdirGroup,
			BashComplete: completeArgsUsage,
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
			Name:         "mkdir",
			Category:     wdirGroup,
			Usage:        "Create an empty directory",
			ArgsUsage:    "<dirname>",
			Description:  "Create a empty directory",
			BashComplete: completeArgsUsage,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "parents, p",
					Usage: "Create parent directories as needed",
				},
			},
			Action: withArgCheck(needAtLeast(1), withDaemon(handleMkdir, true)),
		},
		cli.Command{
			Name:         "mv",
			Category:     wdirGroup,
			Usage:        "Move a specific file",
			ArgsUsage:    "<source> <destination>",
			Description:  "Move a file from SOURCE to DEST",
			Action:       withArgCheck(needAtLeast(2), withDaemon(handleMv, true)),
			BashComplete: completeArgsUsage,
		},
		cli.Command{
			Name:         "cp",
			Category:     wdirGroup,
			Usage:        "Copy a file or directory elsewhere (reflink)",
			ArgsUsage:    "<source> <dest>",
			Description:  "Copy a file from SOURCE to DEST",
			Action:       withArgCheck(needAtLeast(2), withDaemon(handleCp, true)),
			BashComplete: completeArgsUsage,
		},
		cli.Command{
			Name:         "edit",
			Category:     wdirGroup,
			Usage:        "Edit a file in brig with $EDITOR",
			ArgsUsage:    "<path>",
			Description:  "Edit a file in brig with $EDITOR",
			Action:       withArgCheck(needAtLeast(1), withDaemon(handleEdit, true)),
			BashComplete: completeArgsUsage,
		},
		cli.Command{
			Name:         "daemon",
			Category:     repoGroup,
			Usage:        "Manually run the daemon process",
			BashComplete: completeSubcommands,
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
					Name:         "quit",
					Usage:        "Manually kill the daemon process",
					Description:  "Disconnect from ipfs network, shutdown the daemon and lock the repository",
					Action:       withDaemon(handleDaemonQuit, false),
					BashComplete: completeArgsUsage,
				},
				cli.Command{
					Name:         "ping",
					Usage:        "See if the daemon responds in a timely fashion",
					Description:  "Checks if deamon is running and reports the response time",
					Action:       withDaemon(handleDaemonPing, false),
					BashComplete: completeArgsUsage,
				},
			},
		},
		cli.Command{
			Name:         "config",
			Category:     repoGroup,
			Usage:        "Access, list and modify configuration values",
			BashComplete: completeSubcommands,
			Subcommands: []cli.Command{
				cli.Command{
					Name:         "list",
					Usage:        "Show current config values",
					Description:  "Show the current brig configuration",
					Action:       withDaemon(handleConfigList, true),
					BashComplete: completeArgsUsage,
				},
				cli.Command{
					Name:         "get",
					Usage:        "Get a specific config value",
					Description:  "Get a specific config value and print it to stdout",
					ArgsUsage:    "<configkey>",
					Action:       withArgCheck(needAtLeast(1), withDaemon(handleConfigGet, true)),
					BashComplete: completeArgsUsage,
				},
				cli.Command{
					Name:         "set",
					Usage:        "Set a specific config value",
					Description:  "Set a given config option to the given value",
					ArgsUsage:    "<configkey> <value>",
					Action:       withArgCheck(needAtLeast(2), withDaemon(handleConfigSet, true)),
					BashComplete: completeArgsUsage,
				},
			},
		},
		cli.Command{
			Name:        "mount",
			Category:    repoGroup,
			Usage:       "Mount a brig repository",
			ArgsUsage:   "<mount>",
			Description: "Mount a brig repository as FUSE filesystem",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "umount,u",
					Usage: "Unmount the specified directory",
				},
			},
			Action:       withDaemon(handleMount, true),
			BashComplete: completeArgsUsage,
		},
		cli.Command{
			Name:         "unmount",
			Category:     repoGroup,
			Usage:        "Unmount a previosuly mounted directory",
			ArgsUsage:    "<mount>",
			Description:  "Unmounts a FUSE filesystem",
			Action:       withDaemon(handleUnmount, true),
			BashComplete: completeArgsUsage,
		},
		cli.Command{
			Name:     "version",
			Category: repoGroup,
			Usage:    "Show brig and backend (ipfs) version info",
			Action:   withDaemon(handleVersion, false),
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
			Action:       withDaemon(handleGc, true),
			BashComplete: completeArgsUsage,
		},
	})

	if err := app.Run(args); err != nil {
		return 1
	}
	return 0
}
