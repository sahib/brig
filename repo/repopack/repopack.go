package repopack

import (
	"fmt"
	"io"
	"os"

	"github.com/sahib/brig/catfs/mio/encrypt"
	"github.com/sahib/brig/util"
)

func tarAndEncrypt(dir string, key []byte, w io.Writer) error {
	encw, err := encrypt.NewWriter(w, key)
	if err != nil {
		return err
	}

	defer encw.Close()
	return util.Tar(
		dir,
		fmt.Sprintf("brig repo archive of %s", dir),
		encw,
	)
}

func untarAndDecrypt(tarReader io.Reader, key []byte, outputDir string) error {
	encr, err := encrypt.NewReader(tarReader, key)
	if err != nil {
		return err
	}

	return util.Untar(encr, outputDir)
}

func keyFromPassword(password string) []byte {
	// NOTE: we would need to add a static salt in the front of the archive...
	//       if that gets damaged we would not be able to unlock it though.
	//       So for now we just a constant salt.
	return util.DeriveKey([]byte(password), []byte("constant"), 32)
}

// PackRepo archives the repository at `folder` to `archivePath` using `password`.
// if `removeRepo` is true, we remove the repository after.
func PackRepo(folder, archivePath, password string, removeRepo bool) error {
	fd, err := os.OpenFile(archivePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	if err := tarAndEncrypt(folder, keyFromPassword(password), fd); err != nil {
		fd.Close()
		return err
	}

	// Make sure to close the handle to catch any errors,
	// before potentially removing the old repo. Would be
	// embarrassing to first remove it & then notice we failed.
	if err := fd.Close(); err != nil {
		return err
	}

	if removeRepo {
		return os.RemoveAll(folder)
	}

	return nil
}

// UnpackRepo unpacks the tar at `archivePath` using `password` and puts the
// resulting repository at `folder`. If `removeArchive` is true the archive is
// removed after.
func UnpackRepo(folder, archivePath, password string, removeArchive bool) error {
	fd, err := os.Open(archivePath)
	if err != nil {
		return err
	}

	defer fd.Close()

	key := keyFromPassword(password)
	if err := untarAndDecrypt(fd, key, folder); err != nil {
		return err
	}

	if removeArchive {
		return os.Remove(archivePath)
	}

	return nil
}
