package wstest_test

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/posener/wstest"
)

func ExampleNewDialer() {
	var (
		// simple echo dialer
		s = &echoServer{}

		// create a dialer to the dialer
		// this send an HTTP request to the http.Handler, and wait for the connection
		// upgrade response.
		// it uses the gorilla's websocket.Dial function, over a fake net.Conn struct.
		// it runs the handler's ServeHTTP function in a goroutine, so the handler can
		// communicate with a client running on the current program flow
		d = wstest.NewDialer(s)

		resp string
	)

	c, _, err := d.Dial("ws://example.org/ws", nil)
	if err != nil {
		panic(err)
	}

	// the client is also a websocket.Conn object, so all websocket functions
	// can be used with it. here we write a JSON string to the connection.
	c.WriteJSON("hello echo server")

	// Reading from the socket is done with the websocket.Conn functions as well.
	c.ReadJSON(&resp)
	fmt.Println(resp)

	// Pass another message in the connection
	c.WriteJSON("byebye")
	c.ReadJSON(&resp)
	fmt.Println(resp)

	// Finally close the connection
	err = c.Close()
	if err != nil {
		panic(err)
	}

	<-s.Done

	// Output: hello echo server!
	// byebye!
}

type echoServer struct {
	upgrader websocket.Upgrader
	Done     chan struct{}
}

func (s *echoServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		err error
	)

	s.Done = make(chan struct{})
	defer close(s.Done)

	if r.URL.Path != "/ws" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var conn *websocket.Conn
	conn, err = s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	for {
		var msg string
		err := conn.ReadJSON(&msg)
		if err != nil {
			return
		}
		conn.WriteJSON(msg + "!")
	}
}
