package endpoints

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/sahib/brig/gateway/db"
)

// MoveHandler implements http.Handler.
type MoveHandler struct {
	*State
}

// NewMoveHandler creates a new move handler.
func NewMoveHandler(s *State) *MoveHandler {
	return &MoveHandler{State: s}
}

// MoveRequest is the request that can be send to this endpoint.
type MoveRequest struct {
	// Source is the path to the old node.
	Source string `json:"source"`
	// Destination is the path of the new node.
	Destination string `json:"destination"`
}

func (mh *MoveHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !checkRights(w, r, db.RightFsEdit) {
		return
	}

	moveReq := MoveRequest{}
	if err := json.NewDecoder(r.Body).Decode(&moveReq); err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "bad json")
		return
	}

	src := prefixRoot(moveReq.Source)
	dst := prefixRoot(moveReq.Destination)

	if !mh.validatePath(src, w, r) {
		jsonifyErrf(w, http.StatusUnauthorized, "source path forbidden")
		return
	}

	if !mh.validatePath(dst, w, r) {
		jsonifyErrf(w, http.StatusUnauthorized, "destination path forbidden")
		return
	}

	// Move does some extended checking before actually moving the file:
	if err := mh.fs.Move(src, dst); err != nil {
		log.Debugf("failed to move %s -> %s: %v", src, dst, err)
		jsonifyErrf(w, http.StatusInternalServerError, "failed to move")
		return
	}

	msg := fmt.Sprintf("moved »%s« to »%s« via gateway", src, dst)
	if !mh.commitChange(msg, w, r) {
		return
	}

	jsonifySuccess(w)
}
