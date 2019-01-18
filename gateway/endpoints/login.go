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
		SameSite: http.SameSiteLaxMode,
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

type LoginHandler struct {
	*State
}

func NewLoginHandler(s *State) *LoginHandler {
	return &LoginHandler{State: s}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Success  bool   `json:"success"`
	Username string `json:"username"`
}

func (lih *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	loginReq := &LoginRequest{}
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

type LogoutHandler struct {
	*State
}

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

type WhoamiHandler struct {
	*State
}

func NewWhoamiHandler(s *State) *WhoamiHandler {
	return &WhoamiHandler{State: s}
}

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
		user := getUserName(am.store, w, r)
		if user == "" {
			jsonifyErrf(w, http.StatusUnauthorized, "not authorized")
			return
		}
	}

	am.SubHandler.ServeHTTP(w, r)
}

func AuthMiddleware(s *State) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return &authMiddleware{State: s, SubHandler: h}
	}
}
