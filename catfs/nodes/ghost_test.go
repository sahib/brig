package nodes

import (
	"bytes"
	"testing"

	h "github.com/sahib/brig/util/hashlib"
	capnp "zombiezen.com/go/capnproto2"
)

func TestGhost(t *testing.T) {
	lkr := NewMockLinker()
	root, err := NewEmptyDirectory(lkr, nil, "", "a", 1)
	if err != nil {
		t.Fatalf("Failed to create root dir: %v", err)
	}
	lkr.AddNode(root)
	lkr.MemSetRoot(root)

	file := NewEmptyFile(root, "x.png", "a", 42)
	file.backend = h.TestDummy(t, 2)
	file.tree = h.TestDummy(t, 3)
	file.size = 13

	if err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	ghost, err := MakeGhost(file, 666)
	if err != nil {
		t.Fatalf("Failed to make root dir a ghost: %v", err)
	}

	ghost.SetGhostPath("/other")

	if ghost.Type() != NodeTypeGhost {
		t.Fatalf("Ghost does not identify itself as ghost: %d", ghost.Type())
	}

	if !bytes.Equal(ghost.OldNode().TreeHash(), file.TreeHash()) {
		t.Fatalf("Ghost and real hash differ (%v - %v)", ghost.TreeHash(), root.TreeHash())
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

	if empty.Path() != "/other" {
		t.Fatalf("Ghost path was not unmarshaled: %v", empty.Path())
	}

	if !bytes.Equal(ghost.OldNode().TreeHash(), file.TreeHash()) {
		t.Fatalf("Ghost and real hash differ (%v - %v)", ghost.TreeHash(), root.TreeHash())
	}

	unmarshaledFile, err := ghost.OldFile()
	if err != nil {
		t.Fatalf("Failed to cast ghost to old file: %v", err)
	}

	if !unmarshaledFile.BackendHash().Equal(file.BackendHash()) {
		t.Fatalf("Hash content differs after unmarshal: %v", unmarshaledFile.BackendHash())
	}

	if !unmarshaledFile.TreeHash().Equal(file.TreeHash()) {
		t.Fatalf("Hash itself differs after unmarshal: %v", unmarshaledFile.TreeHash())
	}

	if unmarshaledFile.Inode() != file.Inode() {
		t.Fatalf("Inodes differ after unmarshal: %d != %d", unmarshaledFile.Inode(), file.Inode())
	}

	if unmarshaledFile.Path() != "/x.png" {
		t.Fatalf("Path differs after unmarshal: %v", unmarshaledFile.Path())
	}

	if empty.Inode() != ghost.Inode() {
		t.Fatalf("Inodes differ after unmarshal: %d != %d", unmarshaledFile.Inode(), file.Inode())
	}
}
