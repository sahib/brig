package catfs

import (
	"errors"
	"fmt"
)

var (
	// ErrStageNotEmpty is returned by Reset() when it was called without force.
	// and there are still changes in the staging area.
	ErrStageNotEmpty = errors.New("there are changes in the staging area; use the --force")

	// ErrNoChange is returned when trying to commit, but there is no change.
	ErrNoChange = errors.New("nothing changed between the given versions")

	// ErrAmbigiousRev is returned when a ref string could mean several commits.
	ErrAmbigiousRev = errors.New("there is more than one rev with this prefix")

	// ErrExists is returned if a node already exists at a path, but should not.
	ErrExists = errors.New("File exists")

	// ErrBadNode is returned when a wrong node type was passed to a method.
	ErrBadNode = errors.New("Cannot convert to concrete type. Broken input data?")
)

//////////////

// ErrNoSuchRef is returned when a bad ref was used
type ErrNoSuchRef string

func (e ErrNoSuchRef) Error() string {
	return fmt.Sprintf("No ref found named `%s`", string(e))
}

// IsErrNoSuchRef checks if `err` is a no such ref error.
func IsErrNoSuchRef(err error) bool {
	_, ok := err.(ErrNoSuchRef)
	return ok
}

/////////////////

// ErrNoSuchCommitIndex is returned when a bad commit was used
type ErrNoSuchCommitIndex struct {
	index int64
}

func (e ErrNoSuchCommitIndex) Error() string {
	return fmt.Sprintf("No commit with index `%d` found", e.index)
}

func NoSuchCommitIndex(ind int64) error {
	return &ErrNoSuchCommitIndex{ind} 
}

// IsErrNoSuchRef checks if `err` is a no such ref error.
func IsErrNoSuchCommitIndex(err error) bool {
	_, ok := err.(*ErrNoSuchCommitIndex)
	return ok
}

/////////////////

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
