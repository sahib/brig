package ipfs

import (
	"io/ioutil"
	"testing"

	"github.com/disorganizer/brig/util/testutil"
)

// TODO: Make this a test.
// func Main() {
// 	alice, aliceId := ipfsWithPort(9000)
// 	bob, bobId := ipfsWithPort(9001)
// 	// bobId := "QmSUu41eAf4TEDernpXR4S4Z33Hrae3LwoFh2v8vuPK2yy"
//
// 	fmt.Println("WAIT 10")
// 	time.Sleep(10 * time.Second)
//
// 	go func() {
// 		fmt.Println("bob listen")
// 		lst, err := bob.Listen("/p2p/bubu")
// 		if err != nil {
// 			fmt.Println("listen failed", err)
// 		}
//
// 		fmt.Println("bob accept")
// 		conn, err := lst.Accept()
// 		if err != nil {
// 			fmt.Println("Accept failed", err)
// 		}
//
// 		fmt.Println("bob write")
// 		_, err = conn.Write([]byte("Hello world"))
// 		if err != nil {
// 			fmt.Println("write failed", err)
// 		}
//
// 		time.Sleep(1 * time.Second)
//
// 		if err := conn.Close(); err != nil {
// 			fmt.Println("conn close fail", err)
// 		}
// 	}()
//
// 	fmt.Println("WAIT 5")
// 	time.Sleep(5 * time.Second)
//
// 	fmt.Println("DIAL", aliceId, "->", bobId)
// 	conn, err := alice.Dial(bobId, "/p2p/bubu")
// 	if err != nil {
// 		fmt.Println("Dial failed", err)
// 	}
//
// 	_, err = conn.Write([]byte("hello you too"))
// 	if err != nil {
// 		fmt.Println("write failed", err)
// 	}
//
// 	data, err := ioutil.ReadAll(conn)
// 	fmt.Println("READ", data, err)
// 	conn.Close()
// }

func WithIpfs(t *testing.T, f func(*Node)) {
	tmpDir, err := ioutil.TempDir("", "brig-ipfs-")
	if err != nil {
		t.Errorf("Cannot create temp dir %v", err)
		return
	}

	WithIpfsAtPath(t, tmpDir, f)
}

func WithIpfsAtPath(t *testing.T, root string, f func(*Node)) {
	WithIpfsAtPathAndPort(t, root, 4001, f)
}

func WithIpfsAtPort(t *testing.T, port int, f func(*Node)) {
	tmpDir, err := ioutil.TempDir("", "brig-ipfs")
	if err != nil {
		t.Errorf("Cannot create temp dir %v", err)
		return
	}

	WithIpfsAtPathAndPort(t, tmpDir, port, f)
}

func WithIpfsAtPathAndPort(t *testing.T, root string, port int, f func(*Node)) {
	WithIpfsRepo(t, root, func(path string) {
		nd, err := NewWithPort(path, port)
		if err != nil {
			t.Fatalf("with ipfs: %v", err)
		}

		f(nd)

		if err := nd.Close(); err != nil {
			t.Errorf("Closing ipfs-daemon failed: %v", err)
		}
	})
}

func WithIpfsRepo(t *testing.T, root string, f func(repoPath string)) {
	if err := Init(root, 1024); err != nil {
		t.Errorf("Could not create ipfs repo: %v", err)
		return
	}

	defer testutil.Remover(t, root)

	f(root)
}
