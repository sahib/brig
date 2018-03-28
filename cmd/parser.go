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
			Usage: "Supply user password. Usage is not recommended.",
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
		}, {
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
		}, {
			Name:     "remote",
			Aliases:  []string{"rmt"},
			Category: netwGroup,
			Subcommands: []cli.Command{
				{
					Name:   "add",
					Action: withArgCheck(needAtLeast(2), withDaemon(handleRemoteAdd, true)),
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "folder,f",
							Usage: "What folder the remote can access",
						},
					},
				}, {
					Name:    "remove",
					Aliases: []string{"rm"},
					Action:  withArgCheck(needAtLeast(1), withDaemon(handleRemoteRemove, true)),
				}, {
					Name:    "list",
					Aliases: []string{"ls"},
					Action:  withDaemon(handleRemoteList, true),
				}, {
					Name:   "clear",
					Action: withDaemon(handleRemoteClear, true),
				}, {
					Name:   "edit",
					Action: withDaemon(handleRemoteEdit, true),
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "yml,y",
							Value: "",
							Usage: "Directly overwrite remote list with yml file",
						},
					},
				}, {
					Name:   "ping",
					Action: withArgCheck(needAtLeast(1), withDaemon(handleRemotePing, true)),
				},
			},
		}, {
			Name:     "pin",
			Category: netwGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handlePin, true)),
			Subcommands: []cli.Command{
				{
					Name:   "add",
					Action: withDaemon(handlePin, true),
				}, {
					Name:   "rm",
					Action: withDaemon(handleUnpin, true),
				},
			},
		}, {
			Name:     "net",
			Category: netwGroup,
			Subcommands: []cli.Command{
				{
					Name:   "offline",
					Action: withDaemon(handleOffline, true),
				}, {
					Name:   "online",
					Action: withDaemon(handleOnline, true),
				}, {
					Name:   "status",
					Action: withDaemon(handleIsOnline, true),
				}, {
					// TODO: Should this go to remotes?
					Name:    "list",
					Aliases: []string{"ls"},
					Action:  withDaemon(handleOnlinePeers, true),
				}, {
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
		}, {
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
		}, {
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
		}, {
			Name:     "tag",
			Category: vcscGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleTag, true)),
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "delete,d",
					Usage: "Delete the tag instead of creating it",
				},
			},
		}, {
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
		}, {
			Name:     "fetch",
			Category: vcscGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleFetch, true)),
		}, {
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
		}, {
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
		}, {
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
		}, {
			Name:     "become",
			Category: vcscGroup,
			Action:   withDaemon(handleBecome, true),
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "self,s",
					Usage: "Become self (i.e. the owner of the repository)",
				},
			},
		}, {
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
		}, {
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
		}, {
			Name:     "touch",
			Aliases:  []string{"t"},
			Category: wdirGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleTouch, true)),
		}, {
			Name:     "cat",
			Category: wdirGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleCat, true)),
		}, {
			Name:     "info",
			Category: wdirGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleInfo, true)),
		}, {
			Name:     "rm",
			Aliases:  []string{"remove"},
			Category: wdirGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleRm, true)),
		}, {
			Name:     "ls",
			Category: wdirGroup,
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
		}, {
			Name:     "tree",
			Category: wdirGroup,
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "depth, d",
					Usage: "Max depth to traverse",
					Value: -1,
				},
			},
			Action: withDaemon(handleTree, true),
		}, {
			Name:     "mkdir",
			Category: wdirGroup,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "parents, p",
					Usage: "Create parent directories as needed",
				},
			},
			Action: withArgCheck(needAtLeast(1), withDaemon(handleMkdir, true)),
		}, {
			Name:     "mv",
			Category: wdirGroup,
			Action:   withArgCheck(needAtLeast(2), withDaemon(handleMv, true)),
		}, {
			Name:     "cp",
			Category: wdirGroup,
			Action:   withArgCheck(needAtLeast(2), withDaemon(handleCp, true)),
		}, {
			Name:     "edit",
			Category: wdirGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleEdit, true)),
		}, {
			// TODO: subcommand brig bug -> collect bug report info
			Name:     "daemon",
			Category: repoGroup,
			Subcommands: []cli.Command{
				{
					Name:   "launch",
					Action: withExit(handleDaemonLaunch),
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "trace,t",
							Usage: "Create tracing output suitable for `go tool trace`",
						},
					},
				}, {
					Name:   "quit",
					Action: withDaemon(handleDaemonQuit, false),
				}, {
					Name:   "ping",
					Action: withDaemon(handleDaemonPing, false),
				},
			},
		}, {
			Name:     "config",
			Category: repoGroup,
			Subcommands: []cli.Command{
				{
					Name:   "list",
					Action: withDaemon(handleConfigList, true),
				}, {
					Name:   "get",
					Action: withArgCheck(needAtLeast(1), withDaemon(handleConfigGet, true)),
				}, {
					Name:   "set",
					Action: withArgCheck(needAtLeast(2), withDaemon(handleConfigSet, true)),
				},
			},
		}, {
			Name:     "mount",
			Category: repoGroup,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "umount,u",
					Usage: "Unmount the specified directory",
				},
			},
			Action: withDaemon(handleMount, true),
		}, {
			Name:     "unmount",
			Category: repoGroup,
			Action:   withDaemon(handleUnmount, true),
		}, {
			Name:     "version",
			Category: repoGroup,
			Action:   withDaemon(handleVersion, false),
		}, {
			Name:     "gc",
			Category: repoGroup,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "aggressive,a",
					Usage: "Also run the garbage collector on all filesystems immediately",
				},
			},
			Action: withDaemon(handleGc, true),
		},
	})

	if err := app.Run(args); err != nil {
		return 1
	}
	return 0
}
