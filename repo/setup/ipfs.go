package setup

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

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

func getAPIAddrForPath(baseDir string) (string, error) {
	apiFile := path.Join(baseDir, defaultAPIFile)
	if _, err := os.Stat(apiFile); err != nil {
		return "", e.Wrap(err, "does the IPFS repository exist? full error")
	}

	api, err := ioutil.ReadFile(apiFile)
	if err != nil {
		return "", err
	}

	return string(api), nil
}

func isRunning(apiAddr string) bool {
	sh := shell.NewShell(apiAddr)
	return sh.IsUp()
}

func getLatestStableVersion() string {
	fallbackStable := "v0.4.18"
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

func initIPFS(ipfsPath string) error {
	cmd := exec.Command("ipfs", "init")
	cmd.Env = append(cmd.Env, "IPFS_PATH="+ipfsPath)
	return cmd.Run()
}

func configureIPFS(out io.Writer, ipfsPath string) error {
	config := [][]string{
		// Required: Required for talking to other nodes.
		{"Experimental.Libp2pStreamMounting", "true"},
		// Optional: Helps save us resources.
		{"Experimental.QUIC", "true"},
		{"Swarm.EnableRelayHop", "true"},
		{"Reprovider.Interval", "\"2h\""},
		{"Swarm.ConnMgr.LowWater", "100"},
		{"Swarm.ConnMgr.HighWater", "200"},
		{"Swarm.ConnMgr.GracePeriod", "\"60s\""},
	}

	for _, args := range config {
		args = append([]string{"config", "--json"}, args...)
		cmd := exec.Command("ipfs", args...)
		cmd.Env = append(cmd.Env, "IPFS_PATH="+ipfsPath)
		fmt.Fprintf(out, "  -- Executing: IPFS_PATH='%s' ipfs %s\n", ipfsPath, strings.Join(args, " "))

		errBuf := &bytes.Buffer{}
		cmd.Stderr = errBuf

		if err := cmd.Run(); err != nil {
			fmt.Fprintf(out, "could not set config: %s", strings.TrimSpace(errBuf.String()))
			fmt.Fprintf(out, "command was: `ipfs %s`", strings.Join(args, " "))
		}
	}

	return nil
}

func startIpfs(ipfsPath string) error {
	// We don't call Wait() on cmd, so the process will survive the
	// exit of your process and gets reparented to init.
	cmd := exec.Command("ipfs", "daemon", "--enable-pubsub-experiment")
	cmd.Env = append(cmd.Env, "IPFS_PATH="+ipfsPath)
	return cmd.Start()
}

func isCommandAvailable(name string) bool {
	path, err := exec.LookPath(name)
	if err != nil {
		return false
	}

	fmt.Println(path)
	return path != ""
}

func waitForOpenPort(port int, maxWaitTime time.Duration) {
	for maxWaitTime <= 0 {
		lst, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			time.Sleep(1 * time.Second)
			maxWaitTime -= time.Second
			continue
		}

		lst.Close()
		break
	}
}

// IPFS setups a IPFS repo at the standard place.
// If there is already a repository and the daemon is running, it will do nothing.
// Otherwise it will install IPFS (if it needs to), init a repo, set config and
// bring up the daemon in a fashion that should work for most cases.
// It will output log messages to `out`.
func IPFS(out io.Writer, doSetup, setDefaultConfig bool, ipfsPath string) error {
	if ipfsPath == "" {
		ipfsPath = guessIPFSRepo()
		fmt.Fprintf(out, "-- Guessed IPFS repository as %s\n", ipfsPath)
	} else {
		fmt.Fprintf(out, "-- IPFS repository is supposed to be at %s\n", ipfsPath)
	}

	if ipfsPath == "" && doSetup {
		if !isCommandAvailable("ipfs") {
			fmt.Fprintf(out, "-- There is no »ipfs« command available.\n")
			if err := installIPFS(out); err != nil {
				fmt.Fprintf(out, "-- Failed to install IPFS: %v", err)
				fmt.Fprintf(out, "-- Please refer to »https://docs.ipfs.io/introduction/install«\n")
				fmt.Fprintf(out, "-- to find out on how to install it manually. It is usually very easy.\n")
				fmt.Fprintf(out, "-- Re-run »brig init« once you're done.\n")
				return err
			}
		} else {
			fmt.Fprintf(out, "-- »ipfs« command is available, but no repo found.\n")
		}

		fmt.Fprintf(out, "-- Creating new IPFS repository.\n")
		initIPFS(ipfsPath)
	}

	apiAddr, err := getAPIAddrForPath(ipfsPath)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "-- The API address of the repo is: %s\n", apiAddr)

	if !isRunning(apiAddr) {
		fmt.Fprintf(out, "-- IPFS Daemon does not seem to be running.\n")
		fmt.Fprintf(out, "-- Will start one for you.\n")
		if err := startIpfs(ipfsPath); err != nil {
			return err
		}

		fmt.Fprintf(out, "-- Waiting up to 60s for it to fully boot up...\n")

		port := 5001
		splitApiAddr := strings.Split(apiAddr, "/")
		if len(splitApiAddr) > 0 {
			port, err = strconv.Atoi(splitApiAddr[len(splitApiAddr)-1])
			if err != nil {
				return fmt.Errorf("api addr seems wrong: %v", err)
			}
		}

		waitForOpenPort(port, 60)
	} else {
		fmt.Fprintf(out, "-- IPFS Daemon seems to be running. Let's go!\n")
	}

	if setDefaultConfig {
		fmt.Fprintf(out, "-- Will set some default settings for IPFS.\n")
		fmt.Fprintf(out, "-- These are required for brig to work smoothly.\n")

		if err := configureIPFS(out, ipfsPath); err != nil {
			fmt.Fprintf(out, "-- Failed to set defaults: %v\n", err)
			return err
		}
	}

	return nil
}
