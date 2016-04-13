package protocol

import (
	"bytes"
	"testing"

	"github.com/disorganizer/brig/util/protocol/wire"
	"github.com/disorganizer/brig/util/testutil"
	"github.com/gogo/protobuf/proto"
)

func testProtocol(t *testing.T, compress bool) {
	b := &bytes.Buffer{}
	p := NewProtocol(b, compress)

	// Test with varying potential for compression:
	for i := 0; i < 5; i++ {
		msg := &wire.Response{
			ReqType: wire.RequestType_DUMMY.Enum(),
			Data:    testutil.CreateDummyBuf(int64(i) * 255),
		}

		if err := p.Send(msg); err != nil {
			t.Errorf("Send failed: %v", err)
			return
		}

		remoteMsg := &wire.Response{
			ReqType: wire.RequestType_DUMMY.Enum(),
			ID:      proto.Int64(0),
		}

		if err := p.Recv(remoteMsg); err != nil {
			t.Errorf("Recv failed: %v", err)
			return
		}

		if !bytes.Equal(msg.GetData(), remoteMsg.GetData()) {
			t.Errorf("Data differs.")
			t.Errorf("\tWANT: %v", msg.GetData())
			t.Errorf("\tGOT:  %v", remoteMsg.GetData())
			return
		}

		if msg.GetReqType() != remoteMsg.GetReqType() {
			t.Errorf("Types differ.")
			return
		}
	}
}

func TestProtocol(t *testing.T) {
	testProtocol(t, true)
	testProtocol(t, false)
}
