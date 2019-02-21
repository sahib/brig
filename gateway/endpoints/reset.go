package endpoints

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/sahib/brig/gateway/db"
)

// ResetHandler implements http.Handler.
type ResetHandler struct {
	*State
}

// NewResetHandler returns a new ResetHandler.
func NewResetHandler(s *State) *ResetHandler {
	return &ResetHandler{State: s}
}

// ResetRequest is a request sent to this endpoint.
type ResetRequest struct {
	Path     string `json:"path"`
	Revision string `json:"revision"`
	Force    bool   `json:"force"`
}

func (rh *ResetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !checkRights(w, r, db.RightFsEdit) {
		return
	}

	resetReq := ResetRequest{}
	if err := json.NewDecoder(r.Body).Decode(&resetReq); err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "bad json")
		return
	}

	path := prefixRoot(resetReq.Path)
	if !rh.validatePath(path, w, r) {
		jsonifyErrf(w, http.StatusUnauthorized, "path forbidden")
		return
	}

	var err error
	if resetReq.Path == "/" {
		err = rh.fs.Checkout(resetReq.Revision, true)
	} else {
		err = rh.fs.Reset(path, resetReq.Revision)
	}

	log.Debugf("reset %s to %s", path, resetReq.Revision)
	if err != nil {
		log.Debugf("failed to reset %s to %s: %v", path, resetReq.Revision, err)
		jsonifyErrf(w, http.StatusInternalServerError, "failed to reset")
		return
	}

	msg := fmt.Sprintf("reverted »%s« to »%s«", path, resetReq.Revision)
	if !rh.commitChange(msg, w, r) {
		return
	}

	jsonifySuccess(w)
}
