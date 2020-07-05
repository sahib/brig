package nodes

import (
	"testing"

	ie "github.com/sahib/brig/catfs/errors"
	"github.com/stretchr/testify/require"
	capnp "zombiezen.com/go/capnproto2"
)

func TestDirectoryBasics(t *testing.T) {
	lkr := NewMockLinker()
	repoDir, err := NewEmptyDirectory(lkr, nil, "", "a", 1)
	lkr.MemSetRoot(repoDir)
	lkr.AddNode(repoDir, true)

	if err != nil {
		t.Fatalf("Failed to create empty dir: %v", err)
	}

	subDir, err := NewEmptyDirectory(lkr, repoDir, "sub", "b", 2)
	if err != nil {
		t.Fatalf("Failed to create empty sub dir: %v", err)
	}
	lkr.AddNode(subDir, true)

	if err := repoDir.Add(lkr, subDir); err != ie.ErrExists {
		t.Fatalf("Adding sub/ to repo/ worked twice: %v", err)
	}

	// Fake size here.
	repoDir.size = 3
	repoDir.cachedSize = 3

	msg, err := repoDir.ToCapnp()
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

	empty := &Directory{}
	if err := empty.FromCapnp(newMsg); err != nil {
		t.Fatalf("From capnp failed: %v", err)
	}

	if empty.size != 3 {
		t.Fatalf("Root size was not loaded correctly: %v", err)
	}

	if empty.parentName != "" {
		t.Fatalf("Root parentName as not loaded correctly: %v", err)
	}

	if empty.Inode() != 1 {
		t.Fatalf("Inode was not loaded correctly: %v != 1", empty.Inode())
	}

	if subHash, ok := empty.children["sub"]; ok {
		if !subHash.Equal(subDir.TreeHash()) {
			t.Fatalf("Unmarshaled hash differs (!= sub): %v", subDir.TreeHash())
		}
	} else {
		t.Fatalf("Root children do not contain sub")
	}

	empty.modTime = repoDir.modTime
	require.Equal(t, empty, repoDir)
}
