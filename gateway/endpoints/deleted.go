package endpoints

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	"github.com/sahib/brig/catfs"
	"github.com/sahib/brig/util"
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

// DeletedRequest is the data sent to this endpoint.
type DeletedRequest struct {
	Offset int64  `json:"offset"`
	Limit  int64  `json:"limit"`
	Filter string `json:"filter"`
}

func matchEntry(info *catfs.StatInfo, filter string) bool {
	return strings.Contains(strings.ToLower(info.Path), filter)
}

func (dh *DeletedPathsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	delReq := DeletedRequest{}
	if err := json.NewDecoder(r.Body).Decode(&delReq); err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "bad json")
		return
	}

	if delReq.Offset < 0 {
		jsonifyErrf(w, http.StatusBadRequest, "negative offset")
		return
	}

	nodes, err := dh.fs.DeletedNodes("/")
	if err != nil {
		jsonifyErrf(w, http.StatusInternalServerError, "failed to list")
		return
	}

	filter := strings.ToLower(delReq.Filter)
	filteredNodes := []*catfs.StatInfo{}
	for _, node := range nodes {
		if !matchEntry(node, filter) {
			continue
		}

		filteredNodes = append(filteredNodes, node)
	}

	entries := []*StatInfo{}
	if delReq.Offset >= int64(len(filteredNodes)) {
		jsonify(w, http.StatusOK, &DeletedPathsResponse{
			Success: true,
			Entries: entries,
		})
		return
	}

	filteredNodes = filteredNodes[delReq.Offset:]
	if delReq.Limit >= 0 {
		filteredNodes = filteredNodes[:util.Min64(int64(len(filteredNodes)), delReq.Limit)]
	}

	for _, node := range filteredNodes {
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
