package net

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
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
	srv *Server
	fs  *catfs.FS
	rp  *repo.Repository
	bk  backend.Backend
}

func withNetServer(t *testing.T, name string, basePath string, fn func(u testUnit)) {
	basePath, err := ioutil.TempDir("", "brig-ctl-test")
	require.Nil(t, err)

	netDbPath := "/tmp/brig-net-test-netdb"
	regDbPath := "/tmp/brig-net-test-reg.yml"
	require.Nil(t, err)

	defer func() {
		require.Nil(t, os.RemoveAll(basePath))
		require.Nil(t, os.RemoveAll(netDbPath))
		require.Nil(t, os.RemoveAll(regDbPath))
	}()

	// The following env vars are only read in FromName.
	require.Nil(t, os.Setenv("BRIG_MOCK_USER", name))
	require.Nil(t, os.Setenv("BRIG_MOCK_NET_DB_PATH", netDbPath))
	bk, err := backend.FromName("mock", basePath, "")
	require.Nil(t, err)

	err = repo.Init(basePath, name, "password", "mock", 6666)
	require.Nil(t, err)

	rp, err := repo.Open(basePath, "password")
	require.Nil(t, err)

	srv, err := NewServer(rp, bk, nil)
	require.Nil(t, err)

	fs, err := rp.FS(name, bk)
	require.Nil(t, err)

	waitForDeath := make(chan bool)
	go func() {
		defer func() {
			waitForDeath <- true
		}()
		require.Nil(t, srv.Serve())
		require.Nil(t, srv.Close())
	}()

	// Allow a short time for the server go routine to fully boot up.
	time.Sleep(50 * time.Millisecond)

	// Run the actual test function:
	fn(testUnit{
		rp:  rp,
		bk:  bk,
		fs:  fs,
		srv: srv,
	})

	// wait until serve was done.
	srv.Quit()
	<-waitForDeath
}

func buildFingerprint(t *testing.T, u testUnit) peer.Fingerprint {
	ownPubKey, err := u.rp.Keyring().OwnPubKey()
	require.Nil(t, err)

	self, err := u.srv.Identity()
	require.Nil(t, err)

	return peer.BuildFingerprint(self.Addr, ownPubKey)
}

func withNetPair(t *testing.T, fn func(a, b testUnit)) {
	basePath, err := ioutil.TempDir("", "brig-net-test")
	require.Nil(t, err)

	defer func() {
		require.Nil(t, os.RemoveAll(basePath))
	}()

	withNetServer(t, "alice", basePath, func(a testUnit) {
		withNetServer(t, "bob", basePath, func(b testUnit) {
			// Add each other's fingerprints:
			require.Nil(t, a.rp.Remotes.AddOrUpdateRemote(repo.Remote{
				Name:        "bob",
				Fingerprint: buildFingerprint(t, b),
			}))
			require.Nil(t, b.rp.Remotes.AddOrUpdateRemote(repo.Remote{
				Name:        "alice",
				Fingerprint: buildFingerprint(t, a),
			}))

			ctx := context.Background()
			aliCtl, err := Dial(ctx, "alice", b.rp, b.bk, nil)
			require.Nil(t, err)

			bobCtl, err := Dial(ctx, "bob", a.rp, a.bk, nil)
			require.Nil(t, err)

			a.ctl = bobCtl
			b.ctl = aliCtl

			fn(a, b)
		})
	})
}

func TestClientPing(t *testing.T) {
	withNetPair(t, func(a, b testUnit) {
		for i := 0; i < 100; i++ {
			if err := a.ctl.Ping(); err != nil {
				t.Fatalf("ping to bob failed: %v", err)
			}

			if err := b.ctl.Ping(); err != nil {
				t.Fatalf("ping to alice failed: %v", err)
			}
		}
	})
}

func TestClientFetchStore(t *testing.T) {
	withNetPair(t, func(a, b testUnit) {
		filePath := "/a/new/name/has/been/born"
		fileData := []byte{1, 2, 3}
		fileSrc := bytes.NewReader(fileData)

		if err := a.fs.Stage(filePath, fileSrc); err != nil {
			t.Fatalf("failed to stage simple file: %v", err)
		}

		data, err := b.ctl.FetchStore()
		if err != nil {
			t.Fatalf("failed to read store: %v", err)
		}

		aliceFsAtBob, err := b.rp.FS("alice", b.bk)
		if err != nil {
			t.Fatalf("Failed to get empty bob fs: %v", err)
		}

		_, err = aliceFsAtBob.Stat(filePath)
		if !ie.IsNoSuchFileError(err) {
			t.Fatalf("File has existed in bob's empty store (wtf?): %v", err)
		}

		if err := aliceFsAtBob.Import(data); err != nil {
			t.Fatalf("Failed to import data: %v", err)
		}

		info, err := aliceFsAtBob.Stat(filePath)
		if err != nil {
			t.Fatalf("Failed to read file exported from alice: %v", err)
		}

		// Check superficially that store was imported right:
		require.Equal(t, info.Path, filePath)
		require.Equal(t, info.User, "alice")
		require.Equal(t, info.Size, uint64(3))
		require.Equal(t, info.IsDir, false)
	})
}

func TestClientFetchPatch(t *testing.T) {
	withNetPair(t, func(a, b testUnit) {
		// Create a new file in alice's fs.
		require.Nil(t, a.fs.Stage("/new_file", bytes.NewReader([]byte{1, 2, 3})))

		// Get a patch from alice:
		patchData, err := b.ctl.FetchPatch(0)
		require.NoError(t, err)
		require.NotNil(t, patchData)

		// Create a new empty FS for alice.
		aliceFsAtBob, err := b.rp.FS("alice", b.bk)
		if err != nil {
			t.Fatalf("Failed to get empty bob fs: %v", err)
		}

		// It should have the initial patch version of 0.
		lastPatchIdx, err := aliceFsAtBob.LastPatchIndex()
		require.NoError(t, err)
		require.Equal(t, int64(0), lastPatchIdx)

		// After applying the patch, we should have bob's data.
		require.NoError(t, aliceFsAtBob.ApplyPatch(patchData))
		newFileInfo, err := aliceFsAtBob.Stat("/new_file")
		require.NoError(t, err)
		require.Equal(t, "/new_file", newFileInfo.Path)
		require.Equal(t, uint64(3), newFileInfo.Size)

		// Check that the new patch version is 2 (i.e. bob has two commits)
		aliceFsAtBob.Log("HEAD", func(c *catfs.Commit) error {
			fmt.Println(c)
			return nil
		})

		lastPatchIdx, err = aliceFsAtBob.LastPatchIndex()
		require.NoError(t, err)
		require.Equal(t, int64(2), lastPatchIdx)

		// If we fetch the same patch again, it will be empty.
		// (data will be not len=0, but no real contents)
		patchData, err = b.ctl.FetchPatch(2)
		require.NoError(t, err)
		require.NotNil(t, patchData)
		require.NoError(t, aliceFsAtBob.ApplyPatch(patchData))

		// Last patch was empty, so should not bump the version.
		lastPatchIdx, err = aliceFsAtBob.LastPatchIndex()
		require.NoError(t, err)
		require.Equal(t, int64(2), lastPatchIdx)

		// Bob's patch index should not have changed.
		// For bob, the patch index does not make really sense,
		// since he's the owner of the fs and always has the latest version.
		lastBobPatchIdx, err := b.fs.LastPatchIndex()
		require.NoError(t, err)
		require.Equal(t, int64(0), lastBobPatchIdx)
	})
}

func TestClientCompleteFetchAllowed(t *testing.T) {
	withNetPair(t, func(a, b testUnit) {
		isAllowed, err := b.ctl.IsCompleteFetchAllowed()
		require.Nil(t, err)
		require.True(t, isAllowed)

		// Make the remote have only access to a specific sub folder:
		rmt, err := a.rp.Remotes.Remote("bob")
		require.Nil(t, err)

		err = a.rp.Remotes.AddOrUpdateRemote(repo.Remote{
			Fingerprint: rmt.Fingerprint,
			Name:        rmt.Name,
			Folders: []repo.Folder{
				{
					Folder: "/photos",
				},
			},
		})
		require.Nil(t, err)

		isAllowed, err = b.ctl.IsCompleteFetchAllowed()
		require.Nil(t, err)
		require.False(t, isAllowed)

		// Try again with the root folder enabled:
		err = a.rp.Remotes.AddOrUpdateRemote(repo.Remote{
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
		isAllowed, err = b.ctl.IsCompleteFetchAllowed()
		require.Nil(t, err)
		require.True(t, isAllowed)
	})
}
