package testlayer

import (
	"sync"
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/disorganizer/brig/repo"
	"github.com/disorganizer/brig/transfer"
	"github.com/disorganizer/brig/transfer/moose"
)

func WithConnector(t *testing.T, user string, fc func(c *transfer.Connector)) {
	repo.WithRepo(t, user, user+"pass", func(rp *repo.Repository) {
		if err := rp.IPFS.Online(); err != nil {
			t.Errorf("Cannot go online with IPFS repo: %v", err)
			return
		}

		lay := moose.NewLayer(rp.IPFS, context.Background())
		con := transfer.NewConnector(lay, rp)

		if err := con.Connect(); err != nil {
			t.Errorf("Cannot connect: %v", err)
			return
		}

		con.WaitForPool()

		t.Logf("Entering test for %s's connector", user)
		fc(con)
		t.Logf("Leaving test for %s's connector", user)

		if err := con.Disconnect(); err != nil {
			t.Errorf("Cannot disconnect: %v", err)
			return
		}
	})
}

func TestConversation(t *testing.T) {
	WithParallelConnectors(t, []string{"alice", "bob"}, func(cs []*transfer.Connector) {
		MakeFriends(t, cs...)
		ac, bc := cs[0], cs[1]

		apc, err := ac.Dial(bc.Repo().Peer())
		if err != nil {
			t.Errorf("Alice cannot dial to bob: %v", err)
			return
		}

		// Spam in some queries:
		for i := 0; i < 10; i++ {
			v, err := apc.QueryStoreVersion()
			if err != nil {
				t.Errorf("Usage of api client failed: %v", err)
				return
			}

			if v <= 0 {
				t.Errorf("Version should be any positive number")
				return
			}
		}
	})
}

func WithManyConnectors(t *testing.T, users []string, fc func(c []*transfer.Connector)) {
	withManyConnectors(t, []*transfer.Connector{}, users, fc)
}

func withManyConnectors(t *testing.T, cons []*transfer.Connector, users []string, fc func(c []*transfer.Connector)) {
	if len(users) == 0 {
		fc(cons)
		return
	}

	WithConnector(t, users[0], func(c *transfer.Connector) {
		withManyConnectors(t, append(cons, c), users[1:], fc)
	})
}

func WithParallelConnectors(t *testing.T, users []string, f func(cs []*transfer.Connector)) {
	// You probably could re-use the same waitgroup, but let's give it clear names:
	cleanupWg := sync.WaitGroup{}
	cleanupWg.Add(len(users))
	waitWg := sync.WaitGroup{}
	waitWg.Add(1)
	setupWg := sync.WaitGroup{}
	setupWg.Add(len(users))

	cns := make(map[string]*transfer.Connector)
	mu := sync.Mutex{}

	// Trigger setup of connectors in parallel:
	for _, user := range users {
		go func(user string) {
			WithConnector(t, user, func(cn *transfer.Connector) {
				// Append it to the connector list:
				mu.Lock()
				cns[user] = cn
				mu.Unlock()

				// Count down one setup'd connector:
				setupWg.Done()

				// Wait for testcase to finish:
				waitWg.Wait()
			})

			// Report that 1 connector was cleaned up.
			cleanupWg.Done()
		}(user)
	}

	// Wait for all connectors to show up:
	setupWg.Wait()

	// Sometimes ipfs does not seem to be fully online yet,
	// wait a short bit therefore.
	time.Sleep(10 * time.Second)

	// Reorder cns, so that order is preserved:
	cnsSorted := []*transfer.Connector{}

	for _, user := range users {
		cnsSorted = append(cnsSorted, cns[user])
	}

	// Call testcase:
	f(cnsSorted)

	// Notify go routines that the testcase finished
	// and we may cleanup the connectors again:
	waitWg.Done()

	// Wait until the cleanup is finished:
	cleanupWg.Wait()
}

func MakeCouple(t *testing.T, userA, userB *transfer.Connector) {
	aName, bName := userA.Repo().ID, userB.Repo().ID
	t.Logf("%s has friend %s", aName, bName)

	br, ar := userB.Repo(), userA.Repo()
	berr := br.Remotes.Insert(repo.NewRemoteFromPeer(ar.Peer()))
	if berr != nil {
		t.Errorf("%s has no friends: %v", bName, berr)
		return
	}

	aerr := ar.Remotes.Insert(repo.NewRemoteFromPeer(br.Peer()))
	if aerr != nil {
		t.Errorf("%s has no friends: %v", aName, aerr)
		return
	}
}

func MakeFriends(t *testing.T, cs ...*transfer.Connector) {
	for i := 0; i < len(cs); i++ {
		for j := i + 1; j < len(cs); j++ {
			MakeCouple(t, cs[i], cs[j])
		}
	}
}

func TestBroadcast(t *testing.T) {
	WithParallelConnectors(t, []string{"alice", "charlie", "bob"}, func(cs []*transfer.Connector) {
		MakeFriends(t, cs...)

		// Might take a little bit to startup fully:
		// (this is the same time as ipfs' "backoff" mechanism)
		time.Sleep(5 * time.Second)

		for i := 0; i < len(cs); i++ {
			for j := i + 1; j < len(cs); j++ {
				a, b := cs[i].Repo().Peer(), cs[j].Repo().Peer()
				aSeesB := cs[i].IsOnline(b)
				bSeesA := cs[j].IsOnline(a)

				if !aSeesB {
					t.Errorf("%s sees %s not as online", a.ID(), b.ID())
					return
				}

				if !bSeesA {
					t.Errorf("%s sees %s not as online", b.ID(), a.ID())
					return
				}
			}
		}

		// Make alice broadcast a FileUpdate to bob and charlie.
		bcaster := cs[0].Broadcaster()
		if err := bcaster.FileUpdate(nil); err != nil {
			t.Errorf("Could not broadcast: %v", err)
			return
		}

		if err := cs[0].Repo().Remotes.Remove("bob"); err != nil {
			t.Errorf("Unable to remove bob from friend list: %v", err)
			return
		}

		// Give bob and charlie a bit time to receive the message
		time.Sleep(1 * time.Second)

		if err := bcaster.FileUpdate(nil); err != nil {
			t.Errorf("Could not broadcast the second time: %v", err)
			return
		}

	})
}
