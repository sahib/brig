// +build !windows

package fuse

import (
	"fmt"
	"time"

	"bazil.org/fuse"
	"github.com/sahib/brig/catfs"
	ie "github.com/sahib/brig/catfs/errors"
	"github.com/sahib/brig/repo/hints"
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

type xattrHandler struct {
	get func(cfs *catfs.FS, info *catfs.StatInfo) ([]byte, error)
	set func(cfs *catfs.FS, path string, value []byte) error
}

var xattrMap = map[string]xattrHandler{
	"user.brig.hash.content": xattrHandler{
		get: func(cfs *catfs.FS, info *catfs.StatInfo) ([]byte, error) {
			return []byte(info.ContentHash.B58String()), nil
		},
	},
	"user.brig.hash.tree": xattrHandler{
		get: func(cfs *catfs.FS, info *catfs.StatInfo) ([]byte, error) {
			return []byte(info.TreeHash.B58String()), nil
		},
	},
	"user.brig.hash.backend": xattrHandler{
		get: func(cfs *catfs.FS, info *catfs.StatInfo) ([]byte, error) {
			return []byte(info.BackendHash.B58String()), nil
		},
	},
	"user.brig.pinned": xattrHandler{
		get: func(cfs *catfs.FS, info *catfs.StatInfo) ([]byte, error) {
			if info.IsPinned {
				return []byte("yes"), nil
			}
			return []byte("no"), nil
		},
	},
	"user.brig.explicitly_pinned": xattrHandler{
		get: func(cfs *catfs.FS, info *catfs.StatInfo) ([]byte, error) {
			if info.IsExplicit {
				return []byte("yes"), nil
			}
			return []byte("no"), nil
		},
	},
	"user.brig.hints.encryption": xattrHandler{
		get: func(cfs *catfs.FS, info *catfs.StatInfo) ([]byte, error) {
			return []byte(cfs.Hints().Lookup(info.Path).EncryptionAlgo), nil
		},
		set: func(cfs *catfs.FS, path string, val []byte) error {
			hint := cfs.Hints().Lookup(path)
			hint.EncryptionAlgo = hints.EncryptionHint(val)
			if !hint.IsValid() {
				return fmt.Errorf("bad encryption algorithm: %s", string(val))
			}

			return cfs.Hints().Set(path, hint)
		},
	},
	"user.brig.hints.compression": xattrHandler{
		get: func(cfs *catfs.FS, info *catfs.StatInfo) ([]byte, error) {
			return []byte(cfs.Hints().Lookup(info.Path).CompressionAlgo), nil
		},
		set: func(cfs *catfs.FS, path string, val []byte) error {
			hint := cfs.Hints().Lookup(path)
			hint.CompressionAlgo = hints.CompressionHint(val)
			if !hint.IsValid() {
				return fmt.Errorf("bad compression algorithm: %s", string(val))
			}

			return cfs.Hints().Set(path, hint)
		},
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
	if !ok || handler.get == nil {
		return nil, fuse.ErrNoXattr
	}

	info, err := cfs.Stat(path)
	if err != nil {
		return nil, errorize("getxattr", err)
	}

	return handler.get(cfs, info)
}

func setXattr(cfs *catfs.FS, name, path string, val []byte) error {
	handler, ok := xattrMap[name]
	if !ok || handler.set == nil {
		return fuse.ErrNoXattr
	}

	if err := handler.set(cfs, path, val); err != nil {
		return fuse.EIO
	}

	return nil
}

func notifyChange(m *Mount, d time.Duration) {
	if m.notifier == nil {
		// this can happen in tests.
		return
	}

	time.AfterFunc(d, m.notifier.PublishEvent)
}
