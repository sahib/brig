package repo

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/sahib/brig/backend/ipfs"
	"github.com/sahib/brig/backend/mock"
	"github.com/sahib/brig/util/testutil"
	"github.com/stretchr/testify/require"
)

var (
	TestRegistryPath = "/tmp/test-registry.yml"
)

func init() {
	os.Setenv("BRIG_REGISTRY_PATH", TestRegistryPath)
}

func TestRepoInit(t *testing.T) {
	testDir := "/tmp/.brig-repo-test"
	require.Nil(t, os.RemoveAll(testDir))

	err := Init(testDir, "alice", "klaus", "mock", 6666)
	require.Nil(t, err)

	rp, err := Open(testDir, "klaus")
	require.Nil(t, err)

	bk := mock.NewMockBackend("", "", 0)
	fs, err := rp.FS(rp.CurrentUser(), bk)
	require.Nil(t, err)
	require.NotNil(t, fs)

	require.Nil(t, fs.Stage("/x", bytes.NewReader([]byte{1, 2, 3})))
	stream, err := fs.Cat("/x")
	require.Nil(t, err)

	data, err := ioutil.ReadAll(stream)
	require.Nil(t, err)
	require.Equal(t, data, []byte{1, 2, 3})

	require.Nil(t, fs.Close())
	require.Nil(t, rp.Close("klaus"))

}

func dirSize(t *testing.T, path string) int64 {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			size += info.Size()
		}

		return err
	})

	if err != nil {
		t.Fatalf("Failed to get directory size of `%s`: %v", path, err)
	}

	return size
}

func TestRepoDeduplication(t *testing.T) {
	testDir := "/tmp/.brig-repo-test"
	require.Nil(t, os.RemoveAll(testDir))
	err := Init(testDir, "alice", "klaus", "ipfs", 6666)
	require.Nil(t, err)

	rp, err := Open(testDir, "klaus")
	require.Nil(t, err)

	ipfsPath := filepath.Join(testDir, "data/ipfs")
	require.Nil(t, ipfs.Init(ipfsPath, 1024))

	bk, err := ipfs.New(ipfsPath, nil)
	require.Nil(t, err)

	fs, err := rp.FS(rp.CurrentUser(), bk)
	require.Nil(t, err)
	require.NotNil(t, fs)

	size := dirSize(t, testDir)
	require.True(t, size < 1*1024*1024)

	data := testutil.CreateDummyBuf(8 * 1024 * 1024)

	// Adding a 8MB file should put the size of the repo
	// at somewhere around this size (+ init bytes)
	fs.Stage("/x", bytes.NewReader(data))

	size = dirSize(t, testDir)
	require.True(t, size < 9*1024*1024)
	require.True(t, size > 7*1024*1024)

	// Adding the same file under a different path should
	// not add to the total size of the repository
	// (except a few bytes for storing the metadata)
	fs.Stage("/y", bytes.NewReader(data))

	size = dirSize(t, testDir)
	require.True(t, size < 9*1024*1024)
	require.True(t, size > 7*1024*1024)

	// Modify the beginning of the data,
	// key did not change so there should be only a minimal
	// size increase in the first block (~+64k)
	data[0] += 1
	fs.Stage("/x", bytes.NewReader(data))

	size = dirSize(t, testDir)
	require.True(t, size < 9*1024*1024)
	require.True(t, size > 7*1024*1024)

	// This case is not covered yet:
	// (i.e. adding the same file as "/x" has anew,
	//  this will cause brig to generate a new key,
	//  resulting in a totally data stream)
	fs.Stage("/z", bytes.NewReader(data))
	size = dirSize(t, testDir)
	require.True(t, size < 18*1024*1024)
	require.True(t, size > 16*1024*1024)

	require.Nil(t, fs.Close())
	require.Nil(t, rp.Close("klaus"))
}
