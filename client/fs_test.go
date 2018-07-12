package client

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/sahib/brig/server"
	colorlog "github.com/sahib/brig/util/log"
	"github.com/stretchr/testify/require"
)

func init() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&colorlog.ColorfulLogFormatter{})
}

func withDaemon(t *testing.T, name string, port, backendPort int, basePath string, fn func(ctl *Client)) {
	if basePath == "" {
		var err error
		basePath, err = ioutil.TempDir("", "brig-ctl-test")
		require.Nil(t, err)

		defer func() {
			os.RemoveAll(basePath)
		}()
	}

	// This is a hacky way to tell the mock backend what port it should use:
	fullPath := fmt.Sprintf("%s/user=%s-port=%d", basePath, name, backendPort)
	require.Nil(t, os.MkdirAll(fullPath, 0700))

	srv, err := server.BootServer(fullPath, "password", "", "localhost", port)
	require.Nil(t, err)

	waitForDeath := make(chan bool)
	go func() {
		defer func() {
			waitForDeath <- true
		}()
		require.Nil(t, srv.Serve())
	}()

	ctl, err := Dial(context.Background(), port)
	require.Nil(t, err)

	err = ctl.Init(fullPath, name, "password", "mock")
	require.Nil(t, err)

	// Run the actual test function:
	fn(ctl)

	// Send
	require.Nil(t, ctl.Quit())

	// wait until serve was done.
	<-waitForDeath
}

func TestStageAndCat(t *testing.T) {
	withDaemon(t, "alice", 6667, 9999, "", func(ctl *Client) {
		fd, err := ioutil.TempFile("", "brig-dummy-data")
		path := fd.Name()

		require.Nil(t, err)
		_, err = fd.Write([]byte("hello"))
		require.Nil(t, err)
		require.Nil(t, fd.Close())

		require.Nil(t, ctl.Stage(path, "/hello"))
		rw, err := ctl.Cat("hello")
		require.Nil(t, err)

		data, err := ioutil.ReadAll(rw)
		require.Nil(t, err)

		require.Equal(t, []byte("hello"), data)
		require.Nil(t, rw.Close())
	})
}

func TestMkdir(t *testing.T) {
	withDaemon(t, "alice", 6667, 9999, "", func(ctl *Client) {
		// Create something nested with -p...
		require.Nil(t, ctl.Mkdir("/a/b/c", true))

		// Create it twice...
		require.Nil(t, ctl.Mkdir("/a/b/c", true))

		// Create something nested without -p
		err := ctl.Mkdir("/x/y/z", false)
		require.Contains(t, err.Error(), "No such file")

		require.Nil(t, ctl.Mkdir("/x", false))
		require.Nil(t, ctl.Mkdir("/x/y", false))
		require.Nil(t, ctl.Mkdir("/x/y/z", false))

		lst, err := ctl.List("/", -1)
		require.Nil(t, err)

		paths := []string{}
		for _, info := range lst {
			paths = append(paths, info.Path)
		}

		sort.Strings(paths)
		require.Equal(t, paths, []string{
			"/",
			"/a",
			"/a/b",
			"/a/b/c",
			"/x",
			"/x/y",
			"/x/y/z",
		})
	})
}

func withConnectedDaemonPair(t *testing.T, fn func(aliCtl, bobCtl *Client)) {
	// Use a shared directory for our shared data:
	basePath, err := ioutil.TempDir("", "brig-test-sync-pair-test")
	require.Nil(t, err)

	defer func() {
		os.RemoveAll(basePath)
	}()

	withDaemon(t, "alice", 6668, 9998, basePath, func(aliCtl *Client) {
		withDaemon(t, "bob", 6669, 9999, basePath, func(bobCtl *Client) {
			aliWhoami, err := aliCtl.Whoami()
			require.Nil(t, err)

			bobWhoami, err := bobCtl.Whoami()
			require.Nil(t, err)

			// add bob to alice as remote
			err = aliCtl.RemoteAdd(Remote{
				Name:        "bob",
				Fingerprint: bobWhoami.Fingerprint,
			})
			require.Nil(t, err)

			// add alice to bob as remote
			err = bobCtl.RemoteAdd(Remote{
				Name:        "alice",
				Fingerprint: aliWhoami.Fingerprint,
			})
			require.Nil(t, err)

			fn(aliCtl, bobCtl)
		})
	})
}

func TestSync(t *testing.T) {
	withConnectedDaemonPair(t, func(aliCtl, bobCtl *Client) {
		err := aliCtl.StageFromReader("/ali_file", bytes.NewReader([]byte{42}))
		require.Nil(t, err)

		err = bobCtl.StageFromReader("/bob_file", bytes.NewReader([]byte{23}))
		require.Nil(t, err)

		err = aliCtl.Sync("bob", true)
		require.Nil(t, err)

		err = bobCtl.Sync("alice", true)
		require.Nil(t, err)

		// We cannot query the file contents, since the mock backend
		// does not yet store the file content anywhere.
		bobFileStat, err := aliCtl.Stat("/bob_file")
		require.Nil(t, err)
		require.Equal(t, "/bob_file", bobFileStat.Path)

		aliFileStat, err := bobCtl.Stat("/ali_file")
		require.Nil(t, err)
		require.Equal(t, "/ali_file", aliFileStat.Path)
	})
}

func TestSyncPartial(t *testing.T) {
	withConnectedDaemonPair(t, func(aliCtl, bobCtl *Client) {
		aliWhoami, err := aliCtl.Whoami()
		require.Nil(t, err)

		bobWhoami, err := bobCtl.Whoami()
		require.Nil(t, err)

		err = aliCtl.RemoteSave([]Remote{
			{
				Name:        "bob",
				Fingerprint: bobWhoami.Fingerprint,
				Folders: []RemoteFolder{
					{
						Folder: "/photos",
					},
				},
			},
		})

		err = bobCtl.RemoteSave([]Remote{
			{
				Name:        "alice",
				Fingerprint: aliWhoami.Fingerprint,
				Folders: []RemoteFolder{
					{
						Folder: "/photos",
					},
				},
			},
		})

		err = aliCtl.StageFromReader("/docs/ali_secret.txt", bytes.NewReader([]byte{0}))
		require.Nil(t, err)
		err = aliCtl.StageFromReader("/photos/ali.png", bytes.NewReader([]byte{42}))
		require.Nil(t, err)

		err = bobCtl.StageFromReader("/docs/bob_secret.txt", bytes.NewReader([]byte{0}))
		require.Nil(t, err)
		err = bobCtl.StageFromReader("/photos/bob.png", bytes.NewReader([]byte{23}))
		require.Nil(t, err)

		err = aliCtl.Sync("bob", true)
		require.Nil(t, err)

		err = bobCtl.Sync("alice", true)
		require.Nil(t, err)

		// We cannot query the file contents, since the mock backend
		// does not yet store the file content anywhere.
		aliLs, err := aliCtl.List("/", -1)
		require.Nil(t, err)

		aliPaths := []string{}
		for _, entry := range aliLs {
			aliPaths = append(aliPaths, entry.Path)
		}

		bobLs, err := bobCtl.List("/", -1)
		require.Nil(t, err)

		bobPaths := []string{}
		for _, entry := range bobLs {
			bobPaths = append(bobPaths, entry.Path)
		}

		require.Equal(
			t,
			[]string{
				"/",
				"/docs",
				"/photos",
				"/docs/ali_secret.txt",
				"/photos/ali.png",
				"/photos/bob.png",
			},
			aliPaths,
		)

		require.Equal(
			t,
			[]string{
				"/",
				"/docs",
				"/photos",
				"/docs/bob_secret.txt",
				"/photos/ali.png",
				"/photos/bob.png",
			},
			bobPaths,
		)
	})
}
