package repo

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cathalgarvey/go-minilock"
	"github.com/cathalgarvey/go-minilock/taber"
)

const (
	// EncFileSuffix is appended to all encrypted in-repo file paths
	EncFileSuffix = ".minilock"
)

// LockFile encrypts `path` with minilock, using pass and ID as email.
// The resulting file is written to `path` + EncFileSuffix,
// the source file is removed.
func LockFile(ID, pass, path string) error {
	keys, err := minilock.GenerateKey(ID, pass)
	if err != nil {
		return err
	}

	return lockFile(keys, ID, pass, path)
}

// LockFiles works like LockFile but generates the key only once.
func LockFiles(ID, pass string, paths []string) error {
	keys, err := minilock.GenerateKey(ID, pass)
	if err != nil {
		return err
	}

	for _, path := range paths {
		if err := lockFile(keys, ID, pass, path); err != nil {
			return err
		}
	}

	return nil
}

func lockFile(keys *taber.Keys, ID, pass, path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	var encData []byte
	dir, base := filepath.Split(path)

	// This seemed to crash minilock otherwise:
	if len(data) != 0 {
		if encData, err = minilock.EncryptFileContents(base, data, keys, keys); err != nil {
			return err
		}
	}

	err = ioutil.WriteFile(filepath.Join(dir, base+EncFileSuffix), encData, 0666)
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil {
		return err
	}

	return nil
}

// unlockFileReal is the actual implementation of TryUnlock/UnlockFile
func unlockFileReal(keys *taber.Keys, ID, pass, path string, write bool) error {
	encPath := path + EncFileSuffix
	data, err := ioutil.ReadFile(encPath)
	if err != nil {
		return err
	}

	_, decName, decData, err := minilock.DecryptFileContents(data, keys)
	if err != nil {
		return err
	}

	if !write {
		return nil
	}

	decPath := filepath.Join(filepath.Dir(encPath), decName)
	err = ioutil.WriteFile(decPath, decData, 0666)
	if err != nil {
		return err
	}

	if err := os.Remove(encPath); err != nil {
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
	keys, err := minilock.GenerateKey(ID, pass)
	if err != nil {
		return err
	}

	return unlockFileReal(keys, ID, pass, path, true)
}

// UnlockFiles works like UnlockFile for many paths, but generates keys just once.
func UnlockFiles(ID, pass string, paths []string) error {
	keys, err := minilock.GenerateKey(ID, pass)
	if err != nil {
		return err
	}

	for _, path := range paths {
		if err := unlockFileReal(keys, ID, pass, path, true); err != nil {
			return err
		}
	}

	return nil
}

// TryUnlock tries to unlock a file, if successful,
// `path` will not be removed and no encrypted output is written.
func TryUnlock(ID, pass, path string) error {
	keys, err := minilock.GenerateKey(ID, pass)
	if err != nil {
		return err
	}

	return unlockFileReal(keys, ID, pass, path, false)
}

// EncryptMinilockMsg encrypts a given plaintext for multiple receivers.
func EncryptMinilockMsg(ID, pass, plaintext string, mid ...string) (string, error) {
	ciphertext, err := minilock.EncryptFileContentsWithStrings(
		"Minilock Filename.",
		[]byte(plaintext),
		ID, pass, false, mid...,
	)
	if err != nil {
		return "", nil
	}
	return string(ciphertext), nil
}

// DecryptMinilockMsg decrypts a given ciphertext.
func DecryptMinilockMsg(ID, pass, ciphertext string) (string, error) {
	userKey, err := minilock.GenerateKey(ID, pass)
	if err != nil {
		return "", nil
	}
	_, _, plaintext, _ := minilock.DecryptFileContents([]byte(ciphertext), userKey)
	return string(plaintext), nil
}

// GenerateMinilockID generates a base58-encoded pubkey + 1-byte blake2s checksum as a string
func GenerateMinilockID(ID, pass string) (string, error) {
	keys, err := minilock.GenerateKey(ID, pass)
	if err != nil {
		return "", err
	}
	return keys.EncodeID()
}
