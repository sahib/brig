package repo

import (
	"fmt"

	h "github.com/sahib/brig/util/hashlib"
)

func (rp *Repository) GC(backend Backend, aggressive bool) (map[string]map[string]h.Hash, error) {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	// `killed` are the content hashes the backend disposed.
	killed, err := backend.GC()
	if err != nil {
		fmt.Println("backend gc error", err)
		return nil, err
	}

	result := make(map[string]map[string]h.Hash)
	if len(killed) == 0 {
		// Shortcut, since running the loop below
		// is currently rather expensive due to FilesByContents.
		// We can optimize that if it turns out to be a problem.
		return result, nil
	}

	for owner, fs := range rp.fsMap {
		if aggressive {
			// Make sure we also clean every bit
			// of memory/space we can find.
			fs.ScheduleGCRun()
		}

		nodeMap, err := fs.FilesByContent(killed)
		if err != nil {
			fmt.Println("get files by content")
			return nil, err
		}

		subResult := make(map[string]h.Hash)
		for content, info := range nodeMap {
			subResult[content] = info.Content
		}

		result[owner] = subResult
	}

	return result, nil
}
