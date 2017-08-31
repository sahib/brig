package nodes

import (
	"bytes"
	"testing"

	h "github.com/disorganizer/brig/util/hashlib"
	capnp "zombiezen.com/go/capnproto2"
)

func TestGhost(t *testing.T) {
	lkr := NewMockLinker()
	root, err := NewEmptyDirectory(lkr, nil, "", 1)
	if err != nil {
		t.Fatalf("Failed to create root dir: %v", err)
	}
	lkr.AddNode(root)
	lkr.MemSetRoot(root)

	file, err := NewEmptyFile(root, "x.png", 42)
	file.content = h.TestDummy(t, 2)
	file.hash = h.TestDummy(t, 3)
	file.size = 13

	if err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	ghost, err := MakeGhost(file, 666)
	if err != nil {
		t.Fatalf("Failed to make root dir a ghost: %v", err)
	}

	if ghost.Type() != NodeTypeGhost {
		t.Fatalf("Ghost does not identify itself as ghost: %d", ghost.Type())
	}

	if !bytes.Equal(ghost.OldNode().Hash(), file.Hash()) {
		t.Fatalf("Ghost and real hash differ (%v - %v)", ghost.Hash(), root.Hash())
	}

	msg, err := ghost.ToCapnp()
	if err != nil {
		t.Fatalf("Ghost ToCapnp failed: %v", err)
	}

	data, err := msg.Marshal()
	if err != nil {
		t.Fatalf("Ghost marshal failed: %v", err)
	}

	newMsg, err := capnp.Unmarshal(data)
	if err != nil {
		t.Fatalf("Ghost unmarshal failed: %v", err)
	}

	empty := &Ghost{}
	if err := empty.FromCapnp(newMsg); err != nil {
		t.Fatalf("Ghost FromCapnp failed: %v", err)
	}

	if !bytes.Equal(ghost.OldNode().Hash(), file.Hash()) {
		t.Fatalf("Ghost and real hash differ (%v - %v)", ghost.Hash(), root.Hash())
	}

	unmarshaledFile, err := ghost.OldFile()
	if err != nil {
		t.Fatalf("Failed to cast ghost to old file: %v", err)
	}

	if !unmarshaledFile.Content().Equal(file.Content()) {
		t.Fatalf("Hash content differs after unmarshal: %v", unmarshaledFile.Content())
	}

	if !unmarshaledFile.Hash().Equal(file.Hash()) {
		t.Fatalf("Hash itself differs after unmarshal: %v", unmarshaledFile.Hash())
	}

	if unmarshaledFile.Inode() != file.Inode() {
		t.Fatalf("Inodes differ after unmarshal: %d != %d", unmarshaledFile.Inode, file.Inode())
	}

	if empty.Inode() != ghost.Inode() {
		t.Fatalf("Inodes differ after unmarshal: %d != %d", unmarshaledFile.Inode, file.Inode())
	}
}
