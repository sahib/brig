package transfer

import "github.com/disorganizer/brig/transfer/proto"

type handler func(*Server, *proto.Request) (*proto.Response, error)

var (
	handlerMap = map[proto.RequestType]handler{
		proto.RequestType_QUIT:  handleQuit,
		proto.RequestType_CLONE: handleClone,
	}
)

func handleQuit(sv *Server, req *proto.Request) (*proto.Response, error) {
	return &proto.Response{Data: []byte("BYE")}, nil
}

func handleClone(sv *Server, req *proto.Request) (*proto.Response, error) {
	return &proto.Response{Data: []byte("CLONE")}, nil
}
