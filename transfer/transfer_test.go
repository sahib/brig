package transfer

import (
	"io"
	"testing"
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
	sv := NewServer(bob)

	// Client side is just a goroutine:
	go func() {
		resp, err := cl.Send(&Command{ID: CmdClone})
		if err != nil {
			t.Errorf("Sending clone failed: %v", err)
			return
		}

		if resp.ID != CmdClone {
			t.Errorf("Got a wrong id from command: %v", resp.ID)
			return
		}

		resp, err = cl.Send(&Command{ID: CmdQuit})
		if err != nil {
			t.Errorf("Sending quit failed: %v", err)
			return
		}

		if resp.ID != CmdQuit {
			t.Errorf("Got a wrong id from command: %v", resp.ID)
			return
		}
	}()

	if err := sv.Serve(); err != nil {
		t.Fatalf("Serve failed with error: %v", err)
	}
}
