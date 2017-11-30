package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig"
	"github.com/disorganizer/brig/util/colors"
	colorlog "github.com/disorganizer/brig/util/log"
	"github.com/urfave/cli"
)

func init() {
	log.SetOutput(os.Stderr)
	log.SetLevel(log.DebugLevel)
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

// ld compares two strings and returns the levenshtein distance between them.
// TODO: Use a proper library for that.
func levenshtein(s, t string) float64 {
	s = strings.ToLower(s)
	t = strings.ToLower(t)

	d := make([][]int, len(s)+1)
	for i := range d {
		d[i] = make([]int, len(t)+1)
	}
	for i := range d {
		d[i][0] = i
	}
	for j := range d[0] {
		d[0][j] = j
	}
	for j := 1; j <= len(t); j++ {
		for i := 1; i <= len(s); i++ {
			if s[i-1] == t[j-1] {
				d[i][j] = d[i-1][j-1]
			} else {
				min := d[i-1][j]
				if d[i][j-1] < min {
					min = d[i][j-1]
				}
				if d[i-1][j-1] < min {
					min = d[i-1][j-1]
				}
				d[i][j] = min + 1
			}
		}

	}

	total_len := len(s)
	if len(t) > total_len {
		total_len = len(t)
	}

	dist := d[len(s)][len(t)]
	return float64(dist) / float64(total_len)
}

func findLastGoodCommands(ctx *cli.Context) ([]string, []cli.Command) {
	for ctx.Parent() != nil {
		ctx = ctx.Parent()
	}

	args := ctx.Args()
	if len(args) == 0 || len(args) == 1 {
		return nil, ctx.App.Commands
	}

	cmd := ctx.App.Command(args[0])
	if cmd == nil {
		return nil, ctx.App.Commands
	}

	validArgs := []string{args[0]}
	args = args[1 : len(args)-1]

	for len(args) != 0 && cmd != nil {
		for _, subCmd := range cmd.Subcommands {
			if subCmd.Name == args[0] {
				cmd = &subCmd
			}
		}

		validArgs = append(validArgs, args[0])
		args = args[1:]
	}

	return validArgs, cmd.Subcommands
}

type suggestion struct {
	name  string
	score float64
}

func findSimilarCommands(cmdName string, cmds []cli.Command) []suggestion {
	similars := []suggestion{}

	for _, cmd := range cmds {
		candidates := []string{cmd.Name}
		candidates = append(candidates, cmd.Aliases...)

		for _, candidate := range candidates {
			score := levenshtein(cmdName, candidate)
			if score <= 0.5 {
				similars = append(similars, suggestion{
					name:  cmd.Name,
					score: score,
				})
				break
			}
		}
	}

	// Special cases for the git inclined:
	staticSuggestions := map[string]string{
		"add":  "stage",
		"pull": "sync",
	}

	for gitName, brigName := range staticSuggestions {
		if cmdName == gitName {
			similars = append(similars, suggestion{
				name:  brigName,
				score: 0.0,
			})
		}
	}

	// Let suggestions be sorted by their similarity:
	sort.Slice(similars, func(i, j int) bool {
		return similars[i].score < similars[j].score
	})

	return similars
}

func commandNotFound(ctx *cli.Context, cmdName string) {
	cmdPath, lastGoodCmds := findLastGoodCommands(ctx)
	similars := findSimilarCommands(cmdName, lastGoodCmds)

	badCmd := colors.Colorize(cmdName, colors.Red)
	if cmdPath == nil {
		// A toplevel command was wrong:
		fmt.Printf("`%s` is not a valid command. ", badCmd)
	} else {
		// A command of a subcommand was wrong:
		lastGoodSubCmd := colors.Colorize(strings.Join(cmdPath, " "), colors.Yellow)
		fmt.Printf("`%s` is not a valid subcommand of `%s`. ", badCmd, lastGoodSubCmd)
	}

	switch len(similars) {
	case 0:
		fmt.Printf("\n")
	case 1:
		suggestion := colors.Colorize(similars[0].name, colors.Green)
		fmt.Printf("Did you maybe mean `%s`?\n", suggestion)
	default:
		fmt.Println("\n\nDid you mean one of those?")
		for _, similar := range similars {
			fmt.Printf("  * %s\n", colors.Colorize(similar.name, colors.Green))
		}
	}
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
		brig.VersionString(),
		brig.BuildTime,
	)
	app.CommandNotFound = commandNotFound

	// Groups:
	repoGroup := formatGroup("repository")
	wdirGroup := formatGroup("working tree")
	vcscGroup := formatGroup("version control")
	advnGroup := formatGroup("advanced")
	miscGroup := formatGroup("misc")

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
			Value:  "stderr",
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
			Name:        "become",
			Category:    advnGroup,
			Usage:       "Act as other user and view the data we synced with",
			Description: "Act as other user and view the data we synced with",
			Action:      withArgCheck(needAtLeast(1), withDaemon(handleBecome, true)),
		},
		cli.Command{
			Name:        "whoami",
			Category:    repoGroup,
			Usage:       "Check at what user's data we're currently looking at",
			Description: "Check at what user's data we're currently looking at",
			Action:      withDaemon(handleWhoami, true),
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "fingerprint,f",
					Usage: "Only print the own fingerprint",
				},
			},
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
			Category:    wdirGroup,
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
					Name:  "is-pinned,i",
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
			Name:        "remote",
			Category:    repoGroup,
			Usage:       "Remote management.",
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
			Name:        "stage",
			Category:    wdirGroup,
			Usage:       "Transer a file into brig's control or update an existing one",
			ArgsUsage:   "<file>",
			Description: "Stage a specific file into the brig repository",
			Action:      withArgCheck(needAtLeast(1), withDaemon(handleStage, true)),
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
			Name:        "mv",
			Category:    wdirGroup,
			Usage:       "Move a specific file",
			ArgsUsage:   "<sourcefile> <destinationfile>",
			Description: "Move a file from SOURCE to DEST",
			Action:      withArgCheck(needAtLeast(2), withDaemon(handleMv, true)),
		},
		cli.Command{
			Name:        "cat",
			Category:    wdirGroup,
			Usage:       "Concatenates a file",
			ArgsUsage:   "<file>",
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
					Action:      withExit(handleDaemonLaunch),
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
			Action: withDaemon(handleMount, true),
		},
		cli.Command{
			Name:        "unmount",
			Category:    miscGroup,
			Usage:       "Unmount a previosuly mounted directory",
			ArgsUsage:   "<mountpath>",
			Description: "Unmounts a FUSE filesystem",
			Action:      withDaemon(handleUnmount, true),
		},
	}

	app.Before = func(ctx *cli.Context) error {
		return setLogPath(ctx.String("log-path"))
	}

	app.Run(args)
	return 0
}
