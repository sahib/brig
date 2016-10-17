package store

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
)

var (
	ErrNoMappingFound = errors.New("No mapping between local and remote path found")
	ErrConflict       = errors.New("Conflicting changes")
)

// candidate is a single candidate that needs some sort of action.
type candidate struct {
	ownPath  string
	bobPath  string
	bobStore *Store
	ownStore *Store
}

type pathToHistory map[string]*History

// collectHistoryMap iterates over all files and creates a pathHistory
func collectHistoryMap(bob *Store) (pathToHistory, error) {
	bobRoot, err := bob.fs.Root()
	if err != nil {
		return nil, err
	}

	fmt.Println("Bob root", bobRoot)

	// Walk over the contents of bob and remember all
	// histories under their path.
	bobMap := make(map[string]*History)
	err = Walk(bobRoot, true, func(child Node) error {
		path := child.Path()
		hist, err := bob.fs.History(child.ID())
		if err != nil {
			return fmt.Errorf("No history from bob for `%s`: %v", path, err)
		}

		bobMap[child.Path()] = &hist
		return nil
	})

	if err != nil {
		return nil, err
	}

	return bobMap, nil
}

// handleFastForward assumes that we already have this path.
func handleFastForward(cnd candidate) error {
	bobFile, err := cnd.bobStore.fs.ResolveFile(cnd.bobPath)
	if err != nil {
		return err
	}

	bobOwner, err := cnd.bobStore.Owner()
	if err != nil {
		return err
	}

	bobHash := bobFile.Hash()
	bobSize := bobFile.Size()
	bobKey := bobFile.Key()

	log.Infof("Fast forwarding %s to %s", cnd.ownPath, bobFile.Hash())

	_, err = stageFile(
		cnd.ownStore.fs,
		cnd.ownPath,
		bobHash,
		bobKey,
		bobSize,
		bobOwner.ID(),
	)

	return err
}

// handleConflict takes a candidate and adds it as conflict file.
// If the conflict file already exists, it will be updated.
func handleConflict(cnd candidate) error {
	log.Infof("Conflicting files: %s (own) <-> %s (remote)", cnd.ownPath, cnd.bobPath)
	bobOwner, err := cnd.bobStore.Owner()
	if err != nil {
		return err
	}

	bobFile, err := cnd.bobStore.fs.LookupFile(cnd.bobPath)
	if err != nil {
		return err
	}

	conflictPath := cnd.ownPath + "." + bobOwner.ID().User() + ".conflict"
	log.Infof("Creating conflict file: %s", conflictPath)

	bobHash := bobFile.Hash()
	bobSize := bobFile.Size()
	bobKey := bobFile.Key()

	_, err = stageFile(
		cnd.ownStore.fs,
		conflictPath,
		bobHash,
		bobKey,
		bobSize,
		bobOwner.ID(),
	)

	if err == ErrNoChange {
		return nil
	}

	return err
}

func (st *Store) syncByMapping(bob *Store, bobMap pathToHistory) error {
	ownRoot, err := st.fs.Root()
	if err != nil {
		return err
	}

	// Collect files that we need to handle. Modifying them while iterating
	// over the tree might is not the brightest idea.
	var conflicts, fastForward []candidate

	// Walk over the paths of alice and guess for each node
	// with which node of bob we have to synchronize.
	err = Walk(ownRoot, true, func(child Node) error {
		// TODO: Make sure to synchronize empty dirs later on:
		if child.GetType() != NodeTypeFile {
			return nil
		}

		ownPath := child.Path()

		fmt.Println("visit", ownPath)
		histA, err := st.fs.History(child.ID())
		if err != nil {
			return fmt.Errorf("No history from alice for `%s`: %v", ownPath, err)
		}

		bobPath, err := st.mapPath(bob, histA, bobMap)
		if err != nil && err != ErrNoMappingFound {
			return err
		}

		if err == ErrNoMappingFound {
			fmt.Println("no mapping found for", bobPath, "to", ownPath)
			return nil
		}

		fmt.Println("Mapping was", bobPath)
		histB, err := bob.fs.HistoryByPath(bobPath)
		if err != nil {
			return err
		}

		if err == ErrNoMappingFound {
			// Bob probably has not such a file.
			// Just ignore it then, but silence the error.
			return nil
		}

		canFF, err := st.decideSingleFile(&histA, &histB)
		if err == ErrConflict {
			// Handle conflict.
			conflicts = append(conflicts, candidate{ownPath, bobPath, bob, st})
			delete(bobMap, bobPath)
			return nil
		}

		if err != nil {
			return err
		}

		if canFF {
			fastForward = append(fastForward, candidate{ownPath, bobPath, bob, st})
			delete(bobMap, bobPath)
			return nil
		}

		// Remember that we handled this file.
		delete(bobMap, bobPath)
		return nil
	})

	if err != nil {
		return err
	}

	// Handle the candidates we've found:

	for _, cnd := range conflicts {
		if err := handleConflict(cnd); err != nil {
			return err
		}
	}

	for _, cnd := range fastForward {
		if err := handleFastForward(cnd); err != nil {
			return err
		}
	}

	return nil
}

// addLeftovers takes the paths from bob that alice doesn't posess.
func (st *Store) addLeftovers(bob *Store, bobMap pathToHistory) error {
	owner, err := st.Owner()
	if err != nil {
		return err
	}

	for path := range bobMap {
		// TODO: what to do with bob's history? Ignore for now?
		node, err := bob.fs.LookupNode(path)
		if err != nil {
			return err
		}

		if node.GetType() != NodeTypeFile {
			continue
		}

		file, ok := node.(*File)
		if !ok {
			log.Warningf("Syncing messed up file types; not a file: %v", file)
			continue
		}

		_, err = stageFile(st.fs, path, file.Hash(), file.Key(), file.Size(), owner.ID())
		if err != nil {
			return err
		}
	}

	return nil
}

// SyncWith synchronizes the two stores `st` and `bob`. `st` takes precendence.
// If succesful, a new merge commit is created.
func (st *Store) SyncWith(bob *Store) error {
	st.mu.Lock()
	defer st.mu.Unlock()

	bob.mu.Lock()
	defer bob.mu.Unlock()

	bobMap, err := collectHistoryMap(bob)
	if err != nil {
		return err
	}

	// NOTE: syncByMapping deletes all entries that were handled from bobMap.
	if err := st.syncByMapping(bob, bobMap); err != nil {
		return err
	}

	return st.addLeftovers(bob, bobMap)
}

// mapPath maps the file described by the history histA from Alice
// to a coressponding file in Bob's set of files.
// If no matching file could be found ErrNoMappingFound is returned.
func (st *Store) mapPath(bobStore *Store, histA History, bobMapping pathToHistory) (string, error) {
	// Iterate over all pathes in alice' history of this file.
	// Usually this is just one path.
	paths, err := histA.AllPaths(st.fs)
	if err != nil {
		return "", err
	}

	for _, path := range paths {
		// Test if bob has a file with this path.
		histB, ok := bobMapping[path]
		if !ok {
			continue
		}

		// If yes, just return the newest path.
		// (i.e. the path after all moves)
		bobPath, err := histB.MostCurrentPath(bobStore.fs)
		if err != nil {
			return "", err
		}

		return bobPath, nil
	}

	// Whoops, no corresponding file found in bob's set.
	// It's very likely that bob does not possess this file.
	return "", ErrNoMappingFound
}

func (st *Store) decideSingleFile(historyA, historyB *History) (bool, error) {
	hasPrefix := historyA.HasSharedPrefix(historyB)
	if len(*historyA) == len(*historyB) && hasPrefix {
		// Both histories are the same. Nothing needs to be done.
		return false, nil
	}

	if hasPrefix && len(*historyA) < len(*historyB) {
		// We are behind B. Fast forward the file.
		return true, nil
	}

	if hasPrefix && len(*historyA) > len(*historyB) {
		// We have more checkpoints than B. Do nothing.
		return false, nil
	}

	// TODO: Check if histories have a common root and if so,
	//       if the checkpoints since then have compatible changes.
	//  rootIdx := historyA.CommonRoot(historyB)
	//  if rootIdx < 0 {
	//     // No root found, check all checkpoints.
	//     rootIdx = 0
	//  }
	//
	// 	if historyA.ConflictingChanges(historyB, rootIdx) == nil {
	// 		return ResolveConflict(historyA, historyB, root)
	// 	}
	// }

	// Can't do much here without asking the user.
	return false, ErrConflict
}
