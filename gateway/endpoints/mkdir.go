package endpoints

import (
	"encoding/json"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/sahib/brig/catfs"
	"github.com/sahib/config"
)

type MkdirHandler struct {
	cfg *config.Config
	fs  *catfs.FS
}

func NewMkdirHandler(cfg *config.Config, fs *catfs.FS) *MkdirHandler {
	return &MkdirHandler{
		cfg: cfg,
		fs:  fs,
	}
}

type MkdirRequest struct {
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

	if !validateUserForPath(mh.cfg, mkdirReq.Path, r) {
		jsonifyErrf(w, http.StatusUnauthorized, "path forbidden")
		return
	}

	if err := mh.fs.Mkdir(mkdirReq.Path, true); err != nil {
		log.Debugf("failed to mkdir %s: %v", mkdirReq.Path, err)
		jsonifyErrf(w, http.StatusInternalServerError, "failed to mkdir")
		return
	}

	jsonifySuccess(w)
}
