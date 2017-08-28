package cafs

import (
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	n "github.com/disorganizer/brig/cafs/nodes"
	h "github.com/disorganizer/brig/util/hashlib"
	e "github.com/pkg/errors"
)

var (
	ErrIsGhost = errors.New("Is a ghost")
)

// mkdirParents takes the dirname of repoPath and makes sure all intermediate
// directories are created. The last directory will be returned.

// If any directory exist already, it will not be touched.
// You can also think of it as mkdir -p.
func mkdirParents(lkr *Linker, repoPath string) (*n.Directory, error) {
	repoPath = path.Clean(repoPath)

	elems := strings.Split(repoPath, "/")
	for idx := 0; idx < len(elems)-1; idx++ {
		dirname := strings.Join(elems[:idx+1], "/")
		if dirname == "" {
			dirname = "/"
		}

		dir, err := mkdir(lkr, dirname, false)
		if err != nil {
			return nil, err
		}

		// Return it, if it's the last path component:
		if idx+1 == len(elems)-1 {
			return dir, nil
		}
	}

	return nil, fmt.Errorf("Empty path given")
}

// mkdir creates the directory at repoPath and any intermediate directories if
// createParents is true. It will fail if there is already a file at `repoPath`
// and it is not a directory.
func mkdir(lkr *Linker, repoPath string, createParents bool) (dir *n.Directory, err error) {
	dirname, basename := path.Split(repoPath)

	// Take special care of the root node:
	if basename == "" {
		return lkr.Root()
	}

	// Check if the parent exists:
	parent, err := lkr.LookupDirectory(dirname)
	if err != nil && !n.IsNoSuchFileError(err) {
		return nil, e.Wrap(err, "dirname lookup failed")
	}

	log.Debugf("mkdir: %s", dirname)

	// If it's nil, we might need to create it:
	if parent == nil {
		if !createParents {
			return nil, n.NoSuchFile(dirname)
		}

		parent, err = mkdirParents(lkr, repoPath)
		if err != nil {
			return nil, err
		}
	}

	child, err := parent.Child(lkr, basename)
	if err != nil {
		return nil, err
	}

	if child != nil {
		if child.Type() != n.NodeTypeDirectory {
			return nil, fmt.Errorf("`%s` exists and is not a directory", repoPath)
		}

		// Nothing to do really. Return the old child.
		return child.(*n.Directory), nil
	}

	// Make sure, NextInode() and StageNode is written in one batch.
	batch := lkr.kv.Batch()
	defer func() {
		if err != nil {
			batch.Rollback()
		} else {
			err = batch.Flush()
		}
	}()

	// Create it then!
	dir, err = n.NewEmptyDirectory(lkr, parent, basename, lkr.NextInode())
	if err != nil {
		return nil, err
	}

	if err := lkr.StageNode(dir); err != nil {
		return nil, err
	}

	return dir, nil
}

// remove removes a single node from a directory.
// `nd` is the node that shall be removed and may not be root.
// The parent directory is returned.
func remove(lkr *Linker, nd n.ModNode, createGhost bool) (parentDir *n.Directory, ghost *n.Ghost, err error) {
	if nd.Type() == n.NodeTypeGhost {
		return nil, nil, ErrIsGhost
	}

	parentDir, err = n.ParentDirectory(lkr, nd)
	if err != nil {
		return nil, nil, err
	}

	// We shouldn't delete the root directory
	// (only directory with a parent)
	if parentDir == nil {
		return nil, nil, fmt.Errorf("Refusing to delete /")
	}

	if err := parentDir.RemoveChild(lkr, nd); err != nil {
		return nil, nil, fmt.Errorf("Failed to remove child: %v", err)
	}

	lkr.MemIndexPurge(nd)

	// Make sure both StageNode() will be written in one batch:
	batch := lkr.kv.Batch()
	defer func() {
		if err != nil {
			batch.Rollback()
		} else {
			err = batch.Flush()
		}
	}()

	if err := lkr.StageNode(parentDir); err != nil {
		return nil, nil, err
	}

	if createGhost {
		ghost, err := n.MakeGhost(nd, nil, lkr.NextInode())
		if err != nil {
			return nil, nil, err
		}

		parentDir.Add(lkr, ghost)
		if err != nil {
			return nil, nil, err
		}

		if err := lkr.StageNode(ghost); err != nil {
			return nil, nil, err
		}

		return parentDir, ghost, nil
	}

	return parentDir, nil, nil
}

func move(lkr *Linker, nd n.ModNode, destPath string) (err error) {
	// Forbid moving a node inside of one of it's subdirectories.
	if strings.HasPrefix(destPath, nd.Path()) {
		return fmt.Errorf("Cannot move `%s` into it's own subdir `%s`", nd.Path(), destPath)
	}

	// Check if the destination already exists:
	destNode, err := lkr.LookupModNode(destPath)
	if err != nil && !n.IsNoSuchFileError(err) {
		return err
	}

	var parentDir *n.Directory

	// Make sure both StageNode() will be written in one batch:
	batch := lkr.kv.Batch()
	defer func() {
		if err != nil {
			batch.Rollback()
		} else {
			err = batch.Flush()
		}
	}()

	if destNode != nil {
		switch destNode.Type() {
		case n.NodeTypeDirectory:
			// Move inside of this directory.
			// Check if there is already a file
			destDir, ok := destNode.(*n.Directory)
			if !ok {
				return n.ErrBadNode
			}

			child, err := destDir.Child(lkr, nd.Name())
			if err != nil {
				return err
			}

			// Oh, something is in there?
			if child != nil {
				if nd.Type() == n.NodeTypeFile {
					// TODO: more details
					return fmt.Errorf("Cannot overwrite a directory with a file")
				}

				childDir, ok := child.(*n.Directory)
				if !ok {
					return n.ErrBadNode
				}

				if childDir.Size() > 0 {
					return fmt.Errorf("Cannot move over: %s; directory is not empty!", child.Path())
				}

				// Okay, there is an empty directory. Let's remove it to
				// replace it with our source node.
				if _, _, err := remove(lkr, childDir, false); err != nil {
					return err
				}
			}

			parentDir = destDir
		case n.NodeTypeFile:
			// Move over this file, making it a Ghost.
			parentDir, _, err = remove(lkr, destNode, false)
			if err != nil {
				return err
			}
		case n.NodeTypeGhost:
			// It is already a ghost. Overwrite it and do not create a new one.
			parentDir, _, err = remove(lkr, destNode, false)
			if err != nil {
				return err
			}
		default:
			return n.ErrBadNode
		}
	} else {
		// No node at this place yet, attempt to look it up.
		parentDir, err = lkr.LookupDirectory(path.Dir(destPath))
		if err != nil {
			return err
		}
	}

	// Remove the old node:
	_, ghost, err := remove(lkr, nd, true)
	if err != nil {
		return err
	}

	nd.SetName(path.Base(destPath))

	// And add it to the right destination dir:
	if err := parentDir.Add(lkr, nd); err != nil {
		return err
	}

	if err := lkr.AddMoveMapping(nd, ghost); err != nil {
		return err
	}

	return lkr.StageNode(nd)
}

type NodeUpdate struct {
	Hash   h.Hash
	Size   uint64
	Author string
	Key    []byte
}

func stage(lkr *Linker, repoPath string, info *NodeUpdate) (file *n.File, err error) {
	node, err := lkr.LookupNode(repoPath)
	if err != nil && !n.IsNoSuchFileError(err) {
		return nil, err
	}

	batch := lkr.kv.Batch()
	defer func() {
		if err != nil {
			batch.Rollback()
		} else {
			err = batch.Flush()
		}
	}()

	if node != nil && node.Type() == n.NodeTypeGhost {
		ghostParent, err := n.ParentDirectory(lkr, node)
		if err != nil {
			return nil, err
		}

		if ghostParent == nil {
			// TODO: Think about this case. stage() should not be called on directories
			//       anyways (only on files or files to ghosts)
			return nil, fmt.Errorf("The ghost is root? TODO")
		}

		if err := ghostParent.RemoveChild(lkr, node); err != nil {
			return nil, err
		}

		// Act like there was no previous node.
		// New node will have a different Inode.
		file = nil
		fmt.Println("is a ghost")
	} else if node != nil {
		var ok bool
		file, ok = node.(*n.File)
		fmt.Println("file exists")
		if !ok {
			return nil, n.ErrBadNode
		}
	} else {
		fmt.Println("nothing exists")
	}

	needRemove := false

	if file != nil {
		// We know this file already.
		log.WithFields(log.Fields{"file": repoPath}).Info("File exists; modifying.")
		needRemove = true

		if file.Hash().Equal(info.Hash) {
			log.Debugf("Hash was not modified. Refusing update.")
			return nil, ErrNoChange
		}
	} else {
		parent, err := mkdirParents(lkr, repoPath)
		if err != nil {
			return nil, err
		}

		// Create a new file at specified path:
		file, err = n.NewEmptyFile(parent, path.Base(repoPath), lkr.NextInode())
		if err != nil {
			return nil, err
		}
	}

	parentDir, err := n.ParentDirectory(lkr, file)
	if err != nil {
		return nil, err
	}

	if parentDir == nil {
		return nil, fmt.Errorf("%s has no parent yet (BUG)", repoPath)
	}

	if needRemove {
		// Remove the child before changing the hash:
		if err := parentDir.RemoveChild(lkr, file); err != nil {
			return nil, err
		}
	}

	file.SetSize(info.Size)
	file.SetModTime(time.Now())
	file.SetContent(lkr, info.Hash)
	file.SetKey(info.Key)

	// Add it again when the hash was changed.
	if err := parentDir.Add(lkr, file); err != nil {
		return nil, err
	}

	if err := lkr.StageNode(file); err != nil {
		return nil, err
	}

	return file, nil
}
