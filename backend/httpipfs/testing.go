package httpipfs

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
	"time"

	shell "github.com/ipfs/go-ipfs-api"
	"github.com/stretchr/testify/require"
)

// WithIpfs starts a new IPFS instance and calls `fn` with the API port to it.
// `portOff` is the offset to add on all standard ports.
func WithIpfs(t *testing.T, portOff int, fn func(t *testing.T, ipfsPath string)) {
	ipfsPath, err := ioutil.TempDir("", "brig-httpipfs-test-")
	require.Nil(t, err)
	defer os.RemoveAll(ipfsPath)

	gwtPort := 8081 + portOff
	swmPort := 4001 + portOff
	apiPort := 5011 + portOff

	os.Setenv("IPFS_PATH", ipfsPath)
	script := [][]string{
		{"ipfs", "init"},
		{"ipfs", "config", "--json", "Addresses.Swarm", fmt.Sprintf("[\"/ip4/127.0.0.1/tcp/%d\"]", swmPort)},
		{"ipfs", "config", "--json", "Experimental.Libp2pStreamMounting", "true"},
		{"ipfs", "config", "Addresses.API", fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", apiPort)},
		{"ipfs", "config", "Addresses.Gateway", fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", gwtPort)},
	}

	for _, line := range script {
		cmd := exec.Command(line[0], line[1:]...)
		cmd.Env = append(cmd.Env, fmt.Sprintf("IPFS_PATH=%s", ipfsPath))
		err := cmd.Run()
		require.NoError(t, err)
	}

	daemonCmd := exec.Command("ipfs", "daemon", "--enable-pubsub-experiment")
	// daemonCmd.Stdout = os.Stdout
	// daemonCmd.Stderr = os.Stdout
	daemonCmd.Env = append(daemonCmd.Env, fmt.Sprintf("IPFS_PATH=%s", ipfsPath))
	require.Nil(t, daemonCmd.Start())

	defer func() {
		require.Nil(t, daemonCmd.Process.Kill())
	}()

	// Wait until the daemon actually offers the API interface:
	localAddr := fmt.Sprintf("localhost:%d", apiPort)
	for tries := 0; tries < 200; tries++ {
		if shell.NewShell(localAddr).IsUp() {
			break
		}

		time.Sleep(100 * time.Millisecond)
	}

	// Actually call the test:
	fn(t, ipfsPath)

}

// WithDoubleIpfs starts two IPFS instances in parallel.
func WithDoubleIpfs(t *testing.T, portOff int, fn func(t *testing.T, ipfsPathA, ipfsPathB string)) {
	chPathA := make(chan string)
	chPathB := make(chan string)
	stop := make(chan bool, 2)

	go WithIpfs(t, portOff, func(t *testing.T, ipfsPathA string) {
		chPathA <- ipfsPathA
		<-stop
	})

	go WithIpfs(t, portOff+1, func(t *testing.T, ipfsPathB string) {
		chPathB <- ipfsPathB
		<-stop
	})

	fn(t, <-chPathA, <-chPathB)
	stop <- true
	stop <- true
}
