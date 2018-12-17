package mock

import (
	"context"
	"sync"

	eventsBackend "github.com/sahib/brig/events/backend"
)

var subs map[string][]*mockSubscription
var subsLock sync.Mutex

func init() {
	subs = make(map[string][]*mockSubscription)
}

// EventsBackend fakes the event backend by setting up a very basic
// message broker in memory and tunneling all messages over it.
type EventsBackend struct {
	ownAddr string
}

// NewEventsBackend returns a new EventsBackend
func NewEventsBackend(ownAddr string) *EventsBackend {
	return &EventsBackend{
		ownAddr: ownAddr,
	}
}

type mockMessage struct {
	data   []byte
	source string
}

func (mm mockMessage) Data() []byte {
	return mm.data
}

func (mm mockMessage) Source() string {
	return mm.source
}

type mockSubscription struct {
	msgs chan mockMessage
}

func (ms *mockSubscription) Next(ctx context.Context) (eventsBackend.Message, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case msg := <-ms.msgs:
		return msg, nil
	}
}

func (ms *mockSubscription) Close() error {
	return nil
}

// Subscribe is a mock implementation meant for testing.
func (mb *EventsBackend) Subscribe(ctx context.Context, topic string) (eventsBackend.Subscription, error) {
	subsLock.Lock()
	defer subsLock.Unlock()

	newSub := &mockSubscription{
		msgs: make(chan mockMessage, 100),
	}

	subs[topic] = append(subs[topic], newSub)
	return newSub, nil
}

// PublishEvent is a mock implementation meant for testing.
func (mb *EventsBackend) PublishEvent(topic string, data []byte) error {
	subsLock.Lock()
	defer subsLock.Unlock()

	subs, ok := subs[topic]
	if !ok {
		return nil
	}

	for _, sub := range subs {
		dataCopy := make([]byte, len(data))
		copy(dataCopy, data)

		sub.msgs <- mockMessage{
			data:   dataCopy,
			source: mb.ownAddr,
		}
	}

	return nil
}
