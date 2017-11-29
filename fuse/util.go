package fuse

import (
	"bazil.org/fuse"
	log "github.com/Sirupsen/logrus"
	ie "github.com/disorganizer/brig/catfs/errors"
)

func errorize(name string, err error) error {
	if ie.IsNoSuchFileError(err) {
		log.Infof("errorize: No such file: %v", err)
		return fuse.ENOENT
	}

	if err != nil {
		log.Warningf("fuse: %s: %v", name, err)
		return fuse.EIO
	}

	return nil
}
