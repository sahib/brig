package ipfsutil

import (
	"bytes"
	"io"
	"io/ioutil"

	log "github.com/Sirupsen/logrus"
)

// Add reads `r` and adds it to ipfs.
// The resulting content hash is returned.
func Add(ctx *Context, r io.Reader) ([]byte, error) {
	adder := ipfsCommand(ctx, "add", "-q")
	// adder := exec.Command("cat")
	stdin, err := adder.StdinPipe()
	if err != nil {
		log.Warning("stdin failed")
		return nil, err
	}

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

	if err := adder.Wait(); err != nil {
		log.Warningf("`ipfs add` failed: %v", err)
	}

	return bytes.TrimSpace(hash), nil
}
