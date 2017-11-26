package net

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/disorganizer/brig/backend"
	"github.com/disorganizer/brig/catfs"
	ie "github.com/disorganizer/brig/catfs/errors"
	"github.com/disorganizer/brig/net/peer"
	"github.com/disorganizer/brig/repo"
	"github.com/stretchr/testify/require"
)

type testUnit struct {
	ctl *Client
	fs  *catfs.FS
	rp  *repo.Repository
	bk  backend.Backend
}

func withClientFor(who string, t *testing.T, fn func(u testUnit)) {
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

	// Allow a short time for the server go routine to fully boot up.
	time.Sleep(50 * time.Millisecond)

	ctx := context.Background()
	ctl, err := Dial(who, rp, bk, ctx)
	if err != nil {
		t.Fatalf("Dial to %v failed: %v", who, err)
	}

	fs, err := rp.OwnFS(bk)
	if err != nil {
		t.Fatalf("Failed to retrieve own fs: %v", err)
	}

	// Actually execute the test...
	fn(testUnit{
		fs:  fs,
		rp:  rp,
		ctl: ctl,
		bk:  bk,
	})

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
	withClientFor("bob", t, func(u testUnit) {
		for i := 0; i < 100; i++ {
			if err := u.ctl.Ping(); err != nil {
				t.Fatalf("Ping to bob failed: %v", err)
			}
		}
	})
}

func TestClientFetchStore(t *testing.T) {
	withClientFor("bob", t, func(u testUnit) {
		filePath := "/a/new/name/has/been/born"
		fileData := []byte{1, 2, 3}
		fileSrc := bytes.NewReader(fileData)

		if err := u.fs.Stage(filePath, fileSrc); err != nil {
			t.Fatalf("Failed to stage simple file: %v", err)
		}

		data, err := u.ctl.FetchStore()
		if err != nil {
			t.Fatalf("Failed to read store: %v", err)
		}

		bobFs, err := u.rp.FS("bob", u.bk)
		if err != nil {
			t.Fatalf("Failed to get empty bob fs: %v", err)
		}

		_, err = bobFs.Stat(filePath)
		if !ie.IsNoSuchFileError(err) {
			t.Fatalf("File has existed in bob's empty store (wtf?)")
		}

		if err := bobFs.Import(data); err != nil {
			t.Fatalf("Failed to import data: %v", err)
		}

		info, err := bobFs.Stat(filePath)
		if err != nil {
			t.Fatalf("Failed to read file exported from alice: %v", err)
		}

		// Check superficially that store was imported right:
		require.Equal(t, info.Path, filePath)
		require.Equal(t, info.Size, 3)
		require.Equal(t, info.IsDir, false)

		r, err := bobFs.Cat(filePath)
		if err != nil {
			t.Fatalf("Failed to cat exported file: %v", err)
		}

		bobData, err := ioutil.ReadAll(r)
		if err != nil {
			t.Fatalf("Failed to read bob data: %v", err)
		}

		require.Equal(t, fileData, bobData)
	})
}
