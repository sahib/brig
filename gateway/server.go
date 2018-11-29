package gateway

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/sahib/brig/catfs"
	ie "github.com/sahib/brig/catfs/errors"
	"github.com/sahib/brig/catfs/mio"
	"github.com/sahib/config"
)

const (
	rateLimit = 50
)

// Backend is the backend that the gateway uses to output files.
// This is conviniently the same API as catfs.FS, but useful for
// testing purposes to separate this.
type Backend interface {
	Stat(nodePath string) (*catfs.StatInfo, error)
	Cat(nodePath string) (mio.Stream, error)
}

// Gateway is a small HTTP server that is able to serve
// files from brig over HTTP. This can be used to share files
// inside of brig with users that do not use brig.
type Gateway struct {
	backend  Backend
	cfg      *config.Config
	srv      *http.Server
	tickets  chan int
	isClosed bool
}

// NewGateway returns a newly built gateway.
// This function does not yet start a server.
func NewGateway(backend Backend, cfg *config.Config) *Gateway {
	gw := &Gateway{
		backend: backend,
		cfg:     cfg,
	}

	// Restarts the gateway on the next possible idle phase:
	reloader := func(key string) {
		log.Debugf("reloading gateway because config key changed: %s", key)
		if err := gw.Stop(); err != nil {
			log.Errorf("failed to reload gateway: %v", err)
		}

		gw.Start()
	}

	// If any of those vars change, we should reload:
	cfg.AddEvent("enabled", reloader)
	cfg.AddEvent("port", reloader)
	return gw
}

// Stop stops the gateway gracefully.
func (gw *Gateway) Stop() error {
	if gw.isClosed {
		return nil
	}

	gw.isClosed = true

	// Wait until all requests were done.
	// We do not want to close downloads just because
	// the user changed the config.
	for {
		if len(gw.tickets) == rateLimit {
			// All requests have been served.
			break
		}

		time.Sleep(10 * time.Millisecond)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := gw.srv.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}

// Start will start the gateway.
// If the gateway is not enabled in the config, this does nothing.
func (gw *Gateway) Start() {
	if !gw.cfg.Bool("enabled") {
		log.Debugf("gateway is disabled in the config; doing nothing until enabled.")
		return
	}

	gw.isClosed = false
	gw.tickets = make(chan int, 50)
	for idx := 0; idx < 50; idx++ {
		gw.tickets <- idx
	}

	addr := fmt.Sprintf("0.0.0.0:%d", gw.cfg.Int("port"))
	log.Debugf("starting gateway on %s", addr)

	gw.srv = &http.Server{
		Addr:    addr,
		Handler: gw,
	}

	go func() {
		gw.srv.ListenAndServe()
	}()
}

func (gw *Gateway) validateUserForPath(nodePath string, rq *http.Request) bool {
	if gw.cfg.Bool("auth.enabled") {
		user, pass, ok := rq.BasicAuth()
		if !ok {
			return false
		}

		cfgUser := gw.cfg.String("auth.user")
		cfgPass := gw.cfg.String("auth.pass")
		return user == cfgUser && pass == cfgPass
	}

	folders := make(map[string]bool)
	for _, folder := range gw.cfg.Strings("folders") {
		folders[folder] = true
	}

	curr := nodePath
	for {
		if ok := folders[curr]; ok {
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

func (gw *Gateway) ServeHTTP(rw http.ResponseWriter, rq *http.Request) {
	if gw.isClosed {
		return
	}

	if rq.Method != "GET" {
		return
	}

	// Do some basic rate limiting.
	// Only process this request if we have a free ticket.
	ticket := <-gw.tickets
	defer func() {
		gw.tickets <- ticket
	}()

	fullURL := rq.URL.EscapedPath()
	if !strings.HasPrefix(fullURL, "/get/") {
		rw.WriteHeader(400)
		return
	}

	// get the file nodePath including the leading slash:
	nodePath, err := url.PathUnescape(fullURL[4:])
	if err != nil {
		log.Debugf("received malformed url: %s", fullURL)
		rw.WriteHeader(400)
		return
	}

	hdr := rw.Header()
	if !gw.validateUserForPath(nodePath, rq) {
		// No auth supplied, if the user is using a browser, we should give
		// him the chance to enter a user/password, if we enabled that.
		if gw.cfg.Bool("auth.enabled") {
			hdr.Set("WWW-Authenticate", "Basic realm=\"brig gateway\"")
		}

		rw.WriteHeader(401)
		return
	}

	info, err := gw.backend.Stat(nodePath)
	if err != nil {
		// Handle a bad nodePath more explicit:
		if ie.IsNoSuchFileError(err) {
			rw.WriteHeader(404)
			return
		}

		log.Errorf("gateway: failed to stat %s: %v", nodePath, err)
		rw.WriteHeader(500)
		return
	}

	stream, err := gw.backend.Cat(nodePath)
	if err != nil {
		// All other error is handled relatively broad.
		log.Errorf("gateway: failed to cat %s: %v", nodePath, err)
		rw.WriteHeader(500)
		return
	}

	// Set the right headers for the stream:
	hdr.Set("Content-Type", "application/octet-stream")
	hdr.Set("Content-Transfer-Encoding", "binary")
	hdr.Set("Content-Length", strconv.FormatUint(info.Size, 10))

	if _, err := io.Copy(rw, stream); err != nil {
		log.Errorf("gateway: failed to stream %s: %v", nodePath, err)
		rw.WriteHeader(500)
		return
	}
}
