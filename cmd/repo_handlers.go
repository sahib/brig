package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log/syslog"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/trace"
	"strings"
	"time"

	"github.com/fatih/color"
	e "github.com/pkg/errors"
	"github.com/sahib/brig/client"
	"github.com/sahib/brig/cmd/pwd"
	"github.com/sahib/brig/cmd/tabwriter"
	"github.com/sahib/brig/gateway"
	"github.com/sahib/brig/repo"
	"github.com/sahib/brig/repo/repopack"
	"github.com/sahib/brig/repo/setup"
	"github.com/sahib/brig/server"
	"github.com/sahib/brig/util"
	formatter "github.com/sahib/brig/util/log"
	"github.com/sahib/brig/version"
	log "github.com/sirupsen/logrus"
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

`

const initBanner = `

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

Your repository is here:

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

	return ctl.MakeCommit("added initial README.md")
}

func handleInit(ctx *cli.Context) error {
	if len(ctx.Args()) == 0 {
		return fmt.Errorf("Please specify a name for the owner of this repository")
	}

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
		var err error
		folder, err = guessRepoFolder(ctx)
		if err != nil {
			return err
		}

		fmt.Printf("-- Guessed folder for init: %s\n", folder)
	}

	// doing init twice can easily break things.
	isInitialized, err := isNonEmptyDir(folder)
	if err != nil {
		return err
	}

	if isInitialized {
		return fmt.Errorf("`%s` already exists and is not empty; refusing to do init", folder)
	}

	ipfsPath := ctx.String("ipfs-path")
	doIpfsSetup := !ctx.Bool("no-ipfs-setup")
	doIpfsConfig := !ctx.Bool("no-ipfs-config")
	doExtraIpfsConfig := !ctx.Bool("no-ipfs-optimization")

	if backend == "httpipfs" {
		var err error
		ipfsPath, err = setup.IPFS(os.Stdout, doIpfsSetup, doIpfsConfig, doExtraIpfsConfig, ipfsPath)
		if err != nil {
			return err
		}
	}

	daemonURL, err := guessFreeDaemonURL(ctx, owner)
	if err != nil {
		log.WithError(err).Warnf("failed to figure out a free daemon url")
	}

	if err := Init(
		ctx,
		ipfsPath,
		repo.InitOptions{
			BaseFolder:  folder,
			Owner:       owner,
			BackendName: backend,
			DaemonURL:   daemonURL,
		},
	); err != nil {
		return ExitCode{UnknownError, fmt.Sprintf("init failed: %v", err)}
	}

	// Start the daemon on the freshly initialized repo:
	ctl, err := startDaemon(ctx, folder, daemonURL)
	if err != nil {
		return ExitCode{
			DaemonNotResponding,
			fmt.Sprintf("Unable to start daemon: %v", err),
		}
	}

	// Run the actual handler:
	defer ctl.Close()

	return handleInitPost(ctx, ctl, folder)
}

func handleInitPost(ctx *cli.Context, ctl *client.Client, folder string) error {
	if !ctx.Bool("empty") {
		if err := createInitialReadme(ctl, folder); err != nil {
			return err
		}
	}

	if !ctx.Bool("no-logo") {
		fmt.Println(brigLogo)

		if !ctx.Bool("empty") {
			fmt.Println(initBanner)
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

	for _, elem := range strings.Split(val, " ;; ") {
		fmt.Println(elem)
	}
	return nil
}

func handleConfigSet(ctx *cli.Context, ctl *client.Client) error {
	key := ctx.Args().Get(0)

	val := ctx.Args().Get(1)
	if len(ctx.Args()) > 2 {
		val = strings.Join(ctx.Args()[1:], " ;; ")
	}

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

func switchToSyslog() {
	wSyslog, err := syslog.New(syslog.LOG_NOTICE, "brig")
	if err != nil {
		log.Warningf("failed to open connection to syslog for brig: %v", err)
		logFd, err := ioutil.TempFile("", "brig-*.log")
		if err != nil {
			log.Warningf("")
		} else {
			log.Warningf("Will log to %s from now on.", logFd.Name())
			log.SetOutput(logFd)
		}

		return
	}

	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&formatter.FancyLogFormatter{
		UseColors: false,
	})

	log.SetOutput(
		io.MultiWriter(
			formatter.NewSyslogWrapper(wSyslog),
			os.Stdout,
		),
	)
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

	repoPath, err := guessRepoFolder(ctx)
	if err != nil {
		return err
	}

	daemonURL, err := guessDaemonURL(ctx)
	if err != nil {
		return err
	}

	// Make sure IPFS is running. Also set required options,
	// but don't bother to set optimizations.
	var ipfsPath string
	cfg, err := openConfig(repoPath)
	if err != nil {
		log.Warningf("failed to read config at %v: %v", repoPath, err)
	} else {
		ipfsPath = cfg.String("daemon.ipfs_path")
	}

	if _, err := setup.IPFS(
		&logWriter{prefix: "ipfs"},
		true,
		true,
		false,
		ipfsPath,
	); err != nil {
		return err
	}

	logToStdout := ctx.Bool("log-to-stdout")
	if !logToStdout {
		log.Infof("all further logs will be also piped to the syslog daemon.")
		log.Infof("Use »journalctl -fet brig« to view logs.")
		switchToSyslog()
	} else {
		log.SetOutput(os.Stdout)
	}

	server, err := server.BootServer(
		repoPath,
		daemonURL,
		logToStdout,
	)

	if err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("failed to boot brigd: %v", err),
		}
	}

	defer util.Closer(server)

	if err := server.Serve(); err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("failed to serve: %v", err),
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
		Offline:  ctx.Bool("offline"),
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
		Offline:  ctx.Bool("offline"),
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
		fmt.Fprintln(tabW, "NAME\tPATH\tREAD_ONLY\tOFFLINE\tROOT\tACTIVE\t")
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
			"%s\t%s\t%s\t%s\t%s\t%s\n",
			entry.Name,
			entry.Path,
			yesify(entry.ReadOnly),
			yesify(entry.Offline),
			entry.Root,
			checkmarkify(entry.Active),
		)
	}

	return tabW.Flush()
}

func handleGatewayStart(ctx *cli.Context, ctl *client.Client) error {
	isEnabled, err := ctl.ConfigGet("gateway.enabled")
	if err != nil {
		return err
	}

	if isEnabled == "false" {
		if err := ctl.ConfigSet("gateway.enabled", "true"); err != nil {
			return err
		}
	} else {
		fmt.Println("Seems like we're running already.")
	}

	port, err := ctl.ConfigGet("gateway.port")
	if err != nil {
		return err
	}

	domain, err := ctl.ConfigGet("gateway.cert.domain")
	if err != nil {
		return err
	}

	if domain == "" {
		domain = "localhost"
	}

	isHTTPS, err := gatewayIsHTTPS(ctl)
	if err != nil {
		return err
	}

	protocol := "http"
	if isHTTPS {
		protocol = "https"
	}

	url := fmt.Sprintf("%s://%s:%s", protocol, domain, port)
	fmt.Printf("The gateway is accessible via %s\n", url)
	return nil
}

func gatewayIsHTTPS(ctl *client.Client) (bool, error) {
	certPath, err := ctl.ConfigGet("gateway.cert.certfile")
	if err != nil {
		return false, err
	}

	keyPath, err := ctl.ConfigGet("gateway.cert.keyfile")
	if err != nil {
		return false, err
	}

	return certPath != "" && keyPath != "", nil
}

func handleGatewayStatus(ctx *cli.Context, ctl *client.Client) error {
	isEnabled, err := ctl.ConfigGet("gateway.enabled")
	if err != nil {
		return err
	}

	if isEnabled == "false" {
		fmt.Println("• The gateway is not running. Use »brig gateway start« to start.")
		return nil
	}

	port, err := ctl.ConfigGet("gateway.port")
	if err != nil {
		return err
	}
	domain, err := ctl.ConfigGet("gateway.cert.domain")
	if err != nil {
		return err
	}

	if domain == "" {
		domain = "localhost"
	}

	isHTTPS, err := gatewayIsHTTPS(ctl)
	if err != nil {
		return err
	}

	protocol := "http"
	if isHTTPS {
		protocol = "https"
	}

	url := fmt.Sprintf("%s://%s:%s", protocol, domain, port)

	fmt.Printf("• Running on %s\n", color.GreenString(url))
	if isHTTPS {
		fmt.Printf("• Using %s. Nice.\n", color.GreenString("https"))
	} else {
		fmt.Printf("• Using %s for transmitting files.\n", color.RedString("http"))
		fmt.Println("  Consider changing this (if possible) by using »brig gateway cert«.")
	}

	uiIsEnabled, err := ctl.ConfigGet("gateway.ui.enabled")
	if err != nil {
		return err
	}

	if uiIsEnabled == "true" {
		fmt.Println("• The Web UI is currently enabled and can be accessed via the URL above.")
		fmt.Println("  If you want to disable the UI (»/get« will still work), then do:")
		fmt.Println("")
		fmt.Println("    $ brig cfg gateway.ui.enabled false")
		fmt.Println("")
	} else {
		fmt.Println("• There is no UI enabled. You can enable it via:")
		fmt.Println("")
		fmt.Println("    $ brig cfg gateway.ui.enabled true")
		fmt.Println("")
	}

	authIsEnabled, err := ctl.ConfigGet("gateway.auth.enabled")
	if err != nil {
		return err
	}

	if authIsEnabled == "true" {
		fmt.Printf("• Password based user authentication is enabled. Good.\n")
		users, err := ctl.GatewayUserList()
		if err != nil {
			return err
		}

		if len(users) == 0 {
			fmt.Printf(
				"• But there are %s users set. Add a user with »brig gw user add <name>«!\n",
				color.RedString("no"),
			)
		} else {
			fmt.Printf(
				"• There are %s users currently. Review them with »brig gw user ls«.\n",
				color.GreenString(fmt.Sprintf("%d", len(users))),
			)
		}
	} else {
		fmt.Printf("• There is %s user authentication enabled.\n", color.YellowString("no"))
		fmt.Printf("  You can enable it by setting the following config keys:\n")
		fmt.Printf("\n")
		fmt.Printf("    $ brig config set gateway.auth.user <user>\n")
		fmt.Printf("    $ brig config set gateway.auth.pass <pass>\n")
		fmt.Printf("\n")
	}

	if isHTTPS {
		redirIsEnabled, err := ctl.ConfigGet("gateway.cert.redirect.enabled")
		if err != nil {
			return err
		}

		if redirIsEnabled == "true" {
			redirPort, err := ctl.ConfigGet("gateway.cert.redirect.http_port")
			if err != nil {
				return err
			}

			fmt.Printf(
				"• All requests on HTTP port %s will be forwarded to HTTPS port %s.\n",
				color.GreenString(redirPort),
				color.GreenString(port),
			)
		} else {
			fmt.Printf("• There is not HTTP port configured that forwards to HTTPS.\n")
		}
	}

	return nil
}

func handleGatewayStop(ctx *cli.Context, ctl *client.Client) error {
	isEnabled, err := ctl.ConfigGet("gateway.enabled")
	if err != nil {
		return err
	}

	if isEnabled == "true" {
		if err := ctl.ConfigSet("gateway.enabled", "false"); err != nil {
			return err
		}

		fmt.Println("The gateway will stop serving after handling all open requests.")
	} else {
		fmt.Println("It seems like the gateway is already stopped.")
	}

	return nil
}

func handleGatewayCert(ctx *cli.Context) error {
	domain := ctx.Args().Get(0)
	if domain == "" {
		return fmt.Errorf("Usage: brig gateway cert your.domain.org ")
	}

	cacheDir := ctx.String("cache-dir")
	if cacheDir == "" {
		var err error
		cacheDir, err = gateway.UserCacheDir()
		if err != nil {
			return err
		}

		cacheDir = filepath.Join(cacheDir, "brig")
	}

	if os.Geteuid() != 0 {
		fmt.Println(
			color.YellowString("You are not root. We need root rights to bind to port 80."),
		)
		fmt.Println(
			color.YellowString("I will re-execute this command for you as:"),
		)
		fmt.Printf("$ sudo brig gateway cert %s --cache-dir %s\n", domain, cacheDir)
		fmt.Println()
		exePath, err := getExecutablePath()
		if err != nil {
			return err
		}

		sudoPath, err := exec.LookPath("sudo")
		if err != nil {
			return err
		}

		// #nosec
		proc := exec.Command(
			sudoPath,
			exePath,
			"gateway", "cert",
			domain,
			cacheDir,
		)

		proc.Stdin = os.Stdin
		proc.Stdout = os.Stdout
		proc.Stderr = os.Stderr

		if err := proc.Start(); err != nil {
			return err
		}

		return proc.Wait()
	}

	privPath, pubPath, err := gateway.FetchTLSCertificate(domain, cacheDir)
	if err != nil {
		fmt.Printf("Failed to download cert: %s\n", err)
		return err
	}

	fmt.Println("A certificate was downloaded successfully.")

	daemonURL, err := guessDaemonURL(ctx)
	if err != nil {
		return err
	}

	ctl, err := client.Dial(context.Background(), daemonURL)
	if err != nil {
		fmt.Println("There does not seem a daemon running currently.")
		fmt.Println("Please execute the following commands when it is running:")
		fmt.Println("")
		fmt.Printf("  $ brig config set gateway.cert.certfile '%s'\n", pubPath)
		fmt.Printf("  $ brig config set gateway.cert.keyfile  '%s'\n", privPath)
		fmt.Println("")
		fmt.Println("Alternatively, just re-run this command again some time else.")
		return nil
	}

	defer ctl.Close()

	if err := ctl.ConfigSet("gateway.cert.domain", domain); err != nil {
		return err
	}

	if err := ctl.ConfigSet("gateway.cert.certfile", pubPath); err != nil {
		return err
	}

	if err := ctl.ConfigSet("gateway.cert.keyfile", privPath); err != nil {
		return err
	}

	fmt.Println("Successfully set the gateway config to use the certificate.")
	fmt.Println("Note that you have to re-run this command every 90 days currently.")
	return nil
}

func handleGatewayURL(ctx *cli.Context, ctl *client.Client) error {
	path := ctx.Args().First()
	if _, err := ctl.Stat(path); err != nil {
		return err
	}

	domain, err := ctl.ConfigGet("gateway.cert.domain")
	if err != nil {
		return err
	}

	if domain == "" {
		domain = "localhost"
	}

	port, err := ctl.ConfigGet("gateway.port")
	if err != nil {
		return err
	}

	if port == "80" || port == "443" {
		port = ""
	} else {
		port = ":" + port
	}

	isHTTPS, err := gatewayIsHTTPS(ctl)
	if err != nil {
		return err
	}

	protocol := "http"
	if domain != "localhost" && isHTTPS {
		protocol = "https"
	}

	escapedPath := url.PathEscape(strings.TrimLeft(path, "/"))
	fmt.Printf("%s://%s%s/get/%s\n", protocol, domain, port, escapedPath)
	return nil
}

func handleGatewayUserAdd(ctx *cli.Context, ctl *client.Client) error {
	nArgs := len(ctx.Args())
	name := ctx.Args().First()

	var password string
	if nArgs > 1 {
		password = ctx.Args().Get(1)
	} else {
		bPassword, err := pwd.PromptNewPassword(14)
		if err != nil {
			return err
		}

		password = string(bPassword)
	}

	folders := []string{"/"}
	if nArgs > 2 {
		folders = ctx.Args()[2:]
	}

	allRights := []string{
		"fs.download",
		"fs.view",
		"fs.edit",
		"remotes.view",
		"remotes.edit",
	}

	rights := []string{}
	if ctx.Bool("role-admin") {
		rights = allRights
	}

	if ctx.Bool("role-editor") {
		rights = allRights[:len(allRights)-1]
	}

	if ctx.Bool("role-collaborator") {
		rights = allRights[:len(allRights)-2]
	}

	if ctx.Bool("role-viewer") {
		rights = allRights[:len(allRights)-3]
	}

	if ctx.Bool("role-link-only") {
		rights = allRights[:len(allRights)-4]
	}

	if r := ctx.String("rights"); r != "" {
		rights = strings.Split(r, ",")
	}

	return ctl.GatewayUserAdd(name, password, folders, rights)
}

func handleGatewayUserRemove(ctx *cli.Context, ctl *client.Client) error {
	for _, name := range ctx.Args() {
		if err := ctl.GatewayUserRemove(name); err != nil {
			fmt.Printf("Failed to remove »%s«: %v\n", name, err)
		}
	}

	return nil
}

func handleGatewayUserList(ctx *cli.Context, ctl *client.Client) error {
	users, err := ctl.GatewayUserList()
	if err != nil {
		return err
	}

	tabW := tabwriter.NewWriter(
		os.Stdout, 0, 0, 2, ' ',
		tabwriter.StripEscape,
	)

	tmpl, err := readFormatTemplate(ctx)
	if err != nil {
		return err
	}

	if tmpl == nil {
		if len(users) == 0 {
			fmt.Println("No users. Add some with »brig gw user add <name> <pass> <folders...>«")
		} else {
			fmt.Fprintln(tabW, "NAME\tFOLDERS\tRIGHTS\t")
		}
	}

	for _, user := range users {
		if tmpl != nil {
			if err := tmpl.Execute(os.Stdout, user); err != nil {
				return err
			}

			continue
		}

		fmt.Fprintf(
			tabW,
			"%s\t%s\t%s\t\n",
			user.Name,
			strings.Join(user.Folders, ","),
			strings.Join(user.Rights, ","),
		)
	}

	return tabW.Flush()
}

func readPassword(ctx *cli.Context, isNew bool) ([]byte, error) {
	if ctx.IsSet("password-command") {
		log.Debugf("reading by password command.")
		cmd := exec.Command("/bin/sh", "-c", ctx.String("password-command"))

		// Make sure sub command can access our streams.
		// Some password managers might ask for a master password.
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		out, err := cmd.Output()
		if err != nil {
			return nil, err
		}

		// Strip any newline produced by the tool.
		// Just hope that nobody really tries to use newlines
		// as part of the password. Would still work though
		// as long only --password-command is used to enter the password.
		return bytes.TrimRight(out, "\n\r"), nil
	}

	if ctx.IsSet("password-file") {
		log.Debugf("reading from password file.")
		return ioutil.ReadFile(ctx.String("password-file"))
	}

	if isNew {
		return pwd.PromptNewPassword(10)
	}

	return pwd.PromptPassword()
}

func handleRepoPack(ctx *cli.Context) error {
	folder, err := guessRepoFolder(ctx)
	if err != nil {
		return err
	}

	isRunning, err := isDaemonRunning(ctx)
	if err != nil {
		return e.Wrap(err, "failed to check if daemon is running")
	}

	if isRunning {
		log.Error("daemon is still running for this repo, please quit it first!")
		log.Errorf("Use »brig --repo %s daemon quit« for this.", folder)
		return errors.New("refusing to pack data, there might be inconsistencies")
	}

	pass, err := readPassword(ctx, true)
	if err != nil {
		return err
	}

	archivePath := ctx.Args().First()
	if archivePath == "" {
		archivePath = folder + ".repopack"
	}

	log.Infof("writing archive to »%s«", archivePath)
	return repopack.PackRepo(
		folder,
		archivePath,
		string(pass),
		!ctx.Bool("no-remove"),
	)
}

func handleRepoUnpack(ctx *cli.Context) error {
	archivePath := ctx.Args().First()
	folder, err := guessRepoFolder(ctx)
	if err != nil {
		// Small convenience hack: if the archive ends in .repopack
		// assume that it was created from a repo with the same path
		// but without the suffix.
		folder = strings.TrimSuffix(archivePath, ".repopack")
	}

	isNonEmpty, err := isNonEmptyDir(folder)
	if err != nil {
		return err
	}

	if isNonEmpty {
		return fmt.Errorf("»%s« is non-empty, refusing to overwrite", folder)
	}

	if archivePath == "" {
		return fmt.Errorf("please specify the location of the packed archive")
	}

	pass, err := readPassword(ctx, false)
	if err != nil {
		return err
	}

	log.Infof("unpacking to »%s«", folder)
	return repopack.UnpackRepo(
		folder,
		archivePath,
		string(pass),
		!ctx.Bool("no-remove"),
	)
}

func handleRepoHintsSet(ctx *cli.Context, ctl *client.Client) error {
	path := ctx.Args().First()
	hint := client.Hint{
		Path:            path,
		EncryptionAlgo:  ctx.String("encryption"),
		CompressionAlgo: ctx.String("compression"),
	}

	return ctl.HintSet(path, hint)
}

func handleRepoHintsList(ctx *cli.Context, ctl *client.Client) error {
	hints, err := ctl.HintList()
	if err != nil {
		return err
	}

	tabW := tabwriter.NewWriter(
		os.Stdout, 0, 0, 2, ' ',
		tabwriter.StripEscape,
	)

	fmt.Fprintln(tabW, "PATH\tENCRYPTION\tCOMPRESSION\t")

	for _, hint := range hints {
		fmt.Fprintf(
			tabW,
			"%s\t%s\t%s\t\n",
			hint.Path,
			hint.EncryptionAlgo,
			hint.CompressionAlgo,
		)
	}

	return tabW.Flush()
}

func handleRepoHintsRemove(ctx *cli.Context, ctl *client.Client) error {
	return ctl.HintRemove(ctx.Args().First())
}
