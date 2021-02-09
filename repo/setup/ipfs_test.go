package setup

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMain(t *testing.T) {
	t.Skip("TODO: probably a bit too much for travis")
	_, err := IPFS(Options{
		LogWriter:        os.Stdout,
		Setup:            true,
		SetDefaultConfig: true,
		SetExtraConfig:   true,
	})
	require.NoError(t, err)
}

func TestInstall(t *testing.T) {
	require.Nil(t, installIPFS(os.Stdout))
}

func TestCommandAvailable(t *testing.T) {
	fmt.Println(isCommandAvailable("ipfs"))
}

func TestRepoInit(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "ipfs-tmp-test-")
	require.Nil(t, err)
	require.Nil(t, initIPFS(os.Stdout, tmpDir))
}
