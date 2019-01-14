package endpoints

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/sahib/brig/catfs"
	"github.com/sahib/brig/events"
	"github.com/sahib/config"
)

type State struct {
	fs    *catfs.FS
	cfg   *config.Config
	ev    *events.Listener
	evHdl *EventsHandler
	store *sessions.CookieStore
}

func readOrInitKeyFromConfig(cfg *config.Config, keyName string, keyLen int) ([]byte, error) {
	keyStr := cfg.String(keyName)
	if keyStr == "" {
		keyData := securecookie.GenerateRandomKey(keyLen)
		cfg.SetString(keyName, base64.StdEncoding.EncodeToString(keyData))
		return keyData, nil
	}

	return base64.StdEncoding.DecodeString(keyStr)
}

func NewState(fs *catfs.FS, cfg *config.Config, ev *events.Listener, evHdl *EventsHandler) (*State, error) {
	authKey, err := readOrInitKeyFromConfig(cfg, "auth.session-authentication-key", 64)
	if err != nil {
		return nil, err
	}

	encKey, err := readOrInitKeyFromConfig(cfg, "auth.session-encryption-key", 32)
	if err != nil {
		return nil, err
	}

	// Generated here, but used by the server:
	_, err = readOrInitKeyFromConfig(cfg, "auth.session-csrf-key", 32)
	if err != nil {
		return nil, err
	}

	return &State{
		fs:    fs,
		cfg:   cfg,
		ev:    ev,
		evHdl: evHdl,
		store: sessions.NewCookieStore(authKey, encKey),
	}, nil
}

func (s *State) publishFsEvent(req *http.Request) {
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
