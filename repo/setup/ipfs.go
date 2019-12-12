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
	homedir "github.com/mitchellh/go-homedir"
	e "github.com/pkg/errors"
	"github.com/sahib/brig/util"
	shell "github.com/sahib/go-ipfs-api"
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

func initIPFS(out io.Writer, ipfsPath string) error {
	cmd := exec.Command("ipfs", "init")
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
			{"Experimental.QUIC", "true"},
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

func startIpfs(out io.Writer, ipfsPath string) error {
	// We don't call Wait() on cmd, so the process will survive the
	// exit of your process and gets reparented to init.
	fmt.Fprintf(out, "-- IPFS_PATH='%s' ipfs daemon --enable-pubsub-experiment\n", ipfsPath)
	cmd := exec.Command("ipfs", "daemon", "--enable-pubsub-experiment")
	cmd.Env = append(cmd.Env, "IPFS_PATH="+ipfsPath)
	return cmd.Start()
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

// IPFS setups a IPFS repo at the standard place.
// If there is already a repository and the daemon is running, it will do nothing.
// Otherwise it will install IPFS (if it needs to), init a repo, set config and
// bring up the daemon in a fashion that should work for most cases.
// It will output log messages to `out`.
func IPFS(out io.Writer, doSetup, setDefaultConfig, setExtraConfig bool, ipfsPath string) (string, error) {
	if ipfsPath == "" {
		ipfsPath = guessIPFSRepo()
		fmt.Fprintf(out, "-- Guessed IPFS repository as %s\n", ipfsPath)
	} else {
		fmt.Fprintf(out, "-- IPFS repository is supposed to be at %s\n", ipfsPath)
	}

	if !dirExistsAndIsNotEmpty(ipfsPath) && doSetup {
		if !isCommandAvailable("ipfs") {
			fmt.Fprintf(out, "-- There is no »ipfs« command available.\n")
			if err := installIPFS(out); err != nil {
				fmt.Fprintf(out, "-- Failed to install IPFS: %v", err)
				fmt.Fprintf(out, "-- Please refer to »https://docs.ipfs.io/introduction/install«\n")
				fmt.Fprintf(out, "-- to find out on how to install it manually. It is usually very easy.\n")
				fmt.Fprintf(out, "-- Re-run »brig init« once you're done.\n")
				return "", err
			}
		} else {
			fmt.Fprintf(out, "-- »ipfs« command is available, but no repo found.\n")
		}

		fmt.Fprintf(out, "-- Creating new IPFS repository.\n")
		initIPFS(out, ipfsPath)
	}

	apiAddr, err := GetAPIAddrForPath(ipfsPath)
	if err != nil {
		return "", err
	}

	fmt.Fprintf(out, "-- The API address of the repo is: %s\n", apiAddr)

	if !isRunning(apiAddr) {
		fmt.Fprintf(out, "-- IPFS Daemon does not seem to be running.\n")
		fmt.Fprintf(out, "-- Will start one for you with the following command:\n")
		if err := startIpfs(out, ipfsPath); err != nil {
			return "", err
		}

		waitForRunningIPFS(out, apiAddr, 60)
		fmt.Fprintf(out, "-- Started IPFS as child of this process.\n")
	} else {
		fmt.Fprintf(out, "-- IPFS Daemon seems to be running. Let's go!\n")
	}

	if setDefaultConfig {
		fmt.Fprintf(out, "-- Will set some default settings for IPFS.\n")
		fmt.Fprintf(out, "-- These are required for brig to work smoothly.\n")

		if err := configureIPFS(out, apiAddr, ipfsPath, setExtraConfig); err != nil {
			fmt.Fprintf(out, "-- Failed to set defaults: %v\n", err)
			return "", err
		}
	}

	return ipfsPath, nil
}
