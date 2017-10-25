package client

import (
	"context"
	"io/ioutil"
	"sort"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/brigd/server"
	colorlog "github.com/disorganizer/brig/util/log"
	"github.com/stretchr/testify/require"
)

func init() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&colorlog.ColorfulLogFormatter{})
}

func withDaemon(t *testing.T, fn func(ctl *Client)) {
	base, err := ioutil.TempDir("", "brig-ctl-test-")
	require.Nil(t, err)

	srv, err := server.BootServer(base, "klaus", 6667)
	require.Nil(t, err)

	waitForDeath := make(chan bool)
	go func() {
		require.Nil(t, srv.Serve())
		require.Nil(t, srv.Close())
		waitForDeath <- true
	}()

	ctl, err := Dial(context.Background(), 6667)
	require.Nil(t, err)

	require.Nil(t, ctl.Init(base, "alice", "klaus", "memory"))
	fn(ctl)

	require.Nil(t, ctl.Quit())
	<-waitForDeath
}

func TestStageAndCat(t *testing.T) {
	withDaemon(t, func(ctl *Client) {
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
	withDaemon(t, func(ctl *Client) {
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
