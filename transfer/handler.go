package transfer

type handler func(*Server, *Command) (*Response, error)

var (
	handlerMap = map[CommandID]handler{
		CmdQuit:  handleQuit,
		CmdClone: handleClone,
	}
)

func handleQuit(sv *Server, cmd *Command) (*Response, error) {
	return &Response{data: []byte("BYE")}, nil
}

func handleClone(sv *Server, cmd *Command) (*Response, error) {
	return &Response{data: []byte("CLONE")}, nil
}
