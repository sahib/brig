package catfs

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"testing"

	log "github.com/Sirupsen/logrus"
	c "github.com/disorganizer/brig/catfs/core"
	ie "github.com/disorganizer/brig/catfs/errors"
	"github.com/disorganizer/brig/catfs/mio"
	"github.com/disorganizer/brig/catfs/mio/compress"
	n "github.com/disorganizer/brig/catfs/nodes"
	h "github.com/disorganizer/brig/util/hashlib"
	"github.com/disorganizer/brig/util/testutil"
	"github.com/stretchr/testify/require"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

func withDummyFS(t *testing.T, fn func(fs *FS)) {
	backend := NewMemFsBackend()
	owner := &Person{
		Name: "alice",
		Hash: h.TestDummy(t, 1),
	}

	dbPath, err := ioutil.TempDir("", "brig-fs-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	defer os.RemoveAll(dbPath)

	fs, err := NewFilesystem(backend, dbPath, owner)
	if err != nil {
		t.Fatalf("Failed to create filesystem: %v", err)
	}

	fn(fs)

	if err := fs.Close(); err != nil {
		t.Fatalf("Failed to close filesystem: %v", err)
	}
}

func TestStat(t *testing.T) {
	withDummyFS(t, func(fs *FS) {
		_, err := fs.Stat("/sub/x")
		require.True(t, ie.IsNoSuchFileError(err))

		c.MustMkdir(t, fs.lkr, "/sub")
		file := c.MustTouch(t, fs.lkr, "/sub/x", 1)

		info, err := fs.Stat("/sub/x")
		require.Nil(t, err)
		require.Equal(t, info.Path, "/sub/x")
		require.Equal(t, info.IsDir, false)
		require.Equal(t, info.Size, uint64(0))
		require.Equal(t, info.Inode, file.Inode())
		require.Equal(t, info.Hash, file.Hash())

		file.SetSize(42)
		require.Nil(t, fs.lkr.StageNode(file))

		info, err = fs.Stat("/sub/x")
		require.Nil(t, err)
		require.Equal(t, info.Size, uint64(42))
		require.Equal(t, info.Hash, file.Hash())

		info, err = fs.Stat("/sub")
		require.Nil(t, err)
		require.Equal(t, info.Path, "/sub")
		require.Equal(t, info.IsDir, true)
		// TODO:
		// require.Equal(t, info.Size, uint64(42))
	})
}

func TestLogAndTag(t *testing.T) {
	withDummyFS(t, func(fs *FS) {
		cmts := []*n.Commit{}
		for idx := 0; idx < 10; idx++ {
			_, cmt := c.MustTouchAndCommit(t, fs.lkr, "/x", byte(idx))
			fs.Tag(cmt.Hash().B58String(), fmt.Sprintf("tag%d", idx))
			cmts = append(cmts, cmt)
		}

		status, err := fs.lkr.Status()
		require.Nil(t, err)

		cmts = append(cmts, status)

		log, err := fs.Log()
		require.Nil(t, err)

		for idx, entry := range log {
			ridx := len(cmts) - idx - 1
			cmt := cmts[ridx]
			require.Equal(t, entry.Hash, cmt.Hash())

			msg := fmt.Sprintf("cmt %d", ridx)
			tags := []string{fmt.Sprintf("tag%d", ridx)}

			// 0 is status, 1 is head, 10 is initial
			switch idx {
			case 0:
				tags = []string{"curr"}
				msg = ""
			case 1:
				tags = append(tags, "head")
			case 10:
				tags = append(tags, "init")
			}

			sort.Sort(sort.Reverse(sort.StringSlice(entry.Tags)))
			require.EqualValues(t, tags, entry.Tags)
			require.Equal(t, entry.Msg, msg)
		}
	})
}

var TestKey = []byte("01234567890ABCDE01234567890ABCDE")

func TestCat(t *testing.T) {
	withDummyFS(t, func(fs *FS) {
		raw := []byte{1, 2, 3}
		rinRaw := bytes.NewBuffer(raw)

		rin, err := mio.NewInStream(rinRaw, TestKey, compress.AlgoSnappy)
		require.Nil(t, err)

		hash, err := fs.bk.Add(rin)
		require.Nil(t, err)

		// Stage the file manually (without fs.Stage)
		_, err = c.Stage(fs.lkr, "/x", &c.NodeUpdate{
			Author: "me",
			Hash:   hash,
			Key:    TestKey,
			Size:   uint64(len(raw)),
		})
		require.Nil(t, err)

		// Cat the file again:
		stream, err := fs.Cat("/x")
		require.Nil(t, err)

		// Check if the returned stream really contains 1,2,3
		result := bytes.NewBuffer(nil)
		_, err = stream.WriteTo(result)
		require.Nil(t, err)
		require.Equal(t, result.Bytes(), raw)
	})
}

func TestStage(t *testing.T) {
	tcs := [][]byte{
		{},
		{1},
		{1, 2, 3},
		testutil.CreateDummyBuf(8 * 1024),
	}

	for idx, tc := range tcs {
		t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			withDummyFS(t, func(fs *FS) {
				require.Nil(t, fs.Stage("/x", bytes.NewBuffer(tc)))

				stream, err := fs.Cat("/x")
				require.Nil(t, err)

				data, err := ioutil.ReadAll(stream)
				require.Nil(t, err)

				require.Equal(t, data, tc)
			})
		})
	}
}

func TestHistory(t *testing.T) {
	withDummyFS(t, func(fs *FS) {
		fs.MakeCommit("hello")

		require.Nil(t, fs.Stage("/x", bytes.NewBuffer([]byte{1})))
		fs.MakeCommit("1")
		require.Nil(t, fs.Stage("/x", bytes.NewBuffer([]byte{2})))
		fs.MakeCommit("2")
		require.Nil(t, fs.Stage("/x", bytes.NewBuffer([]byte{3})))
		fs.MakeCommit("3")

		hist, err := fs.History("/x")
		require.Nil(t, err)

		log, err := fs.Log()
		require.Nil(t, err)

		for idx, entry := range hist {
			require.Equal(t, entry.Path, "/x")

			change := "none"
			switch idx {
			case 1, 2:
				change = "modified"
			case 3:
				change = "added"
			}

			require.Equal(t, entry.Change, change)

			// Third index repeats commit "1" since /x was added in there.
			if idx != 3 {
				require.Equal(t, log[idx+1].Hash.B58String(), entry.Ref.B58String())
			}
		}
	})
}
