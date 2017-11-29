package repo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKeyring(t *testing.T) {
	testDir := filepath.Join(os.TempDir(), "brig-repo-key-test")
	require.Nil(t, os.MkdirAll(testDir, 0755))
	require.Nil(t, createKeyPair("alice", testDir, 1024))

	kr := newKeyringHandle(testDir)
	ownPubKey, err := kr.OwnPubKey()
	require.Nil(t, err)
	require.True(t, len(ownPubKey) > 256)

	// Not very realistic that we encrypt with our own pub key ourselves,
	// but that's good enough for testing if it works.
	testData := []byte("Hello!")
	encTestData, err := kr.Encrypt(testData, ownPubKey)
	require.Nil(t, err)
	require.NotEqual(t, testData, encTestData)

	decTestData, err := kr.Decrypt(encTestData)
	require.Nil(t, err)
	require.Equal(t, testData, decTestData)

	require.Nil(t, kr.SavePubKey("a", []byte{1}))
	require.Nil(t, kr.SavePubKey("a", []byte{1}))
	remotePubKey, err := kr.PubKeyFor("a")
	require.Nil(t, err)
	require.Equal(t, remotePubKey, []byte{1})
}
