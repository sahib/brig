package endpoints

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/sahib/brig/catfs"
	"github.com/sahib/config"
)

type UploadHandler struct {
	cfg *config.Config
	fs  *catfs.FS
}

func NewUploadHandler(cfg *config.Config, fs *catfs.FS) *UploadHandler {
	return &UploadHandler{
		cfg: cfg,
		fs:  fs,
	}
}

func (uh *UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: Currently always put into /. How to select the right path?
	// Hold at max 1 MB in memory:
	if err := r.ParseMultipartForm(1 * 1024 * 1024); err != nil {
		log.Debugf("upload: bad multipartform: %v", err)
		jsonifyErrf(w, http.StatusBadRequest, "failed to parse mutlipart form: %v", err)
		return
	}

	for _, headers := range r.MultipartForm.File {
		for _, header := range headers {
			path := header.Filename
			fd, err := header.Open()
			if err != nil {
				log.Debugf("upload: bad header: %v", err)
				jsonifyErrf(w, http.StatusBadRequest, "failed to open file: %v", path)
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

	jsonifySuccess(w)
}
