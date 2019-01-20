package wstest_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/posener/wstest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClient demonstrate the usage of wstest package
func TestClient(t *testing.T) {
	t.Parallel()
	var (
		s    = &handler{Upgraded: make(chan struct{})}
		d    = wstest.NewDialer(s)
		done = make(chan struct{})
	)

	c, resp, err := d.Dial("ws://example.org/ws", nil)
	require.Nil(t, err)

	<-s.Upgraded

	assert.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)

	for i := 0; i < 3; i++ {
		msg := fmt.Sprintf("hello, world! %d", i)

		go func() {
			err := c.WriteMessage(websocket.TextMessage, []byte(msg))
			require.Nil(t, err)
			done <- struct{}{}
		}()

		mT, m, err := s.ReadMessage()
		require.Nil(t, err)

		assert.Equal(t, msg, string(m))
		assert.Equal(t, websocket.TextMessage, mT)
		<-done

		go func() {
			err := s.WriteMessage(websocket.TextMessage, []byte(msg))
			require.Nil(t, err)
			done <- struct{}{}
		}()

		mT, m, err = c.ReadMessage()
		require.Nil(t, err)

		assert.Equal(t, msg, string(m))
		assert.Equal(t, websocket.TextMessage, mT)
		<-done
	}

	err = c.Close()
	require.Nil(t, err)

	err = s.Close()
	require.Nil(t, err)
}

// TestConcurrent tests concurrent reads and writes from a connection
func TestConcurrent(t *testing.T) {
	t.Parallel()
	var (
		s     = &handler{Upgraded: make(chan struct{})}
		d     = wstest.NewDialer(s)
		count = 20
	)

	c, _, err := d.Dial("ws://example.org/ws", nil)
	require.Nil(t, err)

	<-s.Upgraded

	for _, pair := range []struct{ src, dst *websocket.Conn }{{s.Conn, c}, {c, s.Conn}} {
		go func() {
			for i := 0; i < count; i++ {
				err := pair.src.WriteJSON(i)
				require.Nil(t, err)
			}
		}()

		received := make([]bool, count)

		for i := 0; i < count; i++ {
			var j int
			err := pair.dst.ReadJSON(&j)
			require.Nil(t, err)

			received[j] = true
		}

		var missing []int

		for i := range received {
			if !received[i] {
				missing = append(missing, i)
			}
		}
		assert.Equal(t, 0, len(missing), "%q -> %q: Did not received: %q", pair.src.LocalAddr(), pair.dst.LocalAddr(), missing)
	}

	err = c.Close()
	require.Nil(t, err)

	err = s.Close()
	require.Nil(t, err)
}

func TestBadAddress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		url  string
		code int
	}{

		{
			url:  "ws://example.org/not-ws",
			code: http.StatusNotFound,
		},
		{
			url: "http://example.org/ws",
		},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			s := &handler{Upgraded: make(chan struct{})}
			d := wstest.NewDialer(s)
			c, resp, err := d.Dial(tt.url, nil)
			assert.Nil(t, c)
			assert.NotNil(t, err)
			if tt.code != 0 {
				assert.Equal(t, tt.code, resp.StatusCode)
			}

			err = s.Close()
			require.Nil(t, err)
		})
	}
}

const deadlineExceeded = "deadline exceeded"

// TestConnectDeadline tests connection deadlines
func TestDeadlines(t *testing.T) {
	t.Parallel()
	h := &handler{Upgraded: make(chan struct{})}
	d := wstest.NewDialer(h)

	c, _, err := d.Dial("ws://example.org/ws", nil)
	require.Nil(t, err)

	<-h.Upgraded

	var i int

	for _, pair := range []struct{ src, dst *websocket.Conn }{{h.Conn, c}, {c, h.Conn}} {

		// set the deadline to now, and test for timeout
		pair.dst.SetReadDeadline(time.Now())
		err = pair.dst.ReadJSON(&i)
		assert.Contains(t, err.Error(), deadlineExceeded)

		err = pair.dst.ReadJSON(&i)
		assert.Contains(t, err.Error(), deadlineExceeded)

		go pair.src.WriteJSON(1)
		err = pair.dst.ReadJSON(&i)
		assert.Contains(t, err.Error(), deadlineExceeded)

		// even after updating the deadline, should get an error
		pair.dst.SetReadDeadline(time.Now().Add(time.Second))
		err = pair.dst.ReadJSON(&i)
		assert.Contains(t, err.Error(), deadlineExceeded)
	}
}

// TestConnectDeadline tests connection deadline
func TestConnectDeadline(t *testing.T) {
	t.Parallel()

	tests := []struct {
		path    string
		timeout time.Duration
		wantErr bool
	}{
		{
			path:    "/ws/delay",
			timeout: time.Millisecond,
			wantErr: true,
		},
		{
			path:    "/ws",
			timeout: time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s/%s", tt.path, tt.timeout), func(t *testing.T) {
			s := &handler{Upgraded: make(chan struct{})}
			d := wstest.NewDialer(s)
			d.HandshakeTimeout = tt.timeout
			_, _, err := d.Dial("ws://example.org"+tt.path, nil)
			if tt.wantErr {
				assert.NotNil(t, err)
				return
			}

			assert.Nil(t, err)
			select {
			case <-s.Upgraded:
			case <-time.After(time.Second):
				t.Fatal("connection was not upgraded after 1s")
			}
		})
	}
}

// dialer for test purposes, can't handle multiple websocket connections concurrently
type handler struct {
	*websocket.Conn
	upgrader websocket.Upgrader
	Upgraded chan struct{}
}

func (s *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	switch r.URL.Path {
	case "/ws":
		s.connect(w, r)

	case "/ws/delay":
		<-time.After(500 * time.Millisecond)
		s.connect(w, r)

	default:
		w.WriteHeader(http.StatusNotFound)
	}

}

func (s *handler) connect(w http.ResponseWriter, r *http.Request) {
	defer close(s.Upgraded)
	var err error
	s.Conn, err = s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
}

func (s *handler) Close() error {
	if s.Conn == nil {
		return nil
	}
	return s.Conn.Close()
}
