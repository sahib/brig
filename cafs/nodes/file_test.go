package nodes

import (
	"bytes"
	"testing"
	"time"

	capnp "zombiezen.com/go/capnproto2"
)

func TestFile(t *testing.T) {
	lkr := NewMockLinker()
	root, err := NewEmptyDirectory(lkr, nil, "", 2)
	if err != nil {
		t.Fatalf("Failed to create root dir: %v", err)
	}
	lkr.AddNode(root)
	lkr.MemSetRoot(root)

	file, err := NewEmptyFile(root, "some file", 3)
	lkr.AddNode(file)

	if err != nil {
		t.Fatalf("Failed to create empty dir: %v", err)
	}

	file.SetName("new_name")
	file.SetKey([]byte{1, 2, 3})
	file.SetSize(42)
	file.SetContent(lkr, []byte{4, 5, 6})
	hashBeforeUnmarshal := file.Hash().Clone()

	now := time.Now()
	file.SetModTime(now)

	msg, err := file.ToCapnp()
	if err != nil {
		t.Fatalf("Failed to convert repo dir to capnp: %v", err)
	}

	data, err := msg.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}

	newMsg, err := capnp.Unmarshal(data)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	empty := &File{}
	if err := empty.FromCapnp(newMsg); err != nil {
		t.Fatalf("From capnp failed: %v", err)
	}

	if empty.Name() != "new_name" {
		t.Fatalf("Name differs after unmarshal: %v", empty.Name())
	}

	if empty.ModTime() != now {
		t.Fatalf("modtime differs after unmarshal: %v Want: %v", now, empty.ModTime())
	}

	if empty.Size() != 42 {
		t.Fatalf("size differs after unmarshal: %v", empty.Size())
	}

	if !bytes.Equal(empty.Key(), []byte{1, 2, 3}) {
		t.Fatalf("key differs after unmarshal: %v", empty.Key())
	}

	if !bytes.Equal(empty.Hash(), hashBeforeUnmarshal) {
		t.Fatalf("hash differs after unmarshal: %v", empty.Hash())
	}
}
