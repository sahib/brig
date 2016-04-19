package testlayer

import (
	"fmt"
	"testing"
	"time"

	"github.com/disorganizer/brig/repo"
	"github.com/disorganizer/brig/transfer"
	"github.com/disorganizer/brig/transfer/moose"
	"github.com/disorganizer/brig/util/testwith"
)

func WithConnector(t *testing.T, user string, fc func(c *transfer.Connector)) {
	pass := user + "pass"
	testwith.WithRepo(t, user, pass, func(rp *repo.Repository) {
		if err := rp.IPFS.Online(); err != nil {
			t.Errorf("Cannot go online with IPFS repo: %v", err)
			return
		}
		lay := moose.NewLayer(rp.IPFS)
		con := transfer.NewConnector(lay, rp)

		if err := con.Connect(); err != nil {
			t.Errorf("Cannot connect: %v", err)
			return
		}
		fmt.Println("before fc", rp.InternalFolder)
		fc(con)
		fmt.Println("after fc")

		if err := con.Disconnect(); err != nil {
			t.Errorf("Cannot disconnect: %v", err)
			return
		}
	})
}

func TestConversation(t *testing.T) {
	WithConnector(t, "alice", func(ac *transfer.Connector) {
		WithConnector(t, "bob", func(bc *transfer.Connector) {
			br, ar := bc.Repo(), ac.Repo()
			berr := br.Remotes.Insert(repo.NewRemoteFromPeer(ar.Peer()))
			time.Sleep(0 * time.Second)
			if berr != nil {
				fmt.Println("bob remote add")
				t.Errorf("Bob has no friends: %v", berr)
				return
			}

			aerr := ar.Remotes.Insert(repo.NewRemoteFromPeer(br.Peer()))
			if aerr != nil {
				t.Errorf("Alice has no friends: %v", aerr)
				return
			}

			fmt.Println("Alice %v dials bob %v", ar.Peer(), br.Peer())
			apc, err := ac.Dial(br.Peer())
			if err != nil {
				t.Errorf("Alice cannot dial to bob: %v", err)
				return
			}
			fmt.Println("close api")
			if err := apc.Close(); err != nil {
				t.Errorf("Alice cannot close apiclient to bob: %v", err)
				return
			}
		})
	})
}
