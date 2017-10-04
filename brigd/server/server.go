package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/brigd/capnp"
	"zombiezen.com/go/capnproto2/rpc"
)

const (
	MaxConnections = 10
)

//////////////////////////////

type Server struct {
	lst  net.Listener
	ctx  context.Context
	base *base
}

func (sv *Server) handle(ctx context.Context, conn net.Conn) {
	transport := rpc.StreamTransport(conn)
	srv := capnp.API_ServerToClient(newApiHandler(sv.base))
	rpcConn := rpc.NewConn(transport, rpc.MainInterface(srv.Client))

	if err := rpcConn.Wait(); err != nil {
		log.Warnf("Serving rpc failed: %v", err)
	}

	if err := rpcConn.Close(); err != nil {
		log.Warnf("Failed to close rpc conn: %v", err)
	}
}

func (sv *Server) Accept(rateCh chan struct{}) error {
	deadline := time.Now().Add(500 * time.Millisecond)
	err := sv.lst.(*net.TCPListener).SetDeadline(deadline)

	if err != nil {
		rateCh <- struct{}{}
		return err
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
		sv.handle(handleCtx, conn)
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
	signal.Notify(signalCh, os.Interrupt, os.Kill)

	// Reserve a pool of connections:
	rateCh := make(chan struct{}, MaxConnections)
	for i := 0; i < cap(rateCh); i++ {
		rateCh <- struct{}{}
	}

	for {
		select {
		case sig := <-signalCh:
			log.Warnf("Received %s signal", sig)
			return nil
		case <-rateCh:
			// If this signal can receive something, we have a free connection.
			if err := sv.Accept(rateCh); err != nil {
				log.Errorf("Failed to accept connection: %s", err)
			}
		case <-sv.base.QuitCh:
			log.Infof("Will not accept new connections now")
			return nil
		default:
			// No free connection available.
			time.Sleep(250 * time.Millisecond)
		}
	}

	return nil
}

func BootServer(basePath string) (*Server, error) {
	ctx := context.Background()

	// TODO: Read and instantiate correct backend from
	//       marker in the repository.
	backend := NewDummyBackend()
	base, err := newBase(basePath, backend)
	if err != nil {
		return nil, err
	}

	port := base.Repo.Config.GetInt("daemon.port")
	addr := fmt.Sprintf("localhost:%d", port)

	log.Infof("Listening on %s", addr)
	lst, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &Server{
		ctx:  ctx,
		lst:  lst,
		base: base,
	}, nil
}
