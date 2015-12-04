package ipfsutil

import (
	"os/exec"
	"time"

	log "github.com/Sirupsen/logrus"
)

// StartDaemon executes and watches `ipfs daemon`.
func StartDaemon(ctx *Context) (*exec.Cmd, error) {
	port := 4001 // TODO: read from ctx.Path.config

	daemon := ipfsCommand(ctx, "daemon")
	if err := daemon.Start(); err != nil {
		return nil, err
	}

	go func() {
		err := daemon.Wait()
		log.Warningf("ipfs daemon exit: %v", err)
	}()

	// TODO: Poll until it's available.
	time.Sleep(1 * time.Second)

	log.Infof("ipfs running on :%d", port)
	return daemon, nil
}
