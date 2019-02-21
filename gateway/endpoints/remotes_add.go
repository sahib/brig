package endpoints

import (
	"encoding/json"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/sahib/brig/gateway/db"
	"github.com/sahib/brig/gateway/remotesapi"
	"github.com/sahib/brig/net/peer"
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
	Name              string   `json:"name"`
	Folders           []string `json:"folders"`
	Fingerprint       string   `json:"fingerprint"`
	AcceptAutoUpdates bool     `json:"accept_auto_updates"`
}

func dedupeFolders(folders []string) []string {
	seen := make(map[string]bool)
	deduped := []string{}

	for _, folder := range folders {
		if !strings.HasPrefix(folder, "/") {
			folder = "/" + folder
		}

		if seen[folder] {
			continue
		}

		deduped = append(deduped, folder)
		seen[folder] = true
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

func (rh *RemotesAddHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	remoteAddReq := RemoteAddRequest{}
	if err := json.NewDecoder(r.Body).Decode(&remoteAddReq); err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "bad json")
		return
	}

	if !validateFingerprint(remoteAddReq.Fingerprint, w, r) {
		return
	}

	if _, err := rh.rapi.Get(remoteAddReq.Name); err == nil {
		jsonifyErrf(w, http.StatusBadRequest, "remote does exist already")
		return
	}

	rmt := remotesapi.Remote{
		Name:              remoteAddReq.Name,
		Folders:           dedupeFolders(remoteAddReq.Folders),
		Fingerprint:       remoteAddReq.Fingerprint,
		AcceptAutoUpdates: remoteAddReq.AcceptAutoUpdates,
	}

	if err := rh.rapi.Set(rmt); err != nil {
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
	if !checkRights(w, r, db.RightRemotesEdit) {
		return
	}

	remoteAddReq := RemoteAddRequest{}
	if err := json.NewDecoder(r.Body).Decode(&remoteAddReq); err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "bad json")
		return
	}

	if !validateFingerprint(remoteAddReq.Fingerprint, w, r) {
		return
	}

	if _, err := rh.rapi.Get(remoteAddReq.Name); err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "remote does not exist yet")
		return
	}

	rmt := remotesapi.Remote{
		Name:              remoteAddReq.Name,
		Folders:           dedupeFolders(remoteAddReq.Folders),
		Fingerprint:       remoteAddReq.Fingerprint,
		AcceptAutoUpdates: remoteAddReq.AcceptAutoUpdates,
	}

	if err := rh.rapi.Set(rmt); err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "failed to add")
		return
	}

	jsonifySuccess(w)
}
