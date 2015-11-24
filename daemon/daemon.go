package daemon

import (
	"encoding/binary"
	"fmt"
	"net"

	"github.com/disorganizer/brig/daemon/proto"
	"github.com/disorganizer/brig/repo"
	protobuf "github.com/gogo/protobuf/proto"
)

//////////////////////////
// FRONTEND DAEMON PART //
//////////////////////////

// https://github.com/docker/libchan

// Daemon is the top-level struct of the brig daemon.
type DaemonClient struct {
	// Use this channel to send commands to the daemon
	// Send chan<- proto.Command

	// Responses and errors are sent to this channel
	// Recv <-chan proto.Response

	conn net.Conn
}

func Dial(port int) (*DaemonClient, error) {
	client := &DaemonClient{}

	conn, err := net.Dial("tcp", "127.0.0.1:6666")
	if err != nil {
		return nil, err
	}

	client.conn = conn
	return client, nil
}

// Reach tries to Dial() the daemon, if not there it Launch()'es one.
func Reach(repoPath string, host string, port int) (*DaemonClient, error) {
	return nil, nil
}

// TODO:
func send(conn net.Conn, msg protobuf.Message) (int, error) {
	data, err := protobuf.Marshal(msg)
	if err != nil {
		return 0, nil
	}

	sizeBuf := make([]byte, binary.MaxVarintLen64)
	binary.PutUvarint(sizeBuf, uint64(len(data)))

	n, err := conn.Write(sizeBuf)
	if err != nil {
		return n, err
	}

	wn, err := conn.Write(data)
	if err != nil {
		return n + wn, err
	}

	return n + wn, nil

}

func recv(conn net.Conn, msg protobuf.Message) error {
	sizeBuf := make([]byte, binary.MaxVarintLen64)
	n, err := conn.Read(sizeBuf)
	if err != nil {
		return err
	}

	size, _ := binary.Uvarint(sizeBuf[:n])
	if size > 1*1024*1024 {
		return fmt.Errorf("Message too large: %d", size)
	}

	fmt.Println(size)

	buf := make([]byte, size)
	n, err = conn.Read(buf)
	if err != nil {
		return err
	}

	err = protobuf.Unmarshal(buf, msg)
	if err != nil {
		return err
	}

	return nil
}

func (c *DaemonClient) Send(cmd *proto.Command) (int, error) {
	return send(c.conn, cmd)
}

func (c *DaemonClient) Recv() (*proto.Response, error) {
	resp := &proto.Response{}
	if err := recv(c.conn, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *DaemonClient) Ping() bool {
	cmd := &proto.Command{}
	cmd.CommandType = proto.MessageType_PING.Enum()
	if _, err := c.Send(cmd); err != nil {
		return false
	}

	resp := &proto.Response{}
	if err := recv(c.conn, resp); err != nil {
		return false
	}

	fmt.Println("PONG", resp)

	return true
}

func (c *DaemonClient) Exorcise() {
	cmd := &proto.Command{}
	cmd.CommandType = proto.MessageType_QUIT.Enum()
	c.Send(cmd)
}

func (c *DaemonClient) Close() {
}

/////////////////////////
// BACKEND DAEMON PART //
/////////////////////////

type DaemonServer struct {
	// The repo we're working on
	Repo *repo.FsRepository

	// socket....
	done chan bool
	quit chan bool

	listener net.Listener
}

func (s *DaemonServer) daemonMain(repoPath string) {
	// Actual daemon init here.
}

func Summon() (*DaemonServer, error) {
	// Listen for incoming connections.
	addr := "localhost" + ":" + "6666"
	l, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		return nil, err
	}

	// Close the listener when the application closes.
	fmt.Println("Listening on ", addr)

	daemon := &DaemonServer{
		done:     make(chan bool),
		quit:     make(chan bool),
		listener: l,
	}

	// Daemon mainloop:
	go func() {
		for {
			select {
			case <-daemon.quit:
				break
			default:
				// Listen for an incoming connection.
				conn, err := l.Accept()
				if err != nil {
					fmt.Println("Error accepting: ", err.Error())
					break
				}

				// Handle connections in a new goroutine.
				go daemon.handleRequest(conn)
			}
		}

		daemon.done <- true
	}()

	return daemon, nil
}

func (d *DaemonServer) Serve() {
	<-d.done
	d.listener.Close()
}

// Handles incoming requests.
func (d *DaemonServer) handleRequest(conn net.Conn) {
	defer conn.Close()

	msg := &proto.Command{}
	if err := recv(conn, msg); err != nil {
		fmt.Println("daemon: ", err)
		return
	}

	d.handleCommand(msg, conn)
}

func (d *DaemonServer) handleCommand(cmd *proto.Command, conn net.Conn) {
	fmt.Println("MESSAGE RECEIVED:", cmd)

	switch *(cmd.CommandType) {
	case proto.MessageType_INIT:
		fmt.Println("Init")
	case proto.MessageType_ADD:
	case proto.MessageType_CAT:
	case proto.MessageType_QUIT:
		fmt.Println("PRE QUIT")
		d.quit <- true
		fmt.Println("POST QUIT")
		resp := &proto.Response{}
		resp.ResponseType = cmd.CommandType
		resp.Response = protobuf.String("BYE")
		send(conn, resp)
	case proto.MessageType_PING:
		resp := &proto.Response{}
		resp.ResponseType = cmd.CommandType
		resp.Response = protobuf.String("PONG")
		send(conn, resp)
	default:
		fmt.Println("Unknown message type.")
	}
}
