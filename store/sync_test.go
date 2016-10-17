package store

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/disorganizer/brig/id"
	"github.com/disorganizer/brig/util/ipfsutil"
	"github.com/disorganizer/brig/util/testutil"
	"github.com/disorganizer/brig/util/testwith"
)

// TODO: make the with functions globally available.
func withStore(t *testing.T, ID id.ID, IPFS *ipfsutil.Node, fn func(st *Store)) {
	tempDir, err := ioutil.TempDir("", "brig-store-")
	if err != nil {
		t.Fatalf("Could not create temp dir `%s`: %v", tempDir, err)
		return
	}

	defer testutil.Remover(t, tempDir)

	st, err := Open(tempDir, id.NewPeer(ID, "QmW2jc7k5Ug987QEkUx6tJUTdZov7io39MDCiKKp2f57mD"), IPFS)
	if err != nil {
		t.Fatalf("Could not create store: %v", err)
		return
	}

	fn(st)
}

func withIpfsStore(t *testing.T, ID id.ID, fn func(st *Store)) {
	testwith.WithIpfs(t, func(nd *ipfsutil.Node) {
		withStore(t, ID, nd, fn)
	})
}

func TestStoreSync(t *testing.T) {
	data := testutil.CreateDummyBuf(1024)
	path := "/i/have/a/file.png"

	withIpfsStore(t, "alice", func(alice *Store) {
		withIpfsStore(t, "bob", func(bob *Store) {
			if err := alice.StageFromReader(path, bytes.NewReader(data)); err != nil {
				t.Errorf("Failed to stage alice' file: %v", err)
				return
			}

			if err := bob.StageFromReader(path, bytes.NewReader(data)); err != nil {
				t.Errorf("Failed to stage bob's file: %v", err)
				return
			}
			if err := bob.StageFromReader(path+".surprise", bytes.NewReader(data)); err != nil {
				t.Errorf("Failed to stage bob's file: %v", err)
				return
			}

			if err := alice.SyncWith(bob); err != nil {
				t.Errorf("Sync failed: %v", err)
				return
			}

			fmt.Println("+=======")

			if err := alice.SyncWith(bob); err != nil {
				t.Errorf("Sync failed: %v", err)
				return
			}
		})

		printTree(alice.fs)
	})
}
