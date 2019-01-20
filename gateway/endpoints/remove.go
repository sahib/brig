package endpoints

import (
	"encoding/json"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"
)

// RemoveHandler implements http.Handler.
type RemoveHandler struct {
	*State
}

// NewRemoveHandler returns a new RemoveHandler
func NewRemoveHandler(s *State) *RemoveHandler {
	return &RemoveHandler{State: s}
}

// RemoveRequest is the request that is being sent to the endpoint.
type RemoveRequest struct {
	Paths []string `json:"paths"`
}

func (rh *RemoveHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rmReq := &RemoveRequest{}
	if err := json.NewDecoder(r.Body).Decode(&rmReq); err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "bad json")
		return
	}

	for _, path := range rmReq.Paths {
		if !strings.HasPrefix(path, "/") {
			jsonifyErrf(w, http.StatusBadRequest, "bad path: %s (not absolute)", path)
			return
		}

		if !rh.validatePath(path, w, r) {
			jsonifyErrf(w, http.StatusUnauthorized, "path forbidden")
			return
		}
	}

	hasChanged := false

	for _, path := range rmReq.Paths {
		if err := rh.fs.Remove(path); err != nil {
			log.Debugf("failed to remove %s: %v", path, err)
			jsonifyErrf(w, http.StatusBadRequest, "failed to remove")
			return
		}

		hasChanged = true
	}

	if hasChanged {
		rh.evHdl.Notify(r.Context(), "fs")
	}

	jsonifySuccess(w)
}
