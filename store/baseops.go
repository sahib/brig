package store

import (
	"fmt"
	"path"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/id"
	"github.com/disorganizer/brig/util"
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
	fmt.Println("*****")
	root, err := fs.Root()
	if err != nil {
		return
	}

	err = Walk(root, true, func(child Node) error {
		fmt.Printf("%-47s %s\n", child.Hash().B58String(), NodePath(child))
		return nil
	})
	fmt.Println("+++++", err)
}

func mkdir(fs *FS, repoPath string, createParents bool) (*Directory, error) {
	dirname, basename := path.Split(repoPath)
	if basename == "" {
		return fs.Root()
	}

	// Check if the parent exists:
	parent, err := fs.LookupDirectory(dirname)
	if err != nil {
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

func createFile(fs *FS, repoPath string, newHash *Hash, key []byte, size uint64, author id.ID) (*File, error) {
	var oldHash *Hash

	file, err := fs.LookupFile(repoPath)
	if err != nil {
		return nil, err
	}

	needRemove := false

	if file != nil {
		// We know this file already.
		log.WithFields(log.Fields{"file": repoPath}).Info("File exists; modifying.")
		oldHash = file.Hash().Clone()
		needRemove = true
	} else {
		par, err := mkdirParents(fs, repoPath)
		if err != nil {
			return nil, err
		}

		// Create a new file at specified path:
		file, err = newEmptyFile(fs, par, path.Base(repoPath))
		if err != nil {
			return nil, err
		}
	}

	log.Infof(
		"store-add: %s (hash: %s, key: %x)",
		repoPath,
		newHash.B58String(),
		util.OmitBytes(file.Key(), 10),
	)

	if file.Hash().Equal(newHash) {
		log.Debugf("Hash was not modified. Refusing update.")
		return nil, ErrNoChange
	}

	parNode, err := file.Parent()
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
		if err := parDir.RemoveChild(file); err != nil {
			return nil, err
		}
	}

	file.SetSize(size)
	file.SetModTime(time.Now())
	file.SetKey(key)
	file.SetHash(newHash)

	// Add it again when the hash was changed.
	if err := parDir.Add(file); err != nil {
		return nil, err
	}

	// Create a checkpoint in the version history.
	if err := makeCheckpoint(fs, author, file.ID(), oldHash, newHash, repoPath, repoPath); err != nil {
		return nil, err
	}

	if err := fs.StageNode(file); err != nil {
		return nil, err
	}

	return file, err
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
