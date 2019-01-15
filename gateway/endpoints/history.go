package endpoints

import (
	"encoding/json"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/sahib/brig/catfs"
)

type HistoryHandler struct {
	*State
}

func NewHistoryHandler(s *State) *HistoryHandler {
	return &HistoryHandler{State: s}
}

type HistoryRequest struct {
	Path string `json:"path"`
}

type Commit struct {
	Date int64    `json:"date"`
	Msg  string   `json:"msg"`
	Tags []string `json:"tags"`
	Hash string   `json:"hash"`
}

type HistoryEntry struct {
	Head   Commit `json:"head"`
	Path   string `json:"path"`
	Change string `json:"change"`
}

type HistoryResponse struct {
	Success bool           `json:"success"`
	Entries []HistoryEntry `json:"entries"`
}

func toExternalCommit(cmt *catfs.Commit) Commit {
	ext := Commit{}
	ext.Date = cmt.Date.Unix() * 1000
	ext.Hash = cmt.Hash.B58String()
	ext.Msg = cmt.Msg
	ext.Tags = cmt.Tags

	// Make sure we set an empty list,
	// otherwise .Tags gets serialized as null
	// which breaks frontend.
	if ext.Tags == nil {
		ext.Tags = []string{}
	}
	return ext
}

func toExternalChange(c catfs.Change) HistoryEntry {
	e := HistoryEntry{}
	e.Change = c.Change
	e.Head = toExternalCommit(c.Head)
	e.Path = c.Path
	return e
}

func (hh *HistoryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	histReq := &HistoryRequest{}
	if err := json.NewDecoder(r.Body).Decode(&histReq); err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "bad json")
		return
	}

	if !validateUserForPath(hh.store, hh.cfg, histReq.Path, w, r) {
		jsonifyErrf(w, http.StatusUnauthorized, "path forbidden")
		return
	}

	hist, err := hh.fs.History(histReq.Path)
	if err != nil {
		log.Debugf("failed to check history for %s: %v", histReq.Path, err)
		jsonifyErrf(w, http.StatusInternalServerError, "failed to check history")
		return
	}

	entries := []HistoryEntry{}
	for _, change := range hist {
		if change.Change == "none" {
			continue
		}

		entries = append(entries, toExternalChange(change))
	}

	jsonify(w, http.StatusOK, &HistoryResponse{
		Success: true,
		Entries: entries,
	})
}
