package ipfsutil

import "os/exec"

// Context remembers the settings needed for accessing the ipfs daemon.
type Context struct {
	// TODO!
	Path string
}

func ipfsCommand(ctx *Context, args ...string) *exec.Cmd {
	cmd := exec.Command("ipfs", args...)
	cmd.Env = []string{"IPFS_PATH=" + ctx.Path}
	return cmd
}
