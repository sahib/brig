package fusetest

import (
	"os"
	"os/exec"
)

// LaunchAsProcess will start the fusemock server in another process.
// This will serve a fuse mount on the specified directory and will listen
// to commands on a http socket.
//
// NOTE: This will only work if you call this from a part of the main brig
//       executable. This relies on the executable to do start the server
//       when being called as »$0 debug fusemock«. Therefore this will not
//       work in tests, but it should be easy to adapt.
//
// The returned process can be used to terminate the program.
// You should use the provided Dial() / Quit method to cleanup though.
func LaunchAsProcess(opts Options) (*os.Process, error) {
	myself, err := os.Executable()
	if err != nil {
		return nil, err
	}

	args := []string{
		"debug",
		"fusemock",
		"--mount-path", opts.MountPath,
		"--catfs-path", opts.CatfsPath,
		"--ipfs-path-or-url", opts.IpfsPathOrURL,
		"--url", opts.URL,
	}

	if opts.MountReadOnly {
		args = append(args, "--mount-ro")
	}

	if opts.MountOffline {
		args = append(args, "--mount-offline")
	}

	cmd := exec.Command(myself, args...)
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return cmd.Process, nil
}
