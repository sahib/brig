package repo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/disorganizer/brig/util/testutil"
)

func TestRemote(t *testing.T) {
	path := filepath.Join(os.TempDir(), "brig-test-remote.yml")
	fd, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		t.Errorf("Could not open in memory: %v", err)
		return
	}

	defer fd.Close()
	defer testutil.Remover(t, path)

	rms, err := NewYAMLRemotes(fd)
	if err != nil {
		t.Errorf("Creating yaml store failed: %v", err)
		return
	}

	remoteAlc := NewRemote("alice", "Qm123")
	remoteBob := NewRemote("bob", "Qm321")
	remoteChr := NewRemote("chris", "QmABC")
	remoteMal := NewRemote("micrathene", "Qm123")

	for _, rm := range []Remote{remoteAlc, remoteBob, remoteChr} {
		if err := rms.Insert(rm); err != nil {
			t.Errorf("Insert(%v) into the remote store failed: %v", rm.ID(), err)
			return
		}

		retrievedRemote, err := rms.Get(rm.ID())
		if err != nil {
			t.Errorf("Retrieving remote failed: %v", err)
			return
		}

		if !RemoteIsEqual(rm, retrievedRemote) {
			t.Errorf("Remotes are not equal")
			return
		}
	}

	if err := rms.Insert(remoteMal); err == nil {
		t.Errorf("Insert(malicious_micra) into the remote store worked")
		return
	}

	if err := rms.Remove("alice"); err != nil {
		t.Errorf("Removing remote failed: %v", err)
		return
	}

	if r, err := rms.Get("alice"); err == nil || r != nil {
		t.Errorf("removed remote still there: %v (%v)", err, r)
		return
	}

	lst := AsList(rms)
	if lst[0].ID() != "bob" {
		t.Errorf("Not bob")
	}

	if lst[1].ID() != "chris" {
		t.Errorf("Not chris")
	}

	if err := rms.Close(); err != nil {
		t.Errorf("Closing yaml store failed: %v", err)
		return
	}
}
