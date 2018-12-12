package events

import (
	"fmt"

	capnp_model "github.com/sahib/brig/events/capnp"
	capnp "zombiezen.com/go/capnproto2"
)

const (
	// UnknownEvent should not happen in practice.
	UnknownEvent = EventType(1 << iota)
	// FsEvent tells other remotes that our filesystem changed.
	FsEvent
	// NetEvent indicates to other peers that our network status changed.
	NetEvent
)

// EventType is the type of a
type EventType int

// String returns a human readable representation of the event type
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

// EventFromString tries to parse `ev` as event type.
// If it fails, an error will be returned.
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

// Event is a event that can be published or received by the event subsystem.
type Event struct {
	Type   EventType
	Source string
}

func (msg *Event) encode() ([]byte, error) {
	capMsg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, err
	}

	capEv, err := capnp_model.NewRootEvent(seg)
	if err != nil {
		return nil, err
	}

	if err := capEv.SetType(msg.Type.String()); err != nil {
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

	return &Event{Type: ev}, nil
}

func dedupeEvents(evs []Event) []Event {
	seen := make(map[EventType]map[string]bool)
	dedupEvs := []Event{}

	for _, ev := range evs {
		seenSources, ok := seen[ev.Type]
		if ok {
			if seenSources[ev.Source] {
				continue
			}
		} else {
			seenSources = make(map[string]bool)
			seen[ev.Type] = seenSources
		}

		dedupEvs = append(dedupEvs, ev)
		seen[ev.Type][ev.Source] = true
	}

	return dedupEvs
}
