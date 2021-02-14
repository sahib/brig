package fusetest

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
	"github.com/sahib/brig/backend/httpipfs"
	"github.com/sahib/brig/catfs"
	"github.com/sahib/brig/defaults"
	"github.com/sahib/brig/fuse"
	"github.com/sahib/brig/repo/hints"
	"github.com/sahib/brig/util"
	"github.com/sahib/config"
)

func makeFS(dbPath string, backend catfs.FsBackend) (*catfs.FS, error) {
	// open a dummy default config:
	cfg, err := config.Open(nil, defaults.Defaults, config.StrictnessPanic)
	if err != nil {
		return nil, err
	}

	hintMgr, err := hints.NewManager(nil)
	if err != nil {
		return nil, err
	}

	cfs, err := catfs.NewFilesystem(
		backend,
		dbPath,
		"alice",
		false,
		cfg.Section("fs"),
		hintMgr,
	)

	if err != nil {
		log.Fatalf("Failed to create catfs filesystem: %v", err)
		return nil, err
	}

	return cfs, err
}

func mount(cfs *catfs.FS, mountPath string, opts Options) (*fuse.Mount, error) {
	if err := os.MkdirAll(mountPath, 0700); err != nil {
		return nil, err
	}

	return fuse.NewMount(cfs, mountPath, nil, fuse.MountOptions{
		ReadOnly: opts.MountReadOnly,
		Offline:  opts.MountOffline,
		Root:     "/",
	})
}

func serveHTTPServer(opts Options) error {
	scheme, addr, err := util.URLToSchemeAndAddr(opts.URL)
	if err != nil {
		return err
	}

	lst, err := net.Listen(scheme, addr)
	if err != nil {
		return err
	}

	// Needed for /quit.
	srv := &http.Server{}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Properly exit when Ctrl-C is pressed.
	// (including unmounting!)
	go func() {
		<-sigCh

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		srv.Shutdown(ctx)
	}()

	// TODO: routes for stage / cat (although not really necessary...)
	// TODO: root for unmount.
	router := mux.NewRouter()
	router.HandleFunc("/quit", func(w http.ResponseWriter, r *http.Request) {
		go func() {
			// Close the server in a few ms, just not in th request itself.
			// Otherwise the client will block forever.
			time.Sleep(100 * time.Millisecond)
			if err := srv.Shutdown(r.Context()); err != nil {
				log.WithError(err).Warnf("failed to shutdown server")
			}
		}()
	}).Methods("GET")

	srv.Handler = router
	fmt.Println("serving...")
	defer fmt.Println("serving done...")
	return srv.Serve(lst)
}

// Options can be specified to control the behavior of the fusetest server.
type Options struct {
	// MountPath is where the fuse mount will be available.
	MountPath string

	// CatfsPath is where the metdata is stored.
	CatfsPath string

	// IpfsPath tells us which IPFS repo to use.
	// If empty, use the mock backend.
	IpfsPathOrURL string

	// URL defines where the server can be reached.
	URL string

	// MountReadOnly = true means to not allow modifications.
	MountReadOnly bool

	// MountOffline= true means to not allow online queries.
	MountOffline bool
}

// Launch will launch a fuse test server.
func Launch(opts Options) error {
	tmpDir, err := ioutil.TempDir("", "brig-debug-fuse-*")
	if err != nil {
		return err
	}

	defer os.RemoveAll(tmpDir)

	for _, path := range []string{opts.MountPath, opts.CatfsPath} {
		if err := os.MkdirAll(path, 0700); err != nil {
			return err
		}
	}

	var backend catfs.FsBackend
	if opts.IpfsPathOrURL != "" {
		backend, err = httpipfs.NewNode(opts.IpfsPathOrURL, "")
	} else {
		backend = catfs.NewMemFsBackend()
	}

	if err != nil {
		return err
	}

	cfs, err := makeFS(opts.CatfsPath, backend)
	if err != nil {
		return err
	}

	m, err := mount(cfs, opts.MountPath, opts)
	if err != nil {
		return err
	}

	// make sure it gets closed, even when no unmount is happening.
	defer func() {
		fmt.Println("Closing mount")
		if err := m.Close(); err != nil {
			log.WithError(err).Error("fuse mount close failed")
		}
	}()

	return serveHTTPServer(opts)
}
