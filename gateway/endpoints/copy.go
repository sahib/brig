package endpoints

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
)

// CopyHandler implements http.Handler.
type CopyHandler struct {
	*State
}

// NewCopyHandler creates a new copy handler.
func NewCopyHandler(s *State) *CopyHandler {
	return &CopyHandler{State: s}
}

// CopyRequest is the request that can be send to this endpoint.
type CopyRequest struct {
	// Source is the path to the old node.
	Source string `json="source"`
	// Destination is the path of the new node.
	Destination string `json="destination"`
}

func (ch *CopyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	copyReq := &CopyRequest{}
	if err := json.NewDecoder(r.Body).Decode(&copyReq); err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "bad json")
		return
	}

	if !ch.validatePath(copyReq.Source, w, r) {
		jsonifyErrf(w, http.StatusUnauthorized, "source path forbidden")
		return
	}

	if !ch.validatePath(copyReq.Destination, w, r) {
		jsonifyErrf(w, http.StatusUnauthorized, "destination path forbidden")
		return
	}

	if err := ch.fs.Copy(copyReq.Source, copyReq.Destination); err != nil {
		log.Debugf("failed to copy %s -> %s: %v", copyReq.Source, copyReq.Destination, err)
		jsonifyErrf(w, http.StatusInternalServerError, "failed to copy")
		return
	}

	msg := fmt.Sprintf("copied »%s« to »%s«", copyReq.Source, copyReq.Destination)
	ch.commitChange(msg, w, r)
	jsonifySuccess(w)
}
