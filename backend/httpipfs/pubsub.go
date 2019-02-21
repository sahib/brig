package httpipfs

import (
	"context"

	shell "github.com/ipfs/go-ipfs-api"
	eventsBackend "github.com/sahib/brig/events/backend"
)

type subWrapper struct {
	sub *shell.PubSubSubscription
}

type msgWrapper struct {
	msg shell.PubSubRecord
}

func (msg *msgWrapper) Data() []byte {
	return msg.msg.Data()
}

func (msg *msgWrapper) Source() string {
	return msg.msg.From().Pretty()
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
	sub, err := nd.sh.PubSubSubscribe(topic)
	if err != nil {
		return nil, err
	}

	return &subWrapper{sub: sub}, nil
}

func (nd *Node) PublishEvent(topic string, data []byte) error {
	return nd.sh.PubSubPublish(topic, string(data))
}
