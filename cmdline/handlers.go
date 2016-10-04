package cmdline

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/disorganizer/brig"
	"github.com/disorganizer/brig/daemon"
	"github.com/disorganizer/brig/id"
	"github.com/disorganizer/brig/repo"
	repoconfig "github.com/disorganizer/brig/repo/config"
	"github.com/disorganizer/brig/store"
	storewire "github.com/disorganizer/brig/store/wire"
	"github.com/disorganizer/brig/util"
	"github.com/disorganizer/brig/util/colors"
	pwdutil "github.com/disorganizer/brig/util/pwd"
	"github.com/dustin/go-humanize"
	multihash "github.com/jbenet/go-multihash"
	"github.com/olebedev/config"
)

func handleVersion(ctx *cli.Context) error {
	fmt.Println(brig.VersionString())
	return nil
}

func handleOpen(ctx *cli.Context, client *daemon.Client) error {
	log.Infof("Repository is open now.")
	return nil
}

func handleClose(ctx *cli.Context, client *daemon.Client) error {
	// This is currently the same as `brig daemon-quit`
	return handleDaemonQuit(ctx, client)
}

func handleDaemonPing(ctx *cli.Context, client *daemon.Client) error {
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

	return nil
}

func handleDaemonWait(ctx *cli.Context) error {
	port := guessPort()

	for {
		client, err := daemon.Dial(port)
		if err == nil {
			client.Close()
			return nil
		}

		time.Sleep(500 * time.Millisecond)
	}
}

func handleDaemonQuit(ctx *cli.Context, client *daemon.Client) error {
	client.Exorcise()
	return nil
}

func handleDaemon(ctx *cli.Context) error {
	pwd := ctx.GlobalString("password")
	if pwd == "" {
		var err error
		pwd, err = readPassword()
		if err != nil {
			return ExitCode{
				BadPassword,
				fmt.Sprintf("Could not read password: %v", pwd),
			}
		}
	}

	repoFolder := guessRepoFolder()
	err := repo.CheckPassword(repoFolder, pwd)
	if err != nil {
		return ExitCode{
			BadPassword,
			"Wrong password",
		}
	}

	port := guessPort()
	baal, err := daemon.Summon(pwd, repoFolder, port)
	if err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("Unable to start daemon: %v", err),
		}
	}

	baal.Serve()
	return nil
}

func handleMount(ctx *cli.Context, client *daemon.Client) error {
	mountPath := ""
	if len(ctx.Args()) > 0 {
		mountPath = ctx.Args()[0]
	} else {
		mountPath = filepath.Join(guessRepoFolder(), "mount")
		if err := os.Mkdir(mountPath, 0744); err != nil {
			return err
		}
	}

	var err error

	if ctx.Bool("unmount") {
		err = client.Unmount(mountPath)
	} else {
		err = client.Mount(mountPath)
	}

	if err != nil {
		return ExitCode{UnknownError, fmt.Sprintf("fuse: %v", err)}
	}

	return nil
}

func handleConfigList(cli *cli.Context, cfg *config.Config) error {
	yaml, err := config.RenderYaml(cfg)
	if err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("Unable to render config: %v", err),
		}
	}
	fmt.Println(yaml)
	return nil
}

func handleConfigGet(ctx *cli.Context, cfg *config.Config) error {
	key := ctx.Args()[0]
	value, err := cfg.String(key)
	if err != nil {
		return ExitCode{
			BadArgs,
			fmt.Sprintf("Could not retrieve %s: %v", key, err),
		}
	}
	fmt.Println(value)
	return nil
}

func handleConfigSet(ctx *cli.Context, cfg *config.Config) error {
	key := ctx.Args()[0]
	value := ctx.Args()[1]
	if err := cfg.Set(key, value); err != nil {
		return ExitCode{
			BadArgs,
			fmt.Sprintf("Could not set %s: %v", key, err),
		}
	}

	folder := repo.GuessFolder()
	if _, err := repoconfig.SaveConfig(filepath.Join(folder, ".brig", "config"), cfg); err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("Could not save config: %v", err),
		}
	}
	return nil
}

func handleInit(ctx *cli.Context) error {
	ID, err := id.Cast(ctx.Args()[0])
	if err != nil {
		return ExitCode{
			BadArgs,
			fmt.Sprintf("Bad ID: %v", err),
		}
	}

	// Extract the folder from the resource name by default:
	folder := ctx.GlobalString("path")
	fmt.Println("Folder:", folder)
	if folder == "." {
		folder = ID.AsPath()
	}

	pwd := ctx.GlobalString("password")
	fmt.Println(pwd)
	if pwd == "" {
		// TODO: Lower this or make it at least configurable
		pwdBytes, err := pwdutil.PromptNewPassword(40.0)
		if err != nil {
			return ExitCode{BadPassword, err.Error()}
		}

		pwd = string(pwdBytes)
	}

	log.Warning("============================================")
	log.Warning("Please make sure to remember the passphrase!")
	log.Warning("Your data will not be accessible without it.")
	log.Warning("============================================")

	repo, err := repo.NewRepository(string(ID), pwd, folder)
	if err != nil {
		return ExitCode{UnknownError, err.Error()}
	}

	if err := repo.Close(); err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("close: %v", err),
		}
	}

	if !ctx.GlobalBool("nodaemon") {
		port, err := repo.Config.Int("daemon.port")
		if err != nil {
			return ExitCode{UnknownError, "Unable to find out port"}
		}

		if _, err := daemon.Reach(string(pwd), folder, port); err != nil {
			return ExitCode{
				DaemonNotResponding,
				fmt.Sprintf("Unable to start daemon: %v", err),
			}
		}
	}

	return nil
}

func handleAdd(ctx *cli.Context, client *daemon.Client) error {
	filePath, err := filepath.Abs(ctx.Args()[0])
	if err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("Unable to make abs path: %v: %v", filePath, err),
		}
	}

	// Assume "/file.png" for file.png as repo path, if none given.
	repoPath := "/" + filepath.Base(filePath)
	if ctx.NArg() > 1 {
		repoPath = ctx.Args()[1]
	}

	if err := client.Add(filePath, repoPath); err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("Could not add file: %v: %v", filePath, err),
		}
	}

	fmt.Println(repoPath)
	return nil
}

func handleRm(ctx *cli.Context, client *daemon.Client) error {
	repoPath := prefixSlash(ctx.Args()[0])

	if err := client.Remove(repoPath, ctx.Bool("recursive")); err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("Could not remove file: `%s`: %v", repoPath, err),
		}
	}

	return nil
}

func handleCat(ctx *cli.Context, client *daemon.Client) error {
	repoPath := prefixSlash(ctx.Args()[0])

	filePath := ""
	isStdoutMode := ctx.NArg() < 2

	if isStdoutMode {
		tmpFile, err := ioutil.TempFile("", ".brig-tmp-")
		if err != nil {
			return ExitCode{
				UnknownError,
				fmt.Sprintf("Unable to create temp file: %v", err),
			}
		}

		filePath = tmpFile.Name()
		defer util.Closer(tmpFile)
		defer func() {
			if err := os.Remove(filePath); err != nil {
				log.Warningf("Cannot remove temp-file: %v", err)
			}
		}()
	} else {
		absPath, err := filepath.Abs(ctx.Args()[1])
		if err != nil {
			return ExitCode{
				UnknownError,
				fmt.Sprintf("Unable to make abs path: %v: %v", filePath, err),
			}
		}

		filePath = absPath
	}

	if err := client.Cat(repoPath, filePath); err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("Could not cat file: %v: %v", repoPath, err),
		}
	}

	if isStdoutMode {
		fd, err := os.Open(filePath)
		if err != nil {
			return ExitCode{
				UnknownError,
				"Could not open temp file",
			}
		}

		if _, err := io.Copy(os.Stdout, fd); err != nil {
			return ExitCode{
				UnknownError,
				fmt.Sprintf("Cannot copy to stdout: %v", err),
			}
		}

		if err := fd.Close(); err != nil {
			log.Warningf("Unable to close tmpfile handle: %v", err)
		}
	}

	return nil
}
func printCheckpoint(checkpoint *store.Checkpoint, idx, historylen int) {

	threeWayRune, twoWayRune := treeRuneTri, treeRunePipe
	if idx == historylen-1 {
		threeWayRune, twoWayRune = treeRuneCorner, " "
	}

	fmt.Printf(
		" %s%s %s #%d (%s by %s)\n",
		threeWayRune,
		treeRuneBar,
		colors.Colorize("Checkpoint", colors.Cyan),
		historylen-idx,
		colors.Colorize(checkpoint.ChangeType().String(), colors.Red),
		colors.Colorize(string(checkpoint.Author()), colors.Magenta),
	)

	fmt.Printf(
		" %s   ├─ % 9s: %v\n",
		twoWayRune,
		colors.Colorize("Hash", colors.Green),
		checkpoint.Hash().B58String(),
	)

	fmt.Printf(
		" %s   └─ % 9s: %v\n",
		twoWayRune,
		colors.Colorize("Date", colors.Yellow),
		time.Now(), // TODO: Use actual checkpoint mtime
	)
}

func handleHistory(ctx *cli.Context, client *daemon.Client) error {
	repoPath := prefixSlash(ctx.Args()[0])

	history, err := client.History(repoPath)
	if err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("Unable to retrieve history: %v", err),
		}
	}

	fmt.Println(colors.Colorize(repoPath, colors.Magenta))
	for idx := range history {
		checkpoint := history[len(history)-idx-1]
		printCheckpoint(checkpoint, idx, len(history))

	}
	return nil
}

func handleOffline(ctx *cli.Context, client *daemon.Client) error {
	status, err := client.IsOnline()
	if err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("Failed to check online-status: %v", err),
		}
	}

	if !status {
		log.Infof("Already offline.")
		return nil
	}

	if err := client.Offline(); err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("Failed to go offline: %v", err),
		}
	}

	return nil
}

func handleIsOnline(ctx *cli.Context, client *daemon.Client) error {
	status, err := client.IsOnline()
	if err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("Failed to check online-status: %v", err),
		}
	}

	fmt.Println(status)
	return nil
}

func handleOnline(ctx *cli.Context, client *daemon.Client) error {
	status, err := client.IsOnline()
	if err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("Failed to check online-status: %v", err),
		}
	}

	if status {
		log.Infof("Already online.")
		return nil
	}

	if err := client.Online(); err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("Failed to go online: %v", err),
		}
	}

	return nil
}

func handleList(ctx *cli.Context, client *daemon.Client) error {
	path := "/"
	if ctx.NArg() > 0 {
		path = prefixSlash(ctx.Args()[0])
	}

	depth := ctx.Int("depth")
	if ctx.Bool("recursive") {
		depth = -1
	}

	entries, err := client.List(path, depth)
	if err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("ls: %v", err),
		}
	}

	for _, entry := range entries {
		modTime := time.Time{}
		if err := modTime.UnmarshalText(entry.ModTime); err != nil {
			log.Warningf("Could not parse mtime (%s): %v", entry.ModTime, err)
			continue
		}

		fmt.Printf(
			"%s\t%s\t%s\n",
			colors.Colorize(
				humanize.Bytes(uint64(entry.NodeSize)),
				colors.Green,
			),
			colors.Colorize(
				humanize.Time(modTime),
				colors.Cyan,
			),
			colors.Colorize(
				entry.Path,
				colors.Magenta,
			),
		)
	}

	return nil
}

func handleTree(ctx *cli.Context, client *daemon.Client) error {
	path := "/"
	if ctx.NArg() > 0 {
		path = prefixSlash(ctx.Args()[0])
	}

	depth := ctx.Int("depth")
	dirlist, err := client.List(path, depth)
	if err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("ls: %v", err),
		}
	}

	if err := showTree(dirlist, depth); err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("Printing tree failed: %v", err),
		}
	}

	return nil
}

func handleMv(ctx *cli.Context, client *daemon.Client) error {
	source, dest := prefixSlash(ctx.Args()[0]), prefixSlash(ctx.Args()[1])

	if err := client.Move(source, dest); err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("move failed: %v", err),
		}
	}

	return nil
}

func handleMkdir(ctx *cli.Context, client *daemon.Client) error {
	path := prefixSlash(ctx.Args()[0])

	var err error

	if ctx.Bool("parents") {
		err = client.MkdirAll(path)
	} else {
		err = client.Mkdir(path)
	}

	if err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("mkdir failed: %v", err),
		}
	}

	return nil
}

type changeByType map[int32][]*storewire.Checkpoint

func handleStatus(ctx *cli.Context, client *daemon.Client) error {
	status, err := client.Status()
	if err != nil {
		return err
	}

	statusCmt := status.Commit
	if statusCmt == nil {
		return fmt.Errorf("Empty status commit in response.")
	}

	checkpoints := statusCmt.Checkpoints
	if len(checkpoints) == 0 {
		fmt.Println("Nothing to commit.")
		return nil
	}

	userToChanges := make(map[id.ID]changeByType)
	for _, pckp := range checkpoints {
		byAuthor, ok := userToChanges[id.ID(pckp.Author)]
		if !ok {
			byAuthor = make(changeByType)
			userToChanges[id.ID(pckp.Author)] = byAuthor
		}

		byAuthor[pckp.Change] = append(byAuthor[pckp.Change], pckp)
	}

	for user, changesByuser := range userToChanges {
		fmt.Printf(
			"Changes by %s:\n",
			colors.Colorize(string(user), colors.Magenta),
		)
		printChangesByUser(changesByuser)
	}

	return nil
}

func printChangesByUser(changesByuser changeByType) error {
	for typ, checkpoints := range changesByuser {
		changeType := store.ChangeType(typ)

		fmt.Printf("\n    %s:\n", strings.Title((&changeType).String()))
		for _, change := range checkpoints {
			msg := fmt.Sprintf(
				"        %s (%s)",
				"TODO: resolve checkpoint to path",
				humanize.Bytes(uint64(change.Size())),
			)

			color := 0
			switch typ {
			case store.ChangeAdd:
				color = colors.Green
			case store.ChangeRemove:
				color = colors.Red
			case store.ChangeModify:
				color = colors.Yellow
			case store.ChangeMove:
				color = colors.Cyan
			default:
				return fmt.Errorf("Invalid change type")

			}
			fmt.Println(colors.Colorize(msg, color))
		}
	}

	return nil
}

func handleCommit(ctx *cli.Context, client *daemon.Client) error {
	message := ctx.String("message")
	if message == "" {
		message = fmt.Sprintf("Update on %s", time.Now().String())
	}
	status, err := client.Status()
	if err != nil {
		return err
	}

	statusCmt := status.Commit
	if statusCmt == nil {
		return fmt.Errorf("Empty status commit in response.")
	}

	commitCnt := len(statusCmt.Changeset)
	if commitCnt == 0 {
		fmt.Println("Nothing to commit.")
		return nil
	}
	if commitCnt != 1 {
		fmt.Printf("%d changes commited.\n", commitCnt)
	} else {
		fmt.Printf("%d change commited.\n", commitCnt)
	}

	return client.MakeCommit(message)
}

func handleLog(ctx *cli.Context, client *daemon.Client) error {
	log, err := client.Log(nil, nil)
	if err != nil {
		return err
	}

	for _, pnode := range log.Nodes {
		commitMH, err := multihash.Cast(pnode.Hash)
		if err != nil {
			return err
		}

		pcmt := pnode.Commit
		if pcmt == nil {
			return fmt.Errorf("Empty commit in log-commit")
		}

		rootMH, err := multihash.Cast(pcmt.Root)
		if err != nil {
			return err
		}

		fmt.Printf(
			"%s/%s by %s, %s\n",
			colors.Colorize(commitMH.B58String()[:10], colors.Green),
			colors.Colorize(rootMH.B58String()[:10], colors.Magenta),
			pcmt.Author,
			pcmt.Message,
		)
	}
	return nil
}

func handleDiff(ctx *cli.Context, client *daemon.Client) error {
	// TODO
	return nil
}

func handlePin(ctx *cli.Context, client *daemon.Client) error {
	path := ctx.Args()[0]

	if ctx.Bool("is-pinned") {
		isPinned, err := client.IsPinned(path)
		if err != nil {
			return err
		}

		fmt.Printf("%s: %t\n", path, isPinned)
		return nil
	}

	fn := client.Pin
	if ctx.Bool("unpin") {
		fn = client.Unpin
	}

	if err := fn(path); err != nil {
		return err
	}

	return nil
}

func handleDebugExport(ctx *cli.Context, client *daemon.Client) error {
	who := ctx.Args().First()

	var whoID id.ID

	if who != "" {
		validID, err := id.Cast(who)
		if err != nil {
			return err
		}

		whoID = validID
	}

	data, err := client.Export(whoID)
	if err != nil {
		return err
	}

	if _, err := os.Stdout.Write(data); err != nil {
		return err
	}

	return nil
}

func handleDebugImport(ctx *cli.Context, client *daemon.Client) error {
	buf := &bytes.Buffer{}

	if _, err := io.Copy(buf, os.Stdin); err != nil {
		return err
	}

	return client.Import(buf.Bytes())
}

func handleSyncSingle(ctx *cli.Context, client *daemon.Client, idStr string) error {
	ID, err := id.Cast(idStr)

	if err != nil {
		return ExitCode{
			BadArgs,
			fmt.Sprintf("Bad ID: %v", err),
		}
	}

	log.Infof("Syncing with %s", ID)
	if err := client.Sync(ID); err != nil {
		log.Errorf("Failed: %v", err)
		return ExitCode{UnknownError, err.Error()}
	}

	return nil
}

func handleSync(ctx *cli.Context, client *daemon.Client) error {
	// With no arguments, we should sync with everyone around.
	if len(ctx.Args()) != 0 {
		// Pass over to single ID handling:
		return handleSyncSingle(ctx, client, ctx.Args().First())
	}

	// Iterate over all remotes brigd knows about:
	data, err := client.RemoteList()
	if err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("Unable to list remotes: %v", err),
		}
	}

	for _, entry := range data {
		if err := handleSyncSingle(ctx, client, entry.Ident); err != nil {
			return err
		}
	}

	return nil
}
