package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime/trace"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/fatih/color"
	e "github.com/pkg/errors"
	"github.com/sahib/brig/client"
	"github.com/sahib/brig/cmd/pwd"
	"github.com/sahib/brig/cmd/tabwriter"
	"github.com/sahib/brig/server"
	"github.com/sahib/brig/util"
	"github.com/sahib/brig/util/pwutil"
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

func createInitialReadme(ctl *client.Client, folder string) error {
	text := `Welcome to brig!

Here's what you can do next:

    • Read the official documentation (Just type »brig docs«)
    • Add a few remotes to sync with (See »brig help remote«)
    • Mount your data somewhere convinient (See »brig help fstab«)
    • Sync with the remotes you've added (See »brig help sync«)
    • Have a relaxing day while exploring brig.

Please remember that brig is software in its very early stages,
and you should not rely on it yet for production purposes.

If you're done with this README, you can easily remove it:

    $ brig rm README.md

We recommend highly to install a password manager and hook it up
with brig. For example, with "pass" you can execute the following
to avoid re-entering your password on every daemon startup:

    $ brig cfg set repo.password_command "pass brig/my_pwd_key"

The next start of brig will then read the password from the
standard output of this process. Your repository is here:

    %s

Have a nice day.
`
	fd, err := ioutil.TempFile("", ".brig-init-readme-")
	if err != nil {
		return err
	}

	text = fmt.Sprintf(text, folder)
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
	owner := ctx.Args().First()
	backend := ctx.String("backend")
	folder := ctx.GlobalString("repo")
	if ctx.NArg() == 2 {
		var err error
		folder, err = filepath.Abs(ctx.Args().Get(1))
		if err != nil {
			return fmt.Errorf("failed to get absolute path for %s: %v", folder, err)
		}
	}

	if ctx.NArg() > 2 {
		return fmt.Errorf("too many arguments")
	}

	if folder == "" {
		// Make sure that we do not lookup the global registry:
		folder = guessRepoFolder(ctx, false)
		fmt.Printf("Guessed folder for init: %s\n", folder)
	}

	// If a password helper is set, we should read the password from it directly.
	password := readPasswordFromArgs(folder, ctx)
	if pwHelper := ctx.String("pw-helper"); password == "" && pwHelper != "" {
		var err error
		password, err = pwutil.ReadPasswordFromHelper(folder, pwHelper)
		if err != nil {
			return fmt.Errorf("failed to read password from helper: %s", err)
		}
	}

	// Check if the folder exists...
	// doing init twice can easily break things.
	isInitialized, err := repoIsInitialized(folder)
	if err != nil {
		return err
	}

	if isInitialized {
		return fmt.Errorf("`%s` already exists and is not empty; refusing to do init", folder)
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

	if !ctx.Bool("empty") {
		if err := createInitialReadme(ctl, folder); err != nil {
			return err
		}
	}

	fmt.Println(brigLogo)

	if ctx.Bool("no-password") {
		// Set a command in the config that simply echoes a static password:
		staticPasswordHelper := "echo no-password"
		if err := ctl.ConfigSet("repo.password_command", staticPasswordHelper); err != nil {
			return err
		}
	}

	if pwHelper := ctx.String("pw-helper"); pwHelper != "" {
		if err := ctl.ConfigSet("repo.password_command", pwHelper); err != nil {
			return err
		}
	}

	return nil
}

func printConfigDocEntry(entry client.ConfigEntry) {
	val := entry.Val
	if val == "" {
		val = color.YellowString("(empty)")
	}

	defaultMarker := ""
	if entry.Val == entry.Default {
		defaultMarker = color.CyanString("(default)")
	}

	fmt.Printf("%s: %v %s\n", color.GreenString(entry.Key), val, defaultMarker)

	needsRestart := yesify(entry.NeedsRestart)
	defaultVal := entry.Default
	if entry.Default == "" {
		defaultVal = color.YellowString("(empty)")
	}

	fmt.Printf("  Default:       %v\n", defaultVal)
	fmt.Printf("  Documentation: %v\n", entry.Doc)
	fmt.Printf("  Needs restart: %v\n", needsRestart)
}

func handleConfigList(cli *cli.Context, ctl *client.Client) error {
	all, err := ctl.ConfigAll()
	if err != nil {
		return ExitCode{UnknownError, fmt.Sprintf("config list: %v", err)}
	}

	for _, entry := range all {
		printConfigDocEntry(entry)
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

	entry, err := ctl.ConfigDoc(key)
	if err != nil {
		return ExitCode{UnknownError, fmt.Sprintf("config doc: %v", err)}
	}

	if entry.NeedsRestart {
		fmt.Println("NOTE: You need to restart brig for this option to take effect.")
	}

	return nil
}

func handleConfigDoc(ctx *cli.Context, ctl *client.Client) error {
	key := ctx.Args().Get(0)
	entry, err := ctl.ConfigDoc(key)
	if err != nil {
		return ExitCode{UnknownError, fmt.Sprintf("config get: %v", err)}
	}

	printConfigDocEntry(entry)
	return nil
}

func handleDaemonPing(ctx *cli.Context, ctl *client.Client) error {
	if ctx.Bool("wait-for-init") {
		if err := ctl.WaitForInit(); err != nil {
			return err
		}
	}

	count := ctx.Int("count")
	for i := 0; i < count; i++ {
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
	brigPath := guessRepoFolder(ctx, true)
	isInitialized, err := repoIsInitialized(brigPath)
	if err != nil {
		return err
	}

	port := guessPort(ctx)
	bindHost := ctx.GlobalString("bind")

	var password string
	passwordFn := func() (string, error) {
		if !isInitialized {
			return "", nil
		}

		password, err = readPassword(ctx, brigPath)
		if err != nil {
			return "", ExitCode{
				UnknownError,
				fmt.Sprintf("Failed to read password: %v", err),
			}
		}

		return password, nil
	}

	logToStdout := ctx.Bool("log-to-stdout")
	server, err := server.BootServer(brigPath, passwordFn, bindHost, port, logToStdout)
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

	if !ctx.Bool("no-mkdir") {
		if _, err := os.Stat(absMountPath); os.IsNotExist(err) {
			fmt.Printf(
				"Mount directory »%s« does not exist. Will create it.\n",
				absMountPath,
			)
			if err := os.MkdirAll(absMountPath, 0700); err != nil {
				return e.Wrapf(err, "failed to mkdir mount point")
			}
		}
	}

	options := client.MountOptions{
		ReadOnly: ctx.Bool("readonly"),
		RootPath: ctx.String("root"),
	}

	if err := ctl.Mount(absMountPath, options); err != nil {
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

func handleFstabAdd(ctx *cli.Context, ctl *client.Client) error {
	mountName := ctx.Args().Get(0)
	mountPath := ctx.Args().Get(1)

	options := client.MountOptions{
		ReadOnly: ctx.Bool("readonly"),
		RootPath: ctx.String("root"),
	}

	return ctl.FstabAdd(mountName, mountPath, options)
}

func handleFstabRemove(ctx *cli.Context, ctl *client.Client) error {
	mountName := ctx.Args().Get(0)
	return ctl.FstabRemove(mountName)
}

func handleFstabApply(ctx *cli.Context, ctl *client.Client) error {
	if ctx.Bool("unmount") {
		return ctl.FstabUnmountAll()
	}

	return ctl.FstabApply()
}

func handleFstabUnmounetAll(ctx *cli.Context, ctl *client.Client) error {
	return ctl.FstabUnmountAll()
}

func handleFstabList(ctx *cli.Context, ctl *client.Client) error {
	mounts, err := ctl.FsTabList()
	if err != nil {
		return ExitCode{UnknownError, fmt.Sprintf("config list: %v", err)}
	}

	if len(mounts) == 0 {
		return nil
	}

	tabW := tabwriter.NewWriter(
		os.Stdout, 0, 0, 2, ' ',
		tabwriter.StripEscape,
	)

	tmpl, err := readFormatTemplate(ctx)
	if err != nil {
		return err
	}

	if tmpl == nil && len(mounts) != 0 {
		fmt.Fprintln(tabW, "NAME\tPATH\tREAD_ONLY\tROOT\tACTIVE\t")
	}

	for _, entry := range mounts {
		if tmpl != nil {
			if err := tmpl.Execute(os.Stdout, entry); err != nil {
				return err
			}

			continue
		}

		fmt.Fprintf(
			tabW,
			"%s\t%s\t%s\t%s\t%s\n",
			entry.Name,
			entry.Path,
			yesify(entry.ReadOnly),
			entry.Root,
			checkmarkify(entry.Active),
		)
	}

	return tabW.Flush()
}
