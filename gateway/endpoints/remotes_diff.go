package endpoints

import (
	"encoding/json"
	"net/http"

	"github.com/sahib/brig/catfs"
	"github.com/sahib/brig/gateway/db"
)

// RemotesDiffHandler implements http.Handler
type RemotesDiffHandler struct {
	*State
}

// NewRemotesDiffHandler returns a new RemotesDiffHandler
func NewRemotesDiffHandler(s *State) *RemotesDiffHandler {
	return &RemotesDiffHandler{State: s}
}

// RemoteDiffRequest is the data being sent to this endpoint.
type RemoteDiffRequest struct {
	Name string `json:"name"`
}

type DiffPair struct {
	Src *StatInfo `json:"src"`
	Dst *StatInfo `json:"dst"`
}

type Diff struct {
	Added    []*StatInfo `json:"added"`
	Removed  []*StatInfo `json:"removed"`
	Ignored  []*StatInfo `json:"ignored"`
	Missing  []*StatInfo `json:"missing"`
	Conflict []DiffPair  `json:"conflict"`
	Moved    []DiffPair  `json:"moved"`
	Merged   []DiffPair  `json:"merged"`
}

// RemoteDiffResponse is the data being sent to this endpoint.
type RemoteDiffResponse struct {
	Success bool  `json:"success"`
	Diff    *Diff `json:"diff"`
}

func convertSingles(infos []catfs.StatInfo) []*StatInfo {
	result := []*StatInfo{}
	for _, info := range infos {
		result = append(result, toExternalStatInfo(&info))
	}

	return result
}

func convertPairs(pairs []catfs.DiffPair) []DiffPair {
	result := []DiffPair{}
	for _, pair := range pairs {
		result = append(result, DiffPair{
			Src: toExternalStatInfo(&pair.Src),
			Dst: toExternalStatInfo(&pair.Dst),
		})
	}

	return result

}

func (rh *RemotesDiffHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !checkRights(w, r, db.RightRemotesView) {
		return
	}

	rmtDiffReq := RemoteDiffRequest{}
	if err := json.NewDecoder(r.Body).Decode(&rmtDiffReq); err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "bad json")
		return
	}

	if rmtDiffReq.Name == "" {
		jsonifyErrf(w, http.StatusBadRequest, "empty remote name")
		return
	}

	rawDiff, err := rh.rapi.MakeDiff(rmtDiffReq.Name)
	if err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "failed to diff")
		return
	}

	diff := &Diff{
		Added:    convertSingles(rawDiff.Added),
		Removed:  convertSingles(rawDiff.Removed),
		Ignored:  convertSingles(rawDiff.Ignored),
		Missing:  convertSingles(rawDiff.Missing),
		Conflict: convertPairs(rawDiff.Conflict),
		Moved:    convertPairs(rawDiff.Moved),
		Merged:   convertPairs(rawDiff.Merged),
	}

	jsonify(w, http.StatusOK, RemoteDiffResponse{
		Success: true,
		Diff:    diff,
	})
}
