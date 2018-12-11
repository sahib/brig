package events

import (
	"context"
	"io"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/sahib/brig/backend"
	"github.com/sahib/config"
)

// TODO: Make sure the owner's filesystem does not get changed directly.

const (
	brigEventTopic = "brig/events"
)

type Listener struct {
	bk        backend.Backend
	cfg       *config.Config
	callbacks map[EventType]func(*Event)
}

func NewListener(cfg *config.Config, bk backend.Backend) *Listener {
	return &Listener{
		bk:        bk,
		callbacks: make(map[EventType]func(*Event)),
	}
}

func (lst *Listener) RegisterEventHandler(ev EventType, hdl func(ev *Event)) {
	lst.callbacks[ev] = hdl
}

func (lst *Listener) NotifyEvent(ev Event) error {
	data, err := ev.Encode()
	if err != nil {
		return err
	}

	return lst.bk.PublishEvent(brigEventTopic, data)
}

func (lst *Listener) Listen(ctx context.Context) error {
	sub, err := lst.bk.Subscribe(ctx, brigEventTopic)
	if err != nil {
		return err
	}

	defer sub.Close()

	initial := true

	for {
		if !lst.cfg.Bool("events.enabled") {
			continue
		}

		if !initial {
			time.Sleep(lst.cfg.Duration("events.congestion_timeout"))
			initial = false
		}

		msg, err := sub.Next(ctx)
		if msg == nil {
			continue
		}

		if err == io.EOF || err == context.Canceled {
			return nil
		} else if err != nil {
			return err
		}

		ev, err := decodeMessage(msg.Data())
		if err != nil {
			log.Warningf("received bad message: %v", err)
			continue
		}

		if cb, ok := lst.callbacks[ev.EvType]; ok {
			go cb(ev)
		}
	}

	return nil
}
