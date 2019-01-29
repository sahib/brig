package endpoints

import (
	"encoding/json"
	"net/http"
)

// RemotesSyncHandler implements http.Handler
type RemotesSyncHandler struct {
	*State
}

// NewRemotesSyncHandler returns a new RemotesSyncHandler
func NewRemotesSyncHandler(s *State) *RemotesSyncHandler {
	return &RemotesSyncHandler{State: s}
}

// RemoteSyncRequest is the data being sent to this endpoint.
type RemoteSyncRequest struct {
	Name string `json:"name"`
}

func (rh *RemotesSyncHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rmtSyncReq := RemoteSyncRequest{}
	if err := json.NewDecoder(r.Body).Decode(&rmtSyncReq); err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "bad json")
		return
	}

	if err := rh.rapi.Sync(rmtSyncReq.Name); err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "failed to sync")
		return
	}

	jsonifySuccess(w)
}
