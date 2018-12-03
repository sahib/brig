package gateway

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/NYTimes/gziphandler"
	log "github.com/Sirupsen/logrus"
	"github.com/sahib/brig/catfs"
	ie "github.com/sahib/brig/catfs/errors"
	"github.com/sahib/brig/catfs/mio"
	"github.com/sahib/brig/util"
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
	Tar(nodePath string, w io.Writer) error
}

// Gateway is a small HTTP server that is able to serve
// files from brig over HTTP. This can be used to share files
// inside of brig with users that do not use brig.
type Gateway struct {
	backend     Backend
	cfg         *config.Config
	tickets     chan int
	isClosed    bool
	isReloading bool

	srv      *http.Server
	redirSrv *http.Server
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
		if gw.isClosed {
			return
		}

		// Forbid recursive reloading.
		if gw.isReloading {
			return
		}

		gw.isReloading = true

		log.Debugf("reloading gateway because config key changed: %s", key)
		if err := gw.Stop(); err != nil {
			log.Errorf("failed to reload gateway: %v", err)
		}

		gw.Start()
		gw.isReloading = false
	}

	// If any of those vars change, we should reload.
	// All other config values are read on-demand anyways.
	cfg.AddEvent("enabled", reloader)
	cfg.AddEvent("port", reloader)
	cfg.AddEvent("cert.certfile", reloader)
	cfg.AddEvent("cert.keyfile", reloader)
	cfg.AddEvent("cert.domain", reloader)
	cfg.AddEvent("cert.redirect.enabled", reloader)
	cfg.AddEvent("cert.redirect.http_port", reloader)
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

	if gw.redirSrv != nil {
		if err := gw.redirSrv.Shutdown(ctx); err != nil {
			return err
		}

		gw.redirSrv = nil
	}

	return gw.srv.Shutdown(ctx)
}

type redirHandler struct {
	redirPort int64
}

func (rh redirHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// remove/add not default ports from req.Host
	host, _, err := net.SplitHostPort(req.Host)
	if err != nil {
		w.WriteHeader(400)
		return
	}

	target := fmt.Sprintf("https://%s:%d%s", host, rh.redirPort, req.URL.Path)
	if len(req.URL.RawQuery) > 0 {
		target += "?" + req.URL.RawQuery
	}

	log.Debugf("redirect to: %s", target)
	http.Redirect(w, req, target, http.StatusTemporaryRedirect)
}

func setContentDisposition(info *catfs.StatInfo, hdr http.Header, dispoType string) {
	basename := path.Base(info.Path)
	if info.IsDir {
		if basename == "/" {
			basename = "root"
		}

		basename += ".tar"
	}

	hdr.Set(
		"Content-Disposition",
		fmt.Sprintf(
			"%s; filename*=UTF-8''%s",
			dispoType,
			url.QueryEscape(basename),
		),
	)
}

func mimeTypeFromStream(stream mio.Stream) (io.Reader, string) {
	hdr, newStream, err := util.PeekHeader(stream, 512)
	if err != nil {
		return stream, "application/octet-stream"
	}

	return newStream, http.DetectContentType(hdr)
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

	tlsConfig, err := getTLSConfig(gw.cfg)
	if err != nil {
		log.Errorf("failed to read TLS config: %v", err)
		return
	}

	// If requested, forward all http requests from a different port
	// to the normal https port.
	if tlsConfig != nil && gw.cfg.Bool("cert.redirect.enabled") {
		gw.redirSrv = &http.Server{
			Addr:    fmt.Sprintf(":%d", gw.cfg.Int("cert.redirect.http_port")),
			Handler: redirHandler{redirPort: gw.cfg.Int("port")},
		}

		go func() {
			if err := gw.redirSrv.ListenAndServe(); err != nil {
				log.Errorf("failed to start http redirecter: %v", err)
			}
		}()
	}

	gw.srv = &http.Server{
		Addr:      addr,
		Handler:   gziphandler.GzipHandler(gw),
		TLSConfig: tlsConfig,

		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		if tlsConfig != nil {
			err = gw.srv.ListenAndServeTLS("", "")
		} else {
			err = gw.srv.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			log.Errorf("serve failed: %v", err)
		}
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
	if fullURL == "/" {
		rw.WriteHeader(200)
		rw.Write([]byte("This brig gateway seems to be working."))
		return
	}

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

	hdr.Set("ETag", info.ContentHash.B58String())
	hdr.Set("Last-Modified", info.ModTime.Format(http.TimeFormat))

	if info.IsDir {
		setContentDisposition(info, hdr, "attachment")

		if err := gw.backend.Tar(nodePath, rw); err != nil {
			// All other error is handled relatively broad.
			log.Errorf("gateway: failed to stream %s: %v", nodePath, err)
			rw.WriteHeader(500)
			return
		}
	} else {
		stream, err := gw.backend.Cat(nodePath)
		if err != nil {
			// All other error is handled relatively broad.
			log.Errorf("gateway: failed to stream %s: %v", nodePath, err)
			rw.WriteHeader(500)
			return
		}

		rawStream, mimeType := mimeTypeFromStream(stream)
		hdr.Set("Content-Type", mimeType)
		hdr.Set("Content-Length", strconv.FormatUint(info.Size, 10))

		// Set the content disposition to inline if it looks like something viewable.
		if mimeType == "application/octet-stream" {
			setContentDisposition(info, hdr, "attachment")
		} else {
			setContentDisposition(info, hdr, "inline")
		}

		if _, err := io.Copy(rw, rawStream); err != nil {
			log.Errorf("gateway: failed to stream %s: %v", nodePath, err)
			rw.WriteHeader(500)
			return
		}
	}
}
