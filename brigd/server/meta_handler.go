package server

import (
	"fmt"

	"github.com/disorganizer/brig/brigd/capnp"
	"zombiezen.com/go/capnproto2/server"
)

type metaHandler struct {
	base
}

func (mh *metaHandler) Quit(call capnp.Meta_quit) error {
	server.Ack(call.Options)

	fmt.Println("QUIT CALLED")
	return nil
}

func (mh *metaHandler) Ping(call capnp.Meta_ping) error {
	fmt.Println("server: PING!")
	return call.Results.SetReply("PONG")
}
