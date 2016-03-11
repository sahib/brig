package cmdline

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig"
	"github.com/disorganizer/brig/daemon"
	"github.com/disorganizer/brig/repo"
	"github.com/disorganizer/brig/repo/config"
	"github.com/disorganizer/brig/util"
	"github.com/disorganizer/brig/util/colors"
	yamlConfig "github.com/olebedev/config"
	"github.com/tsuibin/goxmpp2/xmpp"
	"github.com/tucnak/climax"
)

func handleVersion(ctx climax.Context) int {
	fmt.Println(brig.VersionString())
	return Success
}

func handleOpen(ctx climax.Context, client *daemon.Client) int {
	log.Infof("Repository is open now.")
	return Success
}

func handleClose(ctx climax.Context, client *daemon.Client) int {
	// This is currently the same as `brig daemon-quit`
	return handleDaemonQuit(ctx, client)
}

func handleDaemonPing(ctx climax.Context, client *daemon.Client) int {
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

func handleDaemonWait(ctx climax.Context) int {
	port := guessPort()

	for {
		client, err := daemon.Dial(port)
		if err == nil {
			client.Close()
			return Success
		}

		time.Sleep(500 * time.Millisecond)
	}
}

func handleDaemonQuit(ctx climax.Context, client *daemon.Client) int {
	client.Exorcise()
	return Success
}

func handleDaemon(ctx climax.Context) int {
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

	port := guessPort()
	baal, err := daemon.Summon(pwd, repoFolder, port)
	if err != nil {
		log.Warning("Unable to start daemon: ", err)
		return UnknownError
	}

	baal.Serve()
	return Success
}

func handleMount(ctx climax.Context, client *daemon.Client) int {
	mountPath := ctx.Args[0]

	var err error

	if ctx.Is("unmount") {
		_, err = client.Unmount(mountPath)
	} else {
		_, err = client.Mount(mountPath)
	}

	if err != nil {
		log.Errorf("fuse: %v", err)
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
	if err := checkJID(jid); err != nil {
		log.Errorf("Bad Jabber ID: %v", err)
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
		port, err := repo.Config.Int("daemon.port")
		if err != nil {
			log.Errorf("Unable to find out port.")
			return UnknownError
		}

		if _, err := daemon.Reach(string(pwd), folder, port); err != nil {
			log.Errorf("Unable to start daemon: %v", err)
			return DaemonNotResponding
		}
	}

	return Success
}

func handleClone(ctx climax.Context) int {
	remoteJID := xmpp.JID(ctx.Args[1])
	if err := checkJID(remoteJID); err != nil {
		log.Warningf("Bad remote Jabber ID: %v", err)
		return BadArgs
	}

	rc := handleInit(ctx)
	if rc != Success {
		return rc
	}

	// Daemon should be up and running by now:
	port := guessPort()

	// Check if the daemon is running:
	client, err := daemon.Dial(port)
	if err == nil {
		return DaemonNotResponding
	}

	if err := client.Fetch(remoteJID); err != nil {
		log.Errorf("fetch failed: %v", err)
		return UnknownError
	}

	return Success
}

func handleAdd(ctx climax.Context, client *daemon.Client) int {
	filePath, err := filepath.Abs(ctx.Args[0])
	if err != nil {
		log.Errorf("Unable to make abs path: %v: %v", filePath, err)
		return UnknownError
	}

	// Assume "/file.png" for file.png as repo path, if none given.
	repoPath := "/" + filepath.Base(filePath)
	if len(ctx.Args) > 1 {
		repoPath = ctx.Args[1]
	}

	path, err := client.Add(filePath, repoPath)
	if err != nil {
		log.Errorf("Could not add file: %v: %v", filePath, err)
		return UnknownError
	}

	fmt.Println(path)
	return Success
}

func handleRm(ctx climax.Context, client *daemon.Client) int {
	repoPath := prefixSlash(ctx.Args[0])

	_, err := client.Rm(repoPath)
	if err != nil {
		log.Errorf("Could not remove file: `%s`: %v", repoPath, err)
		return UnknownError
	}

	return Success
}

func handleCat(ctx climax.Context, client *daemon.Client) int {
	repoPath := prefixSlash(ctx.Args[0])

	filePath := ""
	isStdoutMode := len(ctx.Args) < 2

	if isStdoutMode {
		tmpFile, err := ioutil.TempFile("", ".brig-tmp-")
		if err != nil {
			log.Errorf("Unable to create temp file: %v", err)
			return UnknownError
		}

		filePath = tmpFile.Name()
		defer util.Closer(tmpFile)
		defer func() {
			if err := os.Remove(filePath); err != nil {
				log.Warningf("Cannot remove temp-file: %v", err)
			}
		}()
	} else {
		absPath, err := filepath.Abs(ctx.Args[1])
		if err != nil {
			log.Errorf("Unable to make abs path: %v: %v", filePath, err)
			return UnknownError
		}

		filePath = absPath
	}

	_, err := client.Cat(repoPath, filePath)
	if err != nil {
		log.Errorf("Could not cat file: %v: %v", repoPath, err)
		return UnknownError
	}

	if isStdoutMode {
		fd, err := os.Open(filePath)
		if err != nil {
			log.Errorf("Could not open temp file")
			return UnknownError
		}

		if _, err := io.Copy(os.Stdout, fd); err != nil {
			log.Errorf("Cannot copy to stdout: %v", err)
			return UnknownError
		}

		if err := fd.Close(); err != nil {
			log.Warningf("Unable to close tmpfile handle: %v", err)
		}
	}

	return Success
}

var (
	treeRunePipe   = "│"
	treeRuneTri    = "├"
	treeRuneBar    = "──"
	treeRuneCorner = "└"
	treeRuneDot    = "⌽" // ⚫
)

func handleHistory(ctx climax.Context, client *daemon.Client) int {
	repoPath := prefixSlash(ctx.Args[0])

	history, err := client.History(repoPath)
	if err != nil {
		log.Errorf("Unable to retrieve history: %v", err)
		return UnknownError
	}

	fmt.Println(colors.Colorize(repoPath, colors.Magenta))
	for idx := range history {
		checkpoint := history[len(history)-idx-1]

		threeWayRune, twoWayRune := treeRuneTri, treeRunePipe
		if idx == len(history)-1 {
			threeWayRune, twoWayRune = treeRuneCorner, " "
		}

		fmt.Printf(
			" %s%s %s #%d (%s by %s)\n",
			threeWayRune,
			treeRuneBar,
			colors.Colorize("Checkpoint", colors.Cyan),
			len(history)-idx,
			colors.Colorize(checkpoint.Change.String(), colors.Red),
			colors.Colorize(string(checkpoint.Author), colors.Magenta),
		)

		fmt.Printf(
			" %s   ├─ % 9s: %v\n",
			twoWayRune,
			colors.Colorize("Hash", colors.Green),
			checkpoint.Hash.B58String(),
		)

		fmt.Printf(
			" %s   └─ % 9s: %v\n",
			twoWayRune,
			colors.Colorize("Date", colors.Yellow),
			checkpoint.ModTime,
		)
	}

	return Success
}

func handleOffline(ctx climax.Context, client *daemon.Client) int {
	status, err := client.IsOnline()
	if err != nil {
		log.Errorf("Failed to check online-status: %v", err)
		return UnknownError
	}

	if !status {
		log.Infof("Already offline.")
		return Success
	}

	if err := client.Offline(); err != nil {
		log.Errorf("Failed to go offline: %v", err)
		return UnknownError
	}

	return Success
}

func handleIsOnline(ctx climax.Context, client *daemon.Client) int {
	status, err := client.IsOnline()
	if err != nil {
		log.Errorf("Failed to check online-status: %v", err)
		return UnknownError
	}

	fmt.Println(status)
	return Success
}

func handleOnline(ctx climax.Context, client *daemon.Client) int {
	status, err := client.IsOnline()
	if err != nil {
		log.Errorf("Failed to check online-status: %v", err)
		return UnknownError
	}

	if status {
		log.Infof("Already online.")
		return Success
	}

	if err := client.Online(); err != nil {
		log.Errorf("Failed to go online: %v", err)
		return UnknownError
	}

	return Success
}

func handleList(ctx climax.Context, client *daemon.Client) int {
	path := "/"
	if len(ctx.Args) > 0 {
		path = prefixSlash(ctx.Args[0])
	}

	depth, err := ctxGetIntWithDefault(ctx, "depth", -1)
	if err != nil {
		log.Warningf("Invalid depth: %v", err)
		return BadArgs
	}

	if ctx.Is("recursive") {
		depth = -1
	}

	dirlist, err := client.List(path, depth)
	if err != nil {
		log.Warningf("ls: %v", err)
		return UnknownError
	}

	// TODO: Nicer formatting.
	for _, dirent := range dirlist {
		fmt.Printf(
			"%5d %s %s\n",
			dirent.GetFileSize(),
			dirent.GetModTime(),
			dirent.GetPath(),
		)
	}

	return Success
}

func handlePull(ctx climax.Context, client *daemon.Client) int {
	remoteJID := xmpp.JID(ctx.Args[0])
	if err := checkJID(remoteJID); err != nil {
		log.Warningf("Bad remote Jabber ID: %v", err)
		return BadArgs
	}

	if err := client.Fetch(remoteJID); err != nil {
		log.Errorf("fetch failed: %v", err)
		return UnknownError
	}

	return Success
}

func handleMv(ctx climax.Context, client *daemon.Client) int {
	source, dest := prefixSlash(ctx.Args[0]), prefixSlash(ctx.Args[1])

	if err := client.Move(source, dest); err != nil {
		log.Warningf("move failed: %v", err)
		return UnknownError
	}

	return Success
}
