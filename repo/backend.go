package repo

type RepoBackend interface {
	Init(path string) error
}

type DummyBackend struct{}

func (db DummyBackend) Init(path string) error {
	return nil
}
