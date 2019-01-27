package endpoints

import (
	"net/http"
	"sort"
	"strings"
)

// DeletedPathsHandler implements http.Handler.
// This endpoint returns all directories that the client may see.
// It is used in the client to offer the user a list of directories
// to move or copy files to.
type DeletedPathsHandler struct {
	*State
}

// NewDeletedPathsHandler returns a new DeletedPathsHandler.
func NewDeletedPathsHandler(s *State) *DeletedPathsHandler {
	return &DeletedPathsHandler{State: s}
}

// DeletedPathsResponse is the response sent to the client.
type DeletedPathsResponse struct {
	Success bool        `json:"success"`
	Entries []*StatInfo `json:"entries"`
}

func (dh *DeletedPathsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if user := getUserName(dh.store, w, r); user == "" {
		jsonifyErrf(w, http.StatusForbidden, "bad user")
		return
	}

	nodes, err := dh.fs.DeletedNodes("/")
	if err != nil {
		jsonifyErrf(w, http.StatusInternalServerError, "failed to list")
		return
	}

	entries := []*StatInfo{}
	for _, node := range nodes {
		if !dh.validatePath(node.Path, w, r) {
			continue
		}

		entries = append(entries, toExternalStatInfo(node))
	}

	// Sort dirs before files and sort each part alphabetically
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}

		return strings.ToLower(entries[i].Path) < strings.ToLower(entries[j].Path)
	})

	jsonify(w, http.StatusOK, &DeletedPathsResponse{
		Success: true,
		Entries: entries,
	})
}
