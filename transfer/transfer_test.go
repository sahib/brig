package transfer

import (
	"fmt"
	"io"
	"testing"

	"github.com/disorganizer/brig/im"
	"github.com/disorganizer/brig/transfer/proto"
	"github.com/disorganizer/brig/util/testutil"
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

func testIO(t *testing.T, cl *Client) {
	// Client side is just a goroutine:
	resp, err := cl.Send(&proto.Request{
		Type: proto.RequestType_PING.Enum(),
		//Data: testutil.CreateDummyBuf(20),
	})

	if err != nil {
		t.Errorf("Sending clone failed: %v", err)
		return
	}

	if resp.GetType() != proto.RequestType_PING {
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
}

func TestMemoryIO(t *testing.T) {
	alice, bob := connect()

	cl := NewClient(alice)
	sv := NewServer(bob, nil)

	go testIO(t, cl)

	if err := sv.Serve(); err != nil {
		t.Errorf("Serve failed with error: %v", err)
		return
	}
}

func TestRealXMPP(t *testing.T) {
	clientAlice, err := im.NewDummyClient(im.AliceJid, im.AlicePwd)
	if err != nil {
		t.Errorf("Creating Alice' client failed: %v", err)
		return
	}

	defer clientAlice.Close()

	clientBob, err := im.NewDummyClient(im.BobJid, im.BobPwd)
	if err != nil {
		t.Errorf("Creating Alice' client failed: %v", err)
		return
	}

	defer clientBob.Close()

	paths, err := im.MakeBuddies(clientAlice, clientBob)
	defer testutil.Remover(t, paths...)

	if err != nil {
		t.Errorf("Could not make buddies: %v", err)
		return
	}

	go func() {
		aliceCnv, err := clientAlice.Dial(im.BobJid)
		if err != nil {
			t.Fatalf("Talking to bob failed: %v", err)
			return
		}

		testIO(t, NewClient(aliceCnv))
	}()

	bobCnv := clientBob.Listen()
	if bobCnv == nil {
		t.Errorf("Incoming conversation is nil.")
		return
	}

	sv := NewServer(bobCnv, nil)

	if err := sv.Serve(); err != nil {
		t.Errorf("Serve failed with error: %v", err)
		return
	}
}
