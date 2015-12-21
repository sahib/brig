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
	// "otr.buddies",
	// "otr.key",
}

func lookupJid(configPath string) (string, error) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return "", fmt.Errorf("Could not load config: %v", err)
	}

	jid, err := cfg.String("repository.jid")
	if err != nil {
		return "", fmt.Errorf("No jid in config: %v", err)
	}

	return jid, nil
}

// Open unencrypts all sensible data in the repository.
func Open(pwd, folder string) (*Repository, error) {
	absFolderPath, err := filepath.Abs(folder)
	brigPath := filepath.Join(absFolderPath, ".brig")

	// Figure out the JID from the config:
	jid, err := lookupJid(filepath.Join(brigPath, "config"))
	if err != nil {
		return nil, err
	}

	// Unlock all files:
	absNames := make([]string, 0)
	for _, name := range filenames {
		absName := filepath.Join(brigPath, name)
		if _, err := os.Stat(absName); err == nil {
			// File exists, this might happen on a crash or killed daemon.
			log.Warningf("File is already unlocked: %s", absName)
			continue
		}

		absNames = append(absNames, absName)
	}

	if err := UnlockFiles(jid, pwd, absNames); err != nil {
		return nil, err
	}

	return LoadRepository(pwd, absFolderPath)
}

// Close encrypts sensible files in the repository.
// The password is taken from Repository.Password.
func (r *Repository) Close() error {
	absNames := make([]string, 0)

	for _, name := range filenames {
		absName := filepath.Join(r.InternalFolder, name)
		if _, err := os.Stat(absName); os.IsNotExist(err) {
			// File does not exist. Might be already locked.
			log.Warningf("File is already locked: %s", absName)
			continue
		}

		log.Infof("Locking file `%v`...", absName)
		absNames = append(absNames, absName)
	}

	fmt.Println(absNames)
	if err := LockFiles(r.Jid, r.Password, absNames); err != nil {
		return err
	}

	return nil
}

// CheckPassword tries to decrypt a file in the repository.
// If that does not work, an error is returned.
func CheckPassword(folder, pwd string) error {
	absFolderPath, err := filepath.Abs(folder)
	brigPath := filepath.Join(absFolderPath, ".brig")

	jid, err := lookupJid(filepath.Join(brigPath, "config"))
	if err != nil {
		return err
	}

	absName := filepath.Join(brigPath, "master.key")
	if err := TryUnlock(jid, pwd, absName); err != nil {
		return err
	}

	return nil
}
