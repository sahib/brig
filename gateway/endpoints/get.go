package endpoints

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/sahib/brig/catfs"
	ie "github.com/sahib/brig/catfs/errors"
	"github.com/sahib/brig/catfs/mio"
	"github.com/sahib/brig/util"
	"github.com/sahib/config"
)

type GetHandler struct {
	cfg *config.Config
	fs  *catfs.FS
}

func NewGetHandler(cfg *config.Config, fs *catfs.FS) *GetHandler {
	return &GetHandler{
		cfg: cfg,
		fs:  fs,
	}
}

func mimeTypeFromStream(stream mio.Stream) (io.ReadSeeker, string) {
	hdr, newStream, err := util.PeekHeader(stream, 512)
	if err != nil {
		return stream, "application/octet-stream"
	}

	return newStream, http.DetectContentType(hdr)
}

// setContentDisposition sets the Content-Disposition header, based on
// the content we are serving. It tells a browser if it should open
// a save dialog or display it inline (and how)
func setContentDisposition(info *catfs.StatInfo, hdr http.Header, dispoType string) {
	basename := path.Base(info.Path)
	if info.IsDir {
		if basename == "/" {
			basename = "root"
		}

		basename += ".tar"
	}

	hdr.Set(
		"Content-Disposition",
		fmt.Sprintf(
			"%s; filename*=UTF-8''%s",
			dispoType,
			url.QueryEscape(basename),
		),
	)
}

func (gh *GetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// get the file nodePath including the leading slash:
	fullURL := r.URL.EscapedPath()
	nodePath, err := url.PathUnescape(fullURL[4:])
	if err != nil {
		log.Debugf("received malformed url: %s", fullURL)
		http.Error(w, "malformed url", http.StatusBadRequest)
		return
	}

	if !validateUserForPath(gh.cfg, nodePath, r) {
		http.Error(w, "not authorized", http.StatusUnauthorized)
		return
	}

	info, err := gh.fs.Stat(nodePath)
	if err != nil {
		// Handle a bad nodePath more explicit:
		if ie.IsNoSuchFileError(err) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		log.Errorf("gateway: failed to stat %s: %v", nodePath, err)
		http.Error(w, "failed to stat file", http.StatusInternalServerError)
		return
	}

	hdr := w.Header()
	hdr.Set("ETag", info.ContentHash.B58String())
	hdr.Set("Last-Modified", info.ModTime.Format(http.TimeFormat))

	if info.IsDir {
		setContentDisposition(info, hdr, "attachment")
		if err := gh.fs.Tar(nodePath, w); err != nil {
			log.Errorf("gateway: failed to stream %s: %v", nodePath, err)
			http.Error(w, "failed to stream", http.StatusInternalServerError)
			return
		}
	} else {
		stream, err := gh.fs.Cat(nodePath)
		if err != nil {
			log.Errorf("gateway: failed to stream %s: %v", nodePath, err)
			http.Error(w, "failed to stream", http.StatusInternalServerError)
			return
		}

		prefixStream, mimeType := mimeTypeFromStream(stream)
		hdr.Set("Content-Type", mimeType)
		hdr.Set("Content-Length", strconv.FormatUint(info.Size, 10))

		// Set the content disposition to inline if it looks like something viewable.
		if mimeType == "application/octet-stream" {
			setContentDisposition(info, hdr, "attachment")
		} else {
			setContentDisposition(info, hdr, "inline")
		}

		http.ServeContent(w, r, path.Base(info.Path), info.ModTime, prefixStream)
	}
}
