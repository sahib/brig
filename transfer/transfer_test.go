package transfer

import (
	"fmt"
	"io"
	"testing"

	"github.com/disorganizer/brig/transfer/proto"
)

type Network struct {
	io.Reader
	io.Writer
}

func (n *Network) Read(p []byte) (int, error) {
	return n.Reader.Read(p)
}

func (n *Network) Write(p []byte) (int, error) {
	return n.Writer.Write(p)
}

func (n *Network) Close() error {
	// NO-OP
	return nil
}

func connect() (io.ReadWriteCloser, io.ReadWriteCloser) {
	// That looks easy, but I spend a very confused hour of
	// debugging to get the order right. Embarassing, I know.
	ar, bw := io.Pipe()
	br, aw := io.Pipe()

	return &Network{Reader: ar, Writer: aw}, &Network{Reader: br, Writer: bw}
}

func TestIO(t *testing.T) {
	alice, bob := connect()

	cl := NewClient(alice)
	sv := NewServer(bob, nil)

	// Client side is just a goroutine:
	go func() {
		resp, err := cl.Send(&proto.Request{
			Type: proto.RequestType_CLONE.Enum(),
			//Data: testutil.CreateDummyBuf(20),
		})

		if err != nil {
			t.Errorf("Sending clone failed: %v", err)
			return
		}

		if resp.GetType() != proto.RequestType_CLONE {
			fmt.Println("SEND SUCCESS NOW QUIOT", resp.GetType())
			t.Errorf("Got a wrong id from command: %v", resp.GetType())
			return
		}

		resp, err = cl.Send(&proto.Request{
			Type: proto.RequestType_QUIT.Enum(),
		})
		if err != nil {
			t.Errorf("Sending quit failed: %v", err)
			return
		}

		if resp.GetType() != proto.RequestType_QUIT {
			t.Errorf("Got a wrong id for the quit command: %v", resp.GetType())
			return
		}
	}()

	if err := sv.Serve(); err != nil {
		t.Fatalf("Serve failed with error: %v", err)
	}
}
