package repo

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/disorganizer/brig/net/peer"
	"github.com/stretchr/testify/require"
)

var (
	bobRemote = Remote{
		Name:        "bob@bobbyland.com/home",
		Fingerprint: peer.Fingerprint("fingerprint"),
		Folders: []Folder{
			{
				Folder: "/Public",
				Perms:  PermRead | PermWrite,
			}, {
				Folder: "/ShowOff",
				Perms:  PermRead,
			},
		},
	}
	charlieRemote = Remote{
		Name:        "charlie",
		Fingerprint: peer.Fingerprint("charliesfp"),
		Folders: []Folder{
			{
				Folder: "/Porns",
				Perms:  PermRead,
			},
		},
	}
)

func TestRemotesReload(t *testing.T) {
	fd, err := ioutil.TempFile("", "brig-test-remotes")
	require.Nil(t, err)

	defer require.Nil(t, os.Remove(fd.Name()))
	defer require.Nil(t, fd.Close())

	rl1, err := NewRemotes(fd.Name())
	require.Nil(t, err)

	require.Nil(t, rl1.AddRemote(bobRemote))

	rl2, err := NewRemotes(fd.Name())
	require.Nil(t, err)

	remotes, err := rl2.ListRemotes()
	require.Nil(t, err)

	require.Equal(t, len(remotes), 1)
	require.Equal(t, remotes[0].Name, "bob@bobbyland.com/home")
	if remotes[0].Fingerprint != "fingerprint" {
		t.Fatalf("Fingerprints are differing: %v", remotes[0].Fingerprint)
	}
	require.Equal(t, remotes[0].Folders, bobRemote.Folders)
}

func TestRemoteOps(t *testing.T) {
	fd, err := ioutil.TempFile("", "brig-test-remotes")
	require.Nil(t, err)

	defer require.Nil(t, os.Remove(fd.Name()))
	defer require.Nil(t, fd.Close())

	rl, err := NewRemotes(fd.Name())
	require.Nil(t, err)

	require.Nil(t, rl.AddRemote(bobRemote))
	require.Nil(t, rl.AddRemote(charlieRemote))

	fetchedBob, err := rl.Remote("bob@bobbyland.com/home")
	require.Nil(t, err)
	require.Equal(t, fetchedBob, bobRemote)

	fetchedCharlie, err := rl.Remote("charlie")
	require.Nil(t, err)
	require.Equal(t, fetchedCharlie, charlieRemote)

	// Check that list is outputting it sorted by name
	remotes, err := rl.ListRemotes()
	require.Nil(t, err)
	require.Equal(t, remotes, []Remote{bobRemote, charlieRemote})

	require.Nil(t, rl.RmRemote("charlie"))
	require.Equal(t, rl.RmRemote("charlie"), ErrNoSuchRemote)

	_, err = rl.Remote("charlie")
	require.Equal(t, err, ErrNoSuchRemote)

	err = rl.SaveList([]Remote{bobRemote, charlieRemote})
	require.Nil(t, err)

	// Check it's the same again after we saved it over:
	remotes, err = rl.ListRemotes()
	require.Nil(t, err)
	require.Equal(t, remotes[0], bobRemote)
	require.Equal(t, remotes[1], charlieRemote)
}
