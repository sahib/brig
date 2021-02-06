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

func StartDaemon(name, backendName string) (*server.Server, error) {
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

func WithDaemon(name string, fn func(ctl *client.Client) error) error {
	srv, err := StartDaemon(name, "mock")
	if err != nil {
		return err
	}

	defer os.RemoveAll(srv.RepoPath())
	defer srv.Close()

	ctl, err := client.Dial(context.Background(), srv.DaemonURL())
	if err != nil {
		return err
	}

	defer ctl.Close()

	return fn(ctl)
}

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
