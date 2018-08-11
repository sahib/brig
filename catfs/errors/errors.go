package catfs

import (
	"errors"
	"fmt"
)

type ExpectedError interface {
	error
	ShouldStopBatch() bool
}

// ExpectedError is a kind of error that
type expectedError struct {
	err             error
	shouldStopBatch bool
}

func (ee *expectedError) ShouldStopBatch() bool {
	return ee.shouldStopBatch
}

func (ee *expectedError) Error() string {
	return ee.err.Error()
}

func NewExpectedError(msg string, shouldStopBatch bool) ExpectedError {
	return &expectedError{
		err:             errors.New(msg),
		shouldStopBatch: shouldStopBatch,
	}
}

var (
	ErrNotEmpty      = NewExpectedError("Cannot remove: Directory is not empty", false)
	ErrStageNotEmpty = NewExpectedError("There are changes in the staging area. Use the --force", false)
	ErrNoChange      = NewExpectedError("Nothing changed between the given versions", false)
	ErrAmbigiousRev  = NewExpectedError("There is more than one rev with this prefix", false)
)

type ErrBadNodeType int

func (e ErrBadNodeType) Error() string {
	return fmt.Sprintf("Bad node type in db: %d", int(e))
}

func (e ErrBadNodeType) ShouldStopBatch() bool {
	return false
}

//////////////

type ErrNoHashFound struct {
	b58hash string
	where   string
}

func (e ErrNoHashFound) Error() string {
	return fmt.Sprintf("No such hash in `%s`: '%s'", e.where, e.b58hash)
}

func (e ErrNoHashFound) ShouldStopBatch() bool {
	return false
}

//////////////

type ErrNoSuchRef string

func (e ErrNoSuchRef) Error() string {
	return fmt.Sprintf("No ref found named `%s`", string(e))
}

func (e ErrNoSuchRef) ShouldStopBatch() bool {
	return false
}

func IsErrNoSuchRef(err error) bool {
	_, ok := err.(ErrNoSuchRef)
	return ok
}

type ErrInvalidRefSpec struct {
	input string
	cause string
}

func (e ErrInvalidRefSpec) Error() string {
	return fmt.Sprintf("Invalid ref `%s`: %s", e.input, e.cause)
}

func (e ErrInvalidRefSpec) ShouldStopBatch() bool {
	return false
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
