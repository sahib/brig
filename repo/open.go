package repo

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/repo/config"
)

// Filenames that will be encrypted on close:
var filenames = []string{
	"index.bolt",
	"master.key",
}

func Open(pwd, folder string) (*Repository, error) {
	absFolderPath, err := filepath.Abs(folder)
	brigPath := filepath.Join(absFolderPath, ".brig")

	// Figure out the JID from the config:
	cfg, err := config.LoadConfig(filepath.Join(brigPath, "config"))
	if err != nil {
		return nil, fmt.Errorf("No jid in config: %v", err)
	}

	jid, err := cfg.String("repository.jid")

	// Unlock all files:
	for _, name := range filenames {
		absName := filepath.Join(brigPath, name)
		if _, err := os.Stat(absName); err == nil {
			// File exists, this might happen on a crash or killed daemon.
			log.Warningf("File is already unlocked: %s", absName)
			continue
		}

		if err := UnlockFile(jid, pwd, absName); err != nil {
			return nil, err
		}
	}

	return LoadFsRepository(pwd, absFolderPath)
}

func (r *Repository) Close() error {
	for _, name := range filenames {
		absName := filepath.Join(r.InternalFolder, name)
		if _, err := os.Stat(absName); os.IsNotExist(err) {
			// File does not exist. Might be already locked.
			log.Warningf("File is already locked: %s", absName)
			continue
		}

		if err := LockFile(r.Jid, r.Password, absName); err != nil {
			return err
		}
	}

	return nil
}

func CheckPassword(folder, pwd string) error {
	absFolderPath, err := filepath.Abs(folder)
	brigPath := filepath.Join(absFolderPath, ".brig")

	// Figure out the JID from the config:
	cfg, err := config.LoadConfig(filepath.Join(brigPath, "config"))
	jid, err := cfg.String("repository.jid")
	if err != nil {
		return fmt.Errorf("No jid in config: %v", err)
	}

	absName := filepath.Join(brigPath, "master.key")
	if err := TryUnlock(jid, pwd, absName); err != nil {
		return err
	}

	return nil
}
