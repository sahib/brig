package mock

import (
	"io"

	h "github.com/sahib/brig/util/hashlib"
)

type MockRepoBackend struct{}

func (mrb *MockRepoBackend) Init(folder string) error {
	return nil
}

func (mrb *MockRepoBackend) ForwardLog(w io.Writer) {
}

func (mrb *MockRepoBackend) GC() ([]h.Hash, error) {
	return nil, nil
}

func (mrb *MockRepoBackend) Name() string {
	return "mock"
}

func NewMockRepoBackend() *MockRepoBackend {
	return &MockRepoBackend{}
}
