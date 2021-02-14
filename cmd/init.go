package cmd

import (
	e "github.com/pkg/errors"
	"github.com/sahib/brig/repo"
	"github.com/urfave/cli"
)

// Init creates a new brig repository at `basePath` with specified options.
func Init(ctx *cli.Context, ipfsPathOrURL string, opts repo.InitOptions) error {
	if err := repo.Init(opts); err != nil {
		return e.Wrapf(err, "repo-init")
	}

	// Remember the ipsf connection details,
	// so we can start it later.
	return repo.OverwriteConfigKey(
		opts.BaseFolder,
		"daemon.ipfs_path_or_url",
		ipfsPathOrURL,
	)
}
