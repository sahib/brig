package ipfsutil

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	log "github.com/Sirupsen/logrus"
)

// StartDaemon executes and watches `ipfs daemon`.
// The exec.Cmd associated to it is returned,
// use it to call Wait() on or to stop it via .Process.Kill()
func StartDaemon(ctx *Context) (*exec.Cmd, error) {
	port := 4001 // TODO: read from ctx.Path.config

	daemon := ipfsCommand(ctx, "daemon")
	stderr, err := daemon.StderrPipe()
	if err != nil {
		log.Warningf("Could not attach to stderr: %v", err)
	}

	if err := daemon.Start(); err != nil {
		return nil, err
	}

	go func() {
		stderrBytes, _ := ioutil.ReadAll(stderr)
		if err := daemon.Wait(); err != nil {
			log.Warningf("ipfs daemon exit: %v", err)
			log.Warningf("Stderr: %v", string(stderrBytes))
		}
	}()

	addr := fmt.Sprintf("localhost:%d", port)
	for i := 0; i < 30; i++ {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		conn.Close()

		// This is pretty stupid. The ipfs daemon first starts the
		// network interface but is not ready for usage yet.
		// It appears to be ready once the api/ dir is created.
		// Maybe better to check for the "Daemon is ready" phrase?
		// (This is actually a TODO)
		if _, err := os.Stat(filepath.Join(ctx.Path, "api")); err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		log.Infof("ipfs running on :%d", port)
		return daemon, nil
	}

	// Something wrong. Kill whatever we launched.
	if err := daemon.Process.Kill(); err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("ipfs daemon startup took too long.")
}
