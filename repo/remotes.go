package transfer

import (
	"errors"

	"github.com/disorganizer/brig/id"
)

var (
	ErrNoSuchRemote = errors.New("No remote found with this id")
)

type Remote interface {
	ID() id.ID
	Hash() string
}

type Remotes interface {
	Add(ID id.ID, hash string) error
	Remove(ID id.ID) error
	Set(ID id.ID, hash string) error
}
