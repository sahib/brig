package endpoints

import (
	"net/http"
)

// PingHandler implements http.Handler.
// This handler checks if a user is already logged in.
type PingHandler struct {
	*State
}

// NewPingHandler returns a new PingHandler.
func NewPingHandler(s *State) *PingHandler {
	return &PingHandler{State: s}
}

// PingResponse is the response sent back by this endpoint.
type PingResponse struct {
	IsOnline bool `json:"is_online"`
}

func (wh *PingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	jsonify(w, http.StatusOK, PingResponse{
		IsOnline: true,
	})
}
