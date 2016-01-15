package ipfsutil

import (
	"io"
	"io/ioutil"
	"os/exec"

	log "github.com/Sirupsen/logrus"
	"github.com/jbenet/go-multihash"
)

type ipfsCatter struct {
	catter *exec.Cmd
	stdout io.Reader
	stderr io.Reader
}

func (i *ipfsCatter) Read(buf []byte) (int, error) {
	return i.stdout.Read(buf)
}

func (i *ipfsCatter) Close() error {
	stderrText, _ := ioutil.ReadAll(i.stderr)
	if err := i.catter.Wait(); err != nil {
		log.Warningf("`ipfs cat` failed: %v", err)
		log.Warningf("Stderr: %v", string(stderrText))
		return err
	}
	return nil
}

// Cat returns an io.Reader that reads from ipfs.
func Cat(ctx *Context, hash multihash.Multihash) (io.ReadCloser, error) {
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
		log.Warningf("`ipfs cat` failed to start: ", err)
		return nil, err
	}

	return &ipfsCatter{
		catter: catter,
		stdout: stdout,
		stderr: stderr,
	}, nil
}
