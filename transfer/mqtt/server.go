package mqtt

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/disorganizer/surgemq/auth"
	"github.com/disorganizer/surgemq/service"
)

type ErrAuthDenied struct {
	id string
}

func (e *ErrAuthDenied) Error() string {
	return fmt.Sprintf("Denying access to `%s`", e.id)
}

type server struct {
	srv      *service.Server
	lay      *layer
	listener net.Listener
	authName string
}

// Running counter; incremented for each authenticator
var globalAuthCount = int32(0)

func newServer(lay *layer, listener net.Listener) (*server, error) {
	// Apply a crude hack: We need to pass data to the
	// authenticator. SurgeMQ has no means to do that (except globals),
	// so we register a new authenticator for each server.
	// (at least in tests we need more than one server)
	authName := fmt.Sprintf(
		"brig-auth-%d",
		atomic.AddInt32(&globalAuthCount, 1),
	)

	server := &server{
		srv:      nil,
		lay:      lay,
		listener: listener,
		authName: authName,
	}

	auth.Register(authName, server)
	return server, nil
}

// Authenticate is called whenever a client needs to be tunnels
// on the server.
func (srv *server) Authenticate(id string, cred interface{}) error {
	credData, ok := cred.(string)
	if !ok {
		log.Debugf("Denying; cred was no string")
		return &ErrAuthDenied{id}
	}

	if err := srv.lay.authMgr.Authenticate(id, []byte(credData)); err != nil {
		log.Debugf("Permission to `%s` denied", id)
		return &ErrAuthDenied{id}
	}

	log.Debugf("`%s` granting access to `%s`", srv.lay.self.ID(), id)
	return nil
}

// Sometimes it might happen that a server is disconnected
// and connected soon again afterwards. The older server
// might not have made the port available yet.
// Use a hack so that older server make younger ones wait
// a bit until the port is ready.
type portReservation struct{}

var portMap = make(map[string]chan portReservation)
var portMapLock sync.Mutex

func (srv *server) connect() (err error) {
	srv.srv = &service.Server{
		KeepAlive:        300, // seconds
		ConnectTimeout:   2,   // seconds
		Authenticator:    srv.authName,
		SessionsProvider: "mem", // keeps sessions in memory
		TopicsProvider:   "mem", // keeps topic subscriptions in memory
	}

	addrKey := srv.listener.Addr().String()

	// portMapLock.Lock()
	// reserved, ok := portMap[addrKey]
	// if ok {
	// 	log.Infof("Waiting for broker addr %s", addrKey)
	// 	<-reserved
	// }
	// portMapLock.Unlock()

	log.Infof("Starting MQTT broker on %s...", addrKey)
	go func() {
		portMapLock.Lock()
		reservation := make(chan portReservation, 1)
		portMap[addrKey] = reservation
		portMapLock.Unlock()

		err = srv.srv.Serve(srv.listener)
		if err != nil {
			log.Warningf("Broker running on addr %s died: %v", addrKey, err)
		} else {
			log.Infof("Broker running on addr %s exited", addrKey)
		}

		// Sometimes some background services might take a bit longer:
		time.Sleep(2000 * time.Millisecond)
		reservation <- portReservation{}
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
