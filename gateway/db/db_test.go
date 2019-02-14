package db

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func withDummyDb(t *testing.T, fn func(db *UserDatabase)) {
	tmpPath, err := ioutil.TempDir("", "brig-gw-userdb")
	require.Nil(t, err)
	defer os.RemoveAll(tmpPath)

	userDb, err := NewUserDatabase(tmpPath)
	require.Nil(t, err)

	fn(userDb)

	require.Nil(t, userDb.Close())
}

func TestAddGet(t *testing.T) {
	withDummyDb(t, func(db *UserDatabase) {
		require.Nil(t, db.Add("hello", "world", []string{"/"}, nil))
		user, err := db.Get("hello")
		require.Nil(t, err)
		require.Equal(t, "hello", user.Name)
		require.NotEmpty(t, user.PasswordHash)
		require.NotEmpty(t, user.Salt)
		require.Empty(t, user.Rights)
		require.Equal(t, []string{"/"}, user.Folders)
	})
}
