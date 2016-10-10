package store

import (
	"errors"
	"fmt"
)

var (
	ErrNoMappingFound = errors.New("No mapping between local and remote path found")
	ErrConflict       = errors.New("Conflicting changes")
)

// Ein assoziatives Array mit dem Pfad zu der Historie
// seit dem letzten gemeinsamen Merge-Point.
type PathToHistory map[string]*History

func indexStore(bob *Store) (PathToHistory, error) {
	bobRoot, err := bob.fs.Root()
	if err != nil {
		return nil, err
	}

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

func (st *Store) syncByMapping(bob *Store, bobMap PathToHistory) error {
	ownRoot, err := st.fs.Root()
	if err != nil {
		return err
	}

	// Walk over the paths of alice and guess for each node
	// with which node of bob we have to synchronize.
	return Walk(ownRoot, true, func(child Node) error {
		path := child.Path()
		histA, err := st.fs.History(child.ID())
		if err != nil {
			return fmt.Errorf("No history from alice for `%s`: %v", path, err)
		}

		bobPath, err := st.mapPath(histA, bobMap)
		if err != nil && err != ErrNoMappingFound {
			return err
		}

		histB, err := bob.fs.HistoryByPath(bobPath)
		if err != nil && err != ErrNoMappingFound {
			return err
		}

		if err == ErrNoMappingFound {
			// Bob probably has not such a file.
			// Just ignore it then, but silence the error.
			return nil
		}

		if err := st.syncSingleFile(&histA, &histB); err != nil {
			return err
		}

		// Remember that we handled this file.
		delete(bobMap, bobPath)
		return nil
	})
}

func (st *Store) addLeftovers(bob *Store, bobMap PathToHistory) error {
	owner, err := st.Owner()
	if err != nil {
		return err
	}

	for path := range bobMap {
		// TODO: what to do with bob's history? Ignore for now?
		file, err := bob.fs.LookupFile(path)
		if err != nil {
			return err
		}

		_, err = touchFile(st.fs, path, file.Hash(), file.Key(), file.Size(), owner.ID())
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

	bobMap, err := indexStore(bob)
	if err != nil {
		return err
	}

	// NOTE: syncByMapping deletes all entries that were handled from bobMap.
	if err := st.syncByMapping(bob, bobMap); err != nil {
		return err
	}

	return nil
}

// mapPath maps the file described by the history HistA from Alice
// to a coressponding file in Bob's set of files.
// If no matching file could be found ErrNoMappingFound is returned.
func (st *Store) mapPath(HistA History, BobMapping PathToHistory) (string, error) {
	// Iterate over all pathes in alice' history of this file.
	// Usually this is just one path.
	for _, path := range HistA.AllPaths() {
		// Test if bob has a file with this path.
		histB, ok := BobMapping[path]
		if !ok {
			continue
		}

		// If yes, just return the newest path.
		// (i.e. the path after all moves)
		return histB.MostCurrentPath(st.fs), nil
	}

	// Whoops, no corresponding file found in bob's set.
	// It's very likely that bob does not possess this file.
	return "", ErrNoMappingFound
}

func (st *Store) syncSingleFile(historyA, historyB *History) error {
	if historyA.Equal(historyB) {
		// Keine weitere Aktion nötig.
		return nil
	}

	// Prüfe, ob historyA mit den Checkpoints von historyB beginnt.
	if historyA.IsPrefix(historyB) {
		// B hängt A hinterher.
		return nil
	}

	if historyB.IsPrefix(historyA) {
		// A hängt B hinterher. Kopiere B zu A ("fast forward").
		// TODO
		// copy(B, A)
		return nil
	}

	if rootIdx := historyA.CommonRoot(historyB); rootIdx >= 0 {
		// A und B haben trotzdem eine gemeinsame Historie,
		// haben sich aber auseinanderentwickelt.
		if historyA.ConflictingChanges(historyB, rootIdx) == nil {
			// Die Änderungen sind verträglich und
			// können automatisch aufgelöst werden.
			// ResolveConflict(historyA, historyB, root)
			return nil
		}
	}

	// Keine gemeinsame Historie.
	// TODO: handle history.
	return ErrConflict
}
