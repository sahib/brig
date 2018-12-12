package events

import (
	"context"
	"testing"
	"time"

	"github.com/sahib/brig/defaults"
	"github.com/sahib/brig/events/mock"
	"github.com/sahib/config"
	"github.com/stretchr/testify/require"
)

func withEventListener(t *testing.T, ownAddr string, fn func(lst *Listener)) {
	cfg, err := config.Open(nil, defaults.Defaults, config.StrictnessPanic)
	require.Nil(t, err)

	cfg.SetDuration("events.send_flush_window", time.Millisecond*50)
	cfg.SetDuration("events.recv_flush_window", time.Millisecond*50)

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
	withEventListenerPair(t, "a", "b", func(lstA, lstB *Listener) {
		eventReceived := false

		lstB.RegisterEventHandler(FsEvent, func(ev *Event) {
			require.Equal(t, "a", ev.Source)
			require.Equal(t, FsEvent, ev.Type)
			eventReceived = true
		})

		require.Nil(t, lstB.SetupListeners(context.Background(), []string{"a"}))
		require.Nil(t, lstA.PublishEvent(Event{Type: FsEvent}))
		time.Sleep(200 * time.Millisecond)
		require.True(t, eventReceived)

		// Do a double close:
		require.Nil(t, lstA.Close())
		require.Nil(t, lstA.PublishEvent(Event{Type: NetEvent}))
		time.Sleep(200 * time.Millisecond)
	})
}
