package net

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/sahib/brig/backend"
	"github.com/sahib/brig/catfs"
	ie "github.com/sahib/brig/catfs/errors"
	"github.com/sahib/brig/net/peer"
	"github.com/sahib/brig/repo"
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

	defer os.RemoveAll(tmpFolder)

	if err := repo.Init(tmpFolder, who, "xxx", "mock"); err != nil {
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

	fs, err := rp.FS(rp.CurrentUser(), bk)
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

		aliceFs, err := u.rp.FS("alice", u.bk)
		if err != nil {
			t.Fatalf("Failed to get empty bob fs: %v", err)
		}

		_, err = aliceFs.Stat(filePath)
		if !ie.IsNoSuchFileError(err) {
			t.Fatalf("File has existed in bob's empty store (wtf?): %v", err)
		}

		if err := aliceFs.Import(data); err != nil {
			t.Fatalf("Failed to import data: %v", err)
		}

		info, err := aliceFs.Stat(filePath)
		if err != nil {
			t.Fatalf("Failed to read file exported from alice: %v", err)
		}

		// Check superficially that store was imported right:
		require.Equal(t, info.Path, filePath)
		require.Equal(t, info.User, "bob")
		require.Equal(t, info.Size, uint64(3))
		require.Equal(t, info.IsDir, false)

		r, err := aliceFs.Cat(filePath)
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

func TestClientFetchPatch(t *testing.T) {
	withClientFor("bob", t, func(u testUnit) {
		// Create a new file in bob's fs.
		require.Nil(t, u.fs.Stage("/new_file", bytes.NewReader([]byte{1, 2, 3})))

		// Get a patch from Bob's FS.
		patchData, err := u.ctl.FetchPatch(0)
		require.Nil(t, err)
		require.NotNil(t, patchData)

		// Create a new empty FS for alice.
		aliceFs, err := u.rp.FS("alice", u.bk)
		if err != nil {
			t.Fatalf("Failed to get empty bob fs: %v", err)
		}

		// It should have the initial patch version of 0.
		lastPatchIdx, err := aliceFs.LastPatchIndex()
		require.Nil(t, err)
		require.Equal(t, int64(0), lastPatchIdx)

		// After applying the patch, we should have bob's data.
		require.Nil(t, aliceFs.ApplyPatch(patchData))
		newFileInfo, err := aliceFs.Stat("/new_file")
		require.Nil(t, err)
		require.Equal(t, "/new_file", newFileInfo.Path)
		require.Equal(t, uint64(3), newFileInfo.Size)

		// Check that the new patch version is 2 (i.e. bob has two commits)
		lastPatchIdx, err = aliceFs.LastPatchIndex()
		require.Nil(t, err)
		require.Equal(t, int64(2), lastPatchIdx)

		// If we fetch the same patch again, it will be empty.
		// (data will be not len=0, but no real contents)
		patchData, err = u.ctl.FetchPatch(2)
		require.Nil(t, err)
		require.NotNil(t, patchData)
		require.Nil(t, aliceFs.ApplyPatch(patchData))

		// Last patch was empty, so should not bump the version.
		lastPatchIdx, err = aliceFs.LastPatchIndex()
		require.Nil(t, err)
		require.Equal(t, int64(2), lastPatchIdx)

		// Bob's patch index should not have changed.
		// For bob, the patch index does not make really sense,
		// since he's the owner of the fs and always has the latest version.
		lastBobPatchIdx, err := u.fs.LastPatchIndex()
		require.Nil(t, err)
		require.Equal(t, int64(0), lastBobPatchIdx)
	})
}

func TestClientCompleteFetchAllowed(t *testing.T) {
	withClientFor("bob", t, func(u testUnit) {
		isAllowed, err := u.ctl.IsCompleteFetchAllowed()
		require.Nil(t, err)

		rmt, err := u.rp.Remotes.Remote("bob")
		require.Nil(t, err)
		require.True(t, isAllowed)

		// Make the remote have only access to a specific sub folder:
		err = u.rp.Remotes.SetRemote("bob", repo.Remote{
			Fingerprint: rmt.Fingerprint,
			Name:        rmt.Name,
			Folders: []repo.Folder{
				{
					Folder: "/photos",
				},
			},
		})
		require.Nil(t, err)

		isAllowed, err = u.ctl.IsCompleteFetchAllowed()
		require.Nil(t, err)
		require.False(t, isAllowed)

		// Try again with the root folder enabled:
		err = u.rp.Remotes.SetRemote("bob", repo.Remote{
			Fingerprint: rmt.Fingerprint,
			Name:        rmt.Name,
			Folders: []repo.Folder{
				{
					Folder: "/",
				},
				{
					Folder: "/photos",
				},
			},
		})
		require.Nil(t, err)

		// This also tests that we are able to change remote config
		// without needing to reconnect.
		isAllowed, err = u.ctl.IsCompleteFetchAllowed()
		require.Nil(t, err)
		require.True(t, isAllowed)
	})
}
