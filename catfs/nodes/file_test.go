package nodes

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	capnp "zombiezen.com/go/capnproto2"
)

func TestFile(t *testing.T) {
	lkr := NewMockLinker()
	root, err := NewEmptyDirectory(lkr, nil, "", "a", 2)
	if err != nil {
		t.Fatalf("Failed to create root dir: %v", err)
	}
	lkr.AddNode(root)
	lkr.MemSetRoot(root)

	file, err := NewEmptyFile(root, "some file", "a", 3)
	lkr.AddNode(file)

	if err != nil {
		t.Fatalf("Failed to create empty dir: %v", err)
	}

	file.SetName("new_name")
	file.SetKey([]byte{1, 2, 3})
	file.SetSize(42)
	file.SetContent(lkr, []byte{4, 5, 6})
	file.SetBackend(lkr, []byte{7, 8, 9})
	hashBeforeUnmarshal := file.TreeHash().Clone()

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

	if empty.ModTime() != now.Truncate(time.Microsecond) {
		t.Fatalf("modtime differs after unmarshal: %v Want: %v", now, empty.ModTime())
	}

	if empty.Size() != 42 {
		t.Fatalf("size differs after unmarshal: %v", empty.Size())
	}

	if !bytes.Equal(empty.Key(), []byte{1, 2, 3}) {
		t.Fatalf("key differs after unmarshal: %v", empty.Key())
	}

	if !bytes.Equal(empty.TreeHash(), hashBeforeUnmarshal) {
		t.Fatalf("tree hash differs after unmarshal: %v", empty.TreeHash())
	}

	if !bytes.Equal(empty.BackendHash(), []byte{7, 8, 9}) {
		t.Fatalf("backend hash differs after unmarshal: %v", empty.BackendHash())
	}

	if !bytes.Equal(empty.ContentHash(), []byte{4, 5, 6}) {
		t.Fatalf("content hash differs after unmarshal: %v", empty.ContentHash())
	}

	empty.modTime = file.modTime
	require.Equal(t, empty, file)
}
