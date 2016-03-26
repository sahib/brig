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
	fmt.Printf("ID %v is registering with cred %v (%T)\n", id, cred, cred)
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

	log.Infof("Starting MQTT broker on port %d...", srv.port)
	go func() {
		err = srv.srv.ListenAndServe(fmt.Sprintf("tcp://:%d", srv.port))
		if err != nil {
			log.Warningf("Broker running on port %d died: %v", srv.port, err)
		} else {
			log.Infof("Broker running on port %d exited", srv.port)
		}

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
