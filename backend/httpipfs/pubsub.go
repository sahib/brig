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

func (nd *Node) Subscribe(ctx context.Context, topic string) (eventsBackend.Subscription, error) {
	if !nd.allowNetOps {
		return nil, ErrOffline
	}

	sub, err := nd.sh.PubSubSubscribe(topic)
	if err != nil {
		return nil, err
	}

	return &subWrapper{sub: sub}, nil
}

func (nd *Node) PublishEvent(topic string, data []byte) error {
	if !nd.allowNetOps {
		return ErrOffline
	}

	return nd.sh.PubSubPublish(topic, string(data))
}
