// Package wstest provides a NewDialer function to test just the
// `http.Handler` that upgrades the connection to a websocket session.
// It runs the handler function in a goroutine without listening on
// any port. The returned `websocket.Dialer` then can be used to dial
// and communicate with the given handler.
package wstest

import (
	"bufio"
	"net"
	"net/http"
	"net/http/httptest"

	"github.com/gorilla/websocket"
)

// NewDialer creates a wstest recorder to an http.Handler which accepts websocket upgrades.
// This send an HTTP request to the http.Handler, and wait for the connection upgrade response.
// it runs the recorder's ServeHTTP function in a goroutine, so recorder can communicate with a
// client running on the current program flow
//
// h is an http.Handler that handles websocket connections.
// It returns a *websocket.Dial struct, which can then be used to dial to the handler.
func NewDialer(h http.Handler) *websocket.Dialer {
	client, server := net.Pipe()
	conn := &recorder{server: server}

	// run the runServer in a goroutine, so when the Dial send the request to
	// the recorder on the connection, it will be parsed as an HTTPRequest and
	// sent to the Handler function.
	go conn.runServer(h)

	// use the websocket.NewDialer.Dial with the fake net.recorder to communicate with the recorder
	// the recorder gets the client which is the client side of the connection
	return &websocket.Dialer{NetDial: func(network, addr string) (net.Conn, error) { return client, nil }}
}

// recorder it similar to httptest.ResponseRecorder, but with Hijack capabilities
type recorder struct {
	httptest.ResponseRecorder
	server net.Conn
}

// runServer reads the request sent on the connection to the recorder
// from the websocket.NewDialer.Dial function, and pass it to the recorder.
// once this is done, the communication is done on the wsConn
func (r *recorder) runServer(h http.Handler) {
	// read from the recorder connection the request sent by the recorder.Dial,
	// and use the handler to serve this request.
	req, err := http.ReadRequest(bufio.NewReader(r.server))
	if err != nil {
		return
	}
	h.ServeHTTP(r, req)
}

// Hijack the connection
func (r *recorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	// return to the recorder the recorder, which is the recorder side of the connection
	rw := bufio.NewReadWriter(bufio.NewReader(r.server), bufio.NewWriter(r.server))
	return r.server, rw, nil
}

// WriteHeader write HTTP header to the client and closes the connection
func (r *recorder) WriteHeader(code int) {
	resp := http.Response{StatusCode: code, Header: r.Header()}
	resp.Write(r.server)
}
