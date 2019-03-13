package endpoints

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/sahib/brig/gateway/db"
	log "github.com/sirupsen/logrus"
)

func getUserName(store *sessions.CookieStore, w http.ResponseWriter, r *http.Request) string {
	sess, err := store.Get(r, "sess")
	if err != nil {
		log.Warningf("failed to get session: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return ""
	}

	userNameIf, ok := sess.Values["name"]
	if !ok {
		return ""
	}

	userName, ok := userNameIf.(string)
	if !ok {
		log.Warningf("failed to convert user name to string: %v", userNameIf)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return ""
	}

	return userName
}

func setSession(store *sessions.CookieStore, userName string, w http.ResponseWriter, r *http.Request) {
	// Ignore the error here, since it will usually trigger when there was a previously
	// outdated session that fails to decode. Since we overwrite the session anyways, it
	// doesn't really matter in this case.
	sess, _ := store.Get(r, "sess")

	isHTTPS := r.TLS != nil
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   31 * 24 * 60 * 60,
		HttpOnly: true,
		Secure:   isHTTPS,
	}

	sess.Values["name"] = userName
	if err := sess.Save(r, w); err != nil {
		log.Warningf("set: failed to save session: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func clearSession(store *sessions.CookieStore, w http.ResponseWriter, r *http.Request) {
	sess, err := store.Get(r, "sess")
	if err != nil {
		log.Warningf("failed to get session: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sess.Options.MaxAge = -1
	if err := sess.Save(r, w); err != nil {
		log.Warningf("clear: failed to save session: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

///////

// LoginHandler implements http.Handler
type LoginHandler struct {
	*State
}

// NewLoginHandler creates a new LoginHandler
func NewLoginHandler(s *State) *LoginHandler {
	return &LoginHandler{State: s}
}

// LoginRequest is the request sent as JSON to this endpoint.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse is what the endpoint will return.
type LoginResponse struct {
	Success       bool     `json:"success"`
	Username      string   `json:"username"`
	Rights        []string `json:"rights"`
	IsAnon        bool     `json:"is_anon"`
	AnonIsAllowed bool     `json:"anon_is_allowed"`
}

func (lih *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	loginReq := LoginRequest{}
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		jsonifyErrf(w, http.StatusBadRequest, "bad json")
		return
	}

	if loginReq.Username == "" || loginReq.Password == "" {
		jsonifyErrf(w, http.StatusBadRequest, "empty password or username")
		return
	}

	dbUser, err := lih.userDb.Get(loginReq.Username)
	if err != nil {
		// No such user.
		jsonifyErrf(w, http.StatusForbidden, "bad credentials")
		return
	}

	if dbUser.Name != loginReq.Username {
		// Bad username. Might be a problem on our side.
		jsonifyErrf(w, http.StatusForbidden, "bad credentials")
		return
	}

	isValid, err := dbUser.CheckPassword(loginReq.Password)
	if err != nil || !isValid {
		if err != nil {
			log.Warningf("check password failed: %v", err)
		}

		jsonifyErrf(w, http.StatusForbidden, "bad credentials")
		return
	}

	anonIsAllowed := lih.cfg.Bool("auth.anon_allowed")
	anonUserName := lih.cfg.String("auth.anon_user")

	setSession(lih.store, dbUser.Name, w, r)
	jsonify(w, http.StatusOK, &LoginResponse{
		Success:       true,
		Username:      loginReq.Username,
		Rights:        dbUser.Rights,
		IsAnon:        anonUserName == loginReq.Username,
		AnonIsAllowed: anonIsAllowed,
	})
}

///////

// LogoutHandler implements http.Handler
type LogoutHandler struct {
	*State
}

// NewLogoutHandler returns a new LogoutHandler
func NewLogoutHandler(s *State) *LogoutHandler {
	return &LogoutHandler{State: s}
}

func (loh *LogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user := getUserName(loh.store, w, r)
	if user == "" {
		jsonifyErrf(w, http.StatusBadRequest, "not logged in")
		return
	}

	clearSession(loh.store, w, r)
	jsonifySuccess(w)
}

///////

// WhoamiHandler implements http.Handler.
// This handler checks if a user is already logged in.
type WhoamiHandler struct {
	*State
}

// NewWhoamiHandler returns a new WhoamiHandler.
func NewWhoamiHandler(s *State) *WhoamiHandler {
	return &WhoamiHandler{State: s}
}

// WhoamiResponse is the response sent back by this endpoint.
type WhoamiResponse struct {
	IsLoggedIn    bool     `json:"is_logged_in"`
	IsAnon        bool     `json:"is_anon"`
	AnonIsAllowed bool     `json:"anon_is_allowed"`
	User          string   `json:"user"`
	Rights        []string `json:"rights"`
}

func (wh *WhoamiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rights := []string{}
	isAnon := false

	anonIsAllowed := wh.cfg.Bool("auth.anon_allowed")
	name := getUserName(wh.store, w, r)

	if name == "" && anonIsAllowed {
		isAnon = true
		name = wh.cfg.String("auth.anon_user")
	}

	if name != "" {
		possiblyAnonUser, err := wh.userDb.Get(name)
		if err != nil {
			log.Warningf("could not get user »%s« : %v", name, err)
		} else {
			rights = possiblyAnonUser.Rights
			setSession(wh.store, name, w, r)
		}
	}

	jsonify(w, http.StatusOK, WhoamiResponse{
		IsLoggedIn:    len(name) > 0,
		IsAnon:        isAnon,
		AnonIsAllowed: anonIsAllowed,
		User:          name,
		Rights:        rights,
	})
}

///////

type authMiddleware struct {
	*State
	SubHandler http.Handler
}

type dbUserKey string

func (am *authMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	anonIsAllowed := am.cfg.Bool("auth.anon_allowed")
	name := getUserName(am.store, w, r)

	if name == "" {
		if !anonIsAllowed {
			// invalid token.
			jsonifyErrf(w, http.StatusUnauthorized, "not authorized")
			return
		}

		name = am.cfg.String("auth.anon_user")
	}

	user, err := am.userDb.Get(name)
	if err != nil {
		// valid token, but invalid user.
		// (user might have been deleted on our side)
		jsonifyErrf(w, http.StatusUnauthorized, "not authorized")
		return
	}

	r = r.WithContext(
		context.WithValue(r.Context(), dbUserKey("brig.db_user"),
			user,
		),
	)

	am.SubHandler.ServeHTTP(w, r)
}

func checkRights(w http.ResponseWriter, r *http.Request, rights ...string) bool {
	user, ok := r.Context().Value(dbUserKey("brig.db_user")).(db.User)

	if !ok {
		jsonifyErrf(w, http.StatusInternalServerError, "could not cast user")
		return false
	}

	rmap := make(map[string]bool)
	for _, right := range user.Rights {
		rmap[right] = true
	}

	for _, right := range rights {
		if !rmap[right] {
			jsonifyErrf(w, http.StatusUnauthorized, "insufficient rights")
			return false
		}
	}

	return true
}

// AuthMiddleware returns a new handler wrapper, that will require
// all calls to the respective handler to have a "sess" cookie with
// a valid user name.
func AuthMiddleware(s *State) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return &authMiddleware{State: s, SubHandler: h}
	}
}
