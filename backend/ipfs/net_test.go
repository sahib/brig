package ipfs

import (
	"bytes"
	"testing"

	"github.com/sahib/brig/net/peer"
)

var (
	TestProtocol = "/brig/unittest"
	hello        = []byte("Hello")
	world        = []byte("World")
)

func TestNet(t *testing.T) {
	WithIpfsAtPort(t, 4002, func(alice *Node) {
		if err := alice.Connect(); err != nil {
			t.Fatalf("alice failed to go online: %v", err)
		}

		aliceID, err := alice.Identity()
		if err != nil {
			t.Fatalf("Could not get alice's identity %v", err)
		}

		var bobId peer.Info

		t.Logf("Alice is online (%v).", aliceID)
		WithIpfsAtPort(t, 4003, func(bob *Node) {
			if err := bob.Connect(); err != nil {
				t.Fatalf("Bob failed to go online: %v", err)
			}

			bobId, err = bob.Identity()
			if err != nil {
				t.Fatalf("Could not get bob's identity: %v", err)
			}

			t.Logf("Bob is online. (%v)", bobId)

			ls, err := bob.Listen(TestProtocol)
			if err != nil {
				t.Fatalf("Failed to listen on ipfs: %v", err)
			}

			go func() {
				conn, err := ls.Accept()
				if err != nil {
					t.Fatalf("Accept() failed: %v", err)
				}

				buf := make([]byte, 5)
				if n, err := conn.Read(buf); err != nil && n != len(buf) {
					t.Fatalf("Listen-Read failed: %v (len: %d)", err, n)
				}

				if !bytes.Equal(buf, []byte(hello)) {
					t.Fatalf("Read data does not match. Expected '%s'; got '%s'", hello, buf)
				}

				if _, err := conn.Write(world); err != nil {
					t.Fatalf("Liste-Write failed: %v", err)
				}

				if err := conn.Close(); err != nil {
					t.Fatalf("Listen-Close conn failed: %v", err)
				}

				if err := ls.Close(); err != nil {
					t.Fatalf("Closing listener failed: %v", err)
				}
			}()

			// Alice sending data to bob:
			conn, err := alice.Dial(bobId.Addr, TestProtocol)
			if err != nil {
				t.Fatalf("Dial(self) did not work: %v", err)
			}

			if _, err := conn.Write([]byte(hello)); err != nil {
				t.Fatalf("Write(self) failed: %v", err)
			}

			buf := make([]byte, 5)
			if n, err := conn.Read(buf); err != nil && n != len(buf) {
				t.Fatalf("Read(self) failed: %v (len: %d)", err, n)
			}

			if !bytes.Equal(buf, []byte(world)) {
				t.Fatalf("Read data does not match. Expected '%s'; got '%s'", world, buf)
			}

			if err := conn.Close(); err != nil {
				t.Fatalf("Closing conn failed: %v", err)
			}
		})
	})
}
