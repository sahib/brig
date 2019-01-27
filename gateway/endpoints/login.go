package endpoints

import (
	"encoding/json"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/sessions"
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
	sess, err := store.Get(r, "sess")
	if err != nil {
		log.Warningf("failed to get session: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   31 * 24 * 60 * 60,
		HttpOnly: true,
		Secure:   true,
		// TODO: This is only available in Go 1.11:
		// SameSite: http.SameSiteLaxMode,
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
	Success  bool   `json:"success"`
	Username string `json:"username"`
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

	setSession(lih.store, dbUser.Name, w, r)
	jsonify(w, http.StatusOK, &LoginResponse{
		Success:  true,
		Username: loginReq.Username,
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
	IsLoggedIn bool   `json:"is_logged_in"`
	User       string `json:"user"`
}

func (wh *WhoamiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if wh.cfg.Bool("auth.enabled") {
		user := getUserName(wh.store, w, r)
		if user != "" {
			// renew the session, if already logged in:
			setSession(wh.store, user, w, r)
		}

		jsonify(w, http.StatusOK, WhoamiResponse{
			IsLoggedIn: len(user) > 0,
			User:       user,
		})

		return
	}

	// If no auth is used, fake a anon session:
	setSession(wh.store, "anon", w, r)
	jsonify(w, http.StatusOK, WhoamiResponse{
		IsLoggedIn: true,
		User:       "anon",
	})
}

///////

type authMiddleware struct {
	*State
	SubHandler http.Handler
}

func (am *authMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if am.cfg.Bool("auth.enabled") {
		name := getUserName(am.store, w, r)
		if name == "" {
			// invalid token.
			jsonifyErrf(w, http.StatusUnauthorized, "not authorized")
			return
		}

		if _, err := am.userDb.Get(name); err != nil {
			// valid token, but invalid user.
			// (user might have been deleted on our side)
			jsonifyErrf(w, http.StatusUnauthorized, "not authorized")
			return
		}
	}

	am.SubHandler.ServeHTTP(w, r)
}

// AuthMiddleware returns a new handler wrapper, that will require
// all calls to the respective handler to have a "sess" cookie with
// a valid user name.
func AuthMiddleware(s *State) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return &authMiddleware{State: s, SubHandler: h}
	}
}
