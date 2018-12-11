package backend

import (
	"context"
	"io"
)

type Message interface {
	Data() []byte
	Source() string
}

type Subscription interface {
	io.Closer

	Next(ctx context.Context) (Message, error)
}

type Backend interface {
	Subscribe(ctx context.Context, topic string) (Subscription, error)
	PublishEvent(topic string, data []byte) error
}
