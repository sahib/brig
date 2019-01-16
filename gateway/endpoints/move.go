package endpoints

import (
	"encoding/json"
	"net/http"

	log "github.com/Sirupsen/logrus"
)

type MoveHandler struct {
	*State
}

func NewMoveHandler(s *State) *MoveHandler {
	return &MoveHandler{State: s}
}

type MoveRequest struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
}

func (mh *MoveHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	moveReq := &MoveRequest{}
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

	mh.evHdl.Notify("fs", r.Context())
	jsonifySuccess(w)
}
