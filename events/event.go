package events

import (
	"fmt"

	capnp_model "github.com/sahib/brig/events/capnp"
	capnp "zombiezen.com/go/capnproto2"
)

const (
	UnknownEvent = 1 << iota
	FsEvent
	NetEvent
)

type EventType int

func (ev EventType) String() string {
	switch ev {
	case FsEvent:
		return "fs"
	case NetEvent:
		return "net"
	default:
		return "unknown"
	}
}

func EventFromString(ev string) (EventType, error) {
	switch ev {
	case "fs":
		return FsEvent, nil
	case "net":
		return NetEvent, nil
	default:
		return UnknownEvent, fmt.Errorf("unknown EventType type: %s", ev)
	}
}

type Event struct {
	EvType EventType
}

func (msg *Event) Encode() ([]byte, error) {
	capMsg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, err
	}

	capEv, err := capnp_model.NewRootEvent(seg)
	if err != nil {
		return nil, err
	}

	if err := capEv.SetType(msg.EvType.String()); err != nil {
		return nil, err
	}

	return capMsg.Marshal()
}

func decodeMessage(data []byte) (*Event, error) {
	if data == nil {
		return nil, fmt.Errorf("received empty event")
	}

	capMsg, err := capnp.Unmarshal(data)
	if err != nil {
		return nil, err
	}

	capEv, err := capnp_model.ReadRootEvent(capMsg)
	if err != nil {
		return nil, err
	}

	capEvType, err := capEv.Type()
	if err != nil {
		return nil, err
	}

	ev, err := EventFromString(capEvType)
	if err != nil {
		return nil, err
	}

	return &Event{EvType: ev}, nil
}
