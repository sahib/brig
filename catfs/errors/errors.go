package catfs

import (
	"errors"
	"fmt"
)

var (
	ErrNotEmpty      = errors.New("Cannot remove: Directory is not empty")
	ErrStageNotEmpty = errors.New("There are changes in the staging area")
	ErrNoChange      = errors.New("Nothing changed between the given versions")
)

type ErrBadNodeType int

func (e ErrBadNodeType) Error() string {
	return fmt.Sprintf("Bad node type in db: %d", int(e))
}

//////////////

type ErrNoHashFound struct {
	b58hash string
	where   string
}

func (e ErrNoHashFound) Error() string {
	return fmt.Sprintf("No such hash in `%s`: '%s'", e.where, e.b58hash)
}

//////////////

type ErrNoSuchRef string

func (e ErrNoSuchRef) Error() string {
	return fmt.Sprintf("No ref found named `%s`", string(e))
}

func IsErrNoSuchRef(err error) bool {
	_, ok := err.(ErrNoSuchRef)
	return ok
}

/////////////////

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
