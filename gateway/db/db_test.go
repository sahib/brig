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
		require.Nil(t, db.Add("hello", User{
			Password: "world",
			Folders:  []string{"/"},
		}))

		user, err := db.Get("hello")
		require.Nil(t, err)
		require.Equal(t, "hello", user.Name)
		require.Equal(t, "world", user.Password)
		require.Equal(t, []string{"/"}, user.Folders)
	})
}
