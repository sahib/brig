package endpoints

import (
	"encoding/json"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/sahib/brig/catfs"
	"github.com/sahib/brig/gateway/db"
)

// HistoryHandler implements http.Handler
type HistoryHandler struct {
	*State
}

// NewHistoryHandler returns a new HistoryHandler
func NewHistoryHandler(s *State) *HistoryHandler {
	return &HistoryHandler{State: s}
}

// HistoryRequest is the request sent to this endpoint.
type HistoryRequest struct {
	Path string `json:"path"`
}

// Commit is the same as catfs.Commit, but JSON friendly
// and with some omitted fields that are not used by the client.
type Commit struct {
	Date  int64    `json:"date"`
	Msg   string   `json:"msg"`
	Tags  []string `json:"tags"`
	Hash  string   `json:"hash"`
	Index int64    `json:"index"`
}

// HistoryEntry is one entry in the response.
type HistoryEntry struct {
	Head   Commit `json:"head"`
	Path   string `json:"path"`
	Change string `json:"change"`
}

// HistoryResponse is the data that is sent back to the client.
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
	ext.Index = cmt.Index

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
	if !checkRights(w, r, db.RightFsView) {
		return
	}

	histReq := HistoryRequest{}
	if err := json.NewDecoder(r.Body).Decode(&histReq); err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "bad json")
		return
	}

	path := prefixRoot(histReq.Path)
	if !hh.validatePath(path, w, r) {
		jsonifyErrf(w, http.StatusUnauthorized, "path forbidden")
		return
	}

	hist, err := hh.fs.History(path)
	if err != nil {
		log.Debugf("failed to check history for %s: %v", path, err)
		jsonifyErrf(w, http.StatusBadRequest, "failed to check history")
		return
	}

	entries := []HistoryEntry{}
	for _, change := range hist {
		// Filter none changes, since they are only neat for debugging.
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
