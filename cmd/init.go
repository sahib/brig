package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	e "github.com/pkg/errors"
	"github.com/sahib/brig/backend"
	"github.com/sahib/brig/defaults"
	"github.com/sahib/brig/repo"
	"github.com/sahib/config"
	"github.com/urfave/cli"
)

func Init(ctx *cli.Context, basePath, owner, password, backendName string, port int) error {
	if !backend.IsValidName(backendName) {
		return fmt.Errorf("invalid backend name: %v", backendName)
	}

	err := repo.Init(basePath, owner, password, backendName, int64(port))
	if err != nil {
		return e.Wrapf(err, "repo-init")
	}

	ipfsPort := ctx.Int("ipfs-port")

	configPath := filepath.Join(basePath, "config.yml")
	cfg, err := defaults.OpenMigratedConfig(configPath)
	if err != nil {
		return e.Wrapf(err, "failed to set ipfs port")
	}

	cfg.SetInt("daemon.ipfs_port", int64(ipfsPort))
	fd, err := os.OpenFile(configPath, os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	defer fd.Close()

	if err := cfg.Save(config.NewYamlEncoder(fd)); err != nil {
		return err
	}

	backendPath := filepath.Join(basePath, "data", backendName)
	if err := backend.InitByName(backendName, backendPath, ipfsPort); err != nil {
		return e.Wrapf(err, "backend-init")
	}

	return nil
}
