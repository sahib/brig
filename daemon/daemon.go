package daemon

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/VividCortex/godaemon"
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
				c.Recv <- nil
				continue
			}

			resp := &proto.Response{}
			if err := recv(c.conn, resp); err != nil {
				log.Warning("CLIENT RECV ", err)
				c.Recv <- nil
				continue
			}

			c.Recv <- resp
		}
	}
}

// Reach tries to Dial() the daemon, if not there it Launch()'es one.
func Reach(repoPath string, port int) (*DaemonClient, error) {

	if daemon, err := Dial(port); err != nil {
		exePath, err := godaemon.GetExecutablePath()
		if err != nil {
			return nil, err
		}

		log.Info("Starting daemon: ", exePath)
		proc, err := os.StartProcess(
			exePath, []string{"brig", "daemon"}, &os.ProcAttr{},
		)

		if err != nil {
			return nil, err
		}

		go func() {
			log.Info("Daemon has PID: ", proc.Pid)
			if _, err := proc.Wait(); err != nil {
				log.Warning("Bad exit state: ", err)
			}
		}()

		// Wait at max 5 seconds for the daemon to start up:
		// (this means, wait till it's network interface is started)
		for i := 0; i < 5; i++ {
			time.Sleep(1 * time.Second)
			client, err := Dial(port)
			if err != nil {
				return nil, err
			}

			if client != nil {
				return client, nil
			}
		}

		return nil, fmt.Errorf("Daemon could not be started or took to long.")
	} else {
		return daemon, nil
	}

	return nil, nil
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

	done chan bool
	quit chan os.Signal

	// TCP Listener for incoming connections:
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
		quit:     make(chan os.Signal, 1),
		listener: l,
	}

	// Daemon mainloop:
	go func() {
		// Forward signals to the quit channel:
		signal.Notify(daemon.quit)

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

// Serve waits until the DaemonServer received a quit event.
func (d *DaemonServer) Serve() {
	<-d.done
	d.listener.Close()
}

// Handles incoming requests:
func (d *DaemonServer) handleRequest(conn net.Conn) {
	defer conn.Close()

	msg := &proto.Command{}
	if err := recv(conn, msg); err != nil {
		log.Warning("daemon recv: ", err)
		return
	}

	d.handleCommand(msg, conn)
}

// Handles the actual incoming commands:
func (d *DaemonServer) handleCommand(cmd *proto.Command, conn net.Conn) {
	log.Info("Processing message: ", cmd)

	// Prepare a response template
	resp := &proto.Response{}
	resp.ResponseType = cmd.CommandType

	switch *(cmd.CommandType) {
	case proto.MessageType_INIT:
		fmt.Println("Init")
	case proto.MessageType_ADD:
	case proto.MessageType_CAT:
	case proto.MessageType_QUIT:
		d.quit <- os.Interrupt
		resp.Response = protobuf.String("BYE")
	case proto.MessageType_PING:
		resp.Response = protobuf.String("PONG")
	default:
		fmt.Println("Unknown message type.")
		return
	}

	if err := send(conn, resp); err != nil {
		log.Warning("Unable to send message back to client: ", err)
	}
}
