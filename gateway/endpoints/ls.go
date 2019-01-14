package endpoints

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	"github.com/sahib/brig/catfs"
)

type LsHandler struct {
	*State
}

func NewLsHandler(s *State) *LsHandler {
	return &LsHandler{State: s}
}

type LsRequest struct {
	Root   string `json:"root"`
	Filter string `json:"filter,omitempty"`
}

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

type LsResponse struct {
	Success    bool        `json:"success"`
	Self       *StatInfo   `json:"self"`
	IsFiltered bool        `json:"is_filtered"`
	Files      []*StatInfo `json:"files"`
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

	if !validateUserForPath(lh.store, lh.cfg, lsReq.Root, w, r) {
		jsonifyErrf(w, http.StatusUnauthorized, "unauthorized")
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
