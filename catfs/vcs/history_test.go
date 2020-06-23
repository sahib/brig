package vcs

import (
	"time"
	"testing"

	c "github.com/sahib/brig/catfs/core"
	"github.com/sahib/brig/catfs/db"
	n "github.com/sahib/brig/catfs/nodes"
	h "github.com/sahib/brig/util/hashlib"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func init() {
	log.SetLevel(log.WarnLevel)
}

type historySetup struct {
	commits []*n.Commit
	paths   []string
	changes []ChangeType
	head    *n.Commit
	node    n.ModNode
}

/////////////// ACTUAL TESTCASES ///////////////

func setupHistoryBasic(t *testing.T, lkr *c.Linker) *historySetup {
	_, c1 := c.MustTouchAndCommit(t, lkr, "/x.png", 1)
	_, c2 := c.MustTouchAndCommit(t, lkr, "/x.png", 2)
	file, c3 := c.MustTouchAndCommit(t, lkr, "/x.png", 3)

	status, err := lkr.Status()
	if err != nil {
		t.Fatalf("Failed to retrieve status: %v", err)
	}

	return &historySetup{
		commits: []*n.Commit{status, c3, c2, c1},
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
		head: status,
		node: file,
	}
}

func setupHistoryBasicHole(t *testing.T, lkr *c.Linker) *historySetup {
	_, c1 := c.MustTouchAndCommit(t, lkr, "/x.png", 1)
	_, c2 := c.MustTouchAndCommit(t, lkr, "/x.png", 2)

	// Needed to have a commit that has changes:
	c.MustTouch(t, lkr, "/other", 23)
	file, c3 := c.MustTouchAndCommit(t, lkr, "/x.png", 2)

	status, err := lkr.Status()
	if err != nil {
		t.Fatalf("Failed to retrieve status: %v", err)
	}

	return &historySetup{
		commits: []*n.Commit{status, c3, c2, c1},
		paths: []string{
			"/x.png",
			"/x.png",
			"/x.png",
			"/x.png",
		},
		changes: []ChangeType{
			ChangeTypeNone,
			ChangeTypeNone,
			ChangeTypeModify,
			ChangeTypeAdd,
		},
		head: status,
		node: file,
	}
}

func setupHistoryRemoveImmediately(t *testing.T, lkr *c.Linker) *historySetup {
	x := c.MustTouch(t, lkr, "/x", 1)
	c.MustRemove(t, lkr, x)

	status, err := lkr.Status()
	if err != nil {
		t.Fatalf("Failed to retrieve status: %v", err)
	}

	ghostX, err := lkr.LookupGhost("/x")
	require.Nil(t, err)

	return &historySetup{
		commits: []*n.Commit{status},
		paths: []string{
			"/x",
		},
		changes: []ChangeType{
			ChangeTypeRemove | ChangeTypeAdd,
		},
		head: status,
		node: ghostX,
	}
}

func setupHistoryRemoved(t *testing.T, lkr *c.Linker) *historySetup {
	_, c1 := c.MustTouchAndCommit(t, lkr, "/x.png", 1)
	file, c2 := c.MustTouchAndCommit(t, lkr, "/x.png", 2)
	c.MustRemove(t, lkr, file)
	c3 := c.MustCommit(t, lkr, "after remove")

	status, err := lkr.Status()
	if err != nil {
		t.Fatalf("Failed to acquire status: %v", err)
	}

	// removing will copy file and make that a ghost.
	// i.e. we need to re-lookup it:
	ghost, err := lkr.LookupGhost(file.Path())
	if err != nil {
		t.Fatalf("Failed to lookup ghost at %s: %v", file.Path(), err)
	}

	return &historySetup{
		commits: []*n.Commit{status, c3, c2, c1},
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
		head: status,
		node: ghost,
	}
}

func setupHistoryMoved(t *testing.T, lkr *c.Linker) *historySetup {
	_, c1 := c.MustTouchAndCommit(t, lkr, "/x.png", 1)
	file, c2 := c.MustTouchAndCommit(t, lkr, "/x.png", 2)
	c.MustMove(t, lkr, file, "/y.png")
	c3 := c.MustCommit(t, lkr, "post-move")

	return &historySetup{
		commits: []*n.Commit{c3, c2, c1},
		paths: []string{
			"/y.png",
			"/x.png",
			"/x.png",
		},
		changes: []ChangeType{
			ChangeTypeMove,
			ChangeTypeModify,
			ChangeTypeAdd,
		},
		head: c3,
		node: file,
	}
}

func setupHistoryMoveStaging(t *testing.T, lkr *c.Linker) *historySetup {
	c.MustTouch(t, lkr, "/x.png", 1)
	_, c1 := c.MustTouchAndCommit(t, lkr, "/x.png", 1)
	file, c2 := c.MustTouchAndCommit(t, lkr, "/x.png", 2)
	c.MustMove(t, lkr, file, "/y.png")

	status, err := lkr.Status()
	if err != nil {
		t.Fatalf("Failed to retrieve status: %v", err)
	}

	return &historySetup{
		commits: []*n.Commit{status, c2, c1},
		paths: []string{
			"/y.png",
			"/x.png",
			"/x.png",
		},
		changes: []ChangeType{
			ChangeTypeMove,
			ChangeTypeModify,
			ChangeTypeAdd,
		},
		head: status,
		node: file,
	}
}

func setupMoveInitial(t *testing.T, lkr *c.Linker) *historySetup {
	file := c.MustTouch(t, lkr, "/x.png", 1)
	c.MustMove(t, lkr, file, "/y.png")

	status, err := lkr.Status()
	if err != nil {
		t.Fatalf("Failed to retrieve status: %v", err)
	}

	// Should act like the node was added as "y.png",
	// even though it was moved.
	return &historySetup{
		commits: []*n.Commit{status},
		paths: []string{
			"/y.png",
		},
		changes: []ChangeType{
			ChangeTypeAdd,
		},
		head: status,
		node: file,
	}
}

func setupHistoryMoveAndModify(t *testing.T, lkr *c.Linker) *historySetup {
	_, c1 := c.MustTouchAndCommit(t, lkr, "/x.png", 1)
	file, c2 := c.MustTouchAndCommit(t, lkr, "/x.png", 2)

	newFile := c.MustMove(t, lkr, file, "/y.png")
	c.MustModify(t, lkr, newFile.(*n.File), 42)
	c3 := c.MustCommit(t, lkr, "post-move-modify")

	return &historySetup{
		commits: []*n.Commit{c3, c2, c1},
		paths: []string{
			"/y.png",
			"/x.png",
			"/x.png",
		},
		changes: []ChangeType{
			ChangeTypeModify | ChangeTypeMove,
			ChangeTypeModify,
			ChangeTypeAdd,
		},
		head: c3,
		node: file,
	}
}

func setupHistoryMoveAndModifyStage(t *testing.T, lkr *c.Linker) *historySetup {
	_, c1 := c.MustTouchAndCommit(t, lkr, "/x.png", 1)
	file, c2 := c.MustTouchAndCommit(t, lkr, "/x.png", 2)
	newFile := c.MustMove(t, lkr, file, "/y.png")
	c.MustModify(t, lkr, newFile.(*n.File), 42)

	status, err := lkr.Status()
	if err != nil {
		t.Fatalf("Failed to retrieve status: %v", err)
	}

	return &historySetup{
		commits: []*n.Commit{status, c2, c1},
		paths: []string{
			"/y.png",
			"/x.png",
			"/x.png",
		},
		changes: []ChangeType{
			ChangeTypeModify | ChangeTypeMove,
			ChangeTypeModify,
			ChangeTypeAdd,
		},
		head: status,
		node: file,
	}
}

func setupHistoryRemoveReadd(t *testing.T, lkr *c.Linker) *historySetup {
	_, c1 := c.MustTouchAndCommit(t, lkr, "/x.png", 1)
	file, c2 := c.MustTouchAndCommit(t, lkr, "/x.png", 2)
	c.MustRemove(t, lkr, file)
	c3 := c.MustCommit(t, lkr, "after remove")
	file, c4 := c.MustTouchAndCommit(t, lkr, "/x.png", 2)

	return &historySetup{
		commits: []*n.Commit{c4, c3, c2, c1},
		paths: []string{
			"/x.png",
			"/x.png",
			"/x.png",
			"/x.png",
		},
		changes: []ChangeType{
			ChangeTypeAdd,
			ChangeTypeRemove,
			ChangeTypeModify,
			ChangeTypeAdd,
		},
		head: c4,
		node: file,
	}
}

func setupHistoryRemoveReaddModify(t *testing.T, lkr *c.Linker) *historySetup {
	_, c1 := c.MustTouchAndCommit(t, lkr, "/x.png", 1)
	file, c2 := c.MustTouchAndCommit(t, lkr, "/x.png", 2)
	c.MustRemove(t, lkr, file)
	c3 := c.MustCommit(t, lkr, "after remove")
	file, c4 := c.MustTouchAndCommit(t, lkr, "/x.png", 255)

	return &historySetup{
		commits: []*n.Commit{c4, c3, c2, c1},
		paths: []string{
			"/x.png",
			"/x.png",
			"/x.png",
			"/x.png",
		},
		changes: []ChangeType{
			ChangeTypeAdd | ChangeTypeModify,
			ChangeTypeRemove,
			ChangeTypeModify,
			ChangeTypeAdd,
		},
		head: c4,
		node: file,
	}
}

func setupHistoryRemoveReaddNoModify(t *testing.T, lkr *c.Linker) *historySetup {
	_, c1 := c.MustTouchAndCommit(t, lkr, "/x.png", 1)
	file, c2 := c.MustTouchAndCommit(t, lkr, "/x.png", 2)
	c.MustRemove(t, lkr, file)
	c3 := c.MustCommit(t, lkr, "after remove")
	file, c4 := c.MustTouchAndCommit(t, lkr, "/x.png", 2)

	return &historySetup{
		commits: []*n.Commit{c4, c3, c2, c1},
		paths: []string{
			"/x.png",
			"/x.png",
			"/x.png",
			"/x.png",
		},
		changes: []ChangeType{
			ChangeTypeAdd,
			ChangeTypeRemove,
			ChangeTypeModify,
			ChangeTypeAdd,
		},
		head: c4,
		node: file,
	}
}

func setupHistoryMoveCircle(t *testing.T, lkr *c.Linker) *historySetup {
	_, c1 := c.MustTouchAndCommit(t, lkr, "/x.png", 1)
	file, c2 := c.MustTouchAndCommit(t, lkr, "/x.png", 2)
	newFile := c.MustMove(t, lkr, file, "/y.png")
	c3 := c.MustCommit(t, lkr, "move to y.png")
	newOldFile := c.MustMove(t, lkr, newFile, "/x.png")
	c4 := c.MustCommit(t, lkr, "move back to x.png")

	return &historySetup{
		commits: []*n.Commit{c4, c3, c2, c1},
		paths: []string{
			"/x.png",
			"/y.png",
			"/x.png",
			"/x.png",
		},
		changes: []ChangeType{
			ChangeTypeMove,
			ChangeTypeMove,
			ChangeTypeModify,
			ChangeTypeAdd,
		},
		head: c4,
		node: newOldFile,
	}
}

func setupHistoryMoveSamePlaceLeft(t *testing.T, lkr *c.Linker) *historySetup {
	x := c.MustTouch(t, lkr, "/x", 1)
	y := c.MustTouch(t, lkr, "/y", 1)
	c1 := c.MustCommit(t, lkr, "pre-move")

	c.MustMove(t, lkr, x, "/z")
	c.MustMove(t, lkr, y, "/z")
	c2 := c.MustCommit(t, lkr, "post-move")

	xGhost, err := lkr.LookupGhost("/x")
	require.Nil(t, err)

	return &historySetup{
		commits: []*n.Commit{c2, c1},
		paths: []string{
			"/x",
			"/x",
		},
		changes: []ChangeType{
			// This file was removed, since the destination "z"
			// was overwritten by "y" and thus we may not count it
			// as moved.
			ChangeTypeRemove,
			ChangeTypeAdd,
		},
		head: c2,
		node: xGhost,
	}
}

func setupHistoryTypeChange(t *testing.T, lkr *c.Linker) *historySetup {
	x := c.MustTouch(t, lkr, "/x", 1)
	c.MustCommit(t, lkr, "added")
	c.MustRemove(t, lkr, x)
	c.MustCommit(t, lkr, "removed")
	dir := c.MustMkdir(t, lkr, "/x")
	c3 := c.MustCommit(t, lkr, "mkdir")

	return &historySetup{
		commits: []*n.Commit{c3},
		paths: []string{
			"/x",
		},
		changes: []ChangeType{
			// This file was removed, since the destination "z"
			// was overwritten by "y" and thus we may not count it
			// as moved.
			ChangeTypeAdd,
		},
		head: c3,
		node: dir,
	}
}

func setupHistoryMoveSamePlaceRight(t *testing.T, lkr *c.Linker) *historySetup {
	x := c.MustTouch(t, lkr, "/x", 1)
	y := c.MustTouch(t, lkr, "/y", 1)
	c1 := c.MustCommit(t, lkr, "pre-move")

	c.MustMove(t, lkr, x, "/z")
	c.MustMove(t, lkr, y, "/z")
	c2 := c.MustCommit(t, lkr, "post-move")

	yGhost, err := lkr.LookupGhost("/y")
	require.Nil(t, err)

	return &historySetup{
		commits: []*n.Commit{c2, c1},
		paths: []string{
			"/y",
			"/y",
		},
		changes: []ChangeType{
			ChangeTypeMove | ChangeTypeRemove,
			ChangeTypeAdd,
		},
		head: c2,
		node: yGhost,
	}
}

func setupHistoryMoveAndReaddFromMoved(t *testing.T, lkr *c.Linker) *historySetup {
	_, c1 := c.MustTouchAndCommit(t, lkr, "/x.png", 1)
	file, c2 := c.MustTouchAndCommit(t, lkr, "/x.png", 2)

	newFile := c.MustMove(t, lkr, file, "/y.png")
	_, c3 := c.MustTouchAndCommit(t, lkr, "/x.png", 23)

	return &historySetup{
		commits: []*n.Commit{c3, c2, c1},
		paths: []string{
			"/y.png",
			"/x.png",
			"/x.png",
		},
		changes: []ChangeType{
			ChangeTypeMove,
			ChangeTypeModify,
			ChangeTypeAdd,
		},
		head: c3,
		node: newFile,
	}
}

func setupHistoryMultipleMovesPerCommit(t *testing.T, lkr *c.Linker) *historySetup {
	// Check if we can track multiple moves per commit:
	fileX, c1 := c.MustTouchAndCommit(t, lkr, "/x.png", 1)
	fileY := c.MustMove(t, lkr, fileX, "/y.png")
	c.MustMove(t, lkr, fileY, "/z.png")

	fileZNew, err := c.Stage(lkr, "/z.png", h.TestDummy(t, 2), h.TestDummy(t, 2), uint64(2), nil, time.Now())
	require.Nil(t, err)

	c2 := c.MustCommit(t, lkr, "Moved around")

	return &historySetup{
		commits: []*n.Commit{c2, c1},
		paths: []string{
			"/z.png",
			"/x.png",
		},
		changes: []ChangeType{
			ChangeTypeMove | ChangeTypeModify,
			ChangeTypeAdd,
		},
		head: c2,
		node: fileZNew,
	}
}

func setupHistoryMultipleMovesInStage(t *testing.T, lkr *c.Linker) *historySetup {
	// Check if we can track multiple moves per commit:
	fileX, c1 := c.MustTouchAndCommit(t, lkr, "/x.png", 1)
	fileY := c.MustMove(t, lkr, fileX, "/y.png")
	fileZ := c.MustMove(t, lkr, fileY, "/z.png")

	status, err := lkr.Status()
	if err != nil {
		t.Fatalf("Failed to acquire status: %v", err)
	}

	return &historySetup{
		commits: []*n.Commit{status, c1},
		paths: []string{
			"/z.png",
			"/x.png",
		},
		changes: []ChangeType{
			ChangeTypeMove,
			ChangeTypeAdd,
		},
		head: status,
		node: fileZ,
	}
}

func setupHistoryMoveAndReaddFromAdded(t *testing.T, lkr *c.Linker) *historySetup {
	_, c1 := c.MustTouchAndCommit(t, lkr, "/x.png", 1)
	file, c2 := c.MustTouchAndCommit(t, lkr, "/x.png", 2)

	c.MustMove(t, lkr, file, "/y.png")
	c3 := c.MustCommit(t, lkr, "move to y.png")
	readdedFile, c4 := c.MustTouchAndCommit(t, lkr, "/x.png", 23)

	return &historySetup{
		commits: []*n.Commit{c4, c3, c2, c1},
		paths: []string{
			"/x.png",
			"/x.png",
			"/x.png",
			"/x.png",
		},

		changes: []ChangeType{
			ChangeTypeAdd | ChangeTypeModify,
			ChangeTypeMove | ChangeTypeRemove,
			ChangeTypeModify,
			ChangeTypeAdd,
		},
		head: c4,
		node: readdedFile,
	}
}

func setupMoveDirectoryWithChild(t *testing.T, lkr *c.Linker) *historySetup {
	dir := c.MustMkdir(t, lkr, "/sub")
	_, c1 := c.MustTouchAndCommit(t, lkr, "/sub/x.png", 1)
	file, c2 := c.MustTouchAndCommit(t, lkr, "/sub/x.png", 2)

	c.MustMove(t, lkr, dir, "/moved-sub")
	c3 := c.MustCommit(t, lkr, "moved")

	status, err := lkr.Status()
	if err != nil {
		t.Fatalf("Failed to get status: %v", err)
	}

	return &historySetup{
		commits: []*n.Commit{status, c3, c2, c1},
		paths: []string{
			"/moved-sub/x.png",
			"/moved-sub/x.png",
			"/sub/x.png",
			"/sub/x.png",
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

func setupDirectoryHistory(t *testing.T, lkr *c.Linker) *historySetup {
	dir := c.MustMkdir(t, lkr, "/src")
	_, c1 := c.MustTouchAndCommit(t, lkr, "/src/x.png", 1)
	_, c2 := c.MustTouchAndCommit(t, lkr, "/src/x.png", 2)

	newDir := c.MustMove(t, lkr, dir, "/dst")
	c3 := c.MustCommit(t, lkr, "move")

	status, err := lkr.Status()
	if err != nil {
		t.Fatalf("Failed to get status: %v", err)
	}

	return &historySetup{
		commits: []*n.Commit{status, c3, c2, c1},
		paths: []string{
			"/dst",
			"/dst",
			"/src",
			"/src",
		},
		changes: []ChangeType{
			ChangeTypeNone,
			ChangeTypeMove,
			ChangeTypeModify,
			ChangeTypeAdd,
		},
		head: status,
		node: newDir,
	}
}

func setupGhostHistory(t *testing.T, lkr *c.Linker) *historySetup {
	_, c1 := c.MustTouchAndCommit(t, lkr, "/x.png", 1)
	file, c2 := c.MustTouchAndCommit(t, lkr, "/x.png", 2)
	c.MustMove(t, lkr, file, "/y.png")
	c3 := c.MustCommit(t, lkr, "move")

	ghost, err := lkr.LookupGhost("/x.png")
	require.Nil(t, err)

	status, err := lkr.Status()
	if err != nil {
		t.Fatalf("Failed to get status: %v", err)
	}

	return &historySetup{
		commits: []*n.Commit{status, c3, c2, c1},
		paths: []string{
			"/x.png",
			"/x.png",
			"/x.png",
			"/x.png",
		},
		changes: []ChangeType{
			ChangeTypeNone,
			// The "ChangeTypeMove" here is a hint that
			// this ghost was part of a move.
			ChangeTypeMove | ChangeTypeRemove,
			ChangeTypeModify,
			ChangeTypeAdd,
		},
		head: status,
		node: ghost,
	}
}

func setupEdgeRoot(t *testing.T, lkr *c.Linker) *historySetup {
	init, err := lkr.Head()
	if err != nil {
		t.Fatalf("could not get head: %v", err)
	}

	_, c1 := c.MustTouchAndCommit(t, lkr, "/x.png", 1)
	_, c2 := c.MustTouchAndCommit(t, lkr, "/x.png", 2)
	_, c3 := c.MustTouchAndCommit(t, lkr, "/x.png", 3)

	status, err := lkr.Status()
	if err != nil {
		t.Fatalf("failed to get status: %v", err)
	}

	root, err := lkr.Root()
	if err != nil {
		t.Fatalf("failed to retrieve root: %v", err)
	}

	return &historySetup{
		commits: []*n.Commit{status, c3, c2, c1, init},
		paths: []string{
			"/",
			"/",
			"/",
			"/",
			"/",
		},
		changes: []ChangeType{
			ChangeTypeNone,
			ChangeTypeModify,
			ChangeTypeModify,
			ChangeTypeModify,
			ChangeTypeAdd,
		},
		head: status,
		node: root,
	}
}

type setupFunc func(t *testing.T, lkr *c.Linker) *historySetup

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
			name:  "holes",
			setup: setupHistoryBasicHole,
		}, {
			name:  "remove-it",
			setup: setupHistoryRemoved,
		}, {
			name:  "remove-readd-simple",
			setup: setupHistoryRemoveReadd,
		}, {
			name:  "remove-immedidately",
			setup: setupHistoryRemoveImmediately,
		}, {
			name:  "remove-readd-modify",
			setup: setupHistoryRemoveReaddModify,
		}, {
			name:  "remove-readd-no-modify",
			setup: setupHistoryRemoveReaddNoModify,
		}, {
			name:  "move-once",
			setup: setupHistoryMoved,
		}, {
			name:  "move-multiple-per-commit",
			setup: setupHistoryMultipleMovesPerCommit,
		}, {
			name:  "move-multiple-per-stage",
			setup: setupHistoryMultipleMovesInStage,
		}, {
			name:  "move-once-stage",
			setup: setupHistoryMoveStaging,
		}, {
			name:  "move-initial",
			setup: setupMoveInitial,
		}, {
			name:  "move-modify",
			setup: setupHistoryMoveAndModify,
		}, {
			name:  "move-to-same-place-left",
			setup: setupHistoryMoveSamePlaceLeft,
		}, {
			name:  "move-to-same-place-right",
			setup: setupHistoryMoveSamePlaceRight,
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
		}, {
			name:  "move-directory-with-child",
			setup: setupMoveDirectoryWithChild,
		}, {
			name:  "directory-simple",
			setup: setupDirectoryHistory,
		}, {
			name:  "ghost-simple",
			setup: setupGhostHistory,
		}, {
			name:  "edge-root",
			setup: setupEdgeRoot,
		}, {
			name:  "edge-type-change",
			setup: setupHistoryTypeChange,
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
func testHistoryRunner(t *testing.T, lkr *c.Linker, setup *historySetup) {
	idx := 0
	walker := NewHistoryWalker(lkr, setup.head, setup.node)
	for walker.Next() {
		state := walker.State()
		// fmt.Println("TYPE", state.Mask)
		// fmt.Println("HEAD", state.Head)
		// fmt.Println("NEXT", state.Next)
		// fmt.Println("===")

		if idx >= len(setup.paths) {
			t.Fatalf("more history entries than expected")
		}

		if setup.paths[idx] != state.Curr.Path() {
			t.Fatalf(
				"Wrong path at index `%d`: %s (want: %s)",
				idx+1, state.Curr.Path(), setup.paths[idx],
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

		if !setup.commits[idx].TreeHash().Equal(state.Head.TreeHash()) {
			t.Fatalf("Hash in commit differs")
		}

		idx++
	}

	if err := walker.Err(); err != nil {
		t.Fatalf(
			"walker failed at index (%d/%d): %v",
			idx+1,
			len(setup.commits),
			err,
		)
	}
}

// Test the History() utility based on HistoryWalker.
func TestHistoryUtil(t *testing.T) {
	c.WithDummyLinker(t, func(lkr *c.Linker) {
		c1File, c1 := c.MustTouchAndCommit(t, lkr, "/x.png", 1)
		c1File = c1File.Copy(c1File.Inode()).(*n.File)

		c2File, c2 := c.MustTouchAndCommit(t, lkr, "/x.png", 2)
		c2File = c2File.Copy(c2File.Inode()).(*n.File)

		c3File := c.MustMove(t, lkr, c2File.Copy(c2File.Inode()), "/y.png")
		c3File = c3File.Copy(c3File.Inode()).(*n.File)
		c3 := c.MustCommit(t, lkr, "move to y.png")

		c4File, c4 := c.MustTouchAndCommit(t, lkr, "/y.png", 23)
		c4File = c4File.Copy(c4File.Inode()).(*n.File)

		states, err := History(lkr, c4File, c4, nil)
		if err != nil {
			t.Fatalf("History without stop commit failed: %v", err)
		}

		expected := []*Change{
			{
				Head: c4,
				Curr: c4File,
				Mask: ChangeTypeModify,
			}, {
				Head: c3,
				Curr: c3File,
				Mask: ChangeTypeMove,
			}, {
				Head: c2,
				Curr: c2File,
				Mask: ChangeTypeModify,
			}, {
				Head: c1,
				Curr: c1File,
				Mask: ChangeTypeAdd,
			},
		}

		for idx, state := range states {
			expect := expected[idx]
			require.Equal(t, state.Mask, expect.Mask, "Mask differs")
			require.Equal(t, state.Head, expect.Head, "Head differs")
			require.Equal(t, state.Curr, expect.Curr, "Curr differs")
		}
	})
}

func TestHistoryWithNoParent(t *testing.T) {
	c.WithDummyKv(t, func(kv db.Database) {
		lkr := c.NewLinker(kv)
		lkr.SetOwner("alice")

		file, head := c.MustTouchAndCommit(t, lkr, "/x", 1)

		hist, err := History(lkr, file, head, nil)
		require.Nil(t, err)
		require.Len(t, hist, 1)
		require.Equal(t, hist[0].Mask, ChangeTypeAdd)
	})
}

// Regression test:
// Directories loose move history operation
// when restarting the daemon in between.
func TestHistoryMovedDirsWithReloadedLinker(t *testing.T) {
	validateHist := func(hist []*Change) {
		require.Len(t, hist, 2)
		require.Equal(t, hist[0].Mask, ChangeTypeMove)
		require.Equal(t, hist[1].Mask, ChangeTypeAdd)
	}

	c.WithReloadingLinker(t, func(lkr *c.Linker) {
		childDir := c.MustMkdir(t, lkr, "/child")
		c.MustCommit(t, lkr, "created")
		movedDir := c.MustMove(t, lkr, childDir, "/moved_child")

		status, err := lkr.Status()
		require.Nil(t, err)

		hist, err := History(lkr, movedDir, status, nil)
		require.Nil(t, err)

		validateHist(hist)
	}, func(lkr *c.Linker) {
		status, err := lkr.Status()
		require.Nil(t, err)

		childDir, err := lkr.LookupDirectory("/moved_child")
		require.Nil(t, err)

		hist, err := History(lkr, childDir, status, nil)
		require.Nil(t, err)

		validateHist(hist)
	})
}

// Regression test:
func TestHistoryOfMovedNestedDir(t *testing.T) {
	c.WithDummyLinker(t, func(lkr *c.Linker) {
		c.MustMkdir(t, lkr, "/src/core")
		c.MustTouch(t, lkr, "/src/core/linker.go", 3)
		c.MustCommit(t, lkr, "added")

		c.MustMove(t, lkr, c.MustLookupDirectory(t, lkr, "/src"), "/dst")
		c.MustCommit(t, lkr, "move")

		status, err := lkr.Status()
		require.Nil(t, err)

		// This raised an error before, since "/dst" was missing
		// in the "added" commit.
		hist, err := History(lkr, c.MustLookupDirectory(t, lkr, "/dst/core"), status, nil)
		require.Nil(t, err)

		require.Equal(t, "/dst/core", hist[0].Curr.Path())
		require.Equal(t, ChangeTypeNone, hist[0].Mask)
		require.Equal(t, "/dst/core", hist[1].Curr.Path())
		require.Equal(t, ChangeTypeMove, hist[1].Mask)
		require.Equal(t, "/src/core", hist[1].WasPreviouslyAt)
		require.Equal(t, "/src/core", hist[2].Curr.Path())
		require.Equal(t, ChangeTypeAdd, hist[2].Mask)

		file, err := lkr.LookupModNode("/dst/core/linker.go")
		require.Nil(t, err)

		hist, err = History(lkr, file, status, nil)
		require.Nil(t, err)

		require.Equal(t, "/dst/core/linker.go", hist[0].Curr.Path())
		require.Equal(t, ChangeTypeNone, hist[0].Mask)
		require.Equal(t, "/dst/core/linker.go", hist[1].Curr.Path())
		require.Equal(t, ChangeTypeMove, hist[1].Mask)
		require.Equal(t, "/src/core/linker.go", hist[1].WasPreviouslyAt)
		require.Equal(t, "/src/core/linker.go", hist[2].Curr.Path())
		require.Equal(t, ChangeTypeAdd, hist[2].Mask)
	})
}
