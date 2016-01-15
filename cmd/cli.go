package cmdline

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig"
	"github.com/disorganizer/brig/daemon"
	"github.com/disorganizer/brig/fuse"
	"github.com/disorganizer/brig/repo"
	"github.com/disorganizer/brig/repo/config"
	"github.com/disorganizer/brig/util/colors"
	colorlog "github.com/disorganizer/brig/util/log"
	yamlConfig "github.com/olebedev/config"
	"github.com/tsuibin/goxmpp2/xmpp"
	"github.com/tucnak/climax"
)

func init() {
	log.SetOutput(os.Stderr)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)

	// Log pretty text
	log.SetFormatter(&colorlog.ColorfulLogFormatter{})
}

///////////////////////
// Utility functions //
///////////////////////

func formatGroup(category string) string {
	return strings.ToUpper(category) + " COMMANDS:"
}

// guessRepoFolder tries to find the repository path
// by using a number of sources.
func guessRepoFolder() string {
	folder := repo.GuessFolder()
	if folder == "" {
		log.Fatalf("This does not like a brig repository (missing .brig)")
	}

	return folder
}

func readPassword() (string, error) {
	repoFolder := guessRepoFolder()
	pwd, err := repo.PromptPasswordMaxTries(4, func(pwd string) bool {
		err := repo.CheckPassword(repoFolder, pwd)
		return err == nil
	})

	return pwd, err
}

///////////////////////
// Handler functions //
///////////////////////

func handleVersion(ctx climax.Context) int {
	fmt.Println(brig.VersionString())
	return Success
}

func handleOpen(ctx climax.Context, client *daemon.Client) int {
	log.Infof("Repository is open now.")
	return Success
}

func handleClose(ctx climax.Context) int {
	// This is currently the same as `brig daemon -q`
	return handleDaemonQuit()
}

func handleDaemonPing() int {
	client, err := daemon.Dial(6666)
	if err != nil {
		log.Warning("Unable to dial to daemon: ", err)
		return DaemonNotResponding
	}
	defer client.Close()

	for i := 0; i < 100; i++ {
		before := time.Now()
		symbol := colors.Colorize("✔", colors.Green)
		if !client.Ping() {
			symbol = colors.Colorize("✘", colors.Red)
		}

		delay := time.Since(before)

		fmt.Printf("#%02d %s ➔ %s: %s (%v)\n",
			i+1,
			client.LocalAddr().String(),
			client.RemoteAddr().String(),
			symbol, delay)
		time.Sleep(1 * time.Second)
	}

	return Success
}

func handleDaemonQuit() int {
	client, err := daemon.Dial(6666)
	if err != nil {
		log.Warning("Unable to dial to daemon: ", err)
		return DaemonNotResponding
	}
	defer client.Close()

	client.Exorcise()
	return Success
}

func handleDaemon(ctx climax.Context) int {
	if ctx.Is("ping") {
		return handleDaemonPing()
	} else if ctx.Is("quit") {
		return handleDaemonQuit()
	}

	pwd, ok := ctx.Get("password")
	if !ok {
		var err error
		pwd, err = readPassword()
		if err != nil {
			log.Errorf("Could not read password: %v", pwd)
			return BadPassword
		}
	}

	repoFolder := guessRepoFolder()
	err := repo.CheckPassword(repoFolder, pwd)
	if err != nil {
		log.Error("Wrong password.")
		return BadPassword
	}

	baal, err := daemon.Summon(pwd, repoFolder, 6666)
	if err != nil {
		log.Warning("Unable to start daemon: ", err)
		return UnknownError
	}

	baal.Serve()
	return Success
}

func handleMount(ctx climax.Context) int {
	mntpath := ctx.Args[0]
	if err := fuse.Mount(mntpath); err != nil {
		log.Errorf("Unable to mount: %v", err)
		return UnknownError
	}

	return Success
}

func handleConfig(ctx climax.Context) int {
	folder := guessRepoFolder()
	cfgPath := filepath.Join(folder, ".brig", "config")

	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		log.Errorf("Could not load config: %v", err)
		return BadArgs
	}

	switch len(ctx.Args) {
	case 0:
		// No key or value. Print whole config as .yaml
		yaml, err := yamlConfig.RenderYaml(cfg)
		if err != nil {
			log.Errorf("Unable to render config: %v", err)
			return UnknownError
		}
		fmt.Println(yaml)
	case 1:
		// Get requested; find value for key.
		key := ctx.Args[0]
		value, err := cfg.String(key)
		if err != nil {
			log.Errorf("Could not retrieve %s: %v", key, err)
			return BadArgs
		}
		fmt.Println(value)
	case 2:
		// Set requested: set key to value.
		key := ctx.Args[0]
		value := ctx.Args[1]
		if err := cfg.Set(key, value); err != nil {
			log.Errorf("Could not set %s: %v", key, err)
			return BadArgs
		}

		if _, err := config.SaveConfig(cfgPath, cfg); err != nil {
			log.Errorf("Could not save config: %v", err)
			return UnknownError
		}
	}

	return Success
}

func handleInit(ctx climax.Context) int {
	jid := xmpp.JID(ctx.Args[0])
	if jid.Domain() == "" {
		log.Error("Your JID needs to have a domain.")
		return BadArgs
	}

	// Extract the folder from the resource name by default:
	folder := jid.Resource()
	if folder == "" {
		log.Error("Need a resource in your JID.")
		return BadArgs
	}

	if envFolder := os.Getenv("BRIG_PATH"); envFolder != "" {
		folder = envFolder
	}

	if ctx.Is("folder") {
		folder, _ = ctx.Get("folder")
	}

	pwd, ok := ctx.Get("password")
	if !ok {
		var err error
		pwdBytes, err := repo.PromptNewPassword(40.0)
		if err != nil {
			log.Error(err)
			return BadPassword
		}

		pwd = string(pwdBytes)
	}

	repo, err := repo.NewRepository(string(jid), pwd, folder)
	if err != nil {
		log.Error(err)
		return UnknownError
	}

	if err := repo.Close(); err != nil {
		log.Errorf("close: %v", err)
		return UnknownError
	}

	if !ctx.Is("nodaemon") {
		if _, err := daemon.Reach(string(pwd), folder, 6666); err != nil {
			log.Errorf("Unable to start daemon: %v", err)
			return DaemonNotResponding
		}
	}

	return Success
}

func handleAdd(ctx climax.Context, client *daemon.Client) int {
	for _, path := range ctx.Args {
		absPath, err := filepath.Abs(path)
		if err != nil {
			log.Errorf("Unable to make abs path: %v: %v", path, err)
			continue
		}

		hash, err := client.Add(absPath)
		if err != nil {
			log.Errorf("Could not add file: %v: %v", absPath, err)
			return UnknownError
		}

		fmt.Println(hash.B58String())
	}

	return Success
}

func handleCat(ctx climax.Context, client *daemon.Client) int {
	dstPath, err := filepath.Abs(ctx.Args[0])
	absPath, err := filepath.Abs(ctx.Args[1])
	if err != nil {
		log.Errorf("Unable to make abs path: %v: %v", absPath, err)
		return UnknownError
	}

	newPath, err := client.Cat(dstPath, absPath)
	if err != nil {
		log.Errorf("Could not cat file: %v: %v", absPath, err)
		return UnknownError
	}

	fmt.Println(newPath)
	return Success
}

////////////////////////////
// Commandline definition //
////////////////////////////

// RunCmdline starts a brig commandline tool.
func RunCmdline() int {
	demo := climax.New("brig")
	demo.Brief = "brig is a decentralized file syncer based on IPFS and XMPP."
	demo.Version = "unstable"

	repoGroup := demo.AddGroup(formatGroup("repository"))
	xmppGroup := demo.AddGroup(formatGroup("xmpp helper"))
	wdirGroup := demo.AddGroup(formatGroup("working"))
	advnGroup := demo.AddGroup(formatGroup("advanced"))
	miscGroup := demo.AddGroup(formatGroup("misc"))

	commands := []climax.Command{
		climax.Command{
			Name:  "init",
			Brief: "Initialize an empty repository and open it",
			Group: repoGroup,
			Usage: `<JID> [<PATH>]`,
			Help:  `Create an empty repository, open it and associate it with the JID`,
			Flags: []climax.Flag{
				{
					Name:     "depth",
					Short:    "o",
					Usage:    `--depth="N"`,
					Help:     `Only clone up to this depth of pinned files`,
					Variable: true,
				}, {
					Name:  "nodaemon",
					Short: "n",
					Help:  `Do not start the daemon.`,
				}, {
					Name:     "password",
					Short:    "x",
					Usage:    `--password PWD`,
					Help:     `Supply password.`,
					Variable: true,
				},
			},
			Examples: []climax.Example{
				{
					Usecase:     `alice@jabber.de/laptop`,
					Description: `Create a folder laptop/ with hidden directories`,
				},
			},
			Handle: withArgCheck(needAtLeast(1), handleInit),
		},
		climax.Command{
			Name:  "clone",
			Brief: "Clone an repository from somebody else",
			Group: repoGroup,
			Usage: `<OTHER_JID> <YOUR_JID> [<PATH>]`,
			Help:  `...`,
			Flags: []climax.Flag{
				{
					Name:     "--depth",
					Short:    "d",
					Usage:    `--depth="N"`,
					Help:     `Only clone up to this depth of pinned files`,
					Variable: true,
				},
			},
			Examples: []climax.Example{
				{
					Usecase:     `alice@jabber.de/laptop bob@jabber.de/desktop`,
					Description: `Clone Alice' contents`,
				},
			},
		},
		climax.Command{
			Name:   "open",
			Group:  repoGroup,
			Brief:  "Open an encrypted port. Asks for passphrase.",
			Handle: withDaemon(handleOpen),
		},
		climax.Command{
			Name:   "close",
			Group:  repoGroup,
			Brief:  "Encrypt all metadata in the port and go offline.",
			Handle: handleClose,
		},
		climax.Command{
			Name:  "sync",
			Group: repoGroup,
			Brief: "Sync with all or selected trusted peers.",
		},
		climax.Command{
			Name:  "push",
			Group: repoGroup,
			Brief: "Push your content to all or selected trusted peers.",
		},
		climax.Command{
			Name:  "pull",
			Group: repoGroup,
			Brief: "Pull content from all or selected trusted peers.",
		},
		climax.Command{
			Name:  "watch",
			Group: repoGroup,
			Brief: "Enable or disable watch mode.",
		},
		climax.Command{
			Name:  "discover",
			Group: xmppGroup,
			Brief: "Try to find other brig users near you.",
		},
		climax.Command{
			Name:  "friends",
			Group: xmppGroup,
			Brief: "List your trusted peers.",
		},
		climax.Command{
			Name:  "beg",
			Group: xmppGroup,
			Brief: "Request authorisation from a buddy.",
		},
		climax.Command{
			Name:  "ban",
			Group: xmppGroup,
			Brief: "Discontinue friendship with a peer.",
		},
		climax.Command{
			Name:  "prio",
			Group: xmppGroup,
			Brief: "Change priority of a peer.",
		},
		climax.Command{
			Name:  "status",
			Group: wdirGroup,
			Brief: "Give an overview of brig's current state.",
		},
		climax.Command{
			Name:   "add",
			Group:  wdirGroup,
			Brief:  "Make file to be managed by brig.",
			Usage:  `FILE_OR_FOLDER [FILE_OR_FOLDER ...]`,
			Help:   `TODO`,
			Handle: withArgCheck(needAtLeast(1), withDaemon(handleAdd)),
		},
		climax.Command{
			Name:   "get",
			Group:  wdirGroup,
			Brief:  "Write ",
			Usage:  `FILE_OR_FOLDER DEST_PATH`,
			Help:   `TODO`,
			Handle: withArgCheck(needAtLeast(2), withDaemon(handleCat)),
		},
		climax.Command{
			Name:  "find",
			Group: wdirGroup,
			Brief: "Find filenames in the fleet.",
		},
		climax.Command{
			Name:  "rm",
			Group: wdirGroup,
			Brief: "Remove file from brig's control.",
		},
		climax.Command{
			Name:  "log",
			Group: wdirGroup,
			Brief: "Visualize changelog tree.",
		},
		climax.Command{
			Name:  "checkout",
			Group: wdirGroup,
			Brief: "Attempt to checkout previous version of a file.",
		},
		climax.Command{
			Name:  "fsck",
			Group: advnGroup,
			Brief: "Verify, and possibly fix, broken files.",
		},
		climax.Command{
			Name:  "daemon",
			Group: advnGroup,
			Brief: "Manually run the daemon process.",
			Flags: []climax.Flag{
				{
					Name:  "ping",
					Short: "p",
					Usage: `--ping`,
					Help:  `Ping the dameon to check if it's running.`,
				},
				{
					Name:  "quit",
					Short: "q",
					Usage: `--quit`,
					Help:  `Kill a running daemon.`,
				},
				{
					Name:     "password",
					Short:    "x",
					Usage:    `--password PWD`,
					Help:     `Supply password.`,
					Variable: true,
				},
			},
			Handle: handleDaemon,
		},
		climax.Command{
			Name:  "passwd",
			Group: advnGroup,
			Brief: "Set your XMPP and access password.",
		},
		climax.Command{
			Name:  "yubi",
			Group: advnGroup,
			Brief: "Manage YubiKeys.",
		},
		climax.Command{
			Name:   "config",
			Group:  miscGroup,
			Brief:  "Access, list and modify configuration values.",
			Handle: handleConfig,
		},
		climax.Command{
			Name:   "mount",
			Group:  miscGroup,
			Brief:  "Handle FUSE mountpoints.",
			Handle: withArgCheck(needAtLeast(1), handleMount),
		},
		climax.Command{
			Name:  "update",
			Group: miscGroup,
			Brief: "Try to securely update brig.",
		},
		climax.Command{
			Name:  "help",
			Group: miscGroup,
			Brief: "Print some help",
			Usage: "Did you really need help on help?",
		},
		climax.Command{
			Name:   "version",
			Group:  miscGroup,
			Brief:  "Print current version.",
			Usage:  "Print current version.",
			Handle: handleVersion,
		},
	}

	for _, command := range commands {
		demo.AddCommand(command)
	}

	// Help topics:
	demo.AddTopic(climax.Topic{
		Name:  "quickstart",
		Brief: "A very short introduction to brig",
		Text:  "TODO: write.",
	})
	demo.AddTopic(climax.Topic{
		Name:  "tutorial",
		Brief: "A slightly longer introduction.",
		Text:  "TODO: write.",
	})
	demo.AddTopic(climax.Topic{
		Name:  "terms",
		Brief: "Cheat sheet for often used terms.",
		Text:  "TODO: write.",
	})

	return demo.Run()
}
