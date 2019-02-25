package httpipfs

import (
	"bytes"
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

func withIpfs(t *testing.T, portOff int, fn func(t *testing.T, apiPort int)) {
	ipfsPath, err := ioutil.TempDir("", "brig-httpipfs-test-")
	require.Nil(t, err)
	defer os.RemoveAll(ipfsPath)

	gwtPort := 8081 + portOff
	swmPort := 4001 + portOff
	apiPort := 5011 + portOff

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

func withDoubleIpfs(t *testing.T, portOff int, fn func(t *testing.T, apiPortA, apiPortB int)) {
	chPortA := make(chan int)
	chPortB := make(chan int)
	stop := make(chan bool)

	go withIpfs(t, portOff, func(t *testing.T, apiPortA int) {
		chPortA <- apiPortA
		<-stop
	})

	go withIpfs(t, portOff+1, func(t *testing.T, apiPortB int) {
		chPortB <- apiPortB
		<-stop
	})

	fn(t, <-chPortA, <-chPortB)
	stop <- true
	stop <- true
}

func TestIpfsStartup(t *testing.T) {
	withIpfs(t, 1, func(t *testing.T, apiPort int) {
		nd, err := NewNode(apiPort)
		require.Nil(t, err)

		hash, err := nd.Add(bytes.NewReader([]byte("hello")))
		require.Nil(t, err, fmt.Sprintf("%v", err))
		require.Equal(t, "QmWfVY9y3xjsixTgbd9AorQxH7VtMpzfx2HaWtsoUYecaX", hash.String())
	})
}

func TestDoubleIpfsStartup(t *testing.T) {
	withDoubleIpfs(t, 1, func(t *testing.T, apiPortA, apiPortB int) {
		ndA, err := NewNode(apiPortA)
		require.Nil(t, err)

		ndB, err := NewNode(apiPortB)
		require.Nil(t, err)

		idA, err := ndA.Identity()
		require.Nil(t, err, fmt.Sprintf("%v", err))

		idB, err := ndB.Identity()
		require.Nil(t, err)

		require.NotEqual(t, idA.Addr, idB.Addr)
	})
}
