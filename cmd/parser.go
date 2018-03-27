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
			Name:     "init",
			Category: repoGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleInit, true)),
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "backend,b",
					Value: "ipfs",
					Usage: "What data backend to use for the new repo",
				},
			},
		},
		cli.Command{
			Name:     "whoami",
			Category: netwGroup,
			Action:   withDaemon(handleWhoami, true),
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
			Name:     "remote",
			Aliases:  []string{"rmt"},
			Category: netwGroup,
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
					Name:    "remove",
					Aliases: []string{"rm"},
					Action:  withArgCheck(needAtLeast(1), withDaemon(handleRemoteRemove, true)),
				},
				cli.Command{
					Name:    "list",
					Aliases: []string{"ls"},
					Action:  withDaemon(handleRemoteList, true),
				},
				cli.Command{
					Name:   "clear",
					Action: withDaemon(handleRemoteClear, true),
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
				},
				cli.Command{
					Name:   "ping",
					Action: withArgCheck(needAtLeast(1), withDaemon(handleRemotePing, true)),
				},
			},
		},
		cli.Command{
			Name:     "pin",
			Category: netwGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handlePin, true)),
			Subcommands: []cli.Command{
				cli.Command{
					Name:   "add",
					Action: withDaemon(handlePin, true),
				},
				cli.Command{
					Name:   "rm",
					Action: withDaemon(handleUnpin, true),
				},
			},
		},
		cli.Command{
			Name:     "net",
			Category: netwGroup,
			Subcommands: []cli.Command{
				cli.Command{
					Name:   "offline",
					Action: withDaemon(handleOffline, true),
				},
				cli.Command{
					Name:   "online",
					Action: withDaemon(handleOnline, true),
				},
				cli.Command{
					Name:   "status",
					Action: withDaemon(handleIsOnline, true),
				},
				// TODO: Should this go to remotes?
				cli.Command{
					Name:    "list",
					Aliases: []string{"ls"},
					Action:  withDaemon(handleOnlinePeers, true),
				},
				cli.Command{
					Name:   "locate",
					Action: withArgCheck(needAtLeast(1), withDaemon(handleNetLocate, true)),
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
			Name:     "status",
			Aliases:  []string{"st"},
			Category: vcscGroup,
			Action:   withDaemon(handleStatus, true),
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "tree,t",
					Usage: "View the status as a tree listing",
				},
			},
		},
		cli.Command{
			// TODO: Do automated fetch by default.
			Name:     "diff",
			Category: vcscGroup,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "list,l",
					Usage: "Output the diff as simple list (like status)",
				},
			},
			Action: withDaemon(handleDiff, true),
		},
		cli.Command{
			Name:     "tag",
			Category: vcscGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleTag, true)),
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "delete,d",
					Usage: "Delete the tag instead of creating it",
				},
			},
		},
		cli.Command{
			Name:     "log",
			Category: vcscGroup,
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
			Name:     "fetch",
			Category: vcscGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleFetch, true)),
		},
		cli.Command{
			// TODO: option to auto-download (parts of?) the synced result.
			Name:     "sync",
			Category: vcscGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleSync, true)),
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "no-fetch,n",
					Usage: "Do not do a fetch before syncing",
				},
			},
		},
		cli.Command{
			// TODO: Do re-pinning of old files only after a commit (to allow safe jump backs)
			// TODO: Have the notion of explicit pins to save them from indirect/automatic unpins?
			//       (is this what ipfs has?)
			Name:     "commit",
			Aliases:  []string{"cmt"},
			Category: vcscGroup,
			// TODO: move bash completion also to help.
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
			// TODO: Figure out/test exact way of pinning and write docs for it.
			Name:     "reset",
			Aliases:  []string{"co"},
			Category: vcscGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleReset, true)),
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "force,f",
					Usage: "Reset even when there are changes in the staging area",
				},
			},
		},
		cli.Command{
			Name:     "become",
			Category: vcscGroup,
			Action:   withDaemon(handleBecome, true),
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "self,s",
					Usage: "Become self (i.e. the owner of the repository)",
				},
			},
		},
		cli.Command{
			Name:     "history",
			Aliases:  []string{"hst", "hist"},
			Category: vcscGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleHistory, true)),
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "empty,e",
					Usage: "Also show commits where nothing happens",
				},
			},
		},
		cli.Command{
			Name:     "stage",
			Aliases:  []string{"stg", "add", "a"},
			Category: wdirGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleStage, true)),
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "stdin,i",
					Usage: "Read data from stdin",
				},
			},
		},
		cli.Command{
			Name:     "touch",
			Aliases:  []string{"t"},
			Category: wdirGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleTouch, true)),
		},
		cli.Command{
			Name:     "cat",
			Category: wdirGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleCat, true)),
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
			ArgsUsage:   "<file>",
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
			ArgsUsage:   "/path",
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
			Name:         "tree",
			Usage:        "List files similar to tree(1)",
			ArgsUsage:    "[/brig-path] [--depth|-d]",
			Description:  "Lists all files of a specific brig path in a tree like-manner",
			Category:     wdirGroup,
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
		},
		cli.Command{
			Name:         "cp",
			Category:     wdirGroup,
			Usage:        "Copy a file or directory elsewhere (reflink)",
			ArgsUsage:    "<source> <dest>",
			Description:  "Copy a file from SOURCE to DEST",
			Action:       withArgCheck(needAtLeast(2), withDaemon(handleCp, true)),
		},
		cli.Command{
			Name:         "edit",
			Category:     wdirGroup,
			Usage:        "Edit a file in brig with $EDITOR",
			ArgsUsage:    "<path>",
			Description:  "Edit a file in brig with $EDITOR",
			Action:       withArgCheck(needAtLeast(1), withDaemon(handleEdit, true)),
		},
		cli.Command{
			Name:         "daemon",
			Category:     repoGroup,
			Usage:        "Manually run the daemon process",
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
				},
				cli.Command{
					Name:         "ping",
					Usage:        "See if the daemon responds in a timely fashion",
					Description:  "Checks if deamon is running and reports the response time",
					Action:       withDaemon(handleDaemonPing, false),
				},
			},
		},
		cli.Command{
			Name:         "config",
			Category:     repoGroup,
			Usage:        "Access, list and modify configuration values",
			Subcommands: []cli.Command{
				cli.Command{
					Name:         "list",
					Usage:        "Show current config values",
					Description:  "Show the current brig configuration",
					Action:       withDaemon(handleConfigList, true),
				},
				cli.Command{
					Name:         "get",
					Usage:        "Get a specific config value",
					Description:  "Get a specific config value and print it to stdout",
					ArgsUsage:    "<configkey>",
					Action:       withArgCheck(needAtLeast(1), withDaemon(handleConfigGet, true)),
				},
				cli.Command{
					Name:         "set",
					Usage:        "Set a specific config value",
					Description:  "Set a given config option to the given value",
					ArgsUsage:    "<configkey> <value>",
					Action:       withArgCheck(needAtLeast(2), withDaemon(handleConfigSet, true)),
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
		},
		cli.Command{
			Name:         "unmount",
			Category:     repoGroup,
			Usage:        "Unmount a previosuly mounted directory",
			ArgsUsage:    "<mount>",
			Description:  "Unmounts a FUSE filesystem",
			Action:       withDaemon(handleUnmount, true),
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
		},
	})

	if err := app.Run(args); err != nil {
		return 1
	}
	return 0
}
