package repo

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/disorganizer/brig/util"
	"github.com/disorganizer/brig/util/security"
)

func hashPassword(salt []byte, password string) []byte {
	return security.Scrypt([]byte(password), salt, 32)
}

func createShadowFile(brigPath string, ID, password string) error {
	shadowPath := filepath.Join(brigPath, "shadow")
	fd, err := os.OpenFile(shadowPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}

	defer util.Closer(fd)

	salt := make([]byte, 8)
	n, err := rand.Reader.Read(salt)
	if err != nil {
		return err
	}

	if n != len(salt) {
		return fmt.Errorf("Inadeqaute salt length from random generator.")
	}

	entry := fmt.Sprintf("%s %x %x\n", ID, salt, hashPassword(salt, password))
	if _, err := fd.Write([]byte(entry)); err != nil {
		return err
	}

	return nil
}

type shadowEntry struct {
	ID   string
	salt []byte
	hash []byte
}

func parseShadowFile(brigPath string, who string) (*shadowEntry, error) {
	fd, err := os.Open(filepath.Join(brigPath, "shadow"))
	if err != nil {
		return nil, err
	}

	defer util.Closer(fd)

	var entry *shadowEntry
	bufd := bufio.NewScanner(fd)
	for bufd.Scan() {
		var ID string
		var salt, hash []byte

		_, err = fmt.Sscanf(bufd.Text(), "%s %x %x", &ID, &salt, &hash)
		if err != nil && err != io.EOF {
			return nil, err
		}

		if ID == who {
			entry = &shadowEntry{ID: ID, salt: salt, hash: hash}
		}
	}

	if err := bufd.Err(); err != nil {
		return nil, err
	}

	// Might be a broken shadow file:
	if entry == nil {
		return nil, fmt.Errorf("No shadow entry found for `%v`.", who)
	}

	return entry, nil
}
