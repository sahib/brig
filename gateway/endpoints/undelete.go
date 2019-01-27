package endpoints

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
)

// UndeleteHandler implements http.Handler.
type UndeleteHandler struct {
	*State
}

// NewUndeleteHandler creates a new undelete handler.
func NewUndeleteHandler(s *State) *UndeleteHandler {
	return &UndeleteHandler{State: s}
}

// UndeleteRequest is the request that can be sent to this endpoint as JSON.
type UndeleteRequest struct {
	// Path to create.
	Path string `json:"path"`
}

func (uh *UndeleteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	undelReq := UndeleteRequest{}
	if err := json.NewDecoder(r.Body).Decode(&undelReq); err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "bad json")
		return
	}

	if !uh.validatePath(undelReq.Path, w, r) {
		jsonifyErrf(w, http.StatusUnauthorized, "path forbidden")
		return
	}

	if err := uh.fs.Undelete(undelReq.Path); err != nil {
		log.Debugf("failed to undelete %s: %v", undelReq.Path, err)
		fmt.Println(err)
		jsonifyErrf(w, http.StatusInternalServerError, "failed to undelete")
		return
	}

	msg := fmt.Sprintf("undeleted »%s«", undelReq.Path)
	uh.commitChange(msg, w, r)
	jsonifySuccess(w)
}
