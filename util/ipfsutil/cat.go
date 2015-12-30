package ipfsutil

import (
	"io"

	log "github.com/Sirupsen/logrus"
)

// Cat returns an io.Reader that reads from ipfs.
func Cat(ctx *Context, hash []byte) (io.Reader, error) {
	catter := ipfsCommand(ctx, "cat", string(hash))
	stdout, err := catter.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := catter.Start(); err != nil {
		log.Warningf("ipfs cat failed: ", err)
		return nil, err
	}

	go func() {
		if err := catter.Wait(); err != nil {
			log.Warningf("`ipfs cat` failed: %v", err)
		}
	}()

	return stdout, nil
}
