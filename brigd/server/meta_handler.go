package server

import (
	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/brigd/capnp"
	"zombiezen.com/go/capnproto2/server"
)

type metaHandler struct {
	base *base
}

func (mh *metaHandler) Quit(call capnp.Meta_quit) error {
	server.Ack(call.Options)
	log.Info("Shutting down brigd due to QUIT command")
	mh.base.QuitCh <- struct{}{}
	return nil
}

func (mh *metaHandler) Ping(call capnp.Meta_ping) error {
	server.Ack(call.Options)
	return call.Results.SetReply("PONG")
}
