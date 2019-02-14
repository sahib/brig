package endpoints

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/sahib/brig/gateway/db"
)

// PinHandler implements http.Handler.
type PinHandler struct {
	*State
	doPin bool
}

// NewPinHandler returns a new PinHandler
func NewPinHandler(s *State) *PinHandler {
	return &PinHandler{State: s, doPin: true}
}

// NewUnpinHandler returns a new PinHandler
func NewUnpinHandler(s *State) *PinHandler {
	return &PinHandler{State: s, doPin: false}
}

// PinRequest is the request that is being sent to the endpoint.
type PinRequest struct {
	Paths []string `json:"paths"`
	DoPin bool     `json:"do_pin"`
}

func (ph *PinHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !checkRights(w, r, db.RightFsEdit) {
		return
	}

	pinReq := PinRequest{}
	if err := json.NewDecoder(r.Body).Decode(&pinReq); err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "bad json")
		return
	}

	for _, path := range pinReq.Paths {
		path = prefixRoot(path)
		if !ph.validatePath(path, w, r) {
			jsonifyErrf(w, http.StatusUnauthorized, "path forbidden")
			return
		}
	}

	paths := []string{}
	for _, path := range pinReq.Paths {
		path = prefixRoot(path)
		op, name := ph.fs.Pin, "pin"
		if ph.doPin == false {
			op, name = ph.fs.Unpin, "unpin"
		}

		if err := op(path); err != nil {
			log.Debugf("failed to %s %s: %v", name, path, err)
			jsonifyErrf(w, http.StatusBadRequest, fmt.Sprintf("failed to %s", name))
			return
		}

		paths = append(paths, path)
	}

	ph.evHdl.Notify(r.Context(), "fs")
	jsonifySuccess(w)
}
