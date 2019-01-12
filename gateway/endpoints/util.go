package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/sahib/brig/catfs"
	"github.com/sahib/brig/events"
	"github.com/sahib/config"
)

type State struct {
	fs    *catfs.FS
	cfg   *config.Config
	ev    *events.Listener
	evHdl *EventsHandler
}

func NewState(fs *catfs.FS, cfg *config.Config, ev *events.Listener, evHdl *EventsHandler) State {
	return State{
		fs:    fs,
		cfg:   cfg,
		ev:    ev,
		evHdl: evHdl,
	}
}

func (s State) publishFsEvent(req *http.Request) {
	if s.evHdl != nil {
		ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
		defer cancel()

		s.evHdl.Notify("fs", ctx)
	}

	if s.ev == nil {
		return
	}

	log.Debugf("publishing fs event from gateway")
	ev := events.Event{
		Type: events.FsEvent,
	}

	if err := s.ev.PublishEvent(ev); err != nil {
		log.Warningf("failed to publish filesystem change event: %v", err)
	}
}

func jsonify(w http.ResponseWriter, statusCode int, data interface{}) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Warningf("failed to encode json: %v", err)
		w.Write([]byte(
			"{\"success\": false, \"message\": \"failed to encode json response\"}",
		))
		w.WriteHeader(500)
		return
	}
}

func jsonifyErrf(w http.ResponseWriter, statusCode int, format string, data ...interface{}) {
	msg := fmt.Sprintf(format, data...)
	success := false
	if statusCode >= 200 && statusCode < 400 {
		success = true
	} else {
		log.Debugf("failed to respond: %v", msg)
	}

	jsonify(w, statusCode, struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}{
		Success: success,
		Message: msg,
	})
}

func jsonifySuccess(w http.ResponseWriter) {
	jsonifyErrf(w, http.StatusOK, "success")
}
