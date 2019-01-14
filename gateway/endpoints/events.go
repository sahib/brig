package endpoints

import (
	"context"
	"net/http"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
	"github.com/sahib/brig/events"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type EventsHandler struct {
	mu  sync.Mutex
	id  int
	chs map[int]chan string
}

func NewEventsHandler(ev *events.Listener) *EventsHandler {
	hdl := &EventsHandler{
		chs: make(map[int]chan string),
	}

	if ev != nil {
		ev.RegisterEventHandler(events.FsEvent, func(ev *events.Event) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			hdl.Notify("fs", ctx)
		})
	}

	return hdl
}

func (eh *EventsHandler) Notify(msg string, ctx context.Context) error {
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

func (eh *EventsHandler) Shutdown() {
	eh.mu.Lock()
	defer eh.mu.Unlock()

	for _, ch := range eh.chs {
		close(ch)
	}
}

func (eh *EventsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Warningf("failed to upgrade to websocket: %v", err)
		return
	}

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
