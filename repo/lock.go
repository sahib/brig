package repo

import (
	"os"
	"path/filepath"

	"github.com/disorganizer/brig/store/encrypt"
	"github.com/disorganizer/brig/util/security"
)

const (
	// EncFileSuffix is appended to all encrypted in-repo file paths
	EncFileSuffix = ".locked"
)

func generateKey(id, pwd string) []byte {
	return security.Scrypt([]byte(pwd), []byte(id), 32)
}

// LockFile encrypts `path` with AES and scrypt, using pass and ID as email.
// The resulting file is written to `path` + EncFileSuffix,
// the source file is removed.
func LockFile(ID, pass, path string) error {
	key := generateKey(ID, pass)
	return lockFile(key, ID, pass, path)
}

// LockFiles works like LockFile but generates the key only once.
func LockFiles(ID, pass string, paths []string) error {
	key := generateKey(ID, pass)

	for _, path := range paths {
		if err := lockFile(key, ID, pass, path); err != nil {
			return err
		}
	}

	return nil
}

func lockFile(key []byte, ID, pass, path string) error {
	srcFd, err := os.Open(path)
	if err != nil {
		return err
	}

	defer srcFd.Close()

	dir, base := filepath.Split(path)
	dstPath := filepath.Join(dir, base+EncFileSuffix)
	dstFd, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}

	defer dstFd.Close()

	if _, err := encrypt.Encrypt(key, srcFd, dstFd); err != nil {
		return err
	}

	if err := os.Remove(path); err != nil {
		return err
	}

	return nil
}

// unlockFileReal is the actual implementation of TryUnlock/UnlockFile
func unlockFileReal(key []byte, ID, pass, path string) error {
	srcPath := path + EncFileSuffix
	srcFd, err := os.Open(srcPath)
	if err != nil {
		return err
	}

	defer srcFd.Close()

	dstPath := path
	dstFd, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}

	defer dstFd.Close()

	if _, err := encrypt.Decrypt(key, srcFd, dstFd); err != nil {
		return err
	}

	if err := os.Remove(srcPath); err != nil {
		return err
	}

	return nil
}

// UnlockFile reverses the effect of LockFile.
//
// NOTE: `path` is the path without EncFileSuffix,
//        i.e. the same path as given to LockFile!
//
// If the operation was successful,
func UnlockFile(ID, pass, path string) error {
	key := generateKey(ID, pass)
	return unlockFileReal(key, ID, pass, path)
}

// UnlockFiles works like UnlockFile for many paths, but generates keys just once.
func UnlockFiles(ID, pass string, paths []string) error {
	key := generateKey(ID, pass)

	for _, path := range paths {
		if err := unlockFileReal(key, ID, pass, path); err != nil {
			return err
		}
	}

	return nil
}
