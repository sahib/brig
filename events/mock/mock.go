package mock

import (
	"context"

	eventsBackend "github.com/sahib/brig/events/backend"
)

// TODO: Actually implement something that works on localhost.

type MockBackend struct {
}

func (mb *MockBackend) Subscribe(ctx context.Context, topic string) (eventsBackend.Subscription, error) {
	return nil, nil
}

func (mb *MockBackend) PublishEvent(topic string, data []byte) error {
	return nil
}
