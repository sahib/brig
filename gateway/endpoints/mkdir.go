package endpoints

import (
	"encoding/json"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"
)

// MkdirHandler implements http.Handler.
type MkdirHandler struct {
	*State
}

// NewMkdirHandler creates a new mkdir handler.
func NewMkdirHandler(s *State) *MkdirHandler {
	return &MkdirHandler{State: s}
}

// MkdirRequest is the request that can be sent to this endpoint as JSON.
type MkdirRequest struct {
	// Path to create.
	Path string `json:"path"`
}

func (mh *MkdirHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mkdirReq := &MkdirRequest{}
	if err := json.NewDecoder(r.Body).Decode(&mkdirReq); err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "bad json")
		return
	}

	if !strings.HasPrefix(mkdirReq.Path, "/") {
		jsonifyErrf(w, http.StatusBadRequest, "absolute path needs to start with /")
		return
	}

	if !mh.validatePath(mkdirReq.Path, w, r) {
		jsonifyErrf(w, http.StatusUnauthorized, "path forbidden")
		return
	}

	if err := mh.fs.Mkdir(mkdirReq.Path, true); err != nil {
		log.Debugf("failed to mkdir %s: %v", mkdirReq.Path, err)
		jsonifyErrf(w, http.StatusInternalServerError, "failed to mkdir")
		return
	}

	mh.evHdl.Notify(r.Context(), "fs")
	jsonifySuccess(w)
}
