package mqtt

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/surgemq/surgemq/auth"
	"github.com/surgemq/surgemq/service"
)

type authenticator struct {
}

func (au *authenticator) Authenticate(id string, cred interface{}) error {
	// fmt.Printf("ID %v is registering with cred %v (%T)\n", id, cred, cred)
	return nil
}

type server struct {
	srv      *service.Server
	port     int
	authMgr  *authenticator
	authName string
}

// Running counter; incremented for each authenticator
var globalAuthCount = int32(0)

func newServer(port int) (*server, error) {
	// Apply a crude hack: We need to pass data to the
	// authenticator. SurgeMQ has no means to do that (except globals),
	// so we register a new authenticator for each server.
	// (at least in tests we need more than one server)
	authMgr := &authenticator{}
	name := fmt.Sprintf(
		"brig-auth-%d",
		atomic.AddInt32(&globalAuthCount, 1),
	)

	auth.Register(name, authMgr)
	return &server{
		srv:     nil,
		port:    port,
		authMgr: authMgr,
	}, nil
}

func (srv *server) addr() net.Addr {
	return &net.TCPAddr{
		IP:   net.IP{127, 0, 0, 1},
		Port: srv.port,
	}
}

// Sometimes it might happen that a server is disconnected
// and connected soon again afterwards. The older server
// might not have made the port available yet.
// Use a hack so that older server make younger ones wait
// a bit until the port is ready.
type portReservation struct{}

var portMap = make(map[int]chan portReservation)
var portMapLock sync.Mutex

func (srv *server) connect() (err error) {
	srv.srv = &service.Server{
		KeepAlive:        300, // seconds
		ConnectTimeout:   2,   // seconds
		Authenticator:    srv.authName,
		SessionsProvider: "mem", // keeps sessions in memory
		TopicsProvider:   "mem", // keeps topic subscriptions in memory
	}

	portMapLock.Lock()
	reserved, ok := portMap[srv.port]
	if ok {
		log.Infof("Waiting for port %d", srv.port)
		<-reserved
	}
	portMapLock.Unlock()

	log.Infof("Starting MQTT broker on port %d...", srv.port)
	go func() {
		portMapLock.Lock()
		reservation := make(chan portReservation)
		portMap[srv.port] = reservation
		portMapLock.Unlock()

		err = srv.srv.ListenAndServe(fmt.Sprintf("tcp://:%d", srv.port))
		if err != nil {
			log.Warningf("Broker running on port %d died: %v", srv.port, err)
		} else {
			log.Infof("Broker running on port %d exited", srv.port)
		}

		// Sometimes some background services might take a bit longer:
		time.Sleep(2000 * time.Millisecond)

		reservation <- portReservation{}

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
		auth.Unregister(srv.authName)
		srv.srv = nil
		return s.Close()
	}

	return nil
}
