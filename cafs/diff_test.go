package cafs

import (
	"fmt"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/cafs/db"
	n "github.com/disorganizer/brig/cafs/nodes"
	h "github.com/disorganizer/brig/util/hashlib"
)

func init() {
	log.SetLevel(log.WarnLevel)
}

func mustMove(t *testing.T, lkr *Linker, nd n.ModNode, destPath string) n.ModNode {
	if err := move(lkr, nd, destPath); err != nil {
		t.Fatalf("move of %s to %s failed: %v", nd.Path(), destPath, err)
	}

	newNd, err := lkr.LookupModNode(destPath)
	if err != nil {
		t.Fatalf("Failed to lookup dest path `%s` of new node: %v", destPath, err)
	}

	return newNd
}

func mustRemove(t *testing.T, lkr *Linker, nd n.ModNode) n.ModNode {
	if _, _, err := remove(lkr, nd, true, false); err != nil {
		t.Fatalf("Failed to remove %s: %v", nd.Path(), err)
	}

	newNd, err := lkr.LookupModNode(nd.Path())
	if err != nil {
		t.Fatalf("Failed to lookup dest path `%s` of deleted node: %v", nd.Path(), err)
	}

	return newNd
}

func mustCommit(t *testing.T, lkr *Linker, msg string) *n.Commit {
	if err := lkr.MakeCommit(n.AuthorOfStage(), msg); err != nil {
		t.Fatalf("Failed to make commit with msg %s: %v", msg, err)
	}

	head, err := lkr.Head()
	if err != nil {
		t.Fatalf("Failed to retrieve head after commit: %v", err)
	}

	return head
}

func makeFileAndCommit(
	t *testing.T, lkr *Linker,
	path string, seed byte) (*n.File, *n.Commit) {

	info := &NodeUpdate{
		Hash:   h.TestDummy(t, seed),
		Size:   uint64(seed),
		Author: "",
		Key:    nil,
	}

	file, err := stage(lkr, path, info)
	if err != nil {
		t.Fatalf("Failed to stage %s at %d: %v", path, seed, err)
	}

	return file, mustCommit(t, lkr, fmt.Sprintf("cmt %d", seed))
}

type moveSetup struct {
	commits []*n.Commit
	paths   []string
	changes []ChangeType
	head    *n.Commit
	node    n.ModNode
}

/////////////// ACTUAL TESTCASES ///////////////

func setupHistoryBasic(t *testing.T, lkr *Linker) *moveSetup {
	file, c1 := makeFileAndCommit(t, lkr, "/x.png", 1)
	file, c2 := makeFileAndCommit(t, lkr, "/x.png", 2)
	file, c3 := makeFileAndCommit(t, lkr, "/x.png", 3)

	head, err := lkr.Head()
	if err != nil {
		t.Fatalf("Failed to retrieve head: %v", err)
	}

	return &moveSetup{
		commits: []*n.Commit{c3, c2, c1},
		paths: []string{
			"/x.png",
			"/x.png",
			"/x.png",
		},
		changes: []ChangeType{
			ChangeTypeNone,
			ChangeTypeModify,
			ChangeTypeAdd,
		},
		head: head,
		node: file,
	}
}

func setupHistoryRemoved(t *testing.T, lkr *Linker) *moveSetup {
	file, c1 := makeFileAndCommit(t, lkr, "/x.png", 1)
	file, c2 := makeFileAndCommit(t, lkr, "/x.png", 2)
	mustRemove(t, lkr, file)
	c3 := mustCommit(t, lkr, "after remove")

	// removing will copy file and make that a ghost.
	// i.e. we need to re-lookup it:
	ghost, err := lkr.LookupGhost(file.Path())
	if err != nil {
		t.Fatalf("Failed to lookup ghost at %s: %v", file.Path(), err)
	}

	return &moveSetup{
		commits: []*n.Commit{c3, c2, c1},
		paths: []string{
			"/x.png",
			"/x.png",
			"/x.png",
		},
		changes: []ChangeType{
			ChangeTypeNone,
			ChangeTypeRemove,
			ChangeTypeAdd,
		},
		head: c3,
		node: ghost,
	}
}

func setupHistoryMoved(t *testing.T, lkr *Linker) *moveSetup {
	file, c1 := makeFileAndCommit(t, lkr, "/x.png", 1)
	file, c2 := makeFileAndCommit(t, lkr, "/x.png", 2)
	mustMove(t, lkr, file, "/y.png")
	c3 := mustCommit(t, lkr, "post-move")

	return &moveSetup{
		commits: []*n.Commit{c3, c2, c1},
		paths: []string{
			"/y.png",
			"/x.png",
			"/x.png",
		},
		changes: []ChangeType{
			ChangeTypeNone,
			ChangeTypeMove,
			ChangeTypeAdd,
		},
		head: c3,
		node: file,
	}
}

func setupHistoryMoveStaging(t *testing.T, lkr *Linker) *moveSetup {
	file, c1 := makeFileAndCommit(t, lkr, "/x.png", 1)
	file, c2 := makeFileAndCommit(t, lkr, "/x.png", 2)
	mustMove(t, lkr, file, "/y.png")

	status, err := lkr.Status()
	if err != nil {
		t.Fatalf("Failed to retrieve status: %v", err)
	}

	return &moveSetup{
		commits: []*n.Commit{status, c2, c1},
		paths: []string{
			"/y.png",
			"/x.png",
			"/x.png",
		},
		changes: []ChangeType{
			ChangeTypeNone,
			ChangeTypeMove,
			ChangeTypeAdd,
		},
		head: status,
		node: file,
	}
}

func setupHistoryMoveAndModify(t *testing.T, lkr *Linker) *moveSetup {
	file, c1 := makeFileAndCommit(t, lkr, "/x.png", 1)
	file, c2 := makeFileAndCommit(t, lkr, "/x.png", 2)

	newFile := mustMove(t, lkr, file, "/y.png")
	modifyFile(t, lkr, newFile.(*n.File), 42)
	c3 := mustCommit(t, lkr, "post-move-modify")

	return &moveSetup{
		commits: []*n.Commit{c3, c2, c1},
		paths: []string{
			"/y.png",
			"/x.png",
			"/x.png",
		},
		changes: []ChangeType{
			ChangeTypeNone,
			ChangeTypeModify | ChangeTypeMove,
			ChangeTypeAdd,
		},
		head: c3,
		node: file,
	}
}

func setupHistoryMoveAndModifyStage(t *testing.T, lkr *Linker) *moveSetup {
	file, c1 := makeFileAndCommit(t, lkr, "/x.png", 1)
	file, c2 := makeFileAndCommit(t, lkr, "/x.png", 2)

	newFile := mustMove(t, lkr, file, "/y.png")

	modifyFile(t, lkr, newFile.(*n.File), 42)

	status, err := lkr.Status()
	if err != nil {
		t.Fatalf("Failed to retrieve status: %v", err)
	}

	return &moveSetup{
		commits: []*n.Commit{status, c2, c1},
		paths: []string{
			"/y.png",
			"/x.png",
			"/x.png",
		},
		changes: []ChangeType{
			ChangeTypeNone,
			ChangeTypeModify | ChangeTypeMove,
			ChangeTypeAdd,
		},
		head: status,
		node: file,
	}
}

func setupHistoryRemoveReadd(t *testing.T, lkr *Linker) *moveSetup {
	file, c1 := makeFileAndCommit(t, lkr, "/x.png", 1)
	file, c2 := makeFileAndCommit(t, lkr, "/x.png", 2)
	mustRemove(t, lkr, file)
	c3 := mustCommit(t, lkr, "after remove")
	file, c4 := makeFileAndCommit(t, lkr, "/x.png", 2)

	return &moveSetup{
		commits: []*n.Commit{c4, c3, c2, c1},
		paths: []string{
			"/x.png",
			"/x.png",
			"/x.png",
			"/x.png",
		},
		changes: []ChangeType{
			ChangeTypeNone,
			ChangeTypeAdd,
			ChangeTypeRemove,
			ChangeTypeAdd,
		},
		head: c4,
		node: file,
	}
}

func setupHistoryRemoveReaddModify(t *testing.T, lkr *Linker) *moveSetup {
	file, c1 := makeFileAndCommit(t, lkr, "/x.png", 1)
	file, c2 := makeFileAndCommit(t, lkr, "/x.png", 2)
	mustRemove(t, lkr, file)
	c3 := mustCommit(t, lkr, "after remove")
	file, c4 := makeFileAndCommit(t, lkr, "/x.png", 255)

	return &moveSetup{
		commits: []*n.Commit{c4, c3, c2, c1},
		paths: []string{
			"/x.png",
			"/x.png",
			"/x.png",
			"/x.png",
		},
		changes: []ChangeType{
			ChangeTypeNone,
			ChangeTypeAdd | ChangeTypeModify,
			ChangeTypeRemove,
			ChangeTypeAdd,
		},
		head: c4,
		node: file,
	}
}

func setupHistoryMoveCircle(t *testing.T, lkr *Linker) *moveSetup {
	file, c1 := makeFileAndCommit(t, lkr, "/x.png", 1)
	file, c2 := makeFileAndCommit(t, lkr, "/x.png", 2)

	newFile := mustMove(t, lkr, file, "/y.png")
	c3 := mustCommit(t, lkr, "move to y.png")

	newOldFile := mustMove(t, lkr, newFile, "/x.png")
	c4 := mustCommit(t, lkr, "move back to x.png")

	return &moveSetup{
		commits: []*n.Commit{c4, c3, c2, c1},
		paths: []string{
			"/x.png",
			"/y.png",
			"/x.png",
			"/x.png",
		},
		changes: []ChangeType{
			ChangeTypeNone,
			ChangeTypeMove,
			ChangeTypeMove,
			ChangeTypeAdd,
		},
		head: c4,
		node: newOldFile,
	}
}

type setupFunc func(t *testing.T, lkr *Linker) *moveSetup

// Registry bank for all testcases:
func TestHistoryWalker(t *testing.T) {
	tcs := []struct {
		name  string
		setup setupFunc
	}{
		{
			name:  "no-frills",
			setup: setupHistoryBasic,
		}, {
			name:  "remove-it",
			setup: setupHistoryRemoved,
		}, {
			name:  "remove-readd",
			setup: setupHistoryRemoveReadd,
		}, {
			name:  "remove-readd-modify",
			setup: setupHistoryRemoveReaddModify,
		}, {
			name:  "move-once",
			setup: setupHistoryMoved,
		}, {
			name:  "move-once-stage",
			setup: setupHistoryMoveStaging,
		}, {
			name:  "move-modify",
			setup: setupHistoryMoveAndModify,
		}, {
			name:  "move-modify-stage",
			setup: setupHistoryMoveAndModifyStage,
		}, {
			name:  "move-circle",
			setup: setupHistoryMoveCircle,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			withDummyKv(t, func(kv db.Database) {
				lkr := NewLinker(kv)
				mustCommit(t, lkr, "init")

				setup := tc.setup(t, lkr)
				testHistoryRunner(t, lkr, setup)
			})
		})
	}
}

// Actual test runner:
func testHistoryRunner(t *testing.T, lkr *Linker, setup *moveSetup) {
	idx := 0
	walker := NewHistoryWalker(lkr, setup.head, setup.node)
	for walker.Next() {
		state := walker.State()
		if setup.paths[idx] != state.Curr.Path() {
			t.Fatalf(
				"Wrong path at index `%d`: %s (want: %s)",
				idx, state.Curr.Path(), setup.paths[idx],
			)
		}

		if state.Mask != setup.changes[idx] {
			t.Errorf(
				"Wrong type of state: %v (want: %s)",
				state.Mask,
				setup.changes[idx],
			)
		}

		if !setup.commits[idx].Hash().Equal(state.Head.Hash()) {
			t.Fatalf("Hash in commit differs")
		}

		idx++
	}

	if err := walker.Err(); err != nil {
		t.Fatalf("walker failed at index (%d/%d): %v", idx, len(setup.commits), err)
	}
}

// TODO:
// - Test move in staging (i.e. inode based lookup)
// - Test move and re-add something on old location.
// - Test move node with same content back and forth.
