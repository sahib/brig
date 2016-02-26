package transfer

import (
	"encoding/json"
	"io"
	"time"

	log "github.com/Sirupsen/logrus"
)

type Server struct {
	im      io.ReadWriter
	errors  chan error
	done    chan bool
	encoder *json.Encoder
	decoder *json.Decoder
}

func (sv *Server) handleCmd() bool {
	cmd := &Command{}
	if err := sv.decoder.Decode(&cmd); err != nil {
		// TODO: Is there a better way than polling?
		if err == io.EOF {
			time.Sleep(100 * time.Millisecond)
			return true
		}

		log.Warningf("Unable to decode item from json stream: %v", err)
		return true
	}

	handler, ok := handlerMap[cmd.ID]
	if !ok {
		log.Warningf("Unknown command id: %d", cmd.ID)
		return true
	}

	resp, err := handler(sv, cmd)
	if err != nil {
		log.Warningf("Handling %s failed: %v", cmd.ID.String(), err)
		return true
	}

	if resp == nil {
		// We don't want to answer.
		return true
	}

	resp.ID = cmd.ID
	if err := sv.encoder.Encode(&resp); err != nil {
		log.Warningf("Casting response to json failed: %v", err)
		return true
	}

	return cmd.ID != CmdQuit
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

func NewServer(im io.ReadWriter) *Server {
	sv := &Server{
		im:      im,
		encoder: json.NewEncoder(im),
		decoder: json.NewDecoder(im),
		done:    make(chan bool, 1),
		errors:  make(chan error),
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
