package endpoints

import (
	"net/http"
	"sort"
	"strings"

	"github.com/sahib/brig/gateway/db"
)

// AllDirsHandler implements http.Handler.
// This endpoint returns all directories that the client may see.
// It is used in the client to offer the user a list of directories
// to move or copy files to.
type AllDirsHandler struct {
	*State
}

// NewAllDirsHandler returns a new AllDirsHandler.
func NewAllDirsHandler(s *State) *AllDirsHandler {
	return &AllDirsHandler{State: s}
}

// AllDirsResponse is the response sent to the client.
type AllDirsResponse struct {
	Success bool     `json:"success"`
	Paths   []string `json:"paths"`
}

func (ah *AllDirsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if user := getUserName(ah.store, w, r); user == "" {
		jsonifyErrf(w, http.StatusForbidden, "bad user")
		return
	}

	if !checkRights(w, r, db.RightFsView) {
		return
	}

	nodes, err := ah.fs.List("/", -1)
	if err != nil {
		jsonifyErrf(w, http.StatusInternalServerError, "failed to list")
		return
	}

	paths := []string{}
	for _, node := range nodes {
		if !node.IsDir || !ah.validatePath(node.Path, w, r) {
			continue
		}

		paths = append(paths, node.Path)
	}

	// Sort dirs before files and sort each part alphabetically
	sort.Slice(paths, func(i, j int) bool {
		return strings.ToLower(paths[i]) < strings.ToLower(paths[j])
	})

	jsonify(w, http.StatusOK, &AllDirsResponse{
		Success: true,
		Paths:   paths,
	})
}
