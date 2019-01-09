package endpoints

import (
	"encoding/json"
	"net/http"
	"strings"

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
	Paths []string `json:"paths"`
}

func (rh *RemoveHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rmReq := &RemoveRequest{}
	if err := json.NewDecoder(r.Body).Decode(&rmReq); err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "bad json")
		return
	}

	for _, path := range rmReq.Paths {
		if !strings.HasPrefix(path, "/") {
			jsonifyErrf(w, http.StatusBadRequest, "bad path: %s (not absolute)", path)
			return
		}

		if !validateUserForPath(rh.cfg, path, r) {
			jsonifyErrf(w, http.StatusUnauthorized, "path forbidden")
			return
		}
	}

	for _, path := range rmReq.Paths {
		if err := rh.fs.Remove(path); err != nil {
			log.Debugf("failed to remove %s: %v", path, err)
			jsonifyErrf(w, http.StatusBadRequest, "failed to remove")
			return
		}
	}

	jsonifySuccess(w)
}
