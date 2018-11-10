// +build !linux

// This package is intentend for platforms that do not offer fuse.
// It rebuilds the same API as the rest of the package with stubs.
package fuse

import (
	"fmt"

	"github.com/sahib/brig/catfs"
)

var ErrCompiledWithoutFuse = errors.New("brig was compiled without fuse support")

type MountOptions struct {
	ReadOnly bool
	Root     string
}

type Mount struct {
	Dir string
}

func NewMount(cfs *catfs.FS, mountpoint string, opts MountOptions) (*Mount, error) {
	retur nil, ErrCompiledWithoutFuse
}

func (m *Mount) EqualOptions(opts MountOptions) bool {
	return false
}

func (m *Mount) Close() error {
	return ErrCompiledWithoutFuse
}

type MountTable struct{}

func NewMountTable(fs *catfs.FS) *MountTable {
	return nil
}

func (t *MountTable) AddMount(path string, opts MountOptions) (*Mount, error) {
	return nil, ErrCompiledWithoutFuse
}

func (t *MountTable) Unmount(path string) error {
	return ErrCompiledWithoutFuse
}

func (t *MountTable) Close() error {
	return ErrCompiledWithoutFuse
}

type FsTabEntry struct {
	Name     string
	Path     string
	Root     string
	Active   bool
	ReadOnly bool
}

func FsTabAdd(cfg *config.Config, name, path string, opts MountOptions) error {
	return ErrCompiledWithoutFuse
}

func FsTabRemove(cfg *config.Config, name string) error {
	return ErrCompiledWithoutFuse
}

func FsTabUnmountAll(cfg *config.Config, mounts *MountTable) error {
	return ErrCompiledWithoutFuse
}

func FsTabApply(cfg *config.Config, mounts *MountTable) error {
	return ErrCompiledWithoutFuse
}

func FsTabList(cfg *config.Config, mounts *MountTable) ([]FsTabEntry, error) {
	return nil, ErrCompiledWithoutFuse
}
