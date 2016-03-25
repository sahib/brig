package mqtt

import (
	"fmt"
	"net"
	"testing"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/id"
	"github.com/disorganizer/brig/transfer"
	"github.com/disorganizer/brig/transfer/wire"
)

type DummyResolver struct {
	Port int
	peer id.Peer
}

func (dr *DummyResolver) Resolve(ctx context.Context) (id.Addresses, error) {
	return id.Addresses{
		&net.TCPAddr{
			IP:   net.IPv4(127, 0, 0, 1),
			Port: dr.Port,
		},
	}, nil
}

func (dr *DummyResolver) Peer() id.Peer {
	return dr.peer
}

var (
	PeerA = id.NewPeer("alice", "QmdEcweCLrhwQCSe5yrYYZ7CP8i1t6PzakQXsf2LoM3eHv")
	PeerB = id.NewPeer("bob", "QmSoLnSGccFuZQJzRadHn95W2CrSFmZuTdDWP8HXaHca9z")
	PortA = 1883
	PortB = 1884
	RslvA = &DummyResolver{PortA, PeerA}
	RslvB = &DummyResolver{PortB, PeerB}
)

func init() {
	log.SetLevel(log.DebugLevel)
}

func withOnlineLayer(t *testing.T, lay transfer.Layer, f func()) {
	name := lay.Self().ID()

	if lay.IsOnlineMode() {
		t.Fatalf("Layer `%s` started online", name)
		return
	}

	if err := lay.Connect(); err != nil {
		t.Fatalf("Layer `%s` could not connect: %v", name, err)
	}

	if !lay.IsOnlineMode() {
		t.Fatalf("Layer `%s` still offline after connect", name)
		return
	}

	f()

	if t.Failed() {
		t.Errorf("Handler for `%s` failed; see output above", name)
	}

	if err := lay.Close(); err != nil {
		t.Errorf("Closing error failed: %v", err)
		return
	}
}

func TestIO(t *testing.T) {
	layA := NewLayer(PeerA, PortA)
	layB := NewLayer(PeerB, PortB)

	withOnlineLayer(t, layA, func() {
		withOnlineLayer(t, layB, func() {
			layB.RegisterHandler(
				wire.RequestType_FETCH,
				func(_ transfer.Layer, req *wire.Request) (*wire.Response, error) {
					fmt.Println("processing response", req)
					return &wire.Response{Data: []byte("World")}, nil
				},
			)

			// Let's talk to shy bob:
			cnv, err := layA.Talk(RslvB)
			if err != nil {
				t.Errorf("Talk(bob) failed: %v", err)
				return
			}

			// Send some dummy request:
			req := &wire.Request{
				ReqType:       wire.RequestType_FETCH.Enum(),
				BroadcastData: []byte("Hello World!"),
			}

			err = cnv.SendAsync(req, func(resp *wire.Response) {
				fmt.Println("Response", resp)
			})

			if err != nil {
				t.Errorf("Send async failed: %v", err)
				return
			}

			layA.Wait()
			// time.Sleep(5 * time.Second)

			if err := cnv.Close(); err != nil {
				t.Errorf("Close failed: %v", err)
				return
			}
		})
	})
}

func TestOnOff(t *testing.T) {
	layA := NewLayer(PeerA, PortA)

	withOnlineLayer(t, layA, func() {
		if err := layA.Disconnect(); err != nil {
			t.Errorf("Disconnect failed: %v", err)
			return
		}

		if layA.IsOnlineMode() {
			t.Errorf("alice is online after disconnect")
			return
		}

		if err := layA.Connect(); err != nil {
			t.Errorf("Disconnect failed: %v", err)
			return
		}

		if !layA.IsOnlineMode() {
			t.Errorf("alice is offline after connect")
			return
		}
	})
}
