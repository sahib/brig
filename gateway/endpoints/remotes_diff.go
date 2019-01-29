package endpoints

import (
	"encoding/json"
	"net/http"

	"github.com/sahib/brig/catfs"
)

// RemotesDiffHandler implements http.Handler
type RemotesDiffHandler struct {
	*State
}

// NewRemotesDiffHandler returns a new RemotesDiffHandler
func NewRemotesDiffHandler(s *State) *RemotesDiffHandler {
	return &RemotesDiffHandler{State: s}
}

// RemoteDiffRequest is the data being sent to this endpoint.
type RemoteDiffRequest struct {
	Name string `json:"name"`
}

// RemoteDiffResponse is the data being sent to this endpoint.
type RemoteDiffResponse struct {
	Success bool        `json:"success"`
	Diff    *catfs.Diff `json:"diff"`
}

func (rh *RemotesDiffHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rmtDiffReq := RemoteDiffRequest{}
	if err := json.NewDecoder(r.Body).Decode(&rmtDiffReq); err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "bad json")
		return
	}

	diff, err := rh.rapi.MakeDiff(rmtDiffReq.Name)
	if err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "failed to sync")
		return
	}

	jsonify(w, http.StatusOK, RemoteDiffResponse{
		Success: true,
		Diff:    diff,
	})
}
