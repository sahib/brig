package events

import (
	"context"
	"testing"
	"time"

	"github.com/sahib/brig/defaults"
	"github.com/sahib/brig/events/mock"
	"github.com/sahib/config"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func withEventListener(t *testing.T, ownAddr string, fn func(lst *Listener)) {
	cfg, err := config.Open(nil, defaults.Defaults, config.StrictnessPanic)
	require.Nil(t, err)

	cfg.SetDuration("events.recv_interval", time.Millisecond*1)
	cfg.SetDuration("events.send_interval", time.Millisecond*1)

	cfg.SetFloat("events.recv_max_events_per_second", 0.1)
	cfg.SetFloat("events.send_max_events_per_second", 0.1)

	evb := mock.NewEventsBackend(ownAddr)
	lst := NewListener(cfg.Section("events"), evb, ownAddr)
	fn(lst)
	require.Nil(t, lst.Close())
}

func withEventListenerPair(t *testing.T, addrA, addrB string, fn func(lstA, lstB *Listener)) {
	withEventListener(t, addrA, func(lstA *Listener) {
		withEventListener(t, addrB, func(lstB *Listener) {
			fn(lstA, lstB)
		})
	})
}

func TestBasicRun(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	withEventListenerPair(t, "a", "b", func(lstA, lstB *Listener) {
		eventReceived := false

		lstB.RegisterEventHandler(FsEvent, false, func(ev *Event) {
			require.Equal(t, "a", ev.Source)
			require.Equal(t, FsEvent, ev.Type)
			eventReceived = true
		})

		require.Nil(t, lstB.SetupListeners(context.Background(), []string{"a"}))

		for i := 0; i < 100; i++ {
			require.Nil(t, lstA.PublishEvent(Event{Type: FsEvent}))
		}

		time.Sleep(500 * time.Millisecond)
		require.True(t, eventReceived)

		// Do a double close:
		require.Nil(t, lstA.Close())
		require.Nil(t, lstA.PublishEvent(Event{Type: NetEvent}))
		time.Sleep(200 * time.Millisecond)
	})
}
