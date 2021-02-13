package setup

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/blang/semver"
	humanize "github.com/dustin/go-humanize"
	shell "github.com/ipfs/go-ipfs-api"
	homedir "github.com/mitchellh/go-homedir"
	e "github.com/pkg/errors"
	"github.com/sahib/brig/util"
	log "github.com/sirupsen/logrus"
)

const (
	defaultPathName = ".ipfs"
	defaultPathRoot = "~/" + defaultPathName
	defaultAPIFile  = "api"
)

func guessIPFSRepo() string {
	baseDir := os.Getenv("IPFS_PATH")
	if baseDir == "" {
		baseDir = defaultPathRoot
	}

	baseDir, err := homedir.Expand(baseDir)
	if err != nil {
		log.Warningf("failed to expand homedir: %v", err)
		return ""
	}

	return baseDir
}

func getAPIAddrFromConfig(baseDir string) (string, error) {
	cfgPath := filepath.Join(baseDir, "config")
	cfgData, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		return "", e.Wrap(err, "does the IPFS repository exist? full error")
	}

	data := struct {
		Addresses struct {
			API string
		}
	}{}

	r := bytes.NewReader(cfgData)
	if err := json.NewDecoder(r).Decode(&data); err != nil {
		return "", err
	}

	return data.Addresses.API, nil
}

// GetAPIAddrForPath returns the API addr of the IPFS repo at `baseDir`.
func GetAPIAddrForPath(baseDir string) (string, error) {
	apiFile := filepath.Join(baseDir, defaultAPIFile)
	if _, err := os.Stat(apiFile); err != nil {
		return getAPIAddrFromConfig(baseDir)
	}

	api, err := ioutil.ReadFile(apiFile)
	if err != nil {
		return getAPIAddrFromConfig(baseDir)
	}

	return string(api), nil
}

func isRunning(apiAddr string) bool {
	return shell.NewShell(apiAddr).IsUp()
}

func getLatestStableVersion() string {
	fallbackStable := "v0.4.22"
	url := "https://dist.ipfs.io/go-ipfs/versions"
	resp, err := http.Get(url)
	if err != nil {
		return fallbackStable
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fallbackStable
	}

	lines := strings.Split(string(data), "\n")
	for idx := len(lines) - 1; idx >= 0; idx-- {
		if !strings.HasPrefix(lines[idx], "v") {
			continue
		}

		if strings.Contains(lines[idx], "-") {
			continue
		}

		return lines[idx]
	}

	return fallbackStable
}

type writeCounter struct {
	total uint64
}

func (wc *writeCounter) Write(p []byte) (int, error) {
	wc.total += uint64(len(p))
	fmt.Printf("\r%s", strings.Repeat(" ", 42))
	fmt.Printf("\r-- Downloading... %s", humanize.Bytes(wc.total))
	return len(p), nil
}

func (wc writeCounter) Close() error {
	fmt.Print("\n")
	return nil
}

func downloadFile(filepath string, url string) error {
	// Create the file, but give it a tmp file extension, this means we won't overwrite a
	// file until it's downloaded, but we'll remove the tmp extension once downloaded.
	out, err := ioutil.TempFile("", "")
	if err != nil {
		return err
	}

	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	// Create our progress reporter and pass it to be used alongside our writer
	progw := &writeCounter{}
	defer progw.Close()

	if _, err = io.Copy(out, io.TeeReader(resp.Body, progw)); err != nil {
		return err
	}

	return os.Rename(out.Name(), filepath)
}

func installIPFS(out io.Writer) error {
	fmt.Fprintf(out, "-- Trying to figure out what IPFS version to install...\n")
	version := getLatestStableVersion()
	fmt.Fprintf(out, "-- Last stable IPFS version is: %s\n", version)

	// Build the IPFS download url:
	url := fmt.Sprintf(
		"https://dist.ipfs.io/go-ipfs/%s/go-ipfs_%s_%s-%s.tar.gz",
		version,
		version,
		runtime.GOOS,
		runtime.GOARCH,
	)

	tgzPath := fmt.Sprintf("/tmp/ipfs-%d.tar.gz", rand.Int63())
	defer os.RemoveAll(tgzPath)

	if err := downloadFile(tgzPath, url); err != nil {
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}

	tgz, err := os.Open(tgzPath)
	if err != nil {
		return err
	}

	tmpDir, err := ioutil.TempDir("", "ipfs-tmp-download-")
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "-- Unpacking to: %s\n", tmpDir)

	defer os.RemoveAll(tmpDir)

	if err := util.Untar(tgz, tmpDir); err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	binPath := filepath.Join(tmpDir, "go-ipfs", "ipfs")
	cwdPath := filepath.Join(cwd, "ipfs")

	fmt.Fprintf(out, "-- Copy binary to: %s\n", cwdPath)
	return util.CopyFile(binPath, cwdPath)
}

func isPortTaken(port int) bool {
	lst, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		// probably cannot take it because it's already taken.
		return true
	}

	lst.Close()
	return false
}

func ipfsSetConfigKey(out io.Writer, ipfsPath, key, value string) {
	args := []string{"config", "--json", key, value}
	cmd := exec.Command("ipfs", args...)
	cmd.Env = append(cmd.Env, "IPFS_PATH="+ipfsPath)
	fmt.Fprintf(out, "  -- Setting config: IPFS_PATH='%s' ipfs %s\n", ipfsPath, strings.Join(args, " "))

	errBuf := &bytes.Buffer{}
	cmd.Stderr = errBuf

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(out, "could not set config: %s\n", strings.TrimSpace(errBuf.String()))
		fmt.Fprintf(out, "command was: `ipfs %s\n`", strings.Join(args, " "))
	}
}

func initIPFS(out io.Writer, ipfsPath, profile string) error {
	cmd := exec.Command("ipfs", "init")
	if profile != "" {
		cmd.Args = append(cmd.Args, "--profile", profile)
	}

	cmd.Env = append(cmd.Env, "IPFS_PATH="+ipfsPath)
	if err := cmd.Run(); err != nil {
		return err
	}

	// default IPFS ports:
	apiPort := 5001
	swarmPort := 4001
	ipfsGwPort := 8080

	// try to find a suitable port for the new ipfs instance.
	// This is used for cases where you have several ipfs instances.
	for off := 0; off < 100; off++ {
		if isPortTaken(apiPort+off) ||
			isPortTaken(swarmPort+off) ||
			isPortTaken(ipfsGwPort+off) {
			continue
		}

		// We have a working port set; build the config keys:
		config := [][]string{
			{
				"Addresses.API",
				fmt.Sprintf("\"/ip4/127.0.0.1/tcp/%d\"", apiPort+off),
			}, {
				"Addresses.Gateway",
				fmt.Sprintf("\"/ip4/127.0.0.1/tcp/%d\"", ipfsGwPort+off),
			}, {
				"Addresses.Swarm",
				fmt.Sprintf(
					"[\"/ip4/0.0.0.0/tcp/%d\", \"/ip6/::/tcp/%d\"]",
					swarmPort+off, swarmPort+off,
				),
			},
		}

		// Go and set the config:
		for _, args := range config {
			ipfsSetConfigKey(out, ipfsPath, args[0], args[1])
		}

		break
	}

	return nil
}

func getIPFSVersion(apiAddr string) (semver.Version, error) {
	vers, _, err := shell.NewShell(apiAddr).Version()
	if err != nil {
		return semver.Version{}, err
	}

	return semver.Parse(vers)
}

func configureIPFS(out io.Writer, apiAddr, ipfsPath string, setExtraConfig bool) error {
	version, err := getIPFSVersion(apiAddr)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "-- The IPFS version is »%s«.\n", version)
	if version.LT(semver.MustParse("0.4.18")) {
		fmt.Fprintf(out, "-- The IPFS version »%s« is quite old. Please update.\n", version)
		fmt.Fprintf(out, "-- We only test on newer versions (>= 0.4.18).\n")
	}

	config := [][]string{
		// Required for talking to other nodes.
		{"Experimental.Libp2pStreamMounting", "true"},
	}

	if setExtraConfig {
		// Optional: Helps save us resources.
		config = append(config, [][]string{
			{"Reprovider.Interval", "\"1h\""},
			{"Swarm.ConnMgr.GracePeriod", "\"60s\""},
		}...)

		if version.GE(semver.MustParse("0.4.19")) {
			config = append(config, [][]string{
				{"Swarm.EnableAutoRelay", "true"},
			}...)
		}
	}

	for _, args := range config {
		ipfsSetConfigKey(out, ipfsPath, args[0], args[1])
	}

	return nil
}

func startIpfs(out io.Writer, ipfsPath string) (int, error) {
	// We don't call Wait() on cmd, so the process will survive the
	// exit of your process and gets reparented to init.
	fmt.Fprintf(out, "-- IPFS_PATH='%s' ipfs daemon --enable-pubsub-experiment\n", ipfsPath)
	cmd := exec.Command("ipfs", "daemon", "--enable-pubsub-experiment")
	cmd.Env = append(cmd.Env, "IPFS_PATH="+ipfsPath)
	if err := cmd.Start(); err != nil {
		return -1, err
	}

	return cmd.Process.Pid, nil
}

func isCommandAvailable(name string) bool {
	path, err := exec.LookPath(name)
	if err != nil {
		return false
	}

	return path != ""
}

func waitForRunningIPFS(out io.Writer, addr string, maxWaitTime time.Duration) {
	waitStart := time.Now()
	for time.Since(waitStart) > maxWaitTime {
		if isRunning(addr) {
			break
		}

		secLeft := float64(maxWaitTime) - float64(time.Since(waitStart))/float64(time.Second)
		fmt.Fprintf(out, "-- Waiting %.2fs for it to fully boot up...   \r", secLeft)
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Fprintf(out, "-- Done waiting.%s\n", strings.Repeat(" ", 30))
}

func dirExistsAndIsNotEmpty(dir string) bool {
	names, err := ioutil.ReadDir(dir)
	if err != nil {
		return false
	}

	return len(names) > 0
}

// Options define how the IPFS setup will be carried out.
type Options struct {
	LogWriter io.Writer

	// Setup, if true, will setup
	Setup bool

	// SetDefaultConfig sets, if true, configuration vital to brig.
	SetDefaultConfig bool

	// SetExtraConfig sets, if true, configuration that helps brig.
	SetExtraConfig bool

	// IpfsPath defines where the new repo should be generated (if needed).
	// If empty we take a guess where the repo could be.
	IpfsPath string

	// InitProfile may be one of the profile names specified here:
	// https://github.com/ipfs/go-ipfs/blob/master/docs/config.md
	InitProfile string
}

// Result details how the setup went.
type Result struct {
	// IpfsPath is always set, even if we only guessed it.
	IpfsPath string

	// PID is the PID of the IPFS daemon if we started it.
	// Otherwise it is set to -1.
	PID int
}

// IPFS setups a IPFS repo at the standard place.
// If there is already a repository and the daemon is running, it will do nothing.
// Otherwise it will install IPFS (if it needs to), init a repo, set config and
// bring up the daemon in a fashion that should work for most cases.
// It will output log messages to `out`.
func IPFS(opts Options) (*Result, error) {
	if opts.LogWriter == nil {
		opts.LogWriter = os.Stdout
	}

	result := &Result{
		PID: -1,
	}

	if opts.IpfsPath == "" {
		opts.IpfsPath = guessIPFSRepo()
		fmt.Fprintf(opts.LogWriter, "-- Guessed IPFS repository as %s\n", opts.IpfsPath)
	} else {
		fmt.Fprintf(opts.LogWriter, "-- IPFS repository is supposed to be at %s\n", opts.IpfsPath)
	}

	// Result should always have the result path set:
	result.IpfsPath = opts.IpfsPath

	if !dirExistsAndIsNotEmpty(opts.IpfsPath) && opts.Setup {
		if !isCommandAvailable("ipfs") {
			fmt.Fprintf(opts.LogWriter, "-- There is no »ipfs« command available.\n")
			if err := installIPFS(opts.LogWriter); err != nil {
				fmt.Fprintf(opts.LogWriter, "-- Failed to install IPFS: %v", err)
				fmt.Fprintf(opts.LogWriter, "-- Please refer to »https://docs.ipfs.io/introduction/install«\n")
				fmt.Fprintf(opts.LogWriter, "-- to find opts.LogWriter on how to install it manually. It is usually very easy.\n")
				fmt.Fprintf(opts.LogWriter, "-- Re-run »brig init« once you're done.\n")
				return nil, err
			}
		} else {
			fmt.Fprintf(opts.LogWriter, "-- »ipfs« command is available, but no repo found.\n")
		}

		fmt.Fprintf(opts.LogWriter, "-- Creating new IPFS repository.\n")
		initIPFS(opts.LogWriter, opts.IpfsPath, opts.InitProfile)
	}

	apiAddr, err := GetAPIAddrForPath(opts.IpfsPath)
	if err != nil {
		return nil, err
	}

	fmt.Fprintf(opts.LogWriter, "-- The API address of the repo is: %s\n", apiAddr)

	if !isRunning(apiAddr) {
		fmt.Fprintf(opts.LogWriter, "-- IPFS Daemon does not seem to be running.\n")
		fmt.Fprintf(opts.LogWriter, "-- Will start one for you with the following command:\n")

		result.PID, err = startIpfs(opts.LogWriter, opts.IpfsPath)
		if err != nil {
			return nil, err
		}

		waitForRunningIPFS(opts.LogWriter, apiAddr, 60)
		fmt.Fprintf(opts.LogWriter, "-- Started IPFS as child of this process.\n")
	} else {
		fmt.Fprintf(opts.LogWriter, "-- IPFS Daemon seems to be running. Let's go!\n")
	}

	if opts.SetDefaultConfig {
		fmt.Fprintf(opts.LogWriter, "-- Will set some default settings for IPFS.\n")
		fmt.Fprintf(opts.LogWriter, "-- These are required for brig to work smoothly.\n")

		if err := configureIPFS(opts.LogWriter, apiAddr, opts.IpfsPath, opts.SetExtraConfig); err != nil {
			fmt.Fprintf(opts.LogWriter, "-- Failed to set defaults: %v\n", err)
			return nil, err
		}
	}

	return result, nil
}
