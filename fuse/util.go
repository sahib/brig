// +build !windows

package fuse

import (
	"bytes"
	"time"

	"bazil.org/fuse"
	"github.com/sahib/brig/catfs"
	ie "github.com/sahib/brig/catfs/errors"
	log "github.com/sirupsen/logrus"
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
// A rather inconvenient behaviour of fuse is to not report panics.
func logPanic(name string) {
	if err := recover(); err != nil {
		log.Errorf("bug: %s panicked: %v", name, err)
	}
}

func listXattr() []byte {
	resp := []byte{}
	resp = append(resp, "user.brig.hash\x00"...)
	resp = append(resp, "user.brig.content\x00"...)
	resp = append(resp, "user.brig.pinned\x00"...)
	resp = append(resp, "user.brig.explicitly_pinned\x00"...)

	return resp
}

func isKnownAttribute(name string) bool {
	knownAttrs := bytes.Split(listXattr(), []byte{0})
	for _, attr := range knownAttrs {
		if string(attr) == name {
			return true
		}
	}
	return false
}

func getXattr(cfs *catfs.FS, name, path string, size uint32) ([]byte, error) {
	if !isKnownAttribute(name) {
		return nil, fuse.ErrNoXattr
	}

	info, err := cfs.Stat(path)
	if err != nil {
		return nil, errorize("getxattr", err)
	}

	resp := []byte{}

	switch name {
	case "user.brig.hash":
		resp = []byte(info.TreeHash.B58String())
	case "user.brig.content":
		resp = []byte(info.ContentHash.B58String())
	case "user.brig.pinned":
		if info.IsPinned {
			resp = []byte("yes")
		} else {
			resp = []byte("no")
		}
	case "user.brig.explicitly_pinned":
		if info.IsExplicit {
			resp = []byte("yes")
		} else {
			resp = []byte("no")
		}
	default:
		return nil, fuse.ErrNoXattr
	}

	// Do not worry about req.Size
	// fuse will cut it to allowed size and report to the caller that buffer need to be larger
	return resp, nil
}

func notifyChange(m *Mount, d time.Duration) {
	if m.notifier == nil {
		// this can happen in tests.
		return
	}

	time.AfterFunc(d, m.notifier.PublishEvent)
}
