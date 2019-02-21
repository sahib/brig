package endpoints

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/sahib/brig/gateway/db"
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
	if !checkRights(w, r, db.RightFsEdit) {
		return
	}

	mkdirReq := MkdirRequest{}
	if err := json.NewDecoder(r.Body).Decode(&mkdirReq); err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "bad json")
		return
	}

	path := prefixRoot(mkdirReq.Path)
	if !mh.validatePath(path, w, r) {
		jsonifyErrf(w, http.StatusUnauthorized, "path forbidden")
		return
	}

	if err := mh.fs.Mkdir(path, true); err != nil {
		log.Debugf("failed to mkdir %s: %v", path, err)
		jsonifyErrf(w, http.StatusInternalServerError, "failed to mkdir")
		return
	}

	msg := fmt.Sprintf("mkdir'd »%s«", path)
	if !mh.commitChange(msg, w, r) {
		return
	}
	jsonifySuccess(w)
}
