package vcs

import (
	"fmt"
	"testing"

	log "github.com/Sirupsen/logrus"
	c "github.com/sahib/brig/catfs/core"
	n "github.com/sahib/brig/catfs/nodes"
	"github.com/stretchr/testify/require"
)

func init() {
	log.SetLevel(log.WarnLevel)
}

type moveSetup struct {
	commits []*n.Commit
	paths   []string
	changes []ChangeType
	head    *n.Commit
	node    n.ModNode
}

/////////////// ACTUAL TESTCASES ///////////////

func setupHistoryBasic(t *testing.T, lkr *c.Linker) *moveSetup {
	file, c1 := c.MustTouchAndCommit(t, lkr, "/x.png", 1)
	file, c2 := c.MustTouchAndCommit(t, lkr, "/x.png", 2)
	file, c3 := c.MustTouchAndCommit(t, lkr, "/x.png", 3)

	head, err := lkr.Head()
	if err != nil {
		t.Fatalf("Failed to retrieve head: %v", err)
	}

	return &moveSetup{
		commits: []*n.Commit{c3, c2, c1, c1},
		paths: []string{
			"/x.png",
			"/x.png",
			"/x.png",
			"/x.png",
		},
		changes: []ChangeType{
			ChangeTypeNone,
			ChangeTypeModify,
			ChangeTypeModify,
			ChangeTypeAdd,
		},
		head: head,
		node: file,
	}
}

func setupHistoryRemoved(t *testing.T, lkr *c.Linker) *moveSetup {
	file, c1 := c.MustTouchAndCommit(t, lkr, "/x.png", 1)
	file, c2 := c.MustTouchAndCommit(t, lkr, "/x.png", 2)
	c.MustRemove(t, lkr, file)
	c3 := c.MustCommit(t, lkr, "after remove")

	// removing will copy file and make that a ghost.
	// i.e. we need to re-lookup it:
	ghost, err := lkr.LookupGhost(file.Path())
	if err != nil {
		t.Fatalf("Failed to lookup ghost at %s: %v", file.Path(), err)
	}

	return &moveSetup{
		commits: []*n.Commit{c3, c2, c1, c1},
		paths: []string{
			"/x.png",
			"/x.png",
			"/x.png",
			"/x.png",
		},
		changes: []ChangeType{
			ChangeTypeNone,
			ChangeTypeRemove,
			ChangeTypeModify,
			ChangeTypeAdd,
		},
		head: c3,
		node: ghost,
	}
}

func setupHistoryMoved(t *testing.T, lkr *c.Linker) *moveSetup {
	file, c1 := c.MustTouchAndCommit(t, lkr, "/x.png", 1)
	file, c2 := c.MustTouchAndCommit(t, lkr, "/x.png", 2)
	c.MustMove(t, lkr, file, "/y.png")
	c3 := c.MustCommit(t, lkr, "post-move")

	return &moveSetup{
		commits: []*n.Commit{c3, c2, c1, c1},
		paths: []string{
			"/y.png",
			"/x.png",
			"/x.png",
			"/x.png",
		},
		changes: []ChangeType{
			ChangeTypeNone,
			ChangeTypeMove,
			ChangeTypeModify,
			ChangeTypeAdd,
		},
		head: c3,
		node: file,
	}
}

func setupHistoryMoveStaging(t *testing.T, lkr *c.Linker) *moveSetup {
	file, c1 := c.MustTouchAndCommit(t, lkr, "/x.png", 1)
	file, c2 := c.MustTouchAndCommit(t, lkr, "/x.png", 2)
	c.MustMove(t, lkr, file, "/y.png")

	status, err := lkr.Status()
	if err != nil {
		t.Fatalf("Failed to retrieve status: %v", err)
	}

	return &moveSetup{
		commits: []*n.Commit{status, c2, c1, c1},
		paths: []string{
			"/y.png",
			"/x.png",
			"/x.png",
			"/x.png",
		},
		changes: []ChangeType{
			ChangeTypeNone,
			ChangeTypeMove,
			ChangeTypeModify,
			ChangeTypeAdd,
		},
		head: status,
		node: file,
	}
}

func setupHistoryMoveAndModify(t *testing.T, lkr *c.Linker) *moveSetup {
	file, c1 := c.MustTouchAndCommit(t, lkr, "/x.png", 1)
	file, c2 := c.MustTouchAndCommit(t, lkr, "/x.png", 2)

	newFile := c.MustMove(t, lkr, file, "/y.png")
	c.MustModify(t, lkr, newFile.(*n.File), 42)
	c3 := c.MustCommit(t, lkr, "post-move-modify")

	return &moveSetup{
		commits: []*n.Commit{c3, c2, c1, c1},
		paths: []string{
			"/y.png",
			"/x.png",
			"/x.png",
			"/x.png",
		},
		changes: []ChangeType{
			ChangeTypeNone,
			ChangeTypeModify | ChangeTypeMove,
			ChangeTypeModify,
			ChangeTypeAdd,
		},
		head: c3,
		node: file,
	}
}

func setupHistoryMoveAndModifyStage(t *testing.T, lkr *c.Linker) *moveSetup {
	file, c1 := c.MustTouchAndCommit(t, lkr, "/x.png", 1)
	file, c2 := c.MustTouchAndCommit(t, lkr, "/x.png", 2)
	newFile := c.MustMove(t, lkr, file, "/y.png")
	c.MustModify(t, lkr, newFile.(*n.File), 42)

	status, err := lkr.Status()
	if err != nil {
		t.Fatalf("Failed to retrieve status: %v", err)
	}

	return &moveSetup{
		commits: []*n.Commit{status, c2, c1, c1},
		paths: []string{
			"/y.png",
			"/x.png",
			"/x.png",
			"/x.png",
		},
		changes: []ChangeType{
			ChangeTypeNone,
			ChangeTypeModify | ChangeTypeMove,
			ChangeTypeModify,
			ChangeTypeAdd,
		},
		head: status,
		node: file,
	}
}

func setupHistoryRemoveReadd(t *testing.T, lkr *c.Linker) *moveSetup {
	file, c1 := c.MustTouchAndCommit(t, lkr, "/x.png", 1)
	file, c2 := c.MustTouchAndCommit(t, lkr, "/x.png", 2)
	c.MustRemove(t, lkr, file)
	c3 := c.MustCommit(t, lkr, "after remove")
	file, c4 := c.MustTouchAndCommit(t, lkr, "/x.png", 2)

	return &moveSetup{
		commits: []*n.Commit{c4, c3, c2, c1, c1},
		paths: []string{
			"/x.png",
			"/x.png",
			"/x.png",
			"/x.png",
			"/x.png",
		},
		changes: []ChangeType{
			ChangeTypeNone,
			ChangeTypeAdd,
			ChangeTypeRemove,
			ChangeTypeModify,
			ChangeTypeAdd,
		},
		head: c4,
		node: file,
	}
}

func setupHistoryRemoveReaddModify(t *testing.T, lkr *c.Linker) *moveSetup {
	file, c1 := c.MustTouchAndCommit(t, lkr, "/x.png", 1)
	file, c2 := c.MustTouchAndCommit(t, lkr, "/x.png", 2)
	c.MustRemove(t, lkr, file)
	c3 := c.MustCommit(t, lkr, "after remove")
	file, c4 := c.MustTouchAndCommit(t, lkr, "/x.png", 255)

	return &moveSetup{
		commits: []*n.Commit{c4, c3, c2, c1, c1},
		paths: []string{
			"/x.png",
			"/x.png",
			"/x.png",
			"/x.png",
			"/x.png",
		},
		changes: []ChangeType{
			ChangeTypeNone,
			ChangeTypeAdd | ChangeTypeModify,
			ChangeTypeRemove,
			ChangeTypeModify,
			ChangeTypeAdd,
		},
		head: c4,
		node: file,
	}
}

func setupHistoryMoveCircle(t *testing.T, lkr *c.Linker) *moveSetup {
	file, c1 := c.MustTouchAndCommit(t, lkr, "/x.png", 1)
	file, c2 := c.MustTouchAndCommit(t, lkr, "/x.png", 2)
	newFile := c.MustMove(t, lkr, file, "/y.png")
	c3 := c.MustCommit(t, lkr, "move to y.png")
	newOldFile := c.MustMove(t, lkr, newFile, "/x.png")
	c4 := c.MustCommit(t, lkr, "move back to x.png")

	return &moveSetup{
		commits: []*n.Commit{c4, c3, c2, c1, c1},
		paths: []string{
			"/x.png",
			"/y.png",
			"/x.png",
			"/x.png",
			"/x.png",
		},
		changes: []ChangeType{
			ChangeTypeNone,
			ChangeTypeMove,
			ChangeTypeMove,
			ChangeTypeModify,
			ChangeTypeAdd,
		},
		head: c4,
		node: newOldFile,
	}
}

func setupHistoryMoveAndReaddFromMoved(t *testing.T, lkr *c.Linker) *moveSetup {
	file, c1 := c.MustTouchAndCommit(t, lkr, "/x.png", 1)
	file, c2 := c.MustTouchAndCommit(t, lkr, "/x.png", 2)

	newFile := c.MustMove(t, lkr, file, "/y.png")
	_, c4 := c.MustTouchAndCommit(t, lkr, "/x.png", 23)

	return &moveSetup{
		commits: []*n.Commit{c4, c2, c1, c1},
		paths: []string{
			"/y.png",
			"/x.png",
			"/x.png",
			"/x.png",
		},
		changes: []ChangeType{
			ChangeTypeNone,
			ChangeTypeMove,
			ChangeTypeModify,
			ChangeTypeAdd,
		},
		head: c4,
		node: newFile,
	}
}

func setupHistoryMoveAndReaddFromAdded(t *testing.T, lkr *c.Linker) *moveSetup {
	file, c1 := c.MustTouchAndCommit(t, lkr, "/x.png", 1)
	file, c2 := c.MustTouchAndCommit(t, lkr, "/x.png", 2)

	c.MustMove(t, lkr, file, "/y.png")
	c3 := c.MustCommit(t, lkr, "move to y.png")
	readdedFile, c4 := c.MustTouchAndCommit(t, lkr, "/x.png", 23)

	return &moveSetup{
		commits: []*n.Commit{c4, c3, c2, c1, c1},
		paths: []string{
			"/x.png",
			"/x.png",
			"/x.png",
			"/x.png",
			"/x.png",
		},
		// TODO: Is this behaviour making sense?
		//       Maybe it makes more sense to "end" the history before the add.
		changes: []ChangeType{
			ChangeTypeNone,
			ChangeTypeAdd | ChangeTypeModify,
			ChangeTypeRemove,
			ChangeTypeModify,
			ChangeTypeAdd,
		},
		head: c4,
		node: readdedFile,
	}
}

type setupFunc func(t *testing.T, lkr *c.Linker) *moveSetup

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
		}, {
			name:  "move-readd-from-moved-perspective",
			setup: setupHistoryMoveAndReaddFromMoved,
		}, {
			name:  "move-readd-from-readded-perspective",
			setup: setupHistoryMoveAndReaddFromAdded,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			c.WithDummyLinker(t, func(lkr *c.Linker) {
				setup := tc.setup(t, lkr)
				testHistoryRunner(t, lkr, setup)
			})
		})
	}
}

// Actual test runner:
func testHistoryRunner(t *testing.T, lkr *c.Linker, setup *moveSetup) {
	idx := 0
	walker := NewHistoryWalker(lkr, setup.head, setup.node)
	for walker.Next() {
		state := walker.State()
		fmt.Println("STATE", idx, state)
		if setup.paths[idx] != state.Curr.Path() {
			t.Fatalf(
				"Wrong path at index `%d`: %s (want: %s)",
				idx, state.Curr.Path(), setup.paths[idx],
			)
		}

		if state.Mask != setup.changes[idx] {
			t.Errorf(
				"%d: Wrong type of state: %v (want: %s)",
				idx,
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

// Test the History() utility based on HistoryWalker.
func TestHistoryUtil(t *testing.T) {
	c.WithDummyLinker(t, func(lkr *c.Linker) {
		c1File, c1 := c.MustTouchAndCommit(t, lkr, "/x.png", 1)

		// c.MustTouchAndCommit will modify c1File for some reason, so copy for
		// expect. That's fine, since catfs is build to re-query nodes freshly.
		c1File = c1File.Copy().(*n.File)

		c2File, c2 := c.MustTouchAndCommit(t, lkr, "/x.png", 2)

		c3FileMoved := c.MustMove(t, lkr, c2File.Copy(), "/y.png")
		c3 := c.MustCommit(t, lkr, "move to y.png")

		_, c4 := c.MustTouchAndCommit(t, lkr, "/x.png", 23)

		states, err := History(lkr, c3FileMoved, c4, nil)
		if err != nil {
			t.Fatalf("History without stop commit failed: %v", err)
		}

		expected := []*NodeState{
			{
				Head: c4,
				Curr: c3FileMoved,
				Mask: ChangeTypeNone,
			}, {
				Head: c3,
				Curr: c3FileMoved,
				Mask: ChangeTypeNone,
			}, {
				Head: c2,
				Curr: c2File,
				Mask: ChangeTypeMove,
			}, {
				Head: c1,
				Curr: c1File,
				Mask: ChangeTypeModify,
			}, {
				Head: c1,
				Curr: c1File,
				Mask: ChangeTypeAdd,
			},
		}

		for idx, state := range states {
			expect := expected[idx]
			require.Equal(t, state.Head, expect.Head, "Head differs")
			require.Equal(t, state.Curr, expect.Curr, "Curr differs")
			require.Equal(t, state.Mask, expect.Mask, "Mask differs")
		}
	})
}

// TODO: Test history for multiple moves in one commit and several commit
