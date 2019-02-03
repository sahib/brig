package endpoints

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/sahib/brig/catfs"
)

// LogHandler implements http.Handler.
type LogHandler struct {
	*State
}

// NewLogHandler returns a new LogHandler
func NewLogHandler(s *State) *LogHandler {
	return &LogHandler{State: s}
}

// LogRequest is the data sent to this endpoint.
type LogRequest struct {
	Offset int64  `json:"offset"`
	Limit  int64  `json:"limit"`
	Filter string `json:"filter"`
}

// LogResponse is the response sent back to the client.
type LogResponse struct {
	Success bool     `json:"success"`
	Commits []Commit `json:"commits"`
}

func matchCommit(cmt *catfs.Commit, filter string) bool {
	return strings.Contains(strings.ToLower(cmt.Msg), filter)
}

func (lh *LogHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logReq := LogRequest{}
	if err := json.NewDecoder(r.Body).Decode(&logReq); err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "bad json")
		return
	}

	status, err := lh.fs.CommitInfo("curr")
	if err != nil {
		jsonifyErrf(w, http.StatusInternalServerError, "could not get status")
		return
	}

	if logReq.Offset < 0 {
		jsonifyErrf(w, http.StatusBadRequest, "negative offsets are not supported")
		return
	}

	if status.Index < logReq.Offset {
		jsonify(w, http.StatusOK, &LogResponse{
			Success: true,
			Commits: []Commit{},
		})
		return
	}

	commits := []Commit{}
	errSkip := errors.New("skip")
	filter := strings.ToLower(logReq.Filter)
	head := fmt.Sprintf("commit[%d]", -(logReq.Offset + 1))

	err = lh.fs.Log(head, func(cmt *catfs.Commit) error {
		if filter != "" && !matchCommit(cmt, filter) {
			return nil
		}

		if int64(len(commits)) >= logReq.Limit {
			return errSkip
		}

		commits = append(commits, toExternalCommit(cmt))
		return nil
	})

	if err != nil && err != errSkip {
		jsonifyErrf(w, http.StatusBadRequest, "failed to query log: %v", err)
		return
	}

	jsonify(w, http.StatusOK, &LogResponse{
		Success: true,
		Commits: commits,
	})
}
