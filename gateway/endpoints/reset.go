package endpoints

import (
	"encoding/json"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"
)

type ResetHandler struct {
	*State
}

func NewResetHandler(s *State) *ResetHandler {
	return &ResetHandler{State: s}
}

type ResetRequest struct {
	Path     string `json:"path"`
	Revision string `json:"revision"`
}

func (rh *ResetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resetReq := &ResetRequest{}
	if err := json.NewDecoder(r.Body).Decode(&resetReq); err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "bad json")
		return
	}

	if !strings.HasPrefix(resetReq.Path, "/") {
		jsonifyErrf(w, http.StatusBadRequest, "absolute path needs to start with /")
		return
	}

	if !rh.validatePath(resetReq.Path, w, r) {
		jsonifyErrf(w, http.StatusUnauthorized, "path forbidden")
		return
	}

	// TODO: Is that a problem when the "new" path (after reset)
	// lies in a forbidden zone? It would be at least confusing for the user.

	if err := rh.fs.Reset(resetReq.Path, resetReq.Revision); err != nil {
		log.Debugf("failed to reset %s to %s: %v", resetReq.Path, resetReq.Revision, err)
		jsonifyErrf(w, http.StatusInternalServerError, "failed to reset")
		return
	}

	if err := rh.fs.MakeCommit("reset of '%s' to '%s' via gateway"); err != nil {
		log.Debugf("failed to make commit: %v", err)
		jsonifyErrf(w, http.StatusInternalServerError, "failed to commit after reset")
		return
	}

	rh.evHdl.Notify("fs", r.Context())
	jsonifySuccess(w)
}
