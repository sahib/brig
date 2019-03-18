package cmd

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	e "github.com/pkg/errors"
	"github.com/sahib/brig/backend"
	"github.com/sahib/brig/repo"
	"github.com/sahib/brig/repo/setup"
	"github.com/urfave/cli"
)

// Init creates a new brig repository at `basePath` with `owner`.
// `password` is used to encrypt it and `backendName` tells `brig` what backend
// to initialize. The port is the port of the brig daemon.
func Init(ctx *cli.Context, basePath, owner, password, backendName, ipfsPath string, port int) error {
	if !backend.IsValidName(backendName) {
		return fmt.Errorf("invalid backend name: %v", backendName)
	}

	err := repo.Init(basePath, owner, password, backendName, int64(port))
	if err != nil {
		return e.Wrapf(err, "repo-init")
	}

	apiAddr, err := setup.GetAPIAddrForPath(ipfsPath)
	if err != nil {
		return e.Wrapf(err, "no config - is »%s« an IPFS repo?", apiAddr)
	}

	splitAPIAddr := strings.Split(string(apiAddr), "/")
	if len(splitAPIAddr) == 0 {
		return fmt.Errorf(
			"failed to read IPFS api port to connect to (at %s): %v",
			ipfsPath,
			err,
		)
	}

	ipfsPort, err := strconv.Atoi(splitAPIAddr[len(splitAPIAddr)-1])
	if err != nil {
		return fmt.Errorf(
			"failed to convert api port to string (at %s): %v",
			ipfsPath,
			err,
		)
	}

	err = repo.OverwriteConfigKey(basePath, "daemon.ipfs_path", ipfsPath)
	if err != nil {
		return err
	}

	backendPath := filepath.Join(basePath, "data", backendName)
	if err := backend.InitByName(backendName, backendPath, ipfsPort); err != nil {
		return e.Wrapf(err, "backend-init")
	}

	return nil
}
