package repo

import (
	"fmt"
	"path/filepath"

	"github.com/disorganizer/brig/repo/config"
)

// Filenames that will be encrypted on close:
var filenames = []string{
	"index.bolt",
	"master.key",
}

func Open(folder, pwd string) (*FsRepository, error) {
	absFolderPath, err := filepath.Abs(folder)
	brigPath := filepath.Join(absFolderPath, ".brig")

	// Figure out the JID from the config:
	cfg, err := config.LoadConfig(filepath.Join(brigPath, "config"))
	jid, err := cfg.String("repository.jid")
	if err != nil {
		return nil, fmt.Errorf("No jid in config: %v", err)
	}

	// Unlock all files:
	for _, name := range filenames {
		absName := filepath.Join(brigPath, name)
		if err := UnlockFile(jid, pwd, absName); err != nil {
			return nil, err
		}
	}

	return LoadFsRepository(absFolderPath)
}

func (r *FsRepository) Close() error {
	for _, name := range filenames {
		absName := filepath.Join(r.InternalFolder, name)

		if err := LockFile(r.Jid, r.Password, absName); err != nil {
			return err
		}
	}

	return nil
}
