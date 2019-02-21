package cmd

import (
	"fmt"
	"path/filepath"

	e "github.com/pkg/errors"
	"github.com/sahib/brig/backend"
	"github.com/sahib/brig/repo"
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

	backendPath := filepath.Join(basePath, "data", backendName)
	if err := backend.InitByName(backendName, backendPath); err != nil {
		return e.Wrapf(err, "backend-init")
	}

	return nil
}
