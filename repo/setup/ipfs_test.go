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
	ipfsPath, _, err := IPFS(os.Stdout, true, true, true, "")
	require.Nil(t, err)
	fmt.Println(ipfsPath)
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
