package vcs

import (
	"fmt"

	c "github.com/sahib/brig/catfs/core"
	ie "github.com/sahib/brig/catfs/errors"
	n "github.com/sahib/brig/catfs/nodes"
)

// Undelete tries to recover the node pointed to by `root`.
// The node must be a ghost, otherwise we will error out.
func Undelete(lkr *c.Linker, root string) error {
	nd, err := lkr.LookupModNode(root)
	if err != nil {
		return err
	}

	if nd.Type() != n.NodeTypeGhost {
		return fmt.Errorf("%s is not a deleted file: %v", root, err)
	}

	cmt, err := lkr.Status()
	if err != nil {
		return err
	}

	var origNd n.ModNode

	// Walk to the last point in history where the ghost
	// was either removed or moved. In theory it could have been
	// modified or added in between, but that would mean that
	// someone played around with the graph.
	walker := NewHistoryWalker(lkr, cmt, nd)
	for walker.Next() {
		state := walker.State()
		typ := state.Curr.Type()
		if typ != n.NodeTypeGhost {
			continue
		}

		if state.Mask&ChangeTypeRemove == 0 {
			continue
		}

		if state.Mask&ChangeTypeMove > 0 {
			continue
		}

		// We know now that we're on the ghost was added after deleting
		// or removing the file. Now go one back to reach the actual node.
		if !walker.Next() {
			break
		}

		origNd = walker.State().Curr
		break
	}

	if origNd == nil {
		return fmt.Errorf("could not find a state where this file was not deleted")
	}

	// Do the actual recovery. Handle the case where we are undeleting a
	// whole directory tree with possibly empty directories inside.
	return lkr.Atomic(func() (bool, error) {
		return true, n.Walk(lkr, origNd, true, func(child n.Node) error {
			switch child.Type() {
			case n.NodeTypeDirectory:
				dir, ok := child.(*n.Directory)
				if !ok {
					return ie.ErrBadNode
				}

				// Create empty directories manually,
				// all other directories will be created implicitly:
				if dir.NChildren() == 0 {
					_, err := c.Mkdir(lkr, dir.Path(), true)
					return err
				}
			case n.NodeTypeFile:
				file, ok := child.(*n.File)
				if !ok {
					return ie.ErrBadNode
				}

				// Stage that old state:
				_, err := c.Stage(
					lkr,
					file.Path(),
					file.ContentHash(),
					file.BackendHash(),
					file.Size(),
					file.Key(),
					file.ModTime(),
				)

				return err
			}
			return nil
		})
	})
}
