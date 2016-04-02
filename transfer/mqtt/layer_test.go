package mqtt

import (
	"bytes"
	"fmt"
	"net"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/id"
	"github.com/disorganizer/brig/transfer"
	"github.com/disorganizer/brig/transfer/wire"
	"github.com/disorganizer/brig/util/testutil"
)

var (
	PeerA = id.NewPeer("alice", "QmdEcweCLrhwQCSe5yrYYZ7CP8i1t6PzakQXsf2LoM3eHv")
	PeerB = id.NewPeer("bob", "QmSoLnSGccFuZQJzRadHn95W2CrSFmZuTdDWP8HXaHca9z")
	PortA = 1883
	PortB = 1884
)

func init() {
	log.SetLevel(log.DebugLevel)
}

type DummyDialer struct{}

func (d *DummyDialer) Dial(peer id.Peer) (net.Conn, error) {
	var port int

	switch h := peer.Hash(); h {
	case PeerA.Hash():
		port = PortA
	case PeerB.Hash():
		port = PortB
	default:
		return nil, fmt.Errorf("Unable to dial %s", h)
	}

	return net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
}

func dummyNetwork(t *testing.T, port int) (net.Listener, transfer.Dialer) {
	var lastError error

	for i := 0; i < 10; i++ {
		l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err != nil {
			time.Sleep(200 * time.Millisecond)
			lastError = err
			fmt.Println(".")
			continue
		}

		return l, &DummyDialer{}
	}

	t.Errorf("Failed to create listener: %v", lastError)
	return nil, nil
}

func withOnlineLayer(t *testing.T, port int, lay transfer.Layer, f func()) {
	name := lay.Self().ID()

	if lay.IsOnlineMode() {
		t.Fatalf("Layer `%s` started online", name)
		return
	}

	l, d := dummyNetwork(t, port)
	if err := lay.Connect(l, d); err != nil {
		t.Fatalf("Layer `%s` could not connect: %v", name, err)
	}

	if !lay.IsOnlineMode() {
		t.Fatalf("Layer `%s` still offline after connect", name)
		return
	}

	f()

	if t.Failed() {
		t.Errorf("Handler for `%s` failed; see output above.", name)
	}

	if err := lay.Close(); err != nil {
		t.Errorf("Closing error failed: %v", err)
		return
	}
}

func withBadRomance(t *testing.T, f func(layA, layB transfer.Layer)) {
	layA := NewLayer(PeerA, transfer.MockAuthSuccess)
	layB := NewLayer(PeerB, transfer.MockAuthSuccess)

	withOnlineLayer(t, PortA, layA, func() {
		withOnlineLayer(t, PortB, layB, func() {
			f(layA, layB)
		})
	})
}

func TestIO(t *testing.T) {
	reqData := testutil.CreateDummyBuf(4 * 1024 * 1024)
	rspData := testutil.CreateDummyBuf(8 * 1024 * 1024)

	withBadRomance(t, func(layA, layB transfer.Layer) {
		layB.RegisterHandler(
			// FETCH is arbitary; Layer does not handle anything of fetch logic.
			wire.RequestType_FETCH,
			func(req *wire.Request) (*wire.Response, error) {
				if !bytes.Equal(req.GetBroadcastData(), reqData) {
					t.Errorf("Request data differs between peers")
				}

				return &wire.Response{Data: rspData}, nil
			},
		)

		// Let's talk to shy bob:
		cnv, err := layA.Talk(PeerB)
		if err != nil {
			t.Errorf("Talk(bob) failed: %v", err)
			return
		}

		// Send some dummy request:
		req := &wire.Request{
			ReqType:       wire.RequestType_FETCH.Enum(),
			BroadcastData: reqData,
		}

		err = cnv.SendAsync(req, func(rsp *wire.Response) {
			if !bytes.Equal(rsp.GetData(), rspData) {
				t.Errorf("Response data differs between peers")
			}
		})

		if err != nil {
			t.Errorf("Send async failed: %v", err)
			return
		}

		// Wait for requests to finish:
		layA.Wait()

		if err := cnv.Close(); err != nil {
			t.Errorf("Close failed: %v", err)
			return
		}
	})
}

func TestOnOff(t *testing.T) {
	layA := NewLayer(PeerA, transfer.MockAuthSuccess)

	withOnlineLayer(t, PortA, layA, func() {
		for i := 0; i < 10; i++ {
			if err := layA.Disconnect(); err != nil {
				t.Errorf("Disconnect failed: %v", err)
				return
			}

			if layA.IsOnlineMode() {
				t.Errorf("alice is online after disconnect")
				return
			}

			l, d := dummyNetwork(t, PortA)
			if err := layA.Connect(l, d); err != nil {
				t.Errorf("Disconnect failed: %v", err)
				return
			}

			if !layA.IsOnlineMode() {
				t.Errorf("alice is offline after connect")
				return
			}
		}
	})
}

func TestBroadcast(t *testing.T) {
	withBadRomance(t, func(layA, layB transfer.Layer) {
		broadcastData := testutil.CreateDummyBuf(8 * 1024 * 1024)

		layB.RegisterHandler(
			wire.RequestType_FETCH,
			func(req *wire.Request) (*wire.Response, error) {
				if !bytes.Equal(req.GetBroadcastData(), broadcastData) {
					t.Errorf("Broadcast data differs between peers")
				}

				// Broadcasts do not need to be answered:
				return nil, nil
			},
		)

		req := &wire.Request{
			ReqType:       wire.RequestType_FETCH.Enum(),
			BroadcastData: broadcastData,
		}

		if err := layA.Broadcast(req); err != nil {
			t.Errorf("Broadcast failed: %v", err)
			return
		}
	})
}

func TestIsOnline(t *testing.T) {
	withBadRomance(t, func(layA, layB transfer.Layer) {
		if !layA.IsOnline(PeerA) {
			t.Errorf("Alice does not see herself online.")
			return
		}
		if !layB.IsOnline(PeerB) {
			t.Errorf("Bob does not see herself online.")
			return
		}

		if layA.IsOnline(PeerB) {
			t.Errorf("Alice can see Bob without talking to him.")
			return
		}

		if layB.IsOnline(PeerA) {
			t.Errorf("Bob can see Alice without talking to him.")
			return
		}

		cnvB, err := layA.Talk(PeerB)
		if err != nil {
			t.Errorf("Talk(bob) failed: %v", err)
			return
		}

		cnvA, err := layB.Talk(PeerA)
		if err != nil {
			t.Errorf("Talk(alice) failed: %v", err)
			return
		}

		if !layA.IsOnline(PeerB) {
			t.Errorf("Alice does not see Bob online.")
			return
		}

		if !layB.IsOnline(PeerA) {
			t.Errorf("Bob does not see Alice online.")
			return
		}

		if err := cnvA.Close(); err != nil {
			t.Errorf("Closing conversation failed")
			return
		}

		if err := cnvB.Close(); err != nil {
			t.Errorf("Closing conversation failed")
			return
		}
	})
}

type authAllowSelf string

// Authenticate just nods yes to everything.
func (s authAllowSelf) Authenticate(id string, cred []byte) error {
	fmt.Println("%s == %s", s, id)
	if string(s) == id {
		return nil
	} else {
		return fmt.Errorf("You shall not pass")
	}
}

func (_ authAllowSelf) Credentials(id id.Peer) ([]byte, error) {
	return nil, nil
}

func (_ authAllowSelf) TunnelFor(id id.Peer) (transfer.AuthTunnel, error) {
	return nil, nil
}

func TestNoAuth(t *testing.T) {
	// TODO: This fails for obscure reasons.
	layA := NewLayer(PeerA, authAllowSelf(PeerA.ID()))
	layB := NewLayer(PeerB, authAllowSelf(PeerB.ID()))

	withOnlineLayer(t, PortA, layA, func() {
		withOnlineLayer(t, PortB, layB, func() {
			_, err := layA.Talk(PeerB)
			if err == nil {
				t.Errorf("Auth was not denied")
			}
		})
	})
}
