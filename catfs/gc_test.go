package catfs

import (
	"testing"

	"github.com/disorganizer/brig/catfs/db"
	n "github.com/disorganizer/brig/catfs/nodes"
)

func assertNodeExists(t *testing.T, kv db.Database, nd n.Node) {
	if _, err := kv.Get("stage", "objects", nd.Hash().B58String()); err != nil {
		t.Fatalf("Stage object %v does not exist: %v", nd, err)
	}
}

func TestGC(t *testing.T) {
	mdb := db.NewMemoryDatabase()
	lkr := NewLinker(mdb)

	killExpected := make(map[string]bool)
	killActual := make(map[string]bool)

	gc := NewGarbageCollector(lkr, mdb, func(nd n.Node) bool {
		killActual[nd.Hash().B58String()] = true
		return true
	})

	root, err := lkr.Root()
	if err != nil {
		t.Fatalf("Failed to retrieve the root: %v", root)
	}

	killExpected[root.Hash().B58String()] = true

	sub1, err := n.NewEmptyDirectory(lkr, root, "a", 2)
	if err != nil {
		t.Fatalf("Creating sub2 failed: %v", err)
	}

	if err := lkr.StageNode(sub1); err != nil {
		t.Fatalf("Staging root failed: %v", err)
	}

	killExpected[root.Hash().B58String()] = true
	killExpected[sub1.Hash().B58String()] = true

	sub2, err := n.NewEmptyDirectory(lkr, sub1, "b", 3)
	if err != nil {
		t.Fatalf("Creating sub2 failed: %v", err)
	}

	if err := lkr.StageNode(sub2); err != nil {
		t.Fatalf("Staging root failed: %v", err)
	}

	if err := gc.Run(true); err != nil {
		t.Fatalf("gc run failed: %v", err)
	}

	if len(killExpected) != len(killActual) {
		t.Fatalf("GC killed %d nodes, but should have killed %d", len(killActual), len(killExpected))
	}

	for killedHash := range killActual {
		if _, ok := killExpected[killedHash]; !ok {
			t.Fatalf("%s was killed, but should not!", killedHash)
		}

		if _, err := mdb.Get("stage", "objects", killedHash); err != db.ErrNoSuchKey {
			t.Fatalf("GC did not wipe key from db: %v", killedHash)
		}
	}

	// Double check that the gc did not delete other stuff from the db:
	assertNodeExists(t, mdb, root)
	assertNodeExists(t, mdb, sub1)
	assertNodeExists(t, mdb, sub2)

	gc = NewGarbageCollector(lkr, mdb, func(nd n.Node) bool {
		t.Fatalf("Second gc run found something, first didn't")
		return true
	})

	if err := gc.Run(true); err != nil {
		t.Fatalf("Second gc run failed: %v", err)
	}

	if err := lkr.MakeCommit(n.AuthorOfStage(), "some message"); err != nil {
		t.Fatalf("MakeCommit() failed: %v", err)
	}

	gc = NewGarbageCollector(lkr, mdb, func(nd n.Node) bool {
		t.Fatalf("Third gc run found something, first didn't")
		return true
	})

	if err := gc.Run(true); err != nil {
		t.Fatalf("Third gc run failed: %v", err)
	}
}
