package endpoints

import (
	"net/http"
	"sort"

	"github.com/sahib/brig/gateway/db"
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
	if !checkRights(w, r, db.RightRemotesView) {
		return
	}

	rmts, err := rh.rapi.List()
	if err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "bad json")
		return
	}

	sort.Slice(rmts, func(i, j int) bool {
		return rmts[i].Name < rmts[j].Name
	})

	jsonify(w, http.StatusOK, &RemoteListResponse{
		Success: true,
		Remotes: rmts,
	})
}
