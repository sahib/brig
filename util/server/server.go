package server

import (
	"context"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	maxConnections = 10
)

// Handler is a interface that needs to be implemented in
// order to react on the requests that Server is getting.
type Handler interface {
	// Handle is called whenever a new connection is accepted.
	Handle(ctx context.Context, conn net.Conn)

	// Quit is being called when the server received a quit signal.
	Quit() error
}

// Server is a generic server implementation that
// listens on a certain port and starts a new go routine
// for each new accepted connection.
// Whatever the goroutine does is defined by the user-defined handler.
type Server struct {
	lst     net.Listener
	ctx     context.Context
	handler Handler
	quitCh  chan bool
}

// DeadlineListener is a listener that allows to set a deadline
type DeadlineListener interface {
	net.Listener

	SetDeadline(deadline time.Time) error
}

type timeoutErr interface {
	Timeout() bool
}

func (sv *Server) accept(rateCh chan struct{}) error {
	deadLst, ok := sv.lst.(DeadlineListener)
	if ok {
		deadline := time.Now().Add(500 * time.Millisecond)
		if err := deadLst.SetDeadline(deadline); err != nil {
			rateCh <- struct{}{}
			return err
		}
	}

	conn, err := sv.lst.Accept()
	if err != nil {
		rateCh <- struct{}{}
		if toutErr, ok := err.(timeoutErr); ok && toutErr.Timeout() {
			return nil
		}

		// Something else happened.
		return err
	}

	// This might happen with broken listeners.
	if conn == nil {
		return nil
	}

	handleCtx, cancel := context.WithTimeout(sv.ctx, 30*time.Second)
	go func() {
		sv.handler.Handle(handleCtx, conn)
		cancel()
		rateCh <- struct{}{}
	}()

	return nil
}

// Close cleans up internal resources
func (sv *Server) Close() error {
	return sv.lst.Close()
}

// Serve blocks to serve requests to the client.
// It can be stopped by calling Quit.
func (sv *Server) Serve() error {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	// Reserve a pool of connections:
	rateCh := make(chan struct{}, maxConnections)
	for i := 0; i < cap(rateCh); i++ {
		rateCh <- struct{}{}
	}

	doServe := true

	for doServe {
		select {
		case sig := <-signalCh:
			log.Warnf("Received %s signal, quitting.", sig)
			doServe = false
		case <-rateCh:
			// If this signal can receive something, we have a free connection.
			if err := sv.accept(rateCh); err != nil {
				if strings.HasSuffix(err.Error(), "use of closed network connection") {
					doServe = false
					break
				}

				log.Errorf("Failed to accept connection: %v", err)
				// prevent spamming log messages in case of repeating errors.
				time.Sleep(100 * time.Millisecond)
			}
		case <-sv.quitCh:
			log.Infof("Will not accept new connections now")
			doServe = false
		default:
			// No free connection available.
			time.Sleep(100 * time.Millisecond)
		}
	}

	return sv.handler.Quit()
}

// Quit stops the blocking of Serve()
func (sv *Server) Quit() {
	sv.quitCh <- true
}

// NewServer creates a new server from the listener in `lst` and will call `handler`
// when receiving requests. It uses `ctx` for handling timeouts.
func NewServer(ctx context.Context, lst net.Listener, handler Handler) (*Server, error) {
	return &Server{
		ctx:     ctx,
		lst:     lst,
		handler: handler,
		quitCh:  make(chan bool, 10),
	}, nil
}
