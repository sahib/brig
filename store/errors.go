package store

import (
	"errors"
	"fmt"
)

var (
	ErrExists   = errors.New("File exists")
	ErrNotEmpty = errors.New("Cannot remove: Directory is not empty")
	ErrBadNode  = errors.New("Cannot convert to concrete type. Broken input data?")
)

type errNoSuchFile struct {
	path string
}

func (e *errNoSuchFile) Error() string {
	return "No such file or directory: " + e.path
}

// NoSuchFile creates a new error that reports `path` as missing
// TODO: move to errors.go?
func NoSuchFile(path string) error {
	return &errNoSuchFile{path}
}

// IsNoSuchFileError asserts that `err` means that the file could not be found
func IsNoSuchFileError(err error) bool {
	_, ok := err.(*errNoSuchFile)
	return ok
}

type ErrBadNodeType int

func (e ErrBadNodeType) Error() string {
	return fmt.Sprintf("Bad node type in db: %d", int(e))
}

type ErrNoHashFound struct {
	b58hash string
	where   string
}

func (e ErrNoHashFound) Error() string {
	return fmt.Sprintf("No such hash in `%s`: '%s'", e.where, e.b58hash)
}
