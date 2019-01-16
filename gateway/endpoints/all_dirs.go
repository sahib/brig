package endpoints

import (
	"net/http"
	"sort"
	"strings"
)

type AllDirsHandler struct {
	*State
}

func NewAllDirsHandler(s *State) *AllDirsHandler {
	return &AllDirsHandler{State: s}
}

type AllDirsResponse struct {
	Success bool     `json:"success"`
	Paths   []string `json:"paths"`
}

func (ah *AllDirsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if user := getUserName(ah.store, w, r); user == "" {
		jsonifyErrf(w, http.StatusForbidden, "bad user")
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
