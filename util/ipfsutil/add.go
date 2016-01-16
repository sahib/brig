package ipfsutil

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/jbenet/go-multihash"

	log "github.com/Sirupsen/logrus"
)

// Add reads `r` and adds it to ipfs.
// The resulting content hash is returned.
func Add(ctx *Context, r io.Reader) (multihash.Multihash, error) {
	adder := ipfsCommand(ctx, "add", "-q")
	// adder := exec.Command("cat")
	stdin, err := adder.StdinPipe()
	if err != nil {
		log.Warning("stdin failed")
		return nil, err
	}

	stderr, err := adder.StderrPipe()
	stdout, err := adder.StdoutPipe()
	if err != nil {
		log.Warning("stdout failed")
		return nil, err
	}

	if err := adder.Start(); err != nil {
		log.Warning("ipfs add failed: ", err)
		return nil, err
	}

	// Copy file to ipfs-add's stdin:
	if _, err = io.Copy(stdin, r); err != nil {
		log.Warning("copy failed")
		return nil, err
	}

	if err := stdin.Close(); err != nil {
		log.Warningf("ipfs add: close failed: %v", err)
	}

	hash, err := ioutil.ReadAll(stdout)
	if err != nil {
		log.Warning("hash failed", stdout)
		return nil, err
	}

	errs, _ := ioutil.ReadAll(stderr)
	if err := adder.Wait(); err != nil {
		log.Warningf("`ipfs add` failed: %v", err)
		log.Warningf("Stderr: %v", string(errs))
	}

	hash = bytes.TrimSpace(hash)
	mh, err := multihash.FromB58String(string(hash))
	if err != nil {
		return nil, err
	}

	return mh, nil
}
