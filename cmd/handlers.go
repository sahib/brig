package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/brigd/client"
	"github.com/disorganizer/brig/brigd/server"
	"github.com/disorganizer/brig/cmd/pwd"
	"github.com/disorganizer/brig/util/colors"
	"github.com/disorganizer/brig/version"
	"github.com/dustin/go-humanize"
	"github.com/urfave/cli"
)

const brigLogo = `
       _____         /  /\        ___          /  /\ 
      /  /::\       /  /::\      /  /\        /  /:/_
     /  /:/\:\     /  /:/\:\    /  /:/       /  /:/ /\ 
    /  /:/~/::\   /  /:/~/:/   /__/::\      /  /:/_/::\ 
   /__/:/ /:/\:| /__/:/ /:/___ \__\/\:\__  /__/:/__\/\:\
   \  \:\/:/~/:/ \  \:\/:::::/    \  \:\/\ \  \:\ /~~/:/
    \  \::/ /:/   \  \::/~~~~      \__\::/  \  \:\  /:/
     \  \:\/:/     \  \:\          /__/:/    \  \:\/:/
      \  \::/       \  \:\         \__\/      \  \::/
       \__\/         \__\/                     \__\/


     A new file README.md was automatically added.
     Use 'brig cat README.md' to view it & get started.

`

func handleVersion(ctx *cli.Context) error {
	fmt.Println(version.String())
	return nil
}

func handleDaemonPing(ctx *cli.Context, ctl *client.Client) error {
	for i := 0; i < 100; i++ {
		before := time.Now()
		symbol := colors.Colorize("✔", colors.Green)

		if err := ctl.Ping(); err != nil {
			symbol = colors.Colorize("✘", colors.Red)
		}

		delay := time.Since(before)
		fmt.Printf("#%02d %s ➔ %s: %s (%v)\n",
			i+1,
			ctl.LocalAddr().String(),
			ctl.RemoteAddr().String(),
			symbol,
			delay,
		)

		time.Sleep(1 * time.Second)
	}

	return nil
}

func handleDaemonQuit(ctx *cli.Context, ctl *client.Client) error {
	if err := ctl.Quit(); err != nil {
		return ExitCode{
			DaemonNotResponding,
			fmt.Sprintf("brigd not responding: %v", err),
		}
	}

	return nil
}

func handleDaemonLaunch(ctx *cli.Context) error {
	brigPath := os.Getenv("BRIG_PATH")
	if brigPath == "" {
		// TODO: Check parent directories to see if we're in some
		//       brig repository.
		brigPath = "."
	}

	// If the repository was not initialized yet,
	// we should not ask for a password, since init
	// will already ask for one. If we recognize the repo
	// wrongly as uninitialized, then it won't unlock without
	// a password though.
	if !repoIsInitialized(brigPath) {
		log.Infof(
			"No repository found at %s. Use `brig init <user>` to create one",
			brigPath,
		)
	}

	password, err := readPassword(ctx, brigPath)
	if err != nil {
		msg := fmt.Sprintf("Failed to read password: %v", err)
		fmt.Println(msg)
		return ExitCode{UnknownError, msg}
	}

	server, err := server.BootServer(brigPath, password, guessPort())
	if err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("Failed to boot brigd: %v", err),
		}
	}

	defer server.Close()

	if err := server.Serve(); err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("Failed to serve: %v", err),
		}
	}

	return nil
}

func handleMount(ctx *cli.Context, ctl *client.Client) error {
	mountPath := ctx.Args().First()
	if err := ctl.Mount(mountPath); err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("Failed to mount: %v", err),
		}
	}

	return nil
}

func handleUnmount(ctx *cli.Context, ctl *client.Client) error {
	mountPath := ctx.Args().First()
	if err := ctl.Unmount(mountPath); err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("Failed to unmount: %v", err),
		}
	}

	return nil
}

func createInitialReadme(ctl *client.Client) error {
	text := `Welcome to brig!

Here's what you can do next:

    * Add a few remotes to sync with (See 'brig remote add -h')
    * Mount your data somewhere convinient (See 'brig mount -h')
    * Have a relaxing day exploring brig's features.

Please remember that brig is software in it's very early stages,
and will currently eat your data with near-certainity.

If you're done with this README, you can easily remove it:

    $ brig rm README.md

`

	fd, err := ioutil.TempFile("", ".brig-init-readme-")
	if err != nil {
		return err
	}

	if _, err := fd.WriteString(text); err != nil {
		return err
	}

	readmePath := fd.Name()

	if err := fd.Close(); err != nil {
		return err
	}

	if err := ctl.Stage(readmePath, "/README.md"); err != nil {
		return err
	}

	return ctl.MakeCommit("Added initial README.md")
}

func handleInit(ctx *cli.Context, ctl *client.Client) error {
	// Accumulate args:
	owner := ctx.Args().First()
	folder := guessRepoFolder()
	backend := ctx.String("backend")
	password := readPasswordFromArgs(ctx)

	if password == "" {
		pwdBytes, err := pwd.PromptNewPassword(25)
		if err != nil {
			msg := fmt.Sprintf("Failed to read password: %v", err)
			fmt.Println(msg)
			return ExitCode{UnknownError, msg}
		}

		password = string(pwdBytes)
	}

	if err := ctl.Init(folder, owner, password, backend); err != nil {
		return ExitCode{UnknownError, fmt.Sprintf("init failed: %v", err)}
	}

	if err := createInitialReadme(ctl); err != nil {
		return err
	}

	fmt.Println(brigLogo)
	return nil
}

func handleConfigList(cli *cli.Context, ctl *client.Client) error {
	all, err := ctl.ConfigAll()
	if err != nil {
		return ExitCode{UnknownError, fmt.Sprintf("config list: %v", err)}
	}

	// Display the output nicely sorted:
	keys := []string{}
	for key := range all {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		fmt.Printf(
			"%s: %s\n",
			colors.Colorize(key, colors.Green),
			all[key],
		)
	}
	return nil
}

func handleConfigGet(ctx *cli.Context, ctl *client.Client) error {
	key := ctx.Args().Get(0)
	val, err := ctl.ConfigGet(key)
	if err != nil {
		return ExitCode{UnknownError, fmt.Sprintf("config get: %v", err)}
	}

	fmt.Println(val)
	return nil
}

func handleConfigSet(ctx *cli.Context, ctl *client.Client) error {
	key := ctx.Args().Get(0)
	val := ctx.Args().Get(1)
	if err := ctl.ConfigSet(key, val); err != nil {
		return ExitCode{UnknownError, fmt.Sprintf("config set: %v", err)}
	}

	return nil
}

func handleStage(ctx *cli.Context, ctl *client.Client) error {
	localPath := ctx.Args().Get(0)

	repoPath := filepath.Base(localPath)
	if len(ctx.Args()) > 1 {
		repoPath = ctx.Args().Get(1)
	}

	if err := ctl.Stage(localPath, repoPath); err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("stage: %v", err),
		}
	}
	return nil
}

func handleCat(ctx *cli.Context, ctl *client.Client) error {
	stream, err := ctl.Cat(ctx.Args().First())
	if err != nil {
		// TODO: Make those exit codes a wrapper function.
		return ExitCode{
			UnknownError,
			fmt.Sprintf("cat: %v", err),
		}
	}

	defer stream.Close()

	if _, err := io.Copy(os.Stdout, stream); err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("cat: %v", err),
		}
	}

	return nil
}

func handleRm(ctx *cli.Context, ctl *client.Client) error {
	path := ctx.Args().First()

	if err := ctl.Remove(path); err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("rm: %v", err),
		}
	}

	return nil
}

func handleMv(ctx *cli.Context, ctl *client.Client) error {
	srcPath := ctx.Args().Get(0)
	dstPath := ctx.Args().Get(1)

	if err := ctl.Move(srcPath, dstPath); err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("mv: %v", err),
		}
	}

	return nil
}

func handleOffline(ctx *cli.Context, ctl *client.Client) error {
	return nil
}

func handleIsOnline(ctx *cli.Context, ctl *client.Client) error {
	return nil
}

func handleOnline(ctx *cli.Context, ctl *client.Client) error {
	return nil
}

func handleList(ctx *cli.Context, ctl *client.Client) error {
	maxDepth := ctx.Int("depth")
	if ctx.Bool("recursive") {
		maxDepth = -1
	}

	root := "/"
	if ctx.Args().Present() {
		root = ctx.Args().First()
	}

	entries, err := ctl.List(root, maxDepth)
	if err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("ls: %v", err),
		}
	}

	for _, entry := range entries {
		fmt.Printf(
			"%6s %8s  %s\n",
			humanize.Bytes(entry.Size),
			entry.ModTime.Format(time.Stamp),
			entry.Path,
		)
	}

	return nil
}

func handleTree(ctx *cli.Context, ctl *client.Client) error {
	entries, err := ctl.List("/", -1)
	if err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("ls: %v", err),
		}
	}

	return showTree(entries, -1)
}

func handleMkdir(ctx *cli.Context, ctl *client.Client) error {
	path := ctx.Args().First()
	createParents := ctx.Bool("parents")

	if err := ctl.Mkdir(path, createParents); err != nil {
		return ExitCode{UnknownError, fmt.Sprintf("mkdir: %v", err)}
	}

	return nil
}

func handleCommit(ctx *cli.Context, ctl *client.Client) error {
	msg := ""
	if msg = ctx.String("message"); msg == "" {
		msg = fmt.Sprintf("manual commit")
	}

	if err := ctl.MakeCommit(msg); err != nil {
		return ExitCode{UnknownError, fmt.Sprintf("commit: %v", err)}
	}

	return nil
}

func handleTag(ctx *cli.Context, ctl *client.Client) error {
	if ctx.Bool("delete") {
		name := ctx.Args().Get(0)

		if err := ctl.Untag(name); err != nil {
			return ExitCode{
				UnknownError,
				fmt.Sprintf("untag: %v", err),
			}
		}
	} else {
		if len(ctx.Args()) < 2 {
			return ExitCode{BadArgs, "tag needs at least two arguments"}
		}

		rev := ctx.Args().Get(0)
		name := ctx.Args().Get(1)

		if err := ctl.Tag(rev, name); err != nil {
			return ExitCode{
				UnknownError,
				fmt.Sprintf("tag: %v", err),
			}
		}
	}

	return nil
}

func handleLog(ctx *cli.Context, ctl *client.Client) error {
	entries, err := ctl.Log()
	if err != nil {
		return ExitCode{UnknownError, fmt.Sprintf("commit: %v", err)}
	}

	for idx, entry := range entries {
		tags := ""
		if len(entry.Tags) > 0 {
			tags = fmt.Sprintf(" (%s)", strings.Join(entry.Tags, ", "))
		}

		msg := entry.Msg
		if msg == "" {
			msg = colors.Colorize("*", colors.Red)
		}

		entry.Hash.ShortB58()

		fmt.Printf(
			"%2d: %s %s %s%s\n",
			idx,
			colors.Colorize(entry.Hash.ShortB58(), colors.Green),
			colors.Colorize(entry.Date.Format(time.Stamp), colors.Yellow),
			msg,
			colors.Colorize(tags, colors.Cyan),
		)
	}

	return nil
}

func handlePin(ctx *cli.Context, ctl *client.Client) error {
	if ctx.Bool("is-pinned") {
		return handleIsPinned(ctx, ctl)
	}

	if ctx.Bool("unpin") {
		return handleUnpin(ctx, ctl)
	}

	path := ctx.Args().First()
	if err := ctl.Pin(path); err != nil {
		return ExitCode{UnknownError, fmt.Sprintf("pin: %v", err)}
	}

	return nil
}

func handleUnpin(ctx *cli.Context, ctl *client.Client) error {
	path := ctx.Args().First()
	if err := ctl.Unpin(path); err != nil {
		return ExitCode{UnknownError, fmt.Sprintf("unpin: %v", err)}
	}

	return nil
}

func handleIsPinned(ctx *cli.Context, ctl *client.Client) error {
	path := ctx.Args().First()
	isPinned, err := ctl.IsPinned(path)
	if err != nil {
		return ExitCode{UnknownError, fmt.Sprintf("unpin: %v", err)}
	}

	fmt.Println(isPinned)
	return nil
}

func handleReset(ctx *cli.Context, ctl *client.Client) error {
	path := ctx.Args().First()
	rev := "HEAD"

	if len(ctx.Args()) > 1 {
		rev = ctx.Args().Get(1)
	}

	if err := ctl.Reset(path, rev); err != nil {
		return ExitCode{UnknownError, fmt.Sprintf("unpin: %v", err)}
	}

	return nil
}

func handleCheckout(ctx *cli.Context, ctl *client.Client) error {
	rev := ctx.Args().First()

	if err := ctl.Checkout(rev, ctx.Bool("force")); err != nil {
		return ExitCode{UnknownError, fmt.Sprintf("checkout: %v", err)}
	}

	return nil
}

func handleHistory(ctx *cli.Context, ctl *client.Client) error {
	path := ctx.Args().First()

	history, err := ctl.History(path)
	if err != nil {
		return ExitCode{UnknownError, fmt.Sprintf("history: %v", err)}
	}

	for _, entry := range history {
		fmt.Printf(
			"%s %-15s %s\n",
			colors.Colorize(entry.Ref.B58String()[:10], colors.Red),
			colors.Colorize(entry.Change, colors.Yellow),
			colors.Colorize(entry.Path, colors.Green),
		)
	}

	return nil
}

func printDiff(diff *client.Diff) {
	simpleSection := func(heading string, infos []client.StatInfo) {
		if len(infos) == 0 {
			return
		}

		fmt.Println(heading)
		for _, info := range diff.Added {
			fmt.Printf("  %s\n", info.Path)
		}

		fmt.Println()
	}

	pairSection := func(heading string, infos []client.DiffPair) {
		if len(infos) == 0 {
			return
		}

		for _, pair := range diff.Merged {
			fmt.Printf("  %s <-> %s\n", pair.Src.Path, pair.Dst.Path)
		}

		fmt.Println()
	}

	simpleSection(colors.Colorize("Added:", colors.Green), diff.Added)
	simpleSection(colors.Colorize("Ignored:", colors.Yellow), diff.Ignored)
	simpleSection(colors.Colorize("Removed:", colors.Red), diff.Removed)

	pairSection(colors.Colorize("Resolveable Conflicts:", colors.Cyan), diff.Merged)
	pairSection(colors.Colorize("Conflicts:", colors.Magenta), diff.Conflict)
}

func handleDiff(ctx *cli.Context, ctl *client.Client) error {
	remoteRev := ctx.Args().Get(0)
	if remoteRev == "" {
		remoteRev = "HEAD"
	}

	localRev := ctx.String("rev")
	remoteName := ctx.String("remote")
	if remoteName == "" {
		self, err := ctl.Whoami()
		if err != nil {
			return err
		}

		remoteName = self.Owner
	}

	diff, err := ctl.MakeDiff(remoteName, localRev, remoteRev)
	if err != nil {
		return ExitCode{UnknownError, fmt.Sprintf("diff: %v", err)}
	}

	printDiff(diff)
	return nil
}

func handleFetch(ctx *cli.Context, ctl *client.Client) error {
	who := ctx.Args().First()
	return ctl.Fetch(who)
}

func handleSync(ctx *cli.Context, ctl *client.Client) error {
	who := ctx.Args().First()

	needFetch := true
	if ctx.Bool("no-fetch") {
		needFetch = false
	}

	return ctl.Sync(who, needFetch)
}

func handleStatus(ctx *cli.Context, ctl *client.Client) error {
	self, err := ctl.Whoami()
	if err != nil {
		return err
	}

	diff, err := ctl.MakeDiff(self.Owner, "HEAD", "CURR")
	if err != nil {
		return err
	}

	printDiff(diff)
	return nil
}

func handleBecome(ctx *cli.Context, ctl *client.Client) error {
	who := ctx.Args().First()
	if err := ctl.Become(who); err != nil {
		return err
	}

	fmt.Printf(
		"You are viewing %s's data now. Changes will be local only.\n",
		colors.Colorize(who, colors.Green),
	)
	return nil
}

func handleWhoami(ctx *cli.Context, ctl *client.Client) error {
	self, err := ctl.Whoami()
	if err != nil {
		return err
	}

	if !ctx.Bool("fingerprint") {
		userName := colors.Colorize(self.CurrentUser, colors.Yellow)
		ownerName := colors.Colorize(self.Owner, colors.Green)
		fmt.Printf("%s", ownerName)
		if self.CurrentUser != self.Owner {
			fmt.Printf(" (viewing %s's data)", userName)
		}

		fmt.Printf(" ")
	}

	fmt.Printf("%s\n", self.Fingerprint)
	return nil
}
