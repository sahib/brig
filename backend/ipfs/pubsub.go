package ipfs

import (
	"context"

	iface "github.com/ipfs/go-ipfs/core/coreapi/interface"
	options "github.com/ipfs/go-ipfs/core/coreapi/interface/options"
	eventsBackend "github.com/sahib/brig/events/backend"
)

// XXX: Works.

type subscription struct {
	sub iface.PubSubSubscription
}

type message struct {
	msg iface.PubSubMessage
}

func (msg *message) Data() []byte {
	return msg.msg.Data()
}

func (msg *message) Source() string {
	return msg.msg.From().Pretty()
}

func (s *subscription) Next(ctx context.Context) (eventsBackend.Message, error) {
	msg, err := s.sub.Next(ctx)
	if err != nil {
		return nil, err
	}

	return &message{msg: msg}, nil
}

func (s *subscription) Close() error {
	return s.sub.Close()
}

// Subscribe is the implementation of the events.Backend Subscribe interface.
func (nd *Node) Subscribe(ctx context.Context, topic string) (eventsBackend.Subscription, error) {
	sub, err := nd.api.PubSub().Subscribe(ctx, topic, options.PubSub.Discover(false))
	if err != nil {
		return nil, err
	}

	return &subscription{sub: sub}, nil
}

// PublishEvent is the implementation of the events.Backend Publish interface.
func (nd *Node) PublishEvent(topic string, data []byte) error {
	return nd.api.PubSub().Publish(nd.ctx, topic, data)
}
