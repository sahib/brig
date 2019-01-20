package endpoints

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	"github.com/sahib/brig/catfs"
)

// LsHandler implements http.Handler.
type LsHandler struct {
	*State
}

// NewLsHandler returns a new LsHandler
func NewLsHandler(s *State) *LsHandler {
	return &LsHandler{State: s}
}

// LsRequest is the data that needs to be sent to this endpoint.
type LsRequest struct {
	Root   string `json:"root"`
	Filter string `json:"filter,omitempty"`
}

// StatInfo is a single node in the list response.
// It is the same as catfs.StatInfo, but is more JSON friendly
// and omits some fields like hashes that are not useful to the client.
type StatInfo struct {
	Path       string `json:"path"`
	User       string `json:"user"`
	Size       uint64 `json:"size"`
	Inode      uint64 `json:"inode"`
	Depth      int    `json:"depth"`
	ModTime    int64  `json:"last_modified_ms"`
	IsDir      bool   `json:"is_dir"`
	IsPinned   bool   `json:"is_pinned"`
	IsExplicit bool   `json:"is_explicit"`
}

func toExternalStatInfo(i *catfs.StatInfo) *StatInfo {
	return &StatInfo{
		Path:       i.Path,
		User:       i.User,
		Size:       i.Size,
		Inode:      i.Inode,
		Depth:      i.Depth,
		ModTime:    i.ModTime.Unix() * 1000,
		IsDir:      i.IsDir,
		IsPinned:   i.IsPinned,
		IsExplicit: i.IsExplicit,
	}
}

// LsResponse is the response sent back to the client.
type LsResponse struct {
	Success    bool        `json:"success"`
	Self       *StatInfo   `json:"self"`
	Files      []*StatInfo `json:"files"`
	IsFiltered bool        `json:"is_filtered"`
}

func doQuery(fs *catfs.FS, req *LsRequest) ([]*catfs.StatInfo, error) {
	if req.Filter == "" {
		return fs.List(req.Root, 1)
	}

	return fs.Filter(req.Root, req.Filter)
}

func (lh *LsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lsReq := &LsRequest{}
	if err := json.NewDecoder(r.Body).Decode(&lsReq); err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "bad json")
		return
	}

	info, err := lh.fs.Stat(lsReq.Root)
	if err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "failed to stat root %s: %v", lsReq.Root, err)
		return
	}

	items, err := doQuery(lh.fs, lsReq)
	if err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "failed to query: %v", err)
		return
	}

	files := []*StatInfo{}
	for _, item := range items {
		if !lh.pathIsVisible(item.Path, w, r) {
			continue
		}

		files = append(files, toExternalStatInfo(item))
	}

	// Sort dirs before files and sort each part alphabetically
	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir != files[j].IsDir {
			return files[i].IsDir
		}

		return strings.ToLower(files[i].Path) < strings.ToLower(files[j].Path)
	})

	jsonify(w, http.StatusOK, &LsResponse{
		Success:    true,
		Files:      files,
		IsFiltered: len(lsReq.Filter) > 0,
		Self:       toExternalStatInfo(info),
	})
}
