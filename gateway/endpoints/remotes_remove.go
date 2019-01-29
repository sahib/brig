package endpoints

import (
	"encoding/json"
	"net/http"
)

// RemotesRemoveHandler implements http.Handler
type RemotesRemoveHandler struct {
	*State
}

// NewRemotesRemoveHandler returns a new RemotesRemoveHandler
func NewRemotesRemoveHandler(s *State) *RemotesRemoveHandler {
	return &RemotesRemoveHandler{State: s}
}

// RemoteRemoveRequest is the data being sent to this endpoint.
type RemoteRemoveRequest struct {
	Name string `json:"name"`
}

func (rh *RemotesRemoveHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rmtRmReq := RemoteRemoveRequest{}
	if err := json.NewDecoder(r.Body).Decode(&rmtRmReq); err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "bad json")
		return
	}

	if err := rh.rapi.Remove(rmtRmReq.Name); err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "failed to remove remote")
		return
	}

	jsonifySuccess(w)
}
