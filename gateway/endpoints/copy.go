package endpoints

import (
	"encoding/json"
	"net/http"

	log "github.com/Sirupsen/logrus"
)

type CopyHandler struct {
	State
}

func NewCopyHandler(s State) *CopyHandler {
	return &CopyHandler{State: s}
}

type CopyRequest struct {
	Source      string `json="source"`
	Destination string `json="destination"`
}

func (ch *CopyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	copyReq := &CopyRequest{}
	if err := json.NewDecoder(r.Body).Decode(&copyReq); err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "bad json")
		return
	}

	if !validateUserForPath(ch.cfg, copyReq.Source, r) {
		jsonifyErrf(w, http.StatusUnauthorized, "source path forbidden")
		return
	}

	if !validateUserForPath(ch.cfg, copyReq.Destination, r) {
		jsonifyErrf(w, http.StatusUnauthorized, "destination path forbidden")
		return
	}

	if err := ch.fs.Copy(copyReq.Source, copyReq.Destination); err != nil {
		log.Debugf("failed to copy %s -> %s: %v", copyReq.Source, copyReq.Destination, err)
		jsonifyErrf(w, http.StatusInternalServerError, "failed to copy")
		return
	}

	ch.evHdl.Notify("fs", r.Context())
	jsonifySuccess(w)
}
