package nodes

import (
	"testing"

	h "github.com/disorganizer/brig/util/hashlib"
	capnp "zombiezen.com/go/capnproto2"
)

func TestCommit(t *testing.T) {
	lkr := NewMockLinker()
	cmt, err := NewCommit(lkr, nil)
	if err != nil {
		t.Fatalf("Failed to create commit: %v", err)
	}

	cmt.root = h.EmptyHash
	cmt.parent = h.EmptyHash
	cmt.Base.name = "some commit"

	if err := cmt.BoxCommit("Hello"); err != nil {
		t.Fatalf("Failed to box commit: %v", err)
	}

	msg, err := cmt.ToCapnp()
	if err != nil {
		t.Fatalf("Failed to convert commit to capnp: %v", err)
	}

	data, err := msg.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}

	// fmt.Println("file write")
	// fd, err := os.OpenFile("/tmp/cmt.capnp", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	// if err != nil {
	// 	t.Fatalf("Something is very wrong")
	// }
	// if _, err := fd.Write(data); err != nil {
	// 	t.Fatalf("Something is very wrong")
	// }
	// if err := fd.Close(); err != nil {
	// 	t.Fatalf("Something is very wrong")
	// }

	newMsg, err := capnp.Unmarshal(data)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	empty := &Commit{}
	if err := empty.FromCapnp(newMsg); err != nil {
		t.Fatalf("From failed: %v", err)
	}

	if empty.message != "Hello" {
		t.Fatalf("Bad message unmarshaled: %v", empty.message)
	}

	if !empty.root.Equal(h.EmptyHash) {
		t.Fatalf("Bad root unmarshaled", empty.root)
	}

	if !empty.parent.Equal(h.EmptyHash) {
		t.Fatalf("Bad parent unmarshaled", empty.root)
	}
}
