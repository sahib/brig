package endpoints

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sahib/brig/events"
	"github.com/sahib/brig/gateway/db"
	"github.com/sahib/brig/gateway/remotesapi"
	log "github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// EventsHandler implements http.Handler
type EventsHandler struct {
	mu         sync.Mutex
	id         int
	chs        map[int]chan string
	rapi       remotesapi.RemotesAPI
	evListener *events.Listener
	changeOnce sync.Once

	// only true while unit tests.
	// circumvents the right check,
	// that can't be mocked away easily.
	testing bool
}

// NewEventsHandler returns a new EventsHandler
func NewEventsHandler(rapi remotesapi.RemotesAPI, ev *events.Listener) *EventsHandler {
	hdl := &EventsHandler{
		chs:  make(map[int]chan string),
		rapi: rapi,
	}

	if ev != nil {
		// Incoming events from our own node:
		ev.RegisterEventHandler(events.FsEvent, true, func(ev *events.Event) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			hdl.notify(ctx, "fs", true, false)
		})

		// Incoming events from other nodes:
		ev.RegisterEventHandler(events.FsEvent, false, func(ev *events.Event) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			hdl.notify(ctx, "fs", false, false)
		})

		hdl.evListener = ev
	}
	return hdl
}

// Notify sends `msg` to all connected clients, but stops in case `ctx`
// was canceled before sending it all.
func (eh *EventsHandler) Notify(ctx context.Context, msg string) error {
	return eh.notify(ctx, msg, true, true)
}

func (eh *EventsHandler) notify(ctx context.Context, msg string, isOwnEvent, triggerPublish bool) error {
	eh.mu.Lock()
	chs := []chan string{}
	for _, ch := range eh.chs {
		chs = append(chs, ch)
	}
	eh.mu.Unlock()

	for _, ch := range chs {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ch <- msg:
			continue
		}
	}

	// We can only trigger fs events in the gateway:
	event := events.Event{
		Type: events.FsEvent,
	}

	if !isOwnEvent && triggerPublish && eh.evListener != nil {
		return eh.evListener.PublishEvent(event)
	}

	return nil
}

// Shutdown closes all open websockets.
func (eh *EventsHandler) Shutdown() {
	eh.mu.Lock()
	defer eh.mu.Unlock()

	for _, ch := range eh.chs {
		close(ch)
	}
}

func (eh *EventsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !eh.testing {
		if !checkRights(w, r, db.RightFsView) {
			return
		}
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Warningf("failed to upgrade to websocket: %v", err)
		return
	}

	// We setup the on change handler only here,
	// since calling OnChange in init might deadlock
	// since the real implementation might call Repo()
	eh.changeOnce.Do(func() {
		eh.rapi.OnChange(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			eh.Notify(ctx, "remotes")
		})
	})

	eh.mu.Lock()
	id := eh.id
	eh.id++
	ch := make(chan string, 20)
	eh.chs[id] = ch
	eh.mu.Unlock()

	defer func() {
		eh.mu.Lock()
		delete(eh.chs, id)
		eh.mu.Unlock()
	}()

	defer conn.Close()

	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return
			}

			if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
				log.Debugf("failed to write to websocket, closing: %v", err)
				return
			}
		}
	}
}
