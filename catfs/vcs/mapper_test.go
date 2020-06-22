package vcs

import (
	"fmt"
	"testing"

	c "github.com/sahib/brig/catfs/core"
	n "github.com/sahib/brig/catfs/nodes"
	"github.com/stretchr/testify/require"
)

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
	dstFile, _ := c.MustTouchAndCommit(t, lkrDst, "/x.png", 42)
	return []MapPair{
		{
			Src:          nil,
			Dst:          dstFile,
			TypeMismatch: false,
		},
	}
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

func mapperMoveNestedDir(t *testing.T, lkrSrc, lkrDst *c.Linker) []MapPair {
	c.MustMkdir(t, lkrSrc, "/old/sub/")
	c.MustMkdir(t, lkrDst, "/old/sub/")
	c.MustTouchAndCommit(t, lkrSrc, "/old/sub/x", 1)
	c.MustTouchAndCommit(t, lkrDst, "/old/sub/x", 1)

	srcDir := c.MustLookupDirectory(t, lkrSrc, "/old")
	dstDir := c.MustLookupDirectory(t, lkrDst, "/old")
	newDstDir := c.MustMove(t, lkrDst, dstDir, "/new")
	c.MustCommit(t, lkrDst, "moved")

	// Test for a special case here:
	// Directories that were moved, but still have identical files.
	return []MapPair{
		{
			Src:           srcDir,
			Dst:           newDstDir,
			SrcWasMoved:   true,
			TypeMismatch:  false,
			SrcWasRemoved: false,
		},
	}
}

func mapperSetupDstMoveDirEmpty(t *testing.T, lkrSrc, lkrDst *c.Linker) []MapPair {
	srcDir := c.MustMkdir(t, lkrSrc, "/x")
	c.MustCommit(t, lkrSrc, "Create src dir")

	dstDirOld := c.MustMkdir(t, lkrDst, "/x")
	dstDir := c.MustMove(t, lkrDst, dstDirOld, "/y")
	c.MustCommit(t, lkrDst, "I like to move it, move it")

	return []MapPair{
		{
			Src:         srcDir,
			Dst:         dstDir,
			SrcWasMoved: true,
		},
	}
}

func mapperSetupDstMoveDir(t *testing.T, lkrSrc, lkrDst *c.Linker) []MapPair {
	c.MustMkdir(t, lkrSrc, "/x")
	srcFile := c.MustTouch(t, lkrSrc, "/x/a.png", 42)
	c.MustCommit(t, lkrSrc, "Create src dir")

	dstDirOld := c.MustMkdir(t, lkrDst, "/x")
	c.MustMove(t, lkrDst, dstDirOld, "/y")
	dstFile := c.MustTouch(t, lkrDst, "/y/a.png", 23)
	c.MustCommit(t, lkrDst, "I like to move it, move it")

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

func mapperSetupMoveDirWithChild(t *testing.T, lkrSrc, lkrDst *c.Linker) []MapPair {
	srcDirOld := c.MustMkdir(t, lkrSrc, "/x")
	srcFile := c.MustTouch(t, lkrSrc, "/x/a.png", 23)
	c.MustMove(t, lkrSrc, srcDirOld, "/y")
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
	dstDir := c.MustMkdir(t, lkrDst, "/x")
	c.MustMkdir(t, lkrDst, "/y")
	c.MustTouch(t, lkrDst, "/x/a.png", 42)
	dstFile := c.MustTouch(t, lkrDst, "/y/a.png", 42)
	c.MustCommit(t, lkrDst, "Create src dir")

	return []MapPair{
		{
			Src:          srcFile,
			Dst:          dstFile,
			TypeMismatch: false,
		}, {
			Src:          nil,
			Dst:          dstDir,
			TypeMismatch: false,
		},
	}
}

func mapperSetupSrcFileMoveToExistingEmptyDir(t *testing.T, lkrSrc, lkrDst *c.Linker) []MapPair {
	c.MustMkdir(t, lkrSrc, "/d1")
	c.MustMkdir(t, lkrSrc, "/d2")
	srcFileOld, _ := c.MustTouchAndCommit(t, lkrSrc, "/d1/t1", 23)
	srcFile := c.MustMove(t, lkrSrc, srcFileOld, "/d2/t1")
	c.MustCommit(t, lkrSrc, "move is done")

	c.MustMkdir(t, lkrDst, "/d1")
	c.MustMkdir(t, lkrDst, "/d2")
	dstFile, _ := c.MustTouchAndCommit(t, lkrDst, "/d1/t1", 23)

	return []MapPair{
		{
			Src:         srcFile,
			Dst:         dstFile,
			SrcWasMoved: true,
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
	dstZ, _ := c.MustTouchAndCommit(t, lkrDst, "/dst-only/z", 23)

	srcZParent, err := n.ParentDirectory(lkrSrc, srcZ)
	if err != nil {
		t.Fatalf("setup failed to get parent dir: %v", err)
	}

	dstZParent, err := n.ParentDirectory(lkrDst, dstZ)
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
		}, {
			Src:          nil,
			Dst:          dstZParent,
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
			Src:           nil,
			Dst:           dstFile,
			TypeMismatch:  false,
			SrcWasRemoved: true,
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
		}, {
			name:  "move-dir-with-child",
			setup: mapperSetupMoveDirWithChild,
		}, {
			name:  "move-nested-dir",
			setup: mapperMoveNestedDir,
		}, {
			name:  "move-src-file-to-existing-empty-dir",
			setup: mapperSetupSrcFileMoveToExistingEmptyDir,
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

				mapper, err := NewMapper(lkrSrc, lkrDst, nil, nil, srcRoot)
				require.Nil(t, err)

				if err := mapper.Map(diffFn); err != nil {
					t.Fatalf("mapping failed: %v", err)
				}

				// DEBUG.
				// for _, pair := range got {
				// 	fmt.Println("-", pair.Src, pair.Dst)
				// }

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
