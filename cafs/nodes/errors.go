package nodes

import "errors"

var (
	// ErrExists is returned if a node already exists at a path, but should not.
	ErrExists = errors.New("File exists")
	// ErrBadNode is returned when a wrong node type was passed to a method.
	ErrBadNode = errors.New("Cannot convert to concrete type. Broken input data?")
)

type errNoSuchFile struct {
	path string
}

// Error will return an error description detailin what path is missing.
func (e *errNoSuchFile) Error() string {
	return "No such file or directory: " + e.path
}

//////////////

// NoSuchFile creates a new error that reports `path` as missing
func NoSuchFile(path string) error {
	return &errNoSuchFile{path}
}

// IsNoSuchFileError asserts that `err` means that the file could not be found
func IsNoSuchFileError(err error) bool {
	_, ok := err.(*errNoSuchFile)
	return ok
}
