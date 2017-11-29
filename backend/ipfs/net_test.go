package ipfs

import (
	"bytes"
	"testing"

	"github.com/disorganizer/brig/net/peer"
)

var (
	TestProtocol = "/brig/unittest"
	hello        = []byte("Hello")
	world        = []byte("World")
)

func TestNet(t *testing.T) {
	WithIpfsAtPort(t, 4002, func(alice *Node) {
		if err := alice.Online(); err != nil {
			t.Errorf("alice failed to go online: %v", err)
			return
		}

		aliceID, err := alice.Identity()
		if err != nil {
			t.Errorf("Could not get alice's identity %v", err)
			return
		}

		var bobId peer.Info

		t.Logf("Alice is online (%v).", aliceID)
		WithIpfsAtPort(t, 4003, func(bob *Node) {
			if err := bob.Online(); err != nil {
				t.Errorf("Bob failed to go online: %v", err)
				return
			}

			bobId, err = bob.Identity()
			if err != nil {
				t.Errorf("Could not get bob's identity: %v", err)
				return
			}

			t.Logf("Bob is online. (%v)", bobId)

			ls, err := bob.Listen(TestProtocol)
			if err != nil {
				t.Errorf("Failed to listen on ipfs: %v", err)
				return
			}

			go func() {
				conn, err := ls.Accept()
				if err != nil {
					t.Errorf("Accept() failed: %v", err)
					return
				}

				buf := make([]byte, 5)
				if n, err := conn.Read(buf); err != nil && n != len(buf) {
					t.Errorf("Listen-Read failed: %v (len: %d)", err, n)
					return
				}

				if !bytes.Equal(buf, []byte(hello)) {
					t.Errorf("Read data does not match. Expected '%s'; got '%s'", hello, buf)
					return
				}

				if _, err := conn.Write(world); err != nil {
					t.Errorf("Liste-Write failed: %v", err)
					return
				}

				if err := conn.Close(); err != nil {
					t.Errorf("Listen-Close conn failed: %v", err)
					return
				}

				if err := ls.Close(); err != nil {
					t.Errorf("Closing listener failed: %v", err)
					return
				}
			}()

			// Alice sending data to bob:
			conn, err := alice.Dial(bobId.Addr, TestProtocol)
			if err != nil {
				t.Errorf("Dial(self) did not work: %v", err)
				return
			}

			if _, err := conn.Write([]byte(hello)); err != nil {
				t.Errorf("Write(self) failed: %v", err)
				return
			}

			buf := make([]byte, 5)
			if n, err := conn.Read(buf); err != nil && n != len(buf) {
				t.Errorf("Read(self) failed: %v (len: %d)", err, n)
				return
			}

			if !bytes.Equal(buf, []byte(world)) {
				t.Errorf("Read data does not match. Expected '%s'; got '%s'", world, buf)
				return
			}

			if err := conn.Close(); err != nil {
				t.Errorf("Closing conn failed: %v", err)
				return
			}
		})
	})
}
