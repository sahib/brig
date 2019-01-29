package endpoints

import (
	"net/http"

	"github.com/sahib/brig/gateway/remotesapi"
)

// RemoteSelfHandler implements http.Handler
type RemoteSelfHandler struct {
	*State
}

// NewRemotesSelfHandler returns a new RemoteSelfHandler
func NewRemotesSelfHandler(s *State) *RemoteSelfHandler {
	return &RemoteSelfHandler{State: s}
}

// RemoteSelfResponse is the data being sent to this endpoint.
type RemoteSelfResponse struct {
	Success bool                `json:"success"`
	Self    remotesapi.Identity `json:"self"`
}

func (rh *RemoteSelfHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	self, err := rh.rapi.Self()
	if err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "failed to get self")
		return
	}

	jsonify(w, http.StatusOK, RemoteSelfResponse{
		Success: true,
		Self:    self,
	})
}
