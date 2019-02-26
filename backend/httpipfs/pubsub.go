package httpipfs

import (
	"context"

	eventsBackend "github.com/sahib/brig/events/backend"
	shell "github.com/sahib/go-ipfs-api"
)

type subWrapper struct {
	sub *shell.PubSubSubscription
}

type msgWrapper struct {
	msg *shell.Message
}

func (msg *msgWrapper) Data() []byte {
	return msg.msg.Data
}

func (msg *msgWrapper) Source() string {
	return msg.msg.From
}

func (s *subWrapper) Next(ctx context.Context) (eventsBackend.Message, error) {
	msg, err := s.sub.Next()
	if err != nil {
		return nil, err
	}

	return &msgWrapper{msg: msg}, nil
}

func (s *subWrapper) Close() error {
	return s.sub.Cancel()
}

// Subscribe will create a subscription for `topic`.
// You can use the subscription to wait for the next incoming message.
// This will only work if the daemon supports/has enabled pub sub.
func (nd *Node) Subscribe(ctx context.Context, topic string) (eventsBackend.Subscription, error) {
	if !nd.isOnline() {
		return nil, ErrOffline
	}

	sub, err := nd.sh.PubSubSubscribe(topic)
	if err != nil {
		return nil, err
	}

	return &subWrapper{sub: sub}, nil
}

// PublishEvent will publish `data` on `topic`.
func (nd *Node) PublishEvent(topic string, data []byte) error {
	if !nd.isOnline() {
		return ErrOffline
	}

	return nd.sh.PubSubPublish(topic, string(data))
}
