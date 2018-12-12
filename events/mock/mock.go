package mock

import (
	"context"

	eventsBackend "github.com/sahib/brig/events/backend"
)

// TODO: Actually implement something that works on localhost.

// EventsBackend fakes the event backend by setting up a very basic
// message broker on localhost and tunneling all messages over it.
type EventsBackend struct {
}

// NewEventsBackend returns a new EventsBackend
func NewEventsBackend() *EventsBackend {
	return &EventsBackend{}
}

// Subscribe is a mock implementation meant for testing.
func (mb *MockBackend) Subscribe(ctx context.Context, topic string) (eventsBackend.Subscription, error) {
	return nil, nil
}

// PublishEvent is a mock implementation meant for testing.
func (mb *MockBackend) PublishEvent(topic string, data []byte) error {
	return nil
}
