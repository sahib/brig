package transfer

import (
	"io"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/repo"
	"github.com/disorganizer/brig/transfer/proto"
)

// Server receives proto.Requests through a io.ReadWriter, processes them
// and writes a proto.Response back to the writer part.
//
// Semantically, it is similar to daemon.Server, but is supposed
// to react to outside commands instead of local ones.
type Server struct {
	// underlying layer - for correct function it's expected
	// that Read() blocks until data is available.
	im io.ReadWriter

	// Serve() waits on this channel
	errors chan error

	// gets signalled once loop() is supposed to break out
	done chan bool

	// Protocol layer for decoding requests and encoding responses.
	ptcl *ServerProtocol

	// Repository reference (required for handlers)
	// (This may be nil for testing purpose)
	rp *repo.Repository
}

func (sv *Server) handleCmd() bool {
	// NOTE: We rely on the underlying stream to not return io.EOF
	//       early (i.e. when no data is yet available)
	req, err := sv.ptcl.Decode()
	if err != nil {
		log.Warningf("Unable to decode item from protobuf stream: %v", err)
		return true
	}

	handler, ok := handlerMap[req.GetType()]
	if !ok {
		log.Warningf("Unknown command id: %d", req.GetType())
		return true
	}

	resp, err := handler(sv, req)
	if err != nil {
		log.Warningf("Handling %s failed: %v", req.GetType().String(), err)
		return true
	}

	if resp == nil {
		// We don't want to answer. That's okay for introverts.
		return true
	}

	resp.Type = req.GetType().Enum()
	if err := sv.ptcl.Encode(resp); err != nil {
		log.Warningf("Casting response to protobuf failed: %v", err)
		return true
	}

	return req.GetType() != proto.RequestType_QUIT
}

func (sv *Server) loop() {
	for {
		select {
		case _ = <-sv.done:
			sv.errors <- nil
			return
		default:
			if !sv.handleCmd() {
				sv.errors <- nil
				return
			}
		}
	}
}

func NewServer(im io.ReadWriter, rp *repo.Repository) *Server {
	sv := &Server{
		im:     im,
		ptcl:   NewServerProtocol(im),
		done:   make(chan bool, 1),
		errors: make(chan error),
		rp:     rp,
	}

	go sv.loop()
	return sv
}

func (sv *Server) Serve() error {
	err := <-sv.errors
	sv.Quit()
	return err
}

func (sv *Server) Quit() {
	sv.done <- true
}
