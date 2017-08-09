package cafs

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
