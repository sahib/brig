package endpoints

import (
	"fmt"
	"net"
	"net/http"

	log "github.com/Sirupsen/logrus"
)

// RedirHandler implements http.Handler.
// It redirects all of its requests to the respective https:// route.
type RedirHandler struct {
	redirPort int64
}

// NewHTTPRedirectHandler returns a new RedirHandler
func NewHTTPRedirectHandler(redirPort int64) *RedirHandler {
	return &RedirHandler{
		redirPort: redirPort,
	}
}

func (rh *RedirHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// remove/add not default ports from req.Host
	host, _, err := net.SplitHostPort(req.Host)
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	target := fmt.Sprintf("https://%s:%d%s", host, rh.redirPort, req.URL.Path)
	if len(req.URL.RawQuery) > 0 {
		target += "?" + req.URL.RawQuery
	}

	log.Debugf("redirect to: %s", target)
	http.Redirect(w, req, target, http.StatusTemporaryRedirect)
}
