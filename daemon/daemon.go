package daemon

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/daemon/proto"
	"github.com/disorganizer/brig/repo"
	protobuf "github.com/gogo/protobuf/proto"
)

///////////////////////
// UTILITY FUNCTIONS //
///////////////////////

func send(conn net.Conn, msg protobuf.Message) error {
	data, err := protobuf.Marshal(msg)
	if err != nil {
		return nil
	}

	sizeBuf := make([]byte, binary.MaxVarintLen64)
	binary.PutUvarint(sizeBuf, uint64(len(data)))

	n, err := conn.Write(sizeBuf)
	if err != nil {
		return err
	}

	if n < len(sizeBuf) {
		return io.ErrShortWrite
	}

	n, err = conn.Write(data)
	if err != nil {
		return err
	}

	if n < len(data) {
		return io.ErrShortWrite
	}

	return nil
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

//////////////////////////
// FRONTEND DAEMON PART //
//////////////////////////

// https://github.com/docker/libchan

// Daemon is the top-level struct of the brig daemon.
type DaemonClient struct {
	// Use this channel to send commands to the daemon
	Send chan *proto.Command

	// Responses are sent to this channel
	Recv chan *proto.Response

	// Underlying tcp connection:
	conn net.Conn

	quit chan bool
}

func Dial(port int) (*DaemonClient, error) {
	client := &DaemonClient{
		Send: make(chan *proto.Command),
		Recv: make(chan *proto.Response),
		quit: make(chan bool, 1),
	}

	conn, err := net.Dial("tcp", "127.0.0.1:6666")
	if err != nil {
		return nil, err
	}

	client.conn = conn

	go client.handleMessages()
	return client, nil
}

func (c *DaemonClient) handleMessages() {
	for {
		select {
		case <-c.quit:
			return
		case msg := <-c.Send:
			if err := send(c.conn, msg); err != nil {
				log.Warning("CLIENT SEND ", err)
				continue
			}

			resp := &proto.Response{}
			if err := recv(c.conn, resp); err != nil {
				log.Warning("CLIENT RECV ", err)
				continue
			}

			c.Recv <- resp
		}
	}
}

// Reach tries to Dial() the daemon, if not there it Launch()'es one.
func Reach(repoPath string, host string, port int) (*DaemonClient, error) {
	// TODO: fork magic.
	return Dial(port)
}

func (c *DaemonClient) Ping() bool {
	cmd := &proto.Command{}
	cmd.CommandType = proto.MessageType_PING.Enum()

	c.Send <- cmd
	resp := <-c.Recv
	if resp != nil {
		return "PONG" == resp.GetResponse()
	}

	return false
}

func (c *DaemonClient) Exorcise() {
	cmd := &proto.Command{}
	cmd.CommandType = proto.MessageType_QUIT.Enum()
	c.Send <- cmd
}

func (c *DaemonClient) Close() {
	if c != nil {
		c.quit <- true
	}
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

func Summon(port int) (*DaemonServer, error) {
	// Listen for incoming connections.
	addr := fmt.Sprintf("localhost:%d", port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Error("Error listening:", err.Error())
		return nil, err
	}

	// Close the listener when the application closes.
	log.Info("Listening on", addr)

	daemon := &DaemonServer{
		done:     make(chan bool, 1),
		quit:     make(chan bool),
		listener: l,
	}

	// Daemon mainloop:
	go func() {
		for {
			select {
			case <-daemon.quit:
				// Break the Serve() loop
				daemon.done <- true
				return
			default:
				// Listen for an incoming connection.
				deadline := time.Now().Add(500 * time.Millisecond)
				err := l.(*net.TCPListener).SetDeadline(deadline)
				if err != nil {
					break
				}

				conn, err := l.Accept()
				if err != nil && err.(*net.OpError).Timeout() {
					continue
				}

				if err != nil {
					fmt.Println("Error accepting: ", err.Error())
					break
				}

				// Handle connections in a new goroutine.
				go daemon.handleRequest(conn)
			}
		}
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
	resp := &proto.Response{}

	switch *(cmd.CommandType) {
	case proto.MessageType_INIT:
		fmt.Println("Init")
	case proto.MessageType_ADD:
	case proto.MessageType_CAT:
	case proto.MessageType_QUIT:
		fmt.Println("PRE QUIT")
		d.quit <- true
		fmt.Println("POST QUIT")
		resp.ResponseType = cmd.CommandType
		resp.Response = protobuf.String("BYE")
		send(conn, resp)
	case proto.MessageType_PING:
		resp.ResponseType = cmd.CommandType
		resp.Response = protobuf.String("PONG")
		send(conn, resp)
	default:
		fmt.Println("Unknown message type.")
	}
}
