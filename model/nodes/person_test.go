package nodes

import (
	"testing"

	h "github.com/disorganizer/brig/util/hashlib"
	capnp "zombiezen.com/go/capnproto2"
)

func TestPerson(t *testing.T) {
	msg, err := AuthorOfStage().ToCapnp()
	if err != nil {
		t.Errorf("Failed to convert author to msg: %v", err)
		return
	}

	data, err := msg.Marshal()
	if err != nil {
		t.Errorf("Failed to marshal message: %v", err)
		return
	}

	newMsg, err := capnp.Unmarshal(data)
	if err != nil {
		t.Errorf("Unmarshal failed: %v", err)
		return
	}

	empty := &Person{}
	if err := empty.FromCapnp(newMsg); err != nil {
		t.Errorf("From failed: %v", err)
		return
	}

	if !empty.Hash.Equal(h.EmptyHash) {
		t.Errorf("Not the empty hash in unmarshaled form")
		return
	}

	if empty.Ident != "unknown" {
		t.Errorf("identity has been messed up")
		return
	}
}
