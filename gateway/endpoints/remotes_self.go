package endpoints

import (
	"net/http"

	"github.com/sahib/brig/gateway/db"
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
	Success                 bool                `json:"success"`
	Self                    remotesapi.Identity `json:"self"`
	DefaultConflictStrategy string              `json:"default_conflict_strategy"`
}

func (rh *RemoteSelfHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !checkRights(w, r, db.RightRemotesView) {
		return
	}

	self, err := rh.rapi.Self()
	if err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "failed to get self")
		return
	}

	jsonify(w, http.StatusOK, RemoteSelfResponse{
		Success:                 true,
		Self:                    self,
		DefaultConflictStrategy: "marker",
	})
}
