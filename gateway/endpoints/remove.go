package endpoints

import (
	"encoding/json"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/sahib/brig/catfs"
	"github.com/sahib/config"
)

type RemoveHandler struct {
	cfg *config.Config
	fs  *catfs.FS
}

func NewRemoveHandler(cfg *config.Config, fs *catfs.FS) *RemoveHandler {
	return &RemoveHandler{
		cfg: cfg,
		fs:  fs,
	}
}

type RemoveRequest struct {
	Path string `json:"path"`
}

func (rh *RemoveHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rmReq := &RemoveRequest{}
	if err := json.NewDecoder(r.Body).Decode(&rmReq); err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "bad json")
		return
	}

	if !validateUserForPath(rh.cfg, rmReq.Path, r) {
		jsonifyErrf(w, http.StatusUnauthorized, "path forbidden")
		return
	}

	if err := rh.fs.Remove(rmReq.Path); err != nil {
		log.Debugf("failed to remove %s: %v", rmReq.Path, err)
		jsonifyErrf(w, http.StatusInternalServerError, "failed to remove")
		return
	}

	jsonifySuccess(w)
}
