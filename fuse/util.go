// +build !windows

package fuse

import (
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

type xattrGetter func(cfs *catfs.FS, info *catfs.StatInfo) ([]byte, error)

var xattrMap = map[string]xattrGetter{
	"user.brig.hash.content": func(cfs *catfs.FS, info *catfs.StatInfo) ([]byte, error) {
		return []byte(info.ContentHash.B58String()), nil
	},
	"user.brig.hash.tree": func(cfs *catfs.FS, info *catfs.StatInfo) ([]byte, error) {
		return []byte(info.TreeHash.B58String()), nil
	},
	"user.brig.hash.backend": func(cfs *catfs.FS, info *catfs.StatInfo) ([]byte, error) {
		return []byte(info.BackendHash.B58String()), nil
	},
	"user.brig.pinned": func(cfs *catfs.FS, info *catfs.StatInfo) ([]byte, error) {
		if info.IsPinned {
			return []byte("yes"), nil
		}
		return []byte("no"), nil
	},
	"user.brig.explicitly_pinned": func(cfs *catfs.FS, info *catfs.StatInfo) ([]byte, error) {
		if info.IsExplicit {
			return []byte("yes"), nil
		}
		return []byte("no"), nil
	},
}

func listXattr() []byte {
	resp := []byte{}
	for k := range xattrMap {
		resp = append(resp, k...)
		resp = append(resp, '\x00')
	}

	return resp
}

func getXattr(cfs *catfs.FS, name, path string) ([]byte, error) {
	handler, ok := xattrMap[name]
	if !ok {
		return nil, fuse.ErrNoXattr
	}

	info, err := cfs.Stat(path)
	if err != nil {
		return nil, errorize("getxattr", err)
	}

	return handler(cfs, info)
}

func notifyChange(m *Mount, d time.Duration) {
	if m.notifier == nil {
		// this can happen in tests.
		return
	}

	time.AfterFunc(d, m.notifier.PublishEvent)
}
