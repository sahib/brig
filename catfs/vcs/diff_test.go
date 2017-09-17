package catfs

import (
	"fmt"
	"testing"

	log "github.com/Sirupsen/logrus"
	c "github.com/disorganizer/brig/catfs/core"
	n "github.com/disorganizer/brig/catfs/nodes"
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

// TODO: Test history for multiple moves in one commit and several commits.

////////////////////////////////////
// TEST FOR DIFFER IMPLEMENTATION //
////////////////////////////////////

func mapperSetupBasicSame(t *testing.T, lkrSrc, lkrDst *c.Linker) []MapPair {
	c.MustTouchAndCommit(t, lkrSrc, "/x.png", 23)
	c.MustTouchAndCommit(t, lkrDst, "/x.png", 23)
	return []MapPair{}
}

func mapperSetupBasicDiff(t *testing.T, lkrSrc, lkrDst *c.Linker) []MapPair {
	srcFile, _ := c.MustTouchAndCommit(t, lkrSrc, "/x.png", 23)
	dstFile, _ := c.MustTouchAndCommit(t, lkrDst, "/x.png", 42)
	return []MapPair{
		{
			Src:          srcFile,
			Dst:          dstFile,
			TypeMismatch: false,
		},
	}
}

func mapperSetupBasicSrcTypeMismatch(t *testing.T, lkrSrc, lkrDst *c.Linker) []MapPair {
	srcDir := c.MustMkdir(t, lkrSrc, "/x")
	c.MustCommit(t, lkrSrc, "add dir")

	dstFile, _ := c.MustTouchAndCommit(t, lkrDst, "/x", 42)

	return []MapPair{
		{
			Src:          srcDir,
			Dst:          dstFile,
			TypeMismatch: true,
		},
	}
}

func mapperSetupBasicDstTypeMismatch(t *testing.T, lkrSrc, lkrDst *c.Linker) []MapPair {
	srcFile, _ := c.MustTouchAndCommit(t, lkrSrc, "/x", 42)
	dstDir := c.MustMkdir(t, lkrDst, "/x")
	c.MustCommit(t, lkrDst, "add dir")

	return []MapPair{
		{
			Src:          srcFile,
			Dst:          dstDir,
			TypeMismatch: true,
		},
	}
}

func mapperSetupBasicSrcAddFile(t *testing.T, lkrSrc, lkrDst *c.Linker) []MapPair {
	srcFile, _ := c.MustTouchAndCommit(t, lkrSrc, "/x.png", 42)

	return []MapPair{
		{
			Src:          srcFile,
			Dst:          nil,
			TypeMismatch: false,
		},
	}
}

func mapperSetupBasicDstAddFile(t *testing.T, lkrSrc, lkrDst *c.Linker) []MapPair {
	c.MustTouchAndCommit(t, lkrDst, "/x.png", 42)
	return []MapPair{}
}

func mapperSetupBasicSrcAddDir(t *testing.T, lkrSrc, lkrDst *c.Linker) []MapPair {
	srcDir := c.MustMkdir(t, lkrSrc, "/x")
	c.MustCommit(t, lkrSrc, "add dir")

	return []MapPair{
		{
			Src:          srcDir,
			Dst:          nil,
			TypeMismatch: false,
		},
	}
}

func mapperSetupBasicDstAddDir(t *testing.T, lkrSrc, lkrDst *c.Linker) []MapPair {
	c.MustMkdir(t, lkrDst, "/x")
	return []MapPair{}
}

func mapperSetupSrcMoveFile(t *testing.T, lkrSrc, lkrDst *c.Linker) []MapPair {
	dstFile, _ := c.MustTouchAndCommit(t, lkrDst, "/x.png", 42)
	srcFileOld, _ := c.MustTouchAndCommit(t, lkrSrc, "/x.png", 23)
	srcFile := c.MustMove(t, lkrSrc, srcFileOld, "/y.png")
	c.MustCommit(t, lkrSrc, "I like to move it")

	return []MapPair{
		{
			Src:          srcFile,
			Dst:          dstFile,
			TypeMismatch: false,
		},
	}
}

func mapperSetupDstMoveFile(t *testing.T, lkrSrc, lkrDst *c.Linker) []MapPair {
	srcFile, _ := c.MustTouchAndCommit(t, lkrSrc, "/x.png", 42)
	dstFileOld, _ := c.MustTouchAndCommit(t, lkrDst, "/x.png", 23)
	dstFile := c.MustMove(t, lkrDst, dstFileOld, "/y.png")
	c.MustCommit(t, lkrDst, "I like to move it, move it")

	return []MapPair{
		{
			Src:          srcFile,
			Dst:          dstFile,
			TypeMismatch: false,
		},
	}
}

func mapperSetupDstMoveDirEmpty(t *testing.T, lkrSrc, lkrDst *c.Linker) []MapPair {
	c.MustMkdir(t, lkrSrc, "/x")
	c.MustCommit(t, lkrSrc, "Create src dir")

	dstDirOld := c.MustMkdir(t, lkrDst, "/x")
	c.MustMove(t, lkrDst, dstDirOld, "/y")
	c.MustCommit(t, lkrDst, "I like to move it, move it")

	return []MapPair{}
}

func mapperSetupDstMoveDir(t *testing.T, lkrSrc, lkrDst *c.Linker) []MapPair {
	c.MustMkdir(t, lkrSrc, "/x")
	srcFile := c.MustTouch(t, lkrSrc, "/x/a.png", 42)
	c.MustCommit(t, lkrSrc, "Create src dir")

	// TODO: There is a bug (likely in move()) when switching move and touch
	fmt.Println("DST MKDIR")
	dstDirOld := c.MustMkdir(t, lkrDst, "/x")
	fmt.Println("DST TOUCH")
	dstFile := c.MustTouch(t, lkrDst, "/x/a.png", 23)
	fmt.Println("DST MOVE")
	c.MustMove(t, lkrDst, dstDirOld, "/y")
	fmt.Println("DST COMMIT")
	c.MustCommit(t, lkrDst, "I like to move it, move it")
	fmt.Println("DST DONE")

	return []MapPair{
		{
			Src:          srcFile,
			Dst:          dstFile,
			TypeMismatch: false,
		},
	}
}

func mapperSetupSrcMoveDir(t *testing.T, lkrSrc, lkrDst *c.Linker) []MapPair {
	srcDirOld := c.MustMkdir(t, lkrSrc, "/x")
	c.MustMove(t, lkrSrc, srcDirOld, "/y")
	srcFile := c.MustTouch(t, lkrSrc, "/y/a.png", 23)
	c.MustCommit(t, lkrSrc, "I like to move it, move it")

	c.MustMkdir(t, lkrDst, "/x")
	dstFile := c.MustTouch(t, lkrDst, "/x/a.png", 42)
	c.MustCommit(t, lkrDst, "Create dst dir")

	return []MapPair{
		{
			Src:          srcFile,
			Dst:          dstFile,
			TypeMismatch: false,
		},
	}
}

func mapperSetupSrcMoveWithExisting(t *testing.T, lkrSrc, lkrDst *c.Linker) []MapPair {
	srcDirOld := c.MustMkdir(t, lkrSrc, "/x")
	c.MustMove(t, lkrSrc, srcDirOld, "/y")
	srcFile := c.MustTouch(t, lkrSrc, "/y/a.png", 23)
	c.MustCommit(t, lkrSrc, "I like to move it, move it")

	// Additionally create an existing file that lives in the place
	// of the moved file. Mapper should favour existing files:
	c.MustMkdir(t, lkrDst, "/x")
	c.MustMkdir(t, lkrDst, "/y")
	c.MustTouch(t, lkrDst, "/x/a.png", 42)
	dstFile := c.MustTouch(t, lkrDst, "/y/a.png", 42)
	c.MustCommit(t, lkrDst, "Create src dir")

	return []MapPair{
		{
			Src:          srcFile,
			Dst:          dstFile,
			TypeMismatch: false,
		},
	}
}

func mapperSetupDstMoveWithExisting(t *testing.T, lkrSrc, lkrDst *c.Linker) []MapPair {
	srcDir := c.MustMkdir(t, lkrSrc, "/x")
	c.MustMkdir(t, lkrSrc, "/y")
	c.MustTouch(t, lkrSrc, "/x/a.png", 42)
	srcFile := c.MustTouch(t, lkrSrc, "/y/a.png", 42)
	c.MustCommit(t, lkrSrc, "Create src dir")

	dstDirOld := c.MustMkdir(t, lkrDst, "/x")
	c.MustMove(t, lkrDst, dstDirOld, "/y")
	dstFile := c.MustTouch(t, lkrDst, "/y/a.png", 23)
	c.MustCommit(t, lkrDst, "I like to move it, move it")

	return []MapPair{
		{
			Src:          srcDir,
			Dst:          nil,
			TypeMismatch: false,
		}, {
			Src:          srcFile,
			Dst:          dstFile,
			TypeMismatch: false,
		},
	}
}

func mapperSetupNested(t *testing.T, lkrSrc, lkrDst *c.Linker) []MapPair {
	srcX, _ := c.MustTouchAndCommit(t, lkrSrc, "/common/a/b/c/x", 42)
	srcY, _ := c.MustTouchAndCommit(t, lkrSrc, "/common/a/b/c/y", 23)
	srcZ, _ := c.MustTouchAndCommit(t, lkrSrc, "/src-only/z", 23)

	dstX, _ := c.MustTouchAndCommit(t, lkrDst, "/common/a/b/c/x", 43)
	dstY, _ := c.MustTouchAndCommit(t, lkrDst, "/common/a/b/c/y", 24)
	c.MustTouchAndCommit(t, lkrDst, "/dst-only/z", 23)

	srcZParent, err := n.ParentDirectory(lkrSrc, srcZ)
	if err != nil {
		t.Fatalf("setup failed to get parent dir: %v", err)
	}

	return []MapPair{
		{
			Src:          srcX,
			Dst:          dstX,
			TypeMismatch: false,
		}, {
			Src:          srcY,
			Dst:          dstY,
			TypeMismatch: false,
		}, {
			Src:          srcZParent,
			Dst:          nil,
			TypeMismatch: false,
		},
	}
}

func mapperSetupSrcRemove(t *testing.T, lkrSrc, lkrDst *c.Linker) []MapPair {
	srcFile := c.MustTouch(t, lkrSrc, "/x.png", 23)
	c.MustCommit(t, lkrSrc, "src: Touched /x.png")
	c.MustRemove(t, lkrSrc, srcFile)
	c.MustCommit(t, lkrSrc, "src: Removed /x.png")

	dstFile := c.MustTouch(t, lkrDst, "/x.png", 23)
	c.MustCommit(t, lkrDst, "dst: Touched /x.png")

	return []MapPair{
		{
			Src:          nil,
			Dst:          dstFile,
			TypeMismatch: false,
		},
	}
}

func mapperSetupDstRemove(t *testing.T, lkrSrc, lkrDst *c.Linker) []MapPair {
	srcFile := c.MustTouch(t, lkrSrc, "/x.png", 23)
	c.MustCommit(t, lkrSrc, "dst: Touched /x.png")

	dstFile := c.MustTouch(t, lkrDst, "/x.png", 23)
	c.MustCommit(t, lkrDst, "src: Touched /x.png")
	c.MustRemove(t, lkrDst, dstFile)
	c.MustCommit(t, lkrDst, "src: Removed /x.png")

	return []MapPair{
		{
			Src:          srcFile,
			Dst:          nil,
			TypeMismatch: false,
		},
	}
}

func mapperSetupMoveOnBothSides(t *testing.T, lkrSrc, lkrDst *c.Linker) []MapPair {
	srcFile := c.MustTouch(t, lkrSrc, "/x", 23)
	c.MustCommit(t, lkrSrc, "src: touched /x")
	srcFileMoved := c.MustMove(t, lkrSrc, srcFile, "/y")
	c.MustCommit(t, lkrSrc, "src: /x moved to /y")

	dstFile := c.MustTouch(t, lkrDst, "/x", 42)
	c.MustCommit(t, lkrDst, "dst: touched /x")
	dstFileMoved := c.MustMove(t, lkrDst, dstFile, "/z")
	c.MustCommit(t, lkrDst, "dst: /x moved to /z")

	return []MapPair{
		{
			Src:          srcFileMoved,
			Dst:          dstFileMoved,
			TypeMismatch: false,
		},
	}
}

func TestMapper(t *testing.T) {
	tcs := []struct {
		name  string
		setup func(t *testing.T, lkrSrc, lkrDst *c.Linker) []MapPair
	}{
		{
			name:  "basic-same",
			setup: mapperSetupBasicSame,
		}, {
			name:  "basic-diff",
			setup: mapperSetupBasicDiff,
		}, {
			name:  "basic-src-add-file",
			setup: mapperSetupBasicSrcAddFile,
		}, {
			name:  "basic-dst-add-file",
			setup: mapperSetupBasicDstAddFile,
		}, {
			name:  "basic-src-add-dir",
			setup: mapperSetupBasicSrcAddDir,
		}, {
			name:  "basic-dst-add-dir",
			setup: mapperSetupBasicDstAddDir,
		}, {
			name:  "basic-src-type-mismatch",
			setup: mapperSetupBasicSrcTypeMismatch,
		}, {
			name:  "basic-dst-type-mismatch",
			setup: mapperSetupBasicDstTypeMismatch,
		}, {
			name:  "basic-nested",
			setup: mapperSetupNested,
		}, {
			name:  "remove-src",
			setup: mapperSetupSrcRemove,
		}, {
			name:  "remove-dst",
			setup: mapperSetupDstRemove,
		}, {
			name:  "move-simple-src-file",
			setup: mapperSetupSrcMoveFile,
		}, {
			name:  "move-simple-dst-file",
			setup: mapperSetupDstMoveFile,
		}, {
			name:  "move-simple-dst-empty-dir",
			setup: mapperSetupDstMoveDirEmpty,
		}, {
			name:  "move-simple-src-dir",
			setup: mapperSetupSrcMoveDir,
		}, {
			name:  "move-simple-dst-dir",
			setup: mapperSetupDstMoveDir,
		}, {
			name:  "move-simple-src-dir-with-existing",
			setup: mapperSetupSrcMoveWithExisting,
		}, {
			name:  "move-simple-dst-dir-with-existing",
			setup: mapperSetupDstMoveWithExisting,
		}, {
			name:  "move-on-both-sides",
			setup: mapperSetupMoveOnBothSides,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			c.WithLinkerPair(t, func(lkrSrc, lkrDst *c.Linker) {
				expect := tc.setup(t, lkrSrc, lkrDst)

				srcRoot, err := lkrSrc.Root()
				if err != nil {
					t.Fatalf("Failed to retrieve root: %v", err)
				}

				got := []MapPair{}
				diffFn := func(pair MapPair) error {
					got = append(got, pair)
					// if pair.Src != nil {
					// 	fmt.Println(".. ", pair.Src.Path())
					// }
					// if pair.Dst != nil {
					// 	fmt.Println("-> ", pair.Dst.Path())
					// }
					return nil
				}

				mapper := NewMapper(lkrSrc, lkrDst, srcRoot)
				if err := mapper.Map(diffFn); err != nil {
					t.Fatalf("mapping failed: %v", err)
				}

				if len(got) != len(expect) {
					t.Fatalf(
						"Got and expect length differ: %d vs %d",
						len(got), len(expect),
					)
				}

				for idx, gotPair := range got {
					expectPair := expect[idx]
					failMsg := fmt.Sprintf("Failed pair %d", idx+1)
					require.Equal(t, expectPair, gotPair, failMsg)
				}
			})
		})
	}
}
