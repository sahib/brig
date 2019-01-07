package endpoints

import (
	"encoding/json"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/sahib/brig/catfs"
	"github.com/sahib/config"
)

type MoveHandler struct {
	cfg *config.Config
	fs  *catfs.FS
}

func NewMoveHandler(cfg *config.Config, fs *catfs.FS) *MoveHandler {
	return &MoveHandler{
		cfg: cfg,
		fs:  fs,
	}
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

	if !validateUserForPath(mh.cfg, moveReq.Source, r) {
		jsonifyErrf(w, http.StatusUnauthorized, "source path forbidden")
		return
	}

	if !validateUserForPath(mh.cfg, moveReq.Destination, r) {
		jsonifyErrf(w, http.StatusUnauthorized, "destination path forbidden")
		return
	}

	if err := mh.fs.Move(moveReq.Source, moveReq.Destination); err != nil {
		log.Debugf("failed to move %s -> %s: %v", moveReq.Source, moveReq.Destination, err)
		jsonifyErrf(w, http.StatusInternalServerError, "failed to move")
		return
	}

	jsonifySuccess(w)
}
