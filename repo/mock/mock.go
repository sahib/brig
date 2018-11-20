package mock

import (
	h "github.com/sahib/brig/util/hashlib"
)

// RepoBackend is used to fake a repo backend in testing.
type RepoBackend struct{}

// Init does nothing to create a mock repo.
func (mrb *RepoBackend) Init(folder string) error {
	return nil
}

// GC does nothing.
func (mrb *RepoBackend) GC() ([]h.Hash, error) {
	return nil, nil
}

// Name returns "mock"
func (mrb *RepoBackend) Name() string {
	return "mock"
}

// NewMockRepoBackend returns a new RepoBackend
func NewMockRepoBackend() *RepoBackend {
	return &RepoBackend{}
}
