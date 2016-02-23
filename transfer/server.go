package transfer

import (
	"encoding/json"
	"fmt"
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

func (sv *Server) handleCmd() {
	cmd := &Command{}
	if err := sv.decoder.Decode(&cmd); err != nil {
		if err == io.EOF {
			time.Sleep(50 * time.Millisecond)
			return
		}

		log.Warningf("Unable to decode item from json stream: %v", err)
		return
	}

	handler, ok := handlerMap[cmd.ID]
	if !ok {
		log.Warningf("Unknown command id: %d", cmd.ID)
		return
	}

	resp, err := handler(sv, cmd)
	if err != nil {
		log.Warningf("Handling %s failed: %v", cmd.ID.String(), err)
		return
	}

	if resp == nil {
		// We don't want to answer.
		return
	}

	resp.ID = cmd.ID
	fmt.Println("Sending", resp)
	if err := sv.encoder.Encode(&resp); err != nil {
		log.Warningf("Casting response to json failed: %v", err)
		return
	}
}

func (sv *Server) loop() {
	for {
		select {
		case _ = <-sv.done:
			fmt.Println("Done. bye")
			return
		default:
			sv.handleCmd()
		}
	}
}

func NewServer(im io.ReadWriter) *Server {
	sv := &Server{
		im:      im,
		encoder: json.NewEncoder(im),
		decoder: json.NewDecoder(im),
	}

	go sv.loop()
	return sv
}

func (sv *Server) Serve() error {
	err := <-sv.errors
	sv.done <- true
	return err
}

func (sv *Server) Quit() {

}
