package ipfs

import (
	"bytes"
	"io"
	"io/ioutil"
	"os/exec"

	log "github.com/Sirupsen/logrus"
)

type Context struct {
	// TODO!
	Path string
}

func ipfsCommand(ctx Context, args ...string) *exec.Cmd {
	cmd := exec.Command("ipfs", args...)
	cmd.Env = []string{"IPFS_PATH=" + ctx.Path}
	return cmd
}

func Add(ctx Context, r io.Reader) ([]byte, error) {
	adder := ipfsCommand(ctx, "add", "-q")
	stdin, err := adder.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := adder.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := adder.Start(); err != nil {
		log.Warning("ipfs add failed: ", err)
		return nil, err
	}

	// Copy file to ipfs-add's stdin:
	if _, err := io.Copy(stdin, r); err != nil {
		return nil, err
	}

	stdin.Close()
	adder.Wait()

	if hash, err := ioutil.ReadAll(stdout); err != nil {
		return nil, err
	} else {
		return bytes.TrimSpace(hash), nil
	}
}

func Cat(ctx Context, hash []byte) (io.Reader, error) {
	catter := ipfsCommand(ctx, "cat", string(hash))
	stdout, err := catter.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := catter.Start(); err != nil {
		log.Warning("ipfs add failed: ", err)
		return nil, err
	}

	return stdout, nil
}
