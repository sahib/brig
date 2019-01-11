package gateway

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/NYTimes/gziphandler"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/sahib/brig/catfs"
	"github.com/sahib/brig/gateway/endpoints"
	"github.com/sahib/config"
	"github.com/ulule/limiter"
	"github.com/ulule/limiter/drivers/middleware/stdlib"
	"github.com/ulule/limiter/drivers/store/memory"

	// Include static resources:
	_ "github.com/sahib/brig/gateway/static"
)

const (
	// TODO: Save somewhere. Config?
	csrfKey = "60b725f10c9c85c70d97880dfe8191b3"
)

// allow at max. 1000 request per hour before limiting.
var rate = limiter.Rate{
	Period: 1 * time.Hour,
	Limit:  1000,
}

// Gateway is a small HTTP server that is able to serve
// files from brig over HTTP. This can be used to share files
// inside of brig with users that do not use brig.
type Gateway struct {
	fs          *catfs.FS
	cfg         *config.Config
	isClosed    bool
	isReloading bool

	srv      *http.Server
	redirSrv *http.Server
}

// NewGateway returns a newly built gateway.
// This function does not yet start a server.
func NewGateway(fs *catfs.FS, cfg *config.Config) *Gateway {
	gw := &Gateway{
		fs:       fs,
		cfg:      cfg,
		isClosed: true,
	}

	// Restarts the gateway on the next possible idle phase:
	reloader := func(key string) {
		// Forbid recursive reloading.
		if gw.isReloading {
			return
		}

		gw.isReloading = true
		defer func() { gw.isReloading = false }()

		log.Debugf("reloading gateway because config key changed: %s", key)
		if err := gw.Stop(); err != nil {
			log.Errorf("failed to stop gateway: %v", err)
			return
		}

		gw.Start()
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

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if gw.redirSrv != nil {
		if err := gw.redirSrv.Shutdown(ctx); err != nil {
			return err
		}

		gw.redirSrv = nil
	}

	if gw.srv != nil {
		return gw.srv.Shutdown(ctx)
	}

	return nil
}

// Start will start the gateway.
// If the gateway is not enabled in the config, this does nothing.
// The gateway is started in the background, this method does not block.
func (gw *Gateway) Start() {
	if !gw.cfg.Bool("enabled") {
		log.Debugf("gateway is disabled in the config; doing nothing until enabled.")
		return
	}

	gw.isClosed = false

	port := gw.cfg.Int("port")
	addr := fmt.Sprintf(":%d", port)
	log.Debugf("starting gateway on %s", addr)

	gw.isReloading = true
	defer func() {
		gw.isReloading = false
	}()

	tlsConfig, err := getTLSConfig(gw.cfg)
	if err != nil {
		log.Errorf("failed to read TLS config: %v", err)
		return
	}

	// If requested, forward all http requests from a different port
	// to the normal https port.
	if tlsConfig != nil && gw.cfg.Bool("cert.redirect.enabled") {
		httpPort := gw.cfg.Int("cert.redirect.http_port")
		gw.redirSrv = &http.Server{
			ReadHeaderTimeout: 10 * time.Second,
			WriteTimeout:      10 * time.Second,
			IdleTimeout:       360 * time.Second,
			Addr:              fmt.Sprintf(":%d", httpPort),
			Handler:           endpoints.NewHTTPRedirectHandler(port),
		}

		go func() {
			if err := gw.redirSrv.ListenAndServe(); err != nil {
				if err != http.ErrServerClosed {
					log.Errorf("failed to start http redirecter: %v", err)
				}
			}
		}()
	}

	router := mux.NewRouter()
	router.Use(csrf.Protect([]byte(csrfKey)))

	// API route definition:
	apiRouter := router.PathPrefix("/api/v0").Methods("POST").Subrouter()
	apiRouter.Handle("/login", endpoints.NewLoginHandler(gw.cfg))
	apiRouter.Handle("/logout", endpoints.NewLogoutHandler())
	apiRouter.Handle("/ls", endpoints.NewLsHandler(gw.cfg, gw.fs))
	apiRouter.Handle("/upload", endpoints.NewUploadHandler(gw.cfg, gw.fs))
	apiRouter.Handle("/move", endpoints.NewMoveHandler(gw.cfg, gw.fs))
	apiRouter.Handle("/mkdir", endpoints.NewMkdirHandler(gw.cfg, gw.fs))
	apiRouter.Handle("/copy", endpoints.NewCopyHandler(gw.cfg, gw.fs))
	apiRouter.Handle("/remove", endpoints.NewRemoveHandler(gw.cfg, gw.fs))
	apiRouter.Handle("/history", endpoints.NewHistoryHandler(gw.cfg, gw.fs))
	apiRouter.Handle("/whoami", endpoints.NewWhoamiHandler())

	// Add the /get endpoint. Since it might contain any path, we have to
	// Use a path prefix so the right handler is called.
	router.PathPrefix("/get/").Handler(endpoints.NewGetHandler(gw.cfg, gw.fs)).Methods("GET")

	// Special case: index.html gets a csrf token:
	router.Handle("/", endpoints.NewIndexHandler()).Methods("GET")
	router.Handle("/index.html", endpoints.NewIndexHandler()).Methods("GET")
	router.PathPrefix("/view").Handler(endpoints.NewIndexHandler()).Methods("GET")

	// Serve all files in the static directory as-is.
	// This has to come last, since it's a wildcard for everything else.
	// The static files are packed inside the binary (for now)
	// TODO: Use this in development mode:
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./gateway/static")))

	// TODO: Use this in release mode:
	// router.PathPrefix("/").Handler(http.FileServer(parcello.ManagerAt("/")))

	// Implement rate limiting:
	router.Use(
		stdlib.NewMiddleware(
			limiter.New(memory.NewStore(), rate),
			stdlib.WithForwardHeader(true),
		).Handler,
	)

	gw.srv = &http.Server{
		Addr:              addr,
		Handler:           gziphandler.GzipHandler(router),
		TLSConfig:         tlsConfig,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       360 * time.Second,
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
