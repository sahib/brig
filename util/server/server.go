package server

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
)

const (
	MaxConnections = 10
)

type Handler interface {
	Handle(ctx context.Context, conn net.Conn)
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

// for calling Accept(). This is used to check periodically for a quit signal.
// DeadListener is a listener that allows to set a deadline
type DeadlineListener interface {
	net.Listener

	SetDeadline(deadline time.Time) error
}

func (sv *Server) accept(rateCh chan struct{}) error {
	deadline := time.Now().Add(500 * time.Millisecond)

	deadLst, ok := sv.lst.(DeadlineListener)
	if ok {
		if err := deadLst.SetDeadline(deadline); err != nil {
			rateCh <- struct{}{}
			return err
		}
	}

	conn, err := sv.lst.Accept()
	if err != nil {
		rateCh <- struct{}{}
		if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
			return nil
		}

		// Something else happened.
		return err
	}

	handleCtx, cancel := context.WithTimeout(sv.ctx, 30*time.Second)
	go func() {
		sv.handler.Handle(handleCtx, conn)
		cancel()
		rateCh <- struct{}{}
	}()

	return nil
}

func (sv *Server) Close() error {
	return sv.lst.Close()
}

func (sv *Server) Serve() error {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	// Reserve a pool of connections:
	rateCh := make(chan struct{}, MaxConnections)
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
				log.Errorf("Failed to accept connection: %s", err)
			}
		case <-sv.quitCh:
			log.Infof("Will not accept new connections now")
			doServe = false
		default:
			// No free connection available.
			time.Sleep(250 * time.Millisecond)
		}
	}

	return sv.handler.Quit()
}

func (sv *Server) Quit() {
	sv.quitCh <- true
}

func NewServer(lst net.Listener, handler Handler, ctx context.Context) (*Server, error) {
	return &Server{
		ctx:     ctx,
		lst:     lst,
		handler: handler,
		quitCh:  make(chan bool),
	}, nil
}
