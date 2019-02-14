package endpoints

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/sahib/brig/catfs"
	ie "github.com/sahib/brig/catfs/errors"
	"github.com/sahib/brig/events"
	"github.com/sahib/brig/gateway/db"
	"github.com/sahib/brig/gateway/remotesapi"
	"github.com/sahib/config"
)

// State is a helper struct that contains all API objects that might be useful
// to the endpoint implementation. It does not serve other purposes.
type State struct {
	fs     *catfs.FS
	rapi   remotesapi.RemotesAPI
	cfg    *config.Config
	ev     *events.Listener
	evHdl  *EventsHandler
	store  *sessions.CookieStore
	userDb *db.UserDatabase
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

// NewState creates a new state object.
// events.Listener can be set later with SetEventListener.
func NewState(
	fs *catfs.FS,
	rapi remotesapi.RemotesAPI,
	cfg *config.Config,
	evHdl *EventsHandler,
	dbPath string,
) (*State, error) {
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

	userDb, err := db.NewUserDatabase(dbPath)
	if err != nil {
		return nil, err
	}

	return &State{
		fs:     fs,
		rapi:   rapi,
		cfg:    cfg,
		evHdl:  evHdl,
		store:  sessions.NewCookieStore(authKey, encKey),
		userDb: userDb,
	}, nil
}

// Close cleans up any potentially open resource.
func (s *State) Close() error {
	s.evHdl.Shutdown()
	return s.userDb.Close()
}

// UserDatabase returns the currently opened user database.
func (s *State) UserDatabase() *db.UserDatabase {
	return s.userDb
}

// SetEventListener sets the event listener.
// Since the gateway can run before (or without) the peer server
// and event listener running, we can set this dynamically.
func (s *State) SetEventListener(ev *events.Listener) {
	s.ev = ev
	s.evHdl.SetEventListener(ev)
}

func (s *State) publishFsEvent(req *http.Request) {
	if s.evHdl != nil {
		ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
		defer cancel()

		s.evHdl.Notify(ctx, "fs")
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

func prefixRoot(nodePath string) string {
	if strings.HasPrefix(nodePath, "/") {
		return nodePath
	}

	return "/" + nodePath
}

func buildFolderCache(folders []string) map[string]bool {
	folderCache := make(map[string]bool)
	for _, folder := range folders {
		folderCache[prefixRoot(path.Clean(folder))] = true
	}

	return folderCache
}

func (s *State) pathIsVisible(nodePath string, w http.ResponseWriter, r *http.Request) bool {
	nodePath = prefixRoot(path.Clean(nodePath))
	if s.validatePath(nodePath, w, r) {
		return true
	}

	name := getUserName(s.store, w, r)
	if name == "" {
		return false
	}

	user, err := s.userDb.Get(name)
	if err != nil {
		return false
	}

	folderCache := buildFolderCache(user.Folders)
	if err != nil {
		log.Debugf("failed to build folder cache: %v", err)
		return false
	}

	// Go over all folders, and see if we have some allowed folder
	// that we need to display "on the way". This could be probably
	// made faster if we ever need to.
	for folder, isValid := range folderCache {
		if !isValid {
			continue
		}

		// Example case:
		// folder   = /nested/something
		// nodePath = /nested
		// (also handles if folder == nodePath)
		//
		// Other case (folder = /nested, nodePath = /nested/something)
		// is already handled by calling validatePath() above.
		if strings.HasPrefix(folder, nodePath) {
			return true
		}
	}

	// There is no valid prefix at all.
	return false
}

func (s *State) validatePath(nodePath string, w http.ResponseWriter, r *http.Request) bool {
	if !strings.HasPrefix(nodePath, "/") {
		return false
	}

	name := getUserName(s.store, w, r)
	if name == "" {
		return false
	}

	user, err := s.userDb.Get(name)
	if err != nil {
		return false
	}

	// At this point we know that the user is logged in.
	return s.validatePathForUser(nodePath, user, w, r)
}

func (s *State) validatePathForUser(nodePath string, user db.User, w http.ResponseWriter, r *http.Request) bool {
	curr := prefixRoot(nodePath)
	folderCache := buildFolderCache(user.Folders)

	for curr != "" {
		if folderCache[curr] {
			return true
		}

		next := path.Dir(curr)
		if curr == "/" && next == curr {
			// We've gone up too much:
			break
		}

		curr = next
	}

	// No fitting path found:
	return false
}

//////////////////////

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
		// TODO: also pass request and output the url in this log message.
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

func (s *State) commitChange(msg string, w http.ResponseWriter, r *http.Request) bool {
	name := getUserName(s.store, w, r)
	fullMsg := fmt.Sprintf("gateway: »%s« %s", name, msg)

	if err := s.fs.MakeCommit(fullMsg); err != nil {
		if err != ie.ErrNoChange {
			log.Warningf("could not commit: %v", err)
			jsonifyErrf(w, http.StatusInternalServerError, "could not commit")
			return false
		}

		// There was no change. No need to notify.
		return true
	}

	s.evHdl.Notify(r.Context(), "fs")
	return true
}

///////

type secureMiddleware struct {
	*State
	SubHandler http.Handler
}

func (sm *secureMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hdr := w.Header()

	// Do not let browsers guess the content type:
	// https://en.wikipedia.org/wiki/Content_sniffing
	hdr.Set("X-Content-Type-Options", "nosniff")

	// https://security.stackexchange.com/questions/121796/what-security-implications-does-dns-prefetching-have
	hdr.Set("X-DNS-Prefetch-Control", "off")

	// Do not allow <iframe> of our website embedded in other sites.
	// This could be changed if we see a valid use case for this.
	hdr.Set("X-Frame-Options", "DENY")

	// Prevents Internet Explorer from executing downloads in site's context
	hdr.Set("X-Download-Options", "noopen")

	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-XSS-Protection
	hdr.Set("X-XSS-Protection", "1; mode=block")

	if r.TLS != nil {
		// https://en.wikipedia.org/wiki/Content_sniffing
		hdr.Set("Strict-Transport-Security", "max-age=5184000; includeSubDomains")
	}

	sm.SubHandler.ServeHTTP(w, r)
}

// SecureMiddleware sets some security related headers suitable for all
// endpoints that are supposed to protect users from some browser quirks.
func SecureMiddleware(s *State) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return &secureMiddleware{State: s, SubHandler: h}
	}
}
