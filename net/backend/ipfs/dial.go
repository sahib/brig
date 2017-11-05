package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	ipfsconfig "github.com/ipfs/go-ipfs/repo/config"
	"github.com/ipfs/go-ipfs/repo/fsrepo"
)

// InitRepo creates an initialized .ipfs directory in the directory `path`.
// The generated RSA key will have `keySize` bits.
func InitRepo(path string, keySize int) error {
	if err := os.MkdirAll(path, 0744); err != nil {
		return err
	}

	cfg, err := ipfsconfig.Init(ioutil.Discard, keySize)
	if err != nil {
		return err
	}

	if err := fsrepo.Init(path, cfg); err != nil {
		return err
	}

	return nil
}

func ipfsWithPort(port int) (*Node, string) {
	repoPath := filepath.Join(os.TempDir(), fmt.Sprintf("ipfs-repo-%d", port))
	os.RemoveAll(repoPath)

	fmt.Println("Init repo", port)
	if err := InitRepo(repoPath, 1024); err != nil {
		fmt.Println("init failed", err)
	}

	fmt.Println("Going online", port)
	nd, err := NewWithPort(repoPath, port)
	if err != nil {
		fmt.Println("Starting daemon failed", err)
	}

	self, err := nd.Identity()
	if err != nil {
		fmt.Println("Failed to guess self", err)
	}

	// fmt.Println("Going offline")
	// if err := nd.Offline(); err != nil {
	// 	fmt.Println("Failed to go offline again", err)
	// }
	return nd, self
}

func main() {
	alice, aliceId := ipfsWithPort(9000)
	bob, bobId := ipfsWithPort(9001)
	// bobId := "QmSUu41eAf4TEDernpXR4S4Z33Hrae3LwoFh2v8vuPK2yy"

	fmt.Println("WAIT 10")
	time.Sleep(10 * time.Second)

	go func() {
		fmt.Println("bob listen")
		lst, err := bob.Listen("/p2p/bubu")
		if err != nil {
			fmt.Println("listen failed", err)
		}

		fmt.Println("bob accept")
		conn, err := lst.Accept()
		if err != nil {
			fmt.Println("Accept failed", err)
		}

		fmt.Println("bob write")
		_, err = conn.Write([]byte("Hello world"))
		if err != nil {
			fmt.Println("write failed", err)
		}

		time.Sleep(1 * time.Second)

		if err := conn.Close(); err != nil {
			fmt.Println("conn close fail", err)
		}
	}()

	fmt.Println("WAIT 5")
	time.Sleep(5 * time.Second)

	fmt.Println("DIAL", aliceId, "->", bobId)
	conn, err := alice.Dial(bobId, "/p2p/bubu")
	if err != nil {
		fmt.Println("Dial failed", err)
	}

	_, err = conn.Write([]byte("hello you too"))
	if err != nil {
		fmt.Println("write failed", err)
	}

	data, err := ioutil.ReadAll(conn)
	fmt.Println("READ", data, err)
	conn.Close()
}
