package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime/trace"
	"sort"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/fatih/color"
	"github.com/sahib/brig/client"
	"github.com/sahib/brig/cmd/pwd"
	"github.com/sahib/brig/cmd/tabwriter"
	"github.com/sahib/brig/server"
	"github.com/sahib/brig/util"
	"github.com/sahib/brig/version"
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

func createInitialReadme(ctl *client.Client) error {
	text := `Welcome to brig!

Here's what you can do next:

    • Add a few remotes to sync with (See 'brig remote add -h')
    • Mount your data somewhere convinient (See 'brig mount -h')
    • Have a relaxing day exploring brig's features.

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

func dirIsInitReady(dir string) (bool, error) {
	fd, err := os.Open(dir)
	if err != nil && os.IsNotExist(err) {
		return true, nil
	}

	if err != nil {
		return false, err
	}

	names, err := fd.Readdirnames(-1)
	if err != nil {
		return false, err
	}

	for _, name := range names {
		switch name {
		case "meta.yml":
			return false, nil
		case "logs":
			// That's okay.
		default:
			// Anything else we do not know:
			return false, nil
		}
	}

	// base case for empty dir:
	return true, nil
}

func handleInit(ctx *cli.Context, ctl *client.Client) error {
	// Accumulate args:
	owner := ctx.Args().First()
	backend := ctx.String("backend")
	password := readPasswordFromArgs(ctx)

	folder := ctx.String("path")
	if folder == "" {
		folder = guessRepoFolder()
	}

	// Check if the folder exists... doing init twice
	// can easily break things.
	isReady, err := dirIsInitReady(folder)
	if err != nil {
		return err
	}

	if !isReady {
		return fmt.Errorf("`%s` already exists and is not empty; refusing to do a init", folder)
	}

	if password == "" {
		pwdBytes, err := pwd.PromptNewPassword(20)
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
			color.GreenString(key),
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

func handleDaemonPing(ctx *cli.Context, ctl *client.Client) error {
	for i := 0; i < 100; i++ {
		before := time.Now()
		symbol := color.GreenString("✔")

		if err := ctl.Ping(); err != nil {
			symbol = color.RedString("✘")
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
	// Enable tracing (for profiling) if required.
	if ctx.Bool("trace") {
		tracePath := fmt.Sprintf("/tmp/brig-%d.trace", os.Getpid())
		log.Debugf("Writing trace output to %s", tracePath)
		fd, err := os.Create(tracePath)
		if err != nil {
			return err
		}

		defer util.Closer(fd)

		if err := trace.Start(fd); err != nil {
			return err
		}

		defer trace.Stop()
	}

	// If the repository was not initialized yet,
	// we should not ask for a password, since init
	// will already ask for one. If we recognize the repo
	// wrongly as uninitialized, then it won't unlock without
	// a password though.
	brigPath := guessRepoFolder()
	isInitialized, err := repoIsInitialized(brigPath)
	if err != nil {
		return err
	}

	if !isInitialized {
		log.Infof(
			"No repository found at %s. Use `brig init <user>` to create one",
			brigPath,
		)
	}

	password, err := readPassword(ctx, brigPath)
	if err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("Failed to read password: %v", err),
		}
	}

	logPath := ""
	if extraLogPath := ctx.GlobalString("log-path"); len(extraLogPath) != 0 {
		logPath = extraLogPath
	}

	port := ctx.GlobalInt("port")
	bindHost := ctx.GlobalString("bind")
	server, err := server.BootServer(brigPath, password, logPath, bindHost, port)
	if err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("Failed to boot brigd: %v", err),
		}
	}

	defer util.Closer(server)

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
	absMountPath, err := filepath.Abs(mountPath)
	if err != nil {
		return err
	}

	if err := ctl.Mount(absMountPath); err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("Failed to mount: %v", err),
		}
	}

	return nil
}

func handleUnmount(ctx *cli.Context, ctl *client.Client) error {
	mountPath := ctx.Args().First()
	absMountPath, err := filepath.Abs(mountPath)
	if err != nil {
		return err
	}

	if err := ctl.Unmount(absMountPath); err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("Failed to unmount: %v", err),
		}
	}

	return nil
}

func handleVersion(ctx *cli.Context, ctl *client.Client) error {
	vInfo, err := ctl.Version()
	if err != nil {
		return err
	}

	row := func(name, value string) {
		fmt.Printf("%25s: %s\n", name, value)
	}

	row("Client Version", version.String())
	row("Client Rev", version.GitRev)
	row("Server Version", vInfo.ServerSemVer)
	row("Server Rev", vInfo.ServerRev)
	row("Backend (ipfs) Version", vInfo.BackendSemVer)
	row("Backend (ipfs) Rev", vInfo.BackendRev)
	row("Build time", version.BuildTime)

	return nil
}

func handleGc(ctx *cli.Context, ctl *client.Client) error {
	aggressive := ctx.Bool("aggressive")
	freed, err := ctl.GarbageCollect(aggressive)
	if err != nil {
		return err
	}

	if len(freed) == 0 {
		fmt.Println("Nothing freed.")
		return nil
	}

	tabW := tabwriter.NewWriter(
		os.Stdout, 0, 0, 2, ' ',
		tabwriter.StripEscape,
	)

	fmt.Fprintln(tabW, "CONTENT\tHASH\tOWNER\t")

	for _, gcItem := range freed {
		fmt.Fprintf(
			tabW,
			"%s\t%s\t%s\t\n",
			color.WhiteString(gcItem.Path),
			color.RedString(gcItem.Content.ShortB58()),
			color.CyanString(gcItem.Owner),
		)
	}

	return tabW.Flush()
}
