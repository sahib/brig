package repo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/disorganizer/brig/util/testutil"
)

func withRemoteStore(t *testing.T, f func(rms RemoteStore)) {
	path := filepath.Join(os.TempDir(), "brig-test-remote.yml")
	defer testutil.Remover(t, path)

	rms, err := NewYAMLRemotes(path)
	if err != nil {
		t.Errorf("Creating yaml store failed: %v", err)
		return
	}

	f(rms)

	if err := rms.Close(); err != nil {
		t.Errorf("Closing yaml store failed: %v", err)
		return
	}
}

func TestRemote(t *testing.T) {
	withRemoteStore(t, func(rms RemoteStore) {
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
	})
}

func TestRemoteObserver(t *testing.T) {
	withRemoteStore(t, func(rms RemoteStore) {
		alice1 := NewRemote("alice", "1")
		alice2 := NewRemote("alice", "2")

		i := 0
		rms.Register(func(change *RemoteChange) {
			i++
			switch i {
			case 1:
				if change.ChangeType != RemoteChangeAdded {
					t.Fatalf("Expected add")
				}

				if change.OldRemote != nil {
					t.Fatalf("Oldremote is not nil after nil")
				}

				if !RemoteIsEqual(change.Remote, alice1) {
					t.Fatalf("Wrong new remote after add")
				}
			case 2:
				if change.ChangeType != RemoteChangeModified {
					t.Fatalf("Expected modified")
				}

				if !RemoteIsEqual(change.OldRemote, alice1) {
					t.Fatalf("Wrong old remote after modify")
				}

				if !RemoteIsEqual(change.Remote, alice2) {
					t.Fatalf("Wrong new remote after modify")
				}
			case 3:
				if change.ChangeType != RemoteChangeRemoved {
					t.Fatalf("Expected removed")
				}

				if !RemoteIsEqual(change.OldRemote, alice2) {
					t.Fatalf("Wrong old remote")
				}

				if change.Remote != nil {
					t.Fatalf(".Remote is nil after remove")
				}
			}
		})

		if err := rms.Insert(alice1); err != nil {
			t.Errorf("First insert failed: %v", err)
			return
		}

		if err := rms.Insert(alice2); err != nil {
			t.Errorf("Second insert failed: %v", err)
			return
		}

		if err := rms.Remove("alice"); err != nil {
			t.Errorf("Remove failed: %v", err)
			return
		}
	})
}
