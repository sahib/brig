package ipfsutil

import (
	"io"
	"io/ioutil"

	log "github.com/Sirupsen/logrus"
	"github.com/jbenet/go-multihash"
)

// Cat returns an io.Reader that reads from ipfs.
func Cat(ctx *Context, hash multihash.Multihash) (io.Reader, error) {
	catter := ipfsCommand(ctx, "cat", hash.B58String())
	stdout, err := catter.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := catter.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := catter.Start(); err != nil {
		log.Warningf("ipfs cat failed: ", err)
		return nil, err
	}

	go func() {
		stderrText, _ := ioutil.ReadAll(stderr)
		if err := catter.Wait(); err != nil {
			log.Warningf("`ipfs cat` failed: %v", err)
			log.Warningf("Stderr: %v", string(stderrText))
		}
	}()

	return stdout, nil
}
