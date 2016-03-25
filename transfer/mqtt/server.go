package mqtt

import (
	"fmt"
	"net"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/surgemq/surgemq/auth"
	"github.com/surgemq/surgemq/service"
)

type authenticator struct{}

func (au *authenticator) Authenticate(id string, cred interface{}) error {
	return nil
}

func init() {
	auth.Register("brigAuth", &authenticator{})
}

type server struct {
	srv  *service.Server
	port int
}

func newServer(port int) (*server, error) {
	return &server{
		srv:  nil,
		port: port,
	}, nil
}

func (srv *server) addr() net.Addr {
	return &net.TCPAddr{
		IP:   net.IP{127, 0, 0, 1},
		Port: srv.port,
	}
}

func (srv *server) connect() (err error) {
	srv.srv = &service.Server{
		KeepAlive:        300,   // seconds
		ConnectTimeout:   2,     // seconds
		SessionsProvider: "mem", // keeps sessions in memory
		Authenticator:    "brigAuth",
		TopicsProvider:   "mem", // keeps topic subscriptions in memory
	}

	log.Infof("Starting server...")
	go func() {
		err = srv.srv.ListenAndServe(fmt.Sprintf("tcp://:%d", srv.port))
		log.Infof("Server stopped...: %v", err)

		// TODO: Initial publish of topcis needed?
		srv.srv = nil
	}()

	// Wait a short time to return errors early.
	time.Sleep(100 * time.Millisecond)
	return
}

func (srv *server) disconnect() error {
	s := srv.srv
	if s != nil {
		srv.srv = nil
		return s.Close()
	}

	return nil
}
