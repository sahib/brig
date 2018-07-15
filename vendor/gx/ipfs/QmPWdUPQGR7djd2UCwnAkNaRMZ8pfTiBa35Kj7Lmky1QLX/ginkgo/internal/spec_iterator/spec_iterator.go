package spec_iterator

import (
	"errors"

	"gx/ipfs/QmPWdUPQGR7djd2UCwnAkNaRMZ8pfTiBa35Kj7Lmky1QLX/ginkgo/internal/spec"
)

var ErrClosed = errors.New("no more specs to run")

type SpecIterator interface {
	Next() (*spec.Spec, error)
	NumberOfSpecsPriorToIteration() int
	NumberOfSpecsToProcessIfKnown() (int, bool)
	NumberOfSpecsThatWillBeRunIfKnown() (int, bool)
}

type Counter struct {
	Index int `json:"index"`
}
