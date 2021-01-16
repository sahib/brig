package lock

import (
	"io"

	"github.com/sahib/brig/catfs/mio/encrypt"
	"github.com/sahib/brig/util"
)

func LockRepo(dir string, key []byte, w io.Writer) error {
	encw, err := encrypt.NewWriter(w, key)
	if err != nil {
		return err
	}

	defer encw.Close()
	return util.Tar(dir, "TODO: archive name?", encw)
}

func UnlockRepo(tarReader io.Reader, key []byte, outputDir string) error {
	encr, err := encrypt.NewReader(tarReader, key)
	if err != nil {
		return err
	}

	return util.Untar(encr, outputDir)
}

func KeyFromPassword(password string) []byte {
	// we have no way to store a salt here... TODO: or do we?
	return util.DeriveKey([]byte(password), []byte("constant"), 32)
}
