package endpoints

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"
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
	resetReq := ResetRequest{}
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

	var err error
	if resetReq.Path == "/" {
		err = rh.fs.Checkout(resetReq.Revision, true)
	} else {
		err = rh.fs.Reset(resetReq.Path, resetReq.Revision)
	}

	log.Debugf("reset %s to %s", resetReq.Path, resetReq.Revision)
	if err != nil {
		log.Debugf("failed to reset %s to %s: %v", resetReq.Path, resetReq.Revision, err)
		jsonifyErrf(w, http.StatusInternalServerError, "failed to reset")
		return
	}

	msg := fmt.Sprintf(
		"reverted »%s« to »%s«",
		resetReq.Path,
		resetReq.Revision,
	)

	rh.commitChange(msg, w, r)
	jsonifySuccess(w)
}
