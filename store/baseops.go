package store

import (
	"fmt"
	"path"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/id"
	"github.com/jbenet/go-multihash"
)

func mkdirParents(fs *FS, repoPath string) (*Directory, error) {
	repoPath = path.Clean(repoPath)

	elems := strings.Split(repoPath, "/")
	for idx := 0; idx < len(elems)-1; idx++ {
		dirname := strings.Join(elems[:idx+1], "/")
		if dirname == "" {
			dirname = "/"
		}

		dir, err := mkdir(fs, dirname, false)
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

func printTree(fs *FS) {
	fmt.Println("=== PRINT ===")
	root, err := fs.Root()
	if err != nil {
		return
	}

	err = Walk(root, true, func(child Node) error {
		fmt.Printf("%-47s %s\n", child.Hash().B58String(), NodePath(child))
		return nil
	})
}

func mkdir(fs *FS, repoPath string, createParents bool) (*Directory, error) {
	dirname, basename := path.Split(repoPath)
	if basename == "" {
		return fs.Root()
	}

	// Check if the parent exists:
	parent, err := fs.LookupDirectory(dirname)
	if err != nil && !IsNoSuchFileError(err) {
		return nil, err
	}
	fmt.Println("mkdir:", repoPath, parent, dirname, basename)

	// If it's nil, we might need to create it:
	if parent == nil {
		if !createParents {
			return nil, NoSuchFile(dirname)
		}

		parent, err = mkdirParents(fs, repoPath)
		if err != nil {
			return nil, err
		}
	}

	child, err := parent.Child(basename)
	if err != nil {
		return nil, err
	}

	if child != nil {
		if child.GetType() != NodeTypeDirectory {
			return nil, fmt.Errorf("`%s` exists and is not a directory", repoPath)
		}

		// Nothing to do really. Return the old child.
		return child.(*Directory), nil
	}

	// Create it then!
	dir, err := newEmptyDirectory(fs, parent, basename)
	if err != nil {
		return nil, err
	}

	if err := fs.StageNode(dir); err != nil {
		return nil, err
	}

	printTree(fs)
	return dir, nil
}

func resetNode(fs *FS, node SettableNode, commit *Commit) error {
	oldRoot, err := fs.DirectoryByHash(commit.Root())
	if err != nil {
		return err
	}

	repoPath := node.Path()
	oldNode, err := oldRoot.Lookup(repoPath)
	if IsNoSuchFileError(err) {
		// Node did not exist back then. Remove the current node.
		parent, err := node.Parent()
		if err != nil {
			return err
		}

		if err := fs.StageNode(parent); err != nil {
			return err
		}

		return makeCheckpoint(
			fs, "TODO",
			node.ID(), node.Hash(), nil,
			repoPath, repoPath,
		)
	}

	// Different error, abort.
	if err != nil {
		return err
	}

	_, err = stageNode(fs, repoPath, oldNode.Hash(), oldNode.Size(), "TODO")
	return err
}

func stageFile(fs *FS, repoPath string, newHash *Hash, size uint64, author id.ID, key []byte) (*File, error) {
	node, err := stageNode(fs, repoPath, newHash, size, author)
	if err != nil {
		return nil, err
	}

	fmt.Println("survived stageNode")

	if node.GetType() != NodeTypeFile {
		return nil, ErrBadNode
	}

	file, ok := node.(*File)
	if !ok {
		return nil, ErrBadNode
	}

	file.SetKey(key)
	return file, nil
}

func stageNode(fs *FS, repoPath string, newHash *Hash, size uint64, author id.ID) (Node, error) {
	var oldHash *Hash

	node, err := fs.LookupSettableNode(repoPath)
	if err != nil && !IsNoSuchFileError(err) {
		return nil, err
	}

	needRemove := false

	if node != nil {
		// We know this file already.
		log.WithFields(log.Fields{"node": repoPath}).Info("File exists; modifying.")
		oldHash = node.Hash().Clone()
		needRemove = true
	} else {
		par, err := mkdirParents(fs, repoPath)
		if err != nil {
			return nil, err
		}

		// Create a new file at specified path:
		node, err = newEmptyFile(fs, par, path.Base(repoPath))
		if err != nil {
			return nil, err
		}
	}

	// TODO: move to coreops
	// log.Infof(
	// 	"store-add: %s (hash: %s, key: %x)",
	// 	repoPath,
	// 	newHash.B58String(),
	// 	util.OmitBytes(file.Key(), 10),
	// )

	if node.Hash().Equal(newHash) {
		log.Debugf("Hash was not modified. Refusing update.")
		return nil, ErrNoChange
	}

	parNode, err := node.Parent()
	if err != nil {
		return nil, err
	}

	if parNode == nil {
		return nil, fmt.Errorf("%s has no parent yet (BUG)", repoPath)
	}

	parDir, ok := parNode.(*Directory)
	if !ok {
		return nil, ErrBadNode
	}

	if needRemove {
		// Remove the child before changing the hash:
		if err := parDir.RemoveChild(node); err != nil {
			return nil, err
		}
	}

	node.SetSize(size)
	node.SetModTime(time.Now())
	node.SetHash(newHash)

	// Add it again when the hash was changed.
	if err := parDir.Add(node); err != nil {
		return nil, err
	}

	// Create a checkpoint in the version history.
	if err := makeCheckpoint(fs, author, node.ID(), oldHash, newHash, repoPath, repoPath); err != nil {
		return nil, err
	}

	if err := fs.StageNode(node); err != nil {
		return nil, err
	}

	return node, err
}

func makeCheckpoint(fs *FS, author id.ID, ID uint64, oldHash, newHash *Hash, oldPath, newPath string) error {
	ckp, err := fs.LastCheckpoint(ID)
	if err != nil {
		return err
	}

	// There was probably no checkpoint yet:
	if ckp == nil {
		ckp = newEmptyCheckpoint(ID, oldHash, author)
	}

	newCkp, err := ckp.Fork(author, oldHash, newHash, oldPath, newPath)
	if err != nil {
		return err
	}

	return fs.StageCheckpoint(newCkp)
}

func resolveCommitRef(fs *FS, commitRef string) (*Commit, error) {
	mh, err := multihash.FromB58String(commitRef)
	if err != nil {
		// Does not look like a hash, maybe a normal ref?
		nd, err := fs.ResolveRef(commitRef)
		if err != nil {
			return nil, err
		}

		cmt, ok := nd.(*Commit)
		if !ok {
			return nil, ErrBadNode
		}

		return cmt, nil
	}

	return fs.CommitByHash(&Hash{mh})
}
