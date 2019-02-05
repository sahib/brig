package endpoints

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
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
	moveReq := MoveRequest{}
	if err := json.NewDecoder(r.Body).Decode(&moveReq); err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "bad json")
		return
	}

	if !mh.validatePath(moveReq.Source, w, r) {
		jsonifyErrf(w, http.StatusUnauthorized, "source path forbidden")
		return
	}

	if !mh.validatePath(moveReq.Destination, w, r) {
		jsonifyErrf(w, http.StatusUnauthorized, "destination path forbidden")
		return
	}

	if err := mh.fs.Move(moveReq.Source, moveReq.Destination); err != nil {
		log.Debugf("failed to move %s -> %s: %v", moveReq.Source, moveReq.Destination, err)
		jsonifyErrf(w, http.StatusInternalServerError, "failed to move")
		return
	}

	msg := fmt.Sprintf("moved »%s« to »%s« via gateway", moveReq.Source, moveReq.Destination)
	if !mh.commitChange(msg, w, r) {
		return
	}

	jsonifySuccess(w)
}
