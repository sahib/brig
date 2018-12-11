package events

import (
	"context"
	"io"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/sahib/brig/backend"
	"github.com/sahib/config"
)

const (
	brigEventTopic = "brig/events"
)

type Listener struct {
	mu sync.Mutex

	bk        backend.Backend
	cfg       *config.Config
	callbacks map[EventType]func(*Event)
	evSendCh  chan Event
	evRecvCh  chan Event
	ownAddr   string
}

func NewListener(cfg *config.Config, bk backend.Backend, ownAddr string) *Listener {
	lst := &Listener{
		bk:        bk,
		cfg:       cfg,
		ownAddr:   ownAddr,
		callbacks: make(map[EventType]func(*Event)),
		evSendCh:  make(chan Event, 10),
		evRecvCh:  make(chan Event, 10),
	}

	go lst.eventSendLoop()
	go lst.eventRecvLoop()
	return lst
}

func (lst *Listener) Close() error {
	lst.mu.Lock()
	defer lst.mu.Unlock()

	close(lst.evSendCh)
	close(lst.evRecvCh)
	return nil
}

func (lst *Listener) RegisterEventHandler(ev EventType, hdl func(ev *Event)) {
	lst.mu.Lock()
	defer lst.mu.Unlock()

	lst.callbacks[ev] = hdl
}

func (lst *Listener) eventSendLoop() {
	events := []Event{}
	tckr := time.NewTicker(lst.cfg.Duration("send_flush_window"))
	defer tckr.Stop()

	for {
		select {
		case <-tckr.C:
			for _, ev := range dedupeEvents(events) {
				data, err := ev.Encode()
				if err != nil {
					log.Errorf("event: failed to encode: %v", err)
					continue
				}

				if err := lst.bk.PublishEvent(brigEventTopic, data); err != nil {
					log.Errorf("event: failed to publish: %v", err)
					continue
				}
			}

			events = []Event{}
		case ev, ok := <-lst.evSendCh:
			if !ok {
				return
			}

			events = append(events, ev)
		}
	}
}

func (lst *Listener) eventRecvLoop() {
	events := []Event{}
	tckr := time.NewTicker(lst.cfg.Duration("recv_flush_window"))
	defer tckr.Stop()

	for {
		select {
		case <-tckr.C:
			for _, ev := range dedupeEvents(events) {
				lst.mu.Lock()
				if cb, ok := lst.callbacks[ev.EvType]; ok {
					go cb(&ev)
				}
				lst.mu.Unlock()
			}

			events = []Event{}
		case ev, ok := <-lst.evRecvCh:
			if !ok {
				return
			}

			events = append(events, ev)
		}
	}
}

func (lst *Listener) PublishEvent(ev Event) error {
	lst.mu.Lock()
	defer lst.mu.Unlock()

	if !lst.cfg.Bool("enabled") {
		return nil
	}

	lst.evSendCh <- ev
	return nil
}

func (lst *Listener) Listen(ctx context.Context) error {
	sub, err := lst.bk.Subscribe(ctx, brigEventTopic)
	if err != nil {
		return err
	}

	defer sub.Close()

	for {
		if !lst.cfg.Bool("enabled") {
			continue
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

		if msg.Source() == lst.ownAddr {
			continue
		}

		ev, err := decodeMessage(msg.Data())
		if err != nil {
			log.Warningf("received bad message: %v", err)
			continue
		}

		ev.Source = msg.Source()
		lst.evRecvCh <- *ev
	}

	return nil
}
