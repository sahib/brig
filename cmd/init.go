package cmd

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	e "github.com/pkg/errors"
	"github.com/sahib/brig/backend"
	"github.com/sahib/brig/repo"
	"github.com/urfave/cli"
)

// Init creates a new brig repository at `basePath` with `owner`.
// `password` is used to encrypt it and `backendName` tells `brig` what backend
// to initialize. The port is the port of the brig daemon.
func Init(ctx *cli.Context, basePath, owner, password, backendName string, port int) error {
	if !backend.IsValidName(backendName) {
		return fmt.Errorf("invalid backend name: %v", backendName)
	}

	err := repo.Init(basePath, owner, password, backendName, int64(port))
	if err != nil {
		return e.Wrapf(err, "repo-init")
	}

	ipfsPath := ctx.String("ipfs-path")
	apiDataPath := filepath.Join(ipfsPath, "api")
	apiData, err := ioutil.ReadFile(apiDataPath)
	if err != nil {
		return err
	}

	splitApiData := strings.Split(string(apiData), "/")
	if len(splitApiData) == 0 {
		return fmt.Errorf(
			"failed to read IPFS api port to connect to (at %s): %v",
			apiDataPath,
			err,
		)
	}

	ipfsPort, err := strconv.Atoi(splitApiData[len(splitApiData)-1])
	if err != nil {
		return fmt.Errorf(
			"failed to convert api port to string (at %s): %v",
			apiDataPath,
			err,
		)
	}

	err = repo.OverwriteConfigKey(basePath, "daemon.ipfs_port", int64(ipfsPort))
	if err != nil {
		return err
	}

	backendPath := filepath.Join(basePath, "data", backendName)
	if err := backend.InitByName(backendName, backendPath, ipfsPort); err != nil {
		return e.Wrapf(err, "backend-init")
	}

	return nil
}
