package endpoints

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/sahib/brig/gateway/db"
	"github.com/sahib/brig/gateway/remotesapi"
	"github.com/sahib/brig/net/peer"
	log "github.com/sirupsen/logrus"
)

// RemotesAddHandler implements http.Handler
type RemotesAddHandler struct {
	*State
}

// NewRemotesAddHandler returns a new RemotesAddHandler
func NewRemotesAddHandler(s *State) *RemotesAddHandler {
	return &RemotesAddHandler{State: s}
}

// RemoteAddRequest is the data being sent to this endpoint.
type RemoteAddRequest struct {
	Name              string              `json:"name"`
	Folders           []remotesapi.Folder `json:"folders"`
	Fingerprint       string              `json:"fingerprint"`
	AcceptAutoUpdates bool                `json:"accept_auto_updates"`
	AcceptPush        bool                `json:"accept_push"`
	ConflictStrategy  string              `json:"conflict_strategy"`
}

func dedupeFolders(folders []remotesapi.Folder) []remotesapi.Folder {
	seen := make(map[string]bool)
	deduped := []remotesapi.Folder{}

	for _, folder := range folders {
		path := folder.Folder
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}

		if seen[path] {
			continue
		}

		deduped = append(deduped, folder)
		seen[path] = true
	}

	return deduped
}

func validateFingerprint(fingerprint string, w http.ResponseWriter, r *http.Request) bool {
	if _, err := peer.CastFingerprint(fingerprint); err != nil {
		log.Debugf("invalid fingerprint: %v", err)
		jsonifyErrf(w, http.StatusBadRequest, "bad fingerprint format")
		return false
	}

	return true
}

func readRemoteRequest(w http.ResponseWriter, r *http.Request) (*remotesapi.Remote, error) {
	if !checkRights(w, r, db.RightRemotesEdit) {
		return nil, fmt.Errorf("bad rights")
	}

	remoteAddReq := RemoteAddRequest{}
	if err := json.NewDecoder(r.Body).Decode(&remoteAddReq); err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "bad json")
		return nil, fmt.Errorf("bad json")
	}

	if !validateFingerprint(remoteAddReq.Fingerprint, w, r) {
		return nil, fmt.Errorf("bad fingerprint")
	}

	return &remotesapi.Remote{
		Name:              remoteAddReq.Name,
		Folders:           dedupeFolders(remoteAddReq.Folders),
		Fingerprint:       remoteAddReq.Fingerprint,
		AcceptAutoUpdates: remoteAddReq.AcceptAutoUpdates,
		AcceptPush:        remoteAddReq.AcceptPush,
		ConflictStrategy:  remoteAddReq.ConflictStrategy,
	}, nil
}

func (rh *RemotesAddHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rmt, err := readRemoteRequest(w, r)
	if err != nil {
		return
	}

	if _, err := rh.rapi.Get(rmt.Name); err == nil {
		jsonifyErrf(w, http.StatusBadRequest, "remote does exist already")
		return
	}

	if err := rh.rapi.Set(*rmt); err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "failed to add")
		return
	}

	jsonifySuccess(w)
}

//////////////

// RemotesModifyHandler implements http.Handler
type RemotesModifyHandler struct {
	*State
}

// NewRemotesModifyHandler returns a new RemotesModifyHandler
func NewRemotesModifyHandler(s *State) *RemotesModifyHandler {
	return &RemotesModifyHandler{State: s}
}

func (rh *RemotesModifyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rmt, err := readRemoteRequest(w, r)
	if err != nil {
		return
	}

	if _, err := rh.rapi.Get(rmt.Name); err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "remote does not exist yet")
		return
	}

	if err := rh.rapi.Set(*rmt); err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "failed to add")
		return
	}

	jsonifySuccess(w)
}
