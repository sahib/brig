package httpipfs

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// WithIpfs starts a new IPFS instance and calls `fn` with the API port to it.
// `portOff` is the offset to add on all standard ports.
func WithIpfs(t *testing.T, portOff int, fn func(t *testing.T, apiPort int)) {
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
		require.Nil(t, err, strings.Join(line, " "))
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
	for tries := 0; tries < 200; tries++ {
		conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", apiPort))
		if err == nil {
			conn.Close()
			break
		}

		time.Sleep(100 * time.Millisecond)
	}

	// Actually call the test:
	fn(t, apiPort)

}

// WithDoubleIpfs starts two IPFS instances in parallel.
func WithDoubleIpfs(t *testing.T, portOff int, fn func(t *testing.T, apiPortA, apiPortB int)) {
	chPortA := make(chan int)
	chPortB := make(chan int)
	stop := make(chan bool)

	go WithIpfs(t, portOff, func(t *testing.T, apiPortA int) {
		chPortA <- apiPortA
		<-stop
	})

	go WithIpfs(t, portOff+1, func(t *testing.T, apiPortB int) {
		chPortB <- apiPortB
		<-stop
	})

	fn(t, <-chPortA, <-chPortB)
	stop <- true
	stop <- true
}
