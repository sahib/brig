package transfer

// TODO: rewrite with Connector and Layer
// import (
// 	"fmt"
// 	"io"
// 	"testing"
//
// 	"github.com/disorganizer/brig/transfer/wire"
// 	"github.com/disorganizer/brig/util/testutil"
// )
//
// type Network struct {
// 	io.Reader
// 	io.Writer
// }
//
// func (n *Network) Read(p []byte) (int, error) {
// 	return n.Reader.Read(p)
// }
//
// func (n *Network) Write(p []byte) (int, error) {
// 	return n.Writer.Write(p)
// }
//
// func (n *Network) Close() error {
// 	// NO-OP
// 	return nil
// }
//
// func connect() (io.ReadWriteCloser, io.ReadWriteCloser) {
// 	// That looks easy, but I spend a very confused hour of
// 	// debugging to get the order right. Embarassing, I know.
// 	ar, bw := io.Pipe()
// 	br, aw := io.Pipe()
//
// 	return &Network{Reader: ar, Writer: aw}, &Network{Reader: br, Writer: bw}
// }
//
// func testIO(t *testing.T, cl *Client) {
// 	// Client side is just a goroutine:
// 	resp, err := cl.Send(&wire.Request{
// 		Type: wire.RequestType_PING.Enum(),
// 		//Data: testutil.CreateDummyBuf(20),
// 	})
//
// 	if err != nil {
// 		t.Errorf("Sending clone failed: %v", err)
// 		return
// 	}
//
// 	if resp.GetType() != wire.RequestType_PING {
// 		fmt.Println("SEND SUCCESS NOW QUIOT", resp.GetType())
// 		t.Errorf("Got a wrong id from command: %v", resp.GetType())
// 		return
// 	}
//
// 	resp, err = cl.Send(&wire.Request{
// 		Type: wire.RequestType_QUIT.Enum(),
// 	})
// 	if err != nil {
// 		t.Errorf("Sending quit failed: %v", err)
// 		return
// 	}
//
// 	if resp.GetType() != wire.RequestType_QUIT {
// 		t.Errorf("Got a wrong id for the quit command: %v", resp.GetType())
// 		return
// 	}
// }
//
// func TestMemoryIO(t *testing.T) {
// 	alice, bob := connect()
//
// 	cl := NewClient(alice)
// 	sv := NewServer(bob, nil)
//
// 	go testIO(t, cl)
//
// 	if err := sv.Serve(); err != nil {
// 		t.Errorf("Serve failed with error: %v", err)
// 		return
// 	}
// }
