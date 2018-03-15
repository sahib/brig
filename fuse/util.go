package fuse

import (
	"bazil.org/fuse"
	log "github.com/Sirupsen/logrus"
	ie "github.com/sahib/brig/catfs/errors"
)

func errorize(name string, err error) error {
	if ie.IsNoSuchFileError(err) {
		log.Infof("errorize: %s: No such file: %v", name, err)
		return fuse.ENOENT
	}

	if err != nil {
		log.Warningf("fuse: %s: %v", name, err)
		return fuse.EIO
	}

	return nil
}

// logPanic logs any panics by being called in a defer.
// A rather inconvinient beahviour of fuse is to not report panics.
func logPanic(name string) {
	if err := recover(); err != nil {
		log.Errorf("bug: %s panicked: %v", name, err)
	}
}
