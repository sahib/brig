package backend

import (
	"context"
	"io"
)

// Message is returned by Subscribe.
// It encapsulates a single event message coming
// from another remote.
type Message interface {
	// Data is the data that is sent alongside the message.
	Data() []byte
	// Source is the addr of the remote.
	Source() string
}

// Subscription is an iterator like interface for accessing and listening
// for messages from other remotes.
type Subscription interface {
	io.Closer

	// Next blocks until receiving a new message or fails with
	// context.Canceled if the cancel func was called.
	Next(ctx context.Context) (Message, error)
}

// Backend is the backend that backends of the event subsystem must fulfill.
type Backend interface {
	// Subscribe returns a new Subscription iterator for `topic`.
	Subscribe(ctx context.Context, topic string) (Subscription, error)
	// PublishEvent sends `data` to all listening remotes on `topic`.
	PublishEvent(topic string, data []byte) error
}
