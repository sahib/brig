package cafs

import (
	"fmt"
	"path"
	"strings"

	log "github.com/Sirupsen/logrus"
	n "github.com/disorganizer/brig/cafs/nodes"
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

func printTree(lkr *Linker) {
	fmt.Println("=== PRINT ===")
	root, err := lkr.Root()
	if err != nil {
		return
	}

	err = n.Walk(lkr, root, true, func(child n.Node) error {
		fmt.Printf("%-47s %s\n", child.Hash().B58String(), child.Path())
		return nil
	})
}

// mkdir creates the directory at repoPath and any intermediate directories if
// createParents is true. It will fail if there is already a file at `repoPath`
// and it is not a directory.
func mkdir(lkr *Linker, repoPath string, createParents bool) (*n.Directory, error) {
	dirname, basename := path.Split(repoPath)

	// Take special care of the root node:
	if basename == "" {
		return lkr.Root()
	}

	// Check if the parent exists:
	parent, err := lkr.LookupDirectory(dirname)
	if err != nil && !n.IsNoSuchFileError(err) {
		return nil, err
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

	// Create it then!
	dir, err := n.NewEmptyDirectory(lkr, parent, basename, lkr.NextInode())
	if err != nil {
		return nil, err
	}

	if err := lkr.StageNode(dir); err != nil {
		return nil, err
	}

	return dir, nil
}

// remove removes a single node from a directory.
func remove(lkr *Linker, nd n.Node, createGhost bool) (*n.Directory, error) {
	parent, err := nd.Parent(lkr)
	if err != nil {
		return nil, err
	}

	// We shouldn't delete the root directory
	// (only directory with a parent)
	if parent == nil {
		return nil, fmt.Errorf("Refusing to delete /")
	}

	parentDir, ok := parent.(*n.Directory)
	if !ok {
		return nil, n.ErrBadNode
	}

	if err := parentDir.RemoveChild(lkr, nd); err != nil {
		return nil, err
	}

	if err := lkr.StageNode(parent); err != nil {
		return nil, err
	}

	if createGhost {
		ghost, err := n.MakeGhost(nd)
		if err != nil {
			return nil, err
		}

		parentDir.Add(lkr, ghost)
		if err != nil {
			return nil, err
		}
	}

	return parentDir, nil
}

func resetNode(lkr *Linker, node n.SettableNode, commit *n.Commit) error {
	oldRoot, err := lkr.DirectoryByHash(commit.Root())
	if err != nil {
		return err
	}

	repoPath := node.Path()
	oldNode, err := oldRoot.Lookup(lkr, repoPath)
	if n.IsNoSuchFileError(err) {
		// Node did not exist back then. Remove the current node.
		_, err = remove(lkr, oldNode, false)
		return err
	}

	// Different error, abort.
	if err != nil {
		return err
	}

	fmt.Println(oldNode)
	// _, err = stageNode(lkr, repoPath, oldNode.Hash(), oldNode.Size(), author)
	return err
}

// func stageFile(lkr *Linker, repoPath string, newHash *h.Hash, size uint64, author string, key []byte) (*n.File, error) {
// 	node, err := stageNode(lkr, repoPath, newHash, size, author)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	if node.GetType() != NodeTypeFile {
// 		return nil, ErrBadNode
// 	}
//
// 	file, ok := node.(*File)
// 	if !ok {
// 		return nil, ErrBadNode
// 	}
//
// 	log.Infof(
// 		"store-add: %s (hash: %s, key: %x)",
// 		repoPath,
// 		newHash.B58String(),
// 		util.OmitBytes(key, 10),
// 	)
//
// 	file.SetKey(key)
// 	return file, nil
// }
//
// func stageNode(lkr *Linker, repoPath string, newHash *h.Hash, size uint64, author string) (Node, error) {
// 	var oldHash *h.Hash
//
// 	node, err := lkr.ResolveSettableNode(repoPath)
// 	if err != nil && !IsNoSuchFileError(err) {
// 		return nil, err
// 	}
//
// 	needRemove := false
//
// 	if node != nil {
// 		// We know this file already.
// 		log.WithFields(log.Fields{"node": repoPath}).Info("File exists; modifying.")
// 		oldHash = node.Hash().Clone()
// 		needRemove = true
// 	} else {
// 		par, err := mkdirParents(lkr, repoPath)
// 		if err != nil {
// 			return nil, err
// 		}
//
// 		// Create a new file at specified path:
// 		node, err = newEmptyFile(lkr, par, path.Base(repoPath))
// 		if err != nil {
// 			return nil, err
// 		}
// 	}
//
// 	if node.Hash().Equal(newHash) {
// 		log.Debugf("Hash was not modified. Refusing update.")
// 		return nil, ErrNoChange
// 	}
//
// 	parNode, err := node.Parent()
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	if parNode == nil {
// 		return nil, fmt.Errorf("%s has no parent yet (BUG)", repoPath)
// 	}
//
// 	parDir, ok := parNode.(*Directory)
// 	if !ok {
// 		return nil, ErrBadNode
// 	}
//
// 	if needRemove {
// 		// Remove the child before changing the hash:
// 		if err := parDir.RemoveChild(node); err != nil && !IsNoSuchFileError(err) {
// 			return nil, err
// 		}
// 	}
//
// 	node.SetSize(size)
// 	node.SetModTime(time.Now())
// 	node.SetHash(newHash)
//
// 	// Add it again when the hash was changed.
// 	if err := parDir.Add(node); err != nil {
// 		return nil, err
// 	}
//
// 	if err := lkr.StageNode(node); err != nil {
// 		return nil, err
// 	}
//
// 	return node, err
// }
