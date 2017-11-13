package mock

type MockRepoBackend struct{}

func (mrb *MockRepoBackend) Init(folder string) error {
	return nil
}

func (mrb *MockRepoBackend) Name() string {
	return "mock"
}

func NewMockRepoBackend() *MockRepoBackend {
	return &MockRepoBackend{}
}
