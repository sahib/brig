package endpoints

import (
	"fmt"
	"net/http"
	"path"

	log "github.com/Sirupsen/logrus"
	"github.com/sahib/brig/gateway/db"
)

// UploadHandler implements http.Handler.
type UploadHandler struct {
	*State
}

// NewUploadHandler returns a new UploadHandler.
func NewUploadHandler(s *State) *UploadHandler {
	return &UploadHandler{State: s}
}

func (uh *UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !checkRights(w, r, db.RightFsEdit) {
		fmt.Println("BAD RIGHTS")
		return
	}

	root := r.URL.Query().Get("root")
	if root == "" {
		root = "/"
	} else {
		root = prefixRoot(root)
	}

	if err := r.ParseMultipartForm(1 * 1024 * 1024); err != nil {
		log.Debugf("upload: bad multipartform: %v", err)
		jsonifyErrf(w, http.StatusBadRequest, "failed to parse mutlipart form: %v", err)
		return
	}

	// Remove the cached files in /tmp
	defer r.MultipartForm.RemoveAll()

	paths := []string{}

	for _, headers := range r.MultipartForm.File {
		for _, header := range headers {
			path := path.Join(root, header.Filename)
			fd, err := header.Open()
			if err != nil {
				log.Debugf("upload: bad header: %v", err)
				jsonifyErrf(w, http.StatusBadRequest, "failed to open file: %v", header.Filename)
				return
			}

			if !uh.validatePath(path, w, r) {
				jsonifyErrf(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			if err := uh.fs.Stage(path, fd); err != nil {
				log.Debugf("upload: could not stage: %v", err)
				jsonifyErrf(w, http.StatusBadRequest, "failed to insert file: %v", path)
				fd.Close()
				return
			}

			paths = append(paths, path)
			fd.Close()
		}
	}

	if len(paths) > 0 {
		msg := fmt.Sprintf("uploaded »%s«", paths[0])
		if len(paths) > 1 {
			msg += fmt.Sprintf(" and %d more", len(paths)-1)
		}

		if !uh.commitChange(msg, w, r) {
			return
		}
	}

	jsonifySuccess(w)
}
