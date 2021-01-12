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

// Init creates a new brig repository at `basePath` with specified options.
func Init(ctx *cli.Context, ipfsPath string, opts repo.InitOptions) error {
	if err := repo.Init(opts); err != nil {
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

	if err := repo.OverwriteConfigKey(
		opts.BaseFolder,
		"daemon.ipfs_path",
		ipfsPath,
	); err != nil {
		return err
	}

	backendPath := filepath.Join(opts.BaseFolder, "data", opts.BackendName)
	if err := backend.InitByName(
		opts.BackendName,
		backendPath,
		ipfsPort,
	); err != nil {
		return e.Wrapf(err, "backend-init")
	}

	return nil
}
