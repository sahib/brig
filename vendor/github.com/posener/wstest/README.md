# wstest

[![Build Status](https://travis-ci.org/posener/wstest.svg?branch=master)](https://travis-ci.org/posener/wstest)
[![codecov](https://codecov.io/gh/posener/wstest/branch/master/graph/badge.svg)](https://codecov.io/gh/posener/wstest)
[![GoDoc](https://godoc.org/github.com/posener/wstest?status.svg)](http://godoc.org/github.com/posener/wstest)
[![Go Report Card](https://goreportcard.com/badge/github.com/posener/wstest)](https://goreportcard.com/report/github.com/posener/wstest)

A websocket client for unit-testing a websocket server

The [gorilla organization](http://www.gorillatoolkit.org/) provides full featured
[websocket implementation](https://github.com/gorilla/websocket) that the standard library lacks.

The standard library provides a `httptest.ResponseRecorder` struct that test
an `http.Handler` without `ListenAndServe`, but is helpless when the connection is being hijacked
by an http upgrader. As for testing websockets, it has the `httptest.NewServer` that actually
listens on a socket on an arbitrary port.

This package provides a NewDialer function to test just the `http.Handler` that upgrades
the connection to a websocket session. It runs the handler function in a goroutine
without listening on any port. The returned `websocket.Dialer` then can be used to dial and communicate
with the given handler.

## Get

`go get -u github.com/posener/wstest`

## Examples

See the [example test](./example_test.go).

An example how to modify a test function from using
`httptest.Server` to use `wstest.NewDialer` function.

```diff
func TestHandler(t *testing.T) {
	var err error

	h := &myHandler{}
-	s := httptest.NewServer(h)
-	defer s.Close()
-	d := websocket.Dialer{}
+	d := wstest.NewDialer(h)

-	c, resp, err := d.Dial("ws://" + s.Listener.Addr().String() + "/ws", nil)
+	c, resp, err := d.Dial("ws://" + "whatever" + "/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	if got, want := resp.StatusCode, http.StatusSwitchingProtocols; got != want {
		t.Errorf("resp.StatusCode = %q, want %q", got, want)
	}
	
	err = c.WriteJSON("test")
	if err != nil {
		t.Fatal(err)
	}
}
```
