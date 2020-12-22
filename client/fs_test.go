package client

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/sahib/brig/repo"
	"github.com/sahib/brig/server"
	"github.com/sahib/brig/util"
	colorLog "github.com/sahib/brig/util/log"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

var CurrBackendPort = 10000

func init() {
	log.SetLevel(log.WarnLevel)
	log.SetFormatter(&colorLog.FancyLogFormatter{
		UseColors: true,
	})
}

func stringify(err error) string {
	if err == nil {
		return ""
	}

	return err.Error()
}

func withDaemon(t *testing.T, name string, fn func(ctl *Client)) {
	port := util.FindFreePort()
	repoPath, err := ioutil.TempDir("", "brig-client-repo")
	require.Nil(t, err)

	defer os.RemoveAll(repoPath)

	err = repo.Init(repoPath, name, "no-pass", "mock", int64(port))
	require.Nil(t, err, stringify(err))

	passwordFn := func() (string, error) {
		return "no-pass", nil
	}

	srv, err := server.BootServer(repoPath, passwordFn, "127.0.0.1", port, true)
	require.Nil(t, err, stringify(err))

	go func() {
		require.Nil(t, srv.Serve())
	}()

	time.Sleep(500 * time.Millisecond)

	ctl, err := Dial(context.Background(), port)
	require.Nil(t, err)

	defer func() {
		require.Nil(t, srv.Close())
	}()

	fn(ctl)

}

func withDaemonPair(t *testing.T, nameA, nameB string, fn func(ctlA, ctlB *Client)) {
	withDaemon(t, nameA, func(ctlA *Client) {
		withDaemon(t, nameB, func(ctlB *Client) {
			aliWhoami, err := ctlA.Whoami()
			require.Nil(t, err, stringify(err))

			bobWhoami, err := ctlB.Whoami()
			require.Nil(t, err, stringify(err))

			// add bob to ali as remote
			err = ctlA.RemoteAddOrUpdate(Remote{
				Name:        nameB,
				Fingerprint: bobWhoami.Fingerprint,
			})
			require.Nil(t, err, stringify(err))

			// add ali to bob as remote
			err = ctlB.RemoteAddOrUpdate(Remote{
				Name:        nameA,
				Fingerprint: aliWhoami.Fingerprint,
			})
			require.Nil(t, err, stringify(err))

			fn(ctlA, ctlB)
		})
	})
}

func TestStageAndCat(t *testing.T) {
	withDaemon(t, "ali", func(ctl *Client) {
		fd, err := ioutil.TempFile("", "brig-dummy-data")
		path := fd.Name()

		require.Nil(t, err, stringify(err))
		_, err = fd.Write([]byte("hello"))
		require.Nil(t, err, stringify(err))
		require.Nil(t, fd.Close())

		require.Nil(t, ctl.Stage(path, "/hello"))
		rw, err := ctl.Cat("hello", false)
		require.Nil(t, err, stringify(err))

		data, err := ioutil.ReadAll(rw)
		require.Nil(t, err, stringify(err))

		require.Equal(t, []byte("hello"), data)
		require.Nil(t, rw.Close())
	})
}

func TestMkdir(t *testing.T) {
	withDaemon(t, "ali", func(ctl *Client) {
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
		require.Nil(t, err, stringify(err))

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

func TestSyncBasic(t *testing.T) {
	withDaemonPair(t, "ali", "bob", func(aliCtl, bobCtl *Client) {
		err := aliCtl.StageFromReader("/ali_file", bytes.NewReader([]byte{42}))
		require.NoError(t, err)

		err = bobCtl.StageFromReader("/bob_file", bytes.NewReader([]byte{23}))
		require.NoError(t, err)

		_, err = aliCtl.Sync("bob", true)
		require.NoError(t, err)

		_, err = bobCtl.Sync("ali", true)
		require.NoError(t, err)

		// We cannot query the file contents, since the mock backend
		// does not yet store the file content anywhere.
		bobFileStat, err := aliCtl.Stat("/bob_file")
		require.NoError(t, err)
		require.Equal(t, "/bob_file", bobFileStat.Path)

		aliFileStat, err := bobCtl.Stat("/ali_file")
		require.NoError(t, err)
		require.Equal(t, "/ali_file", aliFileStat.Path)
	})
}

func pathsFromListing(l []StatInfo) []string {
	result := []string{}
	for _, entry := range l {
		result = append(result, entry.Path)
	}

	return result
}

func TestSyncConflict(t *testing.T) {
	withDaemonPair(t, "ali", "bob", func(aliCtl, bobCtl *Client) {
		// Create two files with the same content on both sides:
		err := aliCtl.StageFromReader("/README", bytes.NewReader([]byte{42}))
		require.Nil(t, err, stringify(err))

		err = bobCtl.StageFromReader("/README", bytes.NewReader([]byte{42}))
		require.Nil(t, err, stringify(err))

		// Sync and check if the files are still equal:
		_, err = bobCtl.Sync("ali", true)
		require.Nil(t, err, stringify(err))

		aliFileStat, err := aliCtl.Stat("/README")
		require.Nil(t, err, stringify(err))
		bobFileStat, err := bobCtl.Stat("/README")
		require.Nil(t, err, stringify(err))
		require.Equal(t, aliFileStat.ContentHash, bobFileStat.ContentHash)

		// Modify bob's side only. A sync should have no effect.
		err = bobCtl.StageFromReader("/README", bytes.NewReader([]byte{43}))
		require.Nil(t, err, stringify(err))

		_, err = bobCtl.Sync("ali", true)
		require.Nil(t, err, stringify(err))

		bobFileStat, err = bobCtl.Stat("/README")
		require.Nil(t, err, stringify(err))

		require.NotEqual(t, aliFileStat.ContentHash, bobFileStat.ContentHash)

		// Modify ali's side additionally. Now we should get a conflicting file.
		err = aliCtl.StageFromReader("/README", bytes.NewReader([]byte{41}))
		require.Nil(t, err, stringify(err))

		dirs, err := bobCtl.List("/", -1)
		require.Nil(t, err, stringify(err))
		require.Equal(t, []string{"/", "/README"}, pathsFromListing(dirs))

		_, err = bobCtl.Sync("ali", true)
		require.Nil(t, err, stringify(err))

		dirs, err = bobCtl.List("/", -1)
		require.Nil(t, err, stringify(err))
		require.Equal(
			t,
			[]string{"/", "/README", "/README.conflict.0"},
			pathsFromListing(dirs),
		)
	})
}

func TestSyncSeveralTimes(t *testing.T) {
	withDaemonPair(t, "ali", "bob", func(aliCtl, bobCtl *Client) {
		err := aliCtl.StageFromReader("/ali_file_1", bytes.NewReader([]byte{1}))
		require.Nil(t, err, stringify(err))

		_, err = bobCtl.Sync("ali", true)
		require.Nil(t, err, stringify(err))

		dirs, err := bobCtl.List("/", -1)
		require.Nil(t, err, stringify(err))
		require.Equal(
			t,
			[]string{"/", "/ali_file_1"},
			pathsFromListing(dirs),
		)

		err = aliCtl.StageFromReader("/ali_file_2", bytes.NewReader([]byte{2}))
		require.Nil(t, err, stringify(err))

		_, err = bobCtl.Sync("ali", true)

		require.Nil(t, err, stringify(err))

		dirs, err = bobCtl.List("/", -1)
		require.Nil(t, err, stringify(err))
		require.Equal(
			t,
			[]string{"/", "/ali_file_1", "/ali_file_2"},
			pathsFromListing(dirs),
		)

		err = aliCtl.StageFromReader("/ali_file_3", bytes.NewReader([]byte{3}))
		require.Nil(t, err, stringify(err))

		_, err = bobCtl.Sync("ali", true)
		require.Nil(t, err, stringify(err))

		dirs, err = bobCtl.List("/", -1)
		require.Nil(t, err, stringify(err))
		require.Equal(
			t,
			[]string{"/", "/ali_file_1", "/ali_file_2", "/ali_file_3"},
			pathsFromListing(dirs),
		)
	})
}

func TestSyncPartial(t *testing.T) {
	withDaemonPair(t, "ali", "bob", func(aliCtl, bobCtl *Client) {
		aliWhoami, err := aliCtl.Whoami()
		require.Nil(t, err, stringify(err))

		bobWhoami, err := bobCtl.Whoami()
		require.Nil(t, err, stringify(err))

		require.Nil(t, aliCtl.RemoteSave([]Remote{
			{
				Name:        "bob",
				Fingerprint: bobWhoami.Fingerprint,
				Folders: []RemoteFolder{
					{
						Folder: "/photos",
					},
				},
			},
		}))

		require.Nil(t, bobCtl.RemoteSave([]Remote{
			{
				Name:        "ali",
				Fingerprint: aliWhoami.Fingerprint,
				Folders: []RemoteFolder{
					{
						Folder: "/photos",
					},
				},
			},
		}))

		err = aliCtl.StageFromReader("/docs/ali_secret.txt", bytes.NewReader([]byte{0}))
		require.Nil(t, err, stringify(err))
		err = aliCtl.StageFromReader("/photos/ali.png", bytes.NewReader([]byte{42}))
		require.Nil(t, err, stringify(err))

		err = bobCtl.StageFromReader("/docs/bob_secret.txt", bytes.NewReader([]byte{0}))
		require.Nil(t, err, stringify(err))
		err = bobCtl.StageFromReader("/photos/bob.png", bytes.NewReader([]byte{23}))
		require.Nil(t, err, stringify(err))

		_, err = aliCtl.Sync("bob", true)
		require.Nil(t, err, stringify(err))

		_, err = bobCtl.Sync("ali", true)
		require.Nil(t, err, stringify(err))

		// We cannot query the file contents, since the mock backend
		// does not yet store the file content anywhere.
		aliLs, err := aliCtl.List("/", -1)
		require.Nil(t, err, stringify(err))

		aliPaths := []string{}
		for _, entry := range aliLs {
			aliPaths = append(aliPaths, entry.Path)
		}

		bobLs, err := bobCtl.List("/", -1)
		require.Nil(t, err, stringify(err))

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

func TestSyncMovedFile(t *testing.T) {
	withDaemonPair(t, "ali", "bob", func(aliCtl, bobCtl *Client) {
		require.NoError(t, aliCtl.StageFromReader("/ali-file", bytes.NewReader([]byte{1, 2, 3})))
		require.NoError(t, bobCtl.StageFromReader("/bob-file", bytes.NewReader([]byte{4, 5, 6})))

		aliDiff, err := aliCtl.Sync("bob", true)
		require.NoError(t, err)

		bobDiff, err := bobCtl.Sync("ali", true)
		require.NoError(t, err)

		require.Equal(t, aliDiff.Added[0].Path, "/bob-file")
		require.Equal(t, bobDiff.Added[0].Path, "/ali-file")

		require.NoError(t, aliCtl.Move("/ali-file", "/bali-file"))

		bobDiffAfter, err := bobCtl.Sync("ali", true)
		require.NoError(t, err)

		require.Len(t, bobDiffAfter.Added, 0)
		require.Len(t, bobDiffAfter.Removed, 0)
		require.Len(t, bobDiffAfter.Moved, 1)
	})
}
