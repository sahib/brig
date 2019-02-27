package parcello

import (
	"os"
	"path/filepath"
)

var _ FileSystemManager = Dir("")

// Dir implements FileSystem using the native file system restricted to a
// specific directory tree.
type Dir string

// Open opens the named file for reading. If successful, methods on
// the returned file can be used for reading; the associated file
// descriptor has mode O_RDONLY.
// If there is an error, it will be of type *PathError.
func (d Dir) Open(name string) (ReadOnlyFile, error) {
	return d.OpenFile(name, os.O_RDONLY, 0)
}

// OpenFile is the generalized open call; most users will use Open
func (d Dir) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	dir := filepath.Join(string(d), filepath.Dir(name))

	if hasFlag(os.O_CREATE, flag) {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, err
		}
	}

	name = filepath.Join(dir, filepath.Base(name))
	return os.OpenFile(name, flag, perm)
}

// Walk walks the file tree rooted at root, calling walkFn for each file or
// directory in the tree, including root.
func (d Dir) Walk(dir string, fn filepath.WalkFunc) error {
	dir = filepath.Join(string(d), dir)

	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		path, _ = filepath.Rel(string(d), path)
		return fn(path, info, err)
	})
}

// Dir returns a sub-manager for given path
func (d Dir) Dir(name string) (FileSystemManager, error) {
	return Dir(filepath.Join(string(d), name)), nil
}

// Add adds resource bundle to the dir. (noop)
func (d Dir) Add(resource *Resource) error {
	return nil
}
