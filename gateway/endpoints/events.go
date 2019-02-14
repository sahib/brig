package endpoints

import (
	"context"
	"net/http"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
	"github.com/sahib/brig/events"
	"github.com/sahib/brig/gateway/db"
	"github.com/sahib/brig/gateway/remotesapi"
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
	changeOnce sync.Once
}

// NewEventsHandler returns a new EventsHandler
func NewEventsHandler(rapi remotesapi.RemotesAPI) *EventsHandler {
	hdl := &EventsHandler{
		chs:  make(map[int]chan string),
		rapi: rapi,
	}

	return hdl
}

// SetEventListener sets the event listener for this event handler.
// See also State.SetEventListener
func (eh *EventsHandler) SetEventListener(ev *events.Listener) {
	ev.RegisterEventHandler(events.FsEvent, func(ev *events.Event) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		eh.Notify(ctx, "fs")
	})
}

// Notify sends `msg` to all connected clients, but stops in case `ctx`
// was canceled before sending it all.
func (eh *EventsHandler) Notify(ctx context.Context, msg string) error {
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
	if !checkRights(w, r, db.RightFsView) {
		return
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
