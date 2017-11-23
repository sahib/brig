package net

import (
	"context"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/disorganizer/brig/backend"
	"github.com/disorganizer/brig/net/peer"
	"github.com/disorganizer/brig/repo"
)

func withClientFor(who string, t *testing.T, fn func(ctl *Client)) {
	tmpFolder, err := ioutil.TempDir("", "brig-net-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	defer os.Remove(tmpFolder)

	if err := repo.Init(tmpFolder, "alice", "xxx", "mock"); err != nil {
		t.Fatalf("Failed to init repo at: %v", err)
	}

	rp, err := repo.Open(tmpFolder, "xxx")
	if err != nil {
		t.Fatalf("Failed to open repository: %v", err)
	}

	ownPubKey, err := rp.Keyring().OwnPubKey()
	if err != nil {
		t.Fatalf("Failed to get own pub key: %v", err)
	}

	rm := repo.Remote{
		Name:        who,
		Fingerprint: peer.BuildFingerprint("addr", ownPubKey),
	}
	if err := rp.Remotes.AddRemote(rm); err != nil {
		t.Fatalf("Failed to add remote: %v", err)
	}

	bk, err := backend.FromName("mock", "")
	if err != nil {
		t.Fatalf("Failed to get mock backend (wtf?): %v", err)
	}

	srv, err := NewServer(rp, bk)
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	wg := &sync.WaitGroup{}
	go func() {
		wg.Add(1)
		defer wg.Done()

		if err := srv.Serve(); err != nil {
			t.Fatalf("Failed to start serving process: %v", err)
		}

		if err := srv.Close(); err != nil {
			t.Fatalf("Failed to close server properly: %v", err)
		}
	}()

	ctx := context.Background()
	ctl, err := Dial(who, rp, bk, ctx)
	if err != nil {
		t.Fatalf("Dial to %v failed: %v", who, err)
	}

	// Actually execute the test...
	fn(ctl)

	if err := ctl.Close(); err != nil {
		t.Fatalf("Failed to close conn")
	}

	if err := rp.Close("xxx"); err != nil {
		t.Fatalf("Failed to close repo: %v", err)
	}

	// Quit the server.
	srv.Quit()
	wg.Wait()
}

func TestClientPing(t *testing.T) {
	withClientFor("bob", t, func(ctl *Client) {
		now := time.Now()

		for i := 0; i < 100; i++ {
			if err := ctl.Ping(); err != nil {
				t.Fatalf("Ping to bob failed: %v", err)
			}
		}
	})
}
