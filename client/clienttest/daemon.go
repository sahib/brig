package clienttest

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/sahib/brig/client"
	"github.com/sahib/brig/repo"
	"github.com/sahib/brig/server"
	log "github.com/sirupsen/logrus"
)

// StartDaemon starts a new daemon with user `name`, using backend defined by
// `backendName` and, if the backend is IPFS, uses the IPFS repository at
// `ipfsPath`.  The resulting server should be closed after use and the
// temporary directory where all data resides should be removed.
func StartDaemon(name, backendName, ipfsPath string) (*server.Server, error) {
	repoPath, err := ioutil.TempDir("", "brig-client-repo")
	if err != nil {
		return nil, err
	}

	daemonURL := "unix:" + filepath.Join(repoPath, "brig.socket")
	if err := repo.Init(repo.InitOptions{
		BaseFolder:  repoPath,
		Owner:       name,
		BackendName: backendName,
		DaemonURL:   daemonURL,
	}); err != nil {
		return nil, err
	}

	if backendName == "httpipfs" {
		if err := repo.OverwriteConfigKey(repoPath, "daemon.ipfs_path", ipfsPath); err != nil {
			return nil, err
		}
	}

	srv, err := server.BootServer(repoPath, daemonURL)
	if err != nil {
		return nil, err
	}

	go func() {
		if err := srv.Serve(); err != nil {
			log.WithError(err).Warnf("failed to serve")
		}
	}()

	// give some time for startup:
	time.Sleep(500 * time.Millisecond)
	return srv, nil
}

// WithDaemon calls `fn` with a readily setup daemon client. `name` is the user.
func WithDaemon(name string, fn func(ctl *client.Client) error) error {
	srv, err := StartDaemon(name, "mock", "")
	if err != nil {
		return err
	}

	defer func() {
		// Somehow there is race condition between
		// srv.Close() from the defer at the very end
		// os.RemoveAll(repoPath).
		// Theoretically, `go` should have closed server
		// but in practice I see that repoPath is removed
		// before server had a chance to close the DB
		// and I see complains in log about DB.Close
		// I introduce this time delay as a crude hack
		time.Sleep(100 * time.Millisecond)
		os.RemoveAll(srv.RepoPath())
	}()
	defer srv.Close()

	ctl, err := client.Dial(context.Background(), srv.DaemonURL())
	if err != nil {
		return err
	}

	defer ctl.Close()

	return fn(ctl)
}

// WithDaemonPair calls `fn` with two readily setup daemon clients.
// `nameA` and `nameB` are the respective names.
func WithDaemonPair(nameA, nameB string, fn func(ctlA, ctlB *client.Client) error) error {
	return WithDaemon(nameA, func(ctlA *client.Client) error {
		return WithDaemon(nameB, func(ctlB *client.Client) error {
			aliWhoami, err := ctlA.Whoami()
			if err != nil {
				return err
			}

			bobWhoami, err := ctlB.Whoami()
			if err != nil {
				return err
			}

			// add bob to ali as remote
			if err := ctlA.RemoteAddOrUpdate(client.Remote{
				Name:        nameB,
				Fingerprint: bobWhoami.Fingerprint,
			}); err != nil {
				return err
			}

			// add ali to bob as remote
			if err := ctlB.RemoteAddOrUpdate(client.Remote{
				Name:        nameA,
				Fingerprint: aliWhoami.Fingerprint,
			}); err != nil {
				return err
			}

			return fn(ctlA, ctlB)
		})
	})
}
