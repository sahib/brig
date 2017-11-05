package repo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKey(t *testing.T) {
	testDir := filepath.Join(os.TempDir(), "brig-repo-key-test")
	require.Nil(t, os.MkdirAll(testDir, 0755))
	require.Nil(t, createKeyPair("alice", testDir, 4096))

	testData := []byte("Hello World")

	encData, err := encryptAsymmetric(testDir, testData)
	require.Nil(t, err)
	require.NotEqual(t, testData, encData)

	decData, err := decryptAsymetric(testDir, encData)
	require.Nil(t, err)
	require.Equal(t, testData, decData)
}
