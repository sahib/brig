package nodes

import (
	"bytes"
	"testing"

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

	ghost, err := MakeGhost(root)
	if err != nil {
		t.Fatalf("Failed to make root dir a ghost: %v", err)
	}

	if ghost.Type() != NodeTypeGhost {
		t.Fatalf("Ghost does not identify itself as ghost: %d", ghost.Type())
	}

	if !bytes.Equal(ghost.OldNode().Hash(), root.Hash()) {
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

	if !bytes.Equal(ghost.OldNode().Hash(), root.Hash()) {
		t.Fatalf("Ghost and real hash differ (%v - %v)", ghost.Hash(), root.Hash())
	}
}
