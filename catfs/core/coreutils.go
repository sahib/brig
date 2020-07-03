package core

import (
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	e "github.com/pkg/errors"
	ie "github.com/sahib/brig/catfs/errors"
	n "github.com/sahib/brig/catfs/nodes"
	h "github.com/sahib/brig/util/hashlib"
	log "github.com/sirupsen/logrus"
)

var (
	// ErrIsGhost is returned by Remove() when calling it on a ghost.
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

		dir, err := Mkdir(lkr, dirname, false)
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

// Mkdir creates the directory at repoPath and any intermediate directories if
// createParents is true. It will fail if there is already a file at `repoPath`
// and it is not a directory.
func Mkdir(lkr *Linker, repoPath string, createParents bool) (dir *n.Directory, err error) {
	dirname, basename := path.Split(repoPath)

	// Take special care of the root node:
	if basename == "" {
		return lkr.Root()
	}

	// Check if the parent exists:
	parent, lerr := lkr.LookupDirectory(dirname)
	if lerr != nil && !ie.IsNoSuchFileError(lerr) {
		err = e.Wrap(lerr, "dirname lookup failed")
		return
	}

	err = lkr.Atomic(func() (bool, error) {
		// If it's nil, we might need to create it:
		if parent == nil {
			if !createParents {
				return false, ie.NoSuchFile(dirname)
			}

			parent, err = mkdirParents(lkr, repoPath)
			if err != nil {
				return true, err
			}
		}

		child, err := parent.Child(lkr, basename)
		if err != nil {
			return true, err
		}

		if child != nil {
			switch child.Type() {
			case n.NodeTypeDirectory:
				// Nothing to do really. Return the old child.
				dir = child.(*n.Directory)
				return false, nil
			case n.NodeTypeFile:
				return true, fmt.Errorf("`%s` exists and is a file", repoPath)
			case n.NodeTypeGhost:
				// Remove the ghost and continue with adding:
				if err := parent.RemoveChild(lkr, child); err != nil {
					return true, err
				}
			default:
				return true, ie.ErrBadNode
			}
		}

		// Create it then!
		dir, err = n.NewEmptyDirectory(lkr, parent, basename, lkr.owner, lkr.NextInode())
		if err != nil {
			return true, err
		}

		if err := lkr.StageNode(dir); err != nil {
			return true, e.Wrapf(err, "stage dir")
		}

		log.Debugf("mkdir: %s", dirname)
		return false, nil
	})

	return
}

// Remove removes a single node from a directory.
// `nd` is the node that shall be removed and may not be root.
// The parent directory is returned.
func Remove(lkr *Linker, nd n.ModNode, createGhost, force bool) (parentDir *n.Directory, ghost *n.Ghost, err error) {
	if !force && nd.Type() == n.NodeTypeGhost {
		err = ErrIsGhost
		return
	}

	parentDir, err = n.ParentDirectory(lkr, nd)
	if err != nil {
		return
	}

	// We shouldn't delete the root directory
	// (only directory with a parent)
	if parentDir == nil {
		err = fmt.Errorf("refusing to delete root")
		return
	}

	err = lkr.Atomic(func() (bool, error) {
		if err := parentDir.RemoveChild(lkr, nd); err != nil {
			return true, fmt.Errorf("failed to remove child: %v", err)
		}

		lkr.MemIndexPurge(nd)

		if err := lkr.StageNode(parentDir); err != nil {
			return true, err
		}

		if createGhost {
			newGhost, err := n.MakeGhost(nd, lkr.NextInode())
			if err != nil {
				return true, err
			}

			if err := parentDir.Add(lkr, newGhost); err != nil {
				return true, err
			}

			if err := lkr.StageNode(newGhost); err != nil {
				return true, err
			}

			ghost = newGhost
			return false, nil
		}

		return false, nil
	})

	return
}

// prepareParent tries to figure out the correct parent directory when attempting
// to move `nd` to `dstPath`. It also removes any nodes that are "in the way" if possible.
func prepareParent(lkr *Linker, nd n.ModNode, dstPath string) (*n.Directory, error) {
	// Check if the destination already exists:
	destNode, err := lkr.LookupModNode(dstPath)
	if err != nil && !ie.IsNoSuchFileError(err) {
		return nil, err
	}

	if destNode == nil {
		// No node at this place yet, attempt to look it up.
		return lkr.LookupDirectory(path.Dir(dstPath))
	}

	switch destNode.Type() {
	case n.NodeTypeDirectory:
		// Move inside of this directory.
		// Check if there is already a file
		destDir, ok := destNode.(*n.Directory)
		if !ok {
			return nil, ie.ErrBadNode
		}

		child, err := destDir.Child(lkr, nd.Name())
		if err != nil {
			return nil, err
		}

		// Oh, something is in there?
		if child != nil {
			if nd.Type() == n.NodeTypeFile {
				return nil, fmt.Errorf(
					"cannot overwrite a directory (%s) with a file (%s)",
					destNode.Path(),
					child.Path(),
				)
			}

			childDir, ok := child.(*n.Directory)
			if !ok {
				return nil, ie.ErrBadNode
			}

			if childDir.Size() > 0 {
				return nil, fmt.Errorf(
					"cannot move over: %s; directory is not empty",
					child.Path(),
				)
			}

			// Okay, there is an empty directory. Let's remove it to
			// replace it with our source node.
			log.Warningf("Remove child dir: %v", childDir)
			if _, _, err := Remove(lkr, childDir, false, false); err != nil {
				return nil, err
			}
		}

		return destDir, nil
	case n.NodeTypeFile:
		log.Infof("Remove file: %v", destNode.Path())
		parentDir, _, err := Remove(lkr, destNode, false, false)
		return parentDir, err
	case n.NodeTypeGhost:
		// It is already a ghost. Overwrite it and do not create a new one.
		log.Infof("Remove ghost: %v", destNode.Path())
		parentDir, _, err := Remove(lkr, destNode, false, true)
		return parentDir, err
	default:
		return nil, ie.ErrBadNode
	}
}

// Copy copies the node `nd` to the path at `dstPath`.
func Copy(lkr *Linker, nd n.ModNode, dstPath string) (newNode n.ModNode, err error) {
	// Forbid moving a node inside of one of it's subdirectories.
	if nd.Path() == dstPath {
		err = fmt.Errorf("source and dest are the same file: %v", dstPath)
		return
	}

	if strings.HasPrefix(path.Dir(dstPath), nd.Path()) {
		err = fmt.Errorf(
			"cannot copy `%s` into it's own subdir `%s`",
			nd.Path(),
			dstPath,
		)
		return
	}

	err = lkr.Atomic(func() (bool, error) {
		parentDir, err := prepareParent(lkr, nd, dstPath)
		if err != nil {
			return true, e.Wrapf(err, "handle parent")
		}

		// We might copy something into a directory.
		// In this case, dstPath specifies the directory we move into,
		// not the file we moved to (which we need here)
		if parentDir.Path() == dstPath {
			dstPath = path.Join(parentDir.Path(), path.Base(nd.Path()))
		}

		// And add it to the right destination dir:
		newNode = nd.Copy(lkr.NextInode())
		newNode.SetName(path.Base(dstPath))
		if err := newNode.SetParent(lkr, parentDir); err != nil {
			return true, e.Wrapf(err, "set parent")
		}

		if err := newNode.NotifyMove(lkr, parentDir, newNode.Path()); err != nil {
			return true, e.Wrapf(err, "notify move")
		}

		return false, lkr.StageNode(newNode)
	})

	return
}

// Move moves the node `nd` to the path at `dstPath` and leaves
// a ghost at the old place.
func Move(lkr *Linker, nd n.ModNode, dstPath string) error {
	// Forbid moving a node inside of one of it's subdirectories.
	if nd.Type() == n.NodeTypeGhost {
		return errors.New("cannot move ghosts")
	}

	if nd.Path() == dstPath {
		return fmt.Errorf("Source and Dest are the same file: %v", dstPath)
	}

	if strings.HasPrefix(path.Dir(dstPath), nd.Path()) {
		return fmt.Errorf(
			"Cannot move `%s` into it's own subdir `%s`",
			nd.Path(),
			dstPath,
		)
	}

	return lkr.Atomic(func() (bool, error) {
		parentDir, err := prepareParent(lkr, nd, dstPath)
		if err != nil {
			return true, err
		}

		// Remove the old node:
		oldPath := nd.Path()
		_, ghost, err := Remove(lkr, nd, true, true)
		if err != nil {
			return true, e.Wrapf(err, "remove old")
		}

		if parentDir.Path() == dstPath {
			dstPath = path.Join(parentDir.Path(), path.Base(oldPath))
		}

		// The node needs to be told that it's path changed,
		// since it might need to change it's hash value now.
		if err := nd.NotifyMove(lkr, parentDir, dstPath); err != nil {
			return true, e.Wrapf(err, "notify move")
		}

		err = n.Walk(lkr, nd, true, func(child n.Node) error {
			return e.Wrapf(lkr.StageNode(child), "stage node")
		})

		if err != nil {
			return true, err
		}

		if err := lkr.AddMoveMapping(nd.Inode(), ghost.Inode()); err != nil {
			return true, e.Wrapf(err, "add move mapping")
		}

		return false, nil
	})
}

// StageFromFileNode is a convinience helper that will call Stage() with all necessary params from `f`.
func StageFromFileNode(lkr *Linker, f *n.File) (*n.File, error) {
	return StageWithFullInfo(lkr, f.Path(), f.ContentHash(), f.BackendHash(), f.Size(), f.CachedSize(), f.Key(), f.ModTime())
}

// Stage adds a file to brigs DAG this is lesser version since it does not use cachedSize
// Do not use it if you can, use StageWithFullInfo couple lines below!
// TODO rename Stage calls everywhere (especially in tests) and then
// rename Stage -> StageWithoutCacheSize, and StageWithFullInfo -> Stage
func Stage(lkr *Linker, repoPath string, contentHash, backendHash h.Hash, size uint64, key []byte, modTime time.Time) (file *n.File, err error) {
	// MaxUint64 indicates that cachedSize is unknown
	MaxUint64 := uint64(1<<64 - 1)
	return StageWithFullInfo(lkr, repoPath, contentHash, backendHash, size, MaxUint64, key, modTime)
}

// Stage adds a file to brigs DAG.
func StageWithFullInfo(lkr *Linker, repoPath string, contentHash, backendHash h.Hash, size, cachedSize uint64, key []byte, modTime time.Time) (file *n.File, err error) {
	node, lerr := lkr.LookupNode(repoPath)
	if lerr != nil && !ie.IsNoSuchFileError(lerr) {
		err = lerr
		return
	}

	err = lkr.Atomic(func() (bool, error) {
		if node != nil {
			if node.Type() == n.NodeTypeGhost {
				ghostParent, err := n.ParentDirectory(lkr, node)
				if err != nil {
					return true, err
				}

				if ghostParent == nil {
					return true, fmt.Errorf(
						"bug: %s has no parent. Is root a ghost?",
						node.Path(),
					)
				}

				if err := ghostParent.RemoveChild(lkr, node); err != nil {
					return true, err
				}

				// Act like there was no previous node.
				// New node will have a different Inode.
				file = nil
			} else {
				var ok bool
				file, ok = node.(*n.File)
				if !ok {
					return true, ie.ErrBadNode
				}
			}
		}

		needRemove := false
		if file != nil {
			// We know this file already.
			log.WithFields(log.Fields{"file": repoPath}).Info("File exists; modifying.")
			needRemove = true

			if file.BackendHash().Equal(backendHash) {
				log.Debugf("Hash was not modified. Not doing any update.")
				return false, nil
			}
		} else {
			parent, err := mkdirParents(lkr, repoPath)
			if err != nil {
				return true, err
			}

			// Create a new file at specified path:
			file = n.NewEmptyFile(parent, path.Base(repoPath), lkr.owner, lkr.NextInode())
		}

		parentDir, err := n.ParentDirectory(lkr, file)
		if err != nil {
			return true, err
		}

		if parentDir == nil {
			return true, fmt.Errorf("%s has no parent yet (BUG)", repoPath)
		}

		if needRemove {
			// Remove the child before changing the hash:
			if err := parentDir.RemoveChild(lkr, file); err != nil {
				return true, err
			}
		}

		file.SetSize(size)
		file.SetCachedSize(cachedSize)
		file.SetModTime(modTime)
		file.SetContent(lkr, contentHash)
		file.SetBackend(lkr, backendHash)
		file.SetKey(key)
		file.SetUser(lkr.owner)

		// Add it again when the hash was changed.
		log.Debugf("adding %s (%v)", file.Path(), file.BackendHash())
		if err := parentDir.Add(lkr, file); err != nil {
			return true, err
		}

		if err := lkr.StageNode(file); err != nil {
			return true, err
		}

		return false, nil
	})

	return
}

// Log will call `fn` on every commit we currently have, starting
// with the most current one (CURR, then HEAD, ...).
// If `fn` will return an error, the iteration is being stopped.
func Log(lkr *Linker, start *n.Commit, fn func(cmt *n.Commit) error) error {
	curr := start
	for curr != nil {
		if err := fn(curr); err != nil {
			return err
		}

		parent, err := curr.Parent(lkr)
		if err != nil {
			return err
		}

		if parent == nil {
			break
		}

		parentCmt, ok := parent.(*n.Commit)
		if !ok {
			return ie.ErrBadNode
		}

		curr = parentCmt
	}

	return nil
}
