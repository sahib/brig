package remotesapi

import (
	"fmt"

	"github.com/sahib/brig/catfs"
)

// Mock is for testing purposes whenever a normal RemotesAPI is needed.
// It stores remotes in memory and does not implement realy syncing or diffing.
type Mock struct {
	name        string
	fingerprint string
	remotes     map[string]*Remote
	callbacks   []func()
}

// NewMock creates a new Mock.
// `name` and `fingerprint` will be returned
func NewMock(name, fingerprint string) *Mock {
	return &Mock{
		name:        name,
		fingerprint: fingerprint,
		remotes:     make(map[string]*Remote),
	}
}

func (m *Mock) List() ([]*Remote, error) {
	rmts := []*Remote{}
	for _, rmt := range m.remotes {
		rmts = append(rmts, rmt)
	}

	return rmts, nil
}

func (m *Mock) Get(name string) (*Remote, error) {
	rm, ok := m.remotes[name]
	if !ok {
		return nil, fmt.Errorf("no such remote: %s", name)
	}

	return rm, nil
}

func (m *Mock) Set(rm Remote) error {
	if rm.Name == "" {
		return fmt.Errorf("empty name")
	}

	if rm.Fingerprint == "" {
		return fmt.Errorf("empty fingerprint")
	}

	if rm.Name == m.name {
		return fmt.Errorf("cannot add remote with own name")
	}

	m.remotes[rm.Name] = &rm
	m.notify()
	return nil
}

func (m *Mock) Remove(name string) error {
	if _, ok := m.remotes[name]; !ok {
		return fmt.Errorf("no such remote: %s", name)
	}

	delete(m.remotes, name)
	m.notify()
	return nil
}

func (m *Mock) Self() (Identity, error) {
	return Identity{
		Name:        m.name,
		Fingerprint: m.fingerprint,
	}, nil
}

func (m *Mock) Sync(name string) error {
	if _, ok := m.remotes[name]; !ok {
		return fmt.Errorf("no such remote: %s", name)
	}

	return nil
}

func (m *Mock) MakeDiff(name string) (*catfs.Diff, error) {
	if _, ok := m.remotes[name]; !ok {
		return nil, fmt.Errorf("no such remote: %s", name)
	}

	// always send an empty diff.
	return &catfs.Diff{}, nil
}

func (m *Mock) notify() {
	for _, fn := range m.callbacks {
		fn()
	}
}

func (m *Mock) OnChange(fn func()) {
	m.callbacks = append(m.callbacks, fn)
}
