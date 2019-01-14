package endpoints

import (
	"net/http"
	"path"

	log "github.com/Sirupsen/logrus"
)

type UploadHandler struct {
	*State
}

func NewUploadHandler(s *State) *UploadHandler {
	return &UploadHandler{State: s}
}

func (uh *UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	root := r.URL.Query().Get("root")
	if root == "" {
		root = "/"
	}

	if err := r.ParseMultipartForm(1 * 1024 * 1024); err != nil {
		log.Debugf("upload: bad multipartform: %v", err)
		jsonifyErrf(w, http.StatusBadRequest, "failed to parse mutlipart form: %v", err)
		return
	}

	defer r.MultipartForm.RemoveAll()

	for _, headers := range r.MultipartForm.File {
		for _, header := range headers {
			path := path.Join(root, header.Filename)
			fd, err := header.Open()
			if err != nil {
				log.Debugf("upload: bad header: %v", err)
				jsonifyErrf(w, http.StatusBadRequest, "failed to open file: %v", header.Filename)
				return
			}

			if !validateUserForPath(uh.store, uh.cfg, path, w, r) {
				jsonifyErrf(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			if err := uh.fs.Stage(path, fd); err != nil {
				log.Debugf("upload: could not stage: %v", err)
				jsonifyErrf(w, http.StatusBadRequest, "failed to insert file: %v", path)
				fd.Close()
				return
			}

			fd.Close()
		}
	}

	uh.evHdl.Notify("fs", r.Context())
	jsonifySuccess(w)
}
