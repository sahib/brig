package transfer

const (
	CmdInvalid = iota
	CmdQuit
	CmdClone
)

type CommandID int

func (i CommandID) String() string {
	switch i {
	case CmdQuit:
		return "quit"
	case CmdClone:
		return "clone"
	}

	return ""
}

// interface?
type Command struct {
	ID CommandID
}

type Response struct {
	ID   CommandID
	data []byte
}
