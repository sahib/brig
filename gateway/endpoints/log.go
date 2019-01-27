package endpoints

import (
	"net/http"
)

// LogHandler implements http.Handler.
type LogHandler struct {
	*State
}

// NewLogHandler returns a new LogHandler
func NewLogHandler(s *State) *LogHandler {
	return &LogHandler{State: s}
}

// LogResponse is the response sent back to the client.
type LogResponse struct {
	Success bool     `json:"success"`
	Commits []Commit `json:"commits"`
}

func (lh *LogHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cmts, err := lh.fs.Log()
	if err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "failed to query log: %v", err)
		return
	}

	exts := []Commit{}
	for _, cmt := range cmts {
		exts = append(exts, toExternalCommit(&cmt))
	}

	jsonify(w, http.StatusOK, &LogResponse{
		Success: true,
		Commits: exts,
	})
}
