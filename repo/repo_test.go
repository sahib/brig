package repo

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/sahib/brig/backend/mock"
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
		if err != nil {
			return err
		}

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
