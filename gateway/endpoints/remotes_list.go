package endpoints

import (
	"net/http"

	"github.com/sahib/brig/gateway/remotesapi"
)

// RemoteListHandler implements http.Handler
type RemoteListHandler struct {
	*State
}

// NewRemotesListHandler returns a new RemoteListHandler
func NewRemotesListHandler(s *State) *RemoteListHandler {
	return &RemoteListHandler{State: s}
}

// RemoteListResponse is the response given by this endpoint.
type RemoteListResponse struct {
	Success bool                 `json:"success"`
	Remotes []*remotesapi.Remote `json:"remotes"`
}

func (rh *RemoteListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rmts, err := rh.rapi.List()
	if err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "bad json")
		return
	}

	jsonify(w, http.StatusOK, &RemoteListResponse{
		Success: true,
		Remotes: rmts,
	})
}
