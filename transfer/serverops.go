package transfer

import (
	"bytes"

	"github.com/disorganizer/brig/transfer/wire"
)

type handler func(*Server, *wire.Request) (*wire.Response, error)

var (
	handlerMap = map[wire.RequestType]handler{
		wire.RequestType_QUIT:  handleQuit,
		wire.RequestType_PING:  handlePing,
		wire.RequestType_FETCH: handleFetch,
	}
)

func handleQuit(sv *Server, req *wire.Request) (*wire.Response, error) {
	return &wire.Response{Data: []byte("BYE")}, nil
}

func handlePing(sv *Server, req *wire.Request) (*wire.Response, error) {
	return &wire.Response{Data: []byte("PONG")}, nil
}

func handleFetch(sv *Server, req *wire.Request) (*wire.Response, error) {
	buf := &bytes.Buffer{}
	if err := sv.rp.OwnStore.Export(buf); err != nil {
		return nil, err
	}

	return &wire.Response{Data: buf.Bytes()}, nil
}
