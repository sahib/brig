package repo

import (
	"os"
	"testing"

	"github.com/disorganizer/brig/backend/memory"
	"github.com/stretchr/testify/require"
)

func TestRepoInit(t *testing.T) {
	testDir := "/tmp/.brig-repo-test"
	require.Nil(t, os.RemoveAll(testDir))

	// Directory does not exist yet:
	err := Init(testDir, "alice", "memory")
	require.NotNil(t, err)

	require.Nil(t, os.Mkdir(testDir, 0700))
	err = Init(testDir, "alice", "memory")
	require.Nil(t, err)

	repo, err := Open(testDir, "klaus")
	require.Nil(t, err)

	bk := memory.NewMemoryBackend()
	fs, err := repo.OwnFS(bk)
	require.Nil(t, err)

	// TODO: Assert a bit more that fs is working.
	require.NotNil(t, fs)

	require.Nil(t, repo.Close("klaus"))
	require.Nil(t, fs.Close())
}
