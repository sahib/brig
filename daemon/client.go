package daemon

import (
	"fmt"
	"io"
	"net"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/VividCortex/godaemon"
	"github.com/disorganizer/brig/daemon/proto"
	"github.com/disorganizer/brig/util/tunnel"
	protobuf "github.com/gogo/protobuf/proto"
	"github.com/jbenet/go-multihash"
)

// Client is the client API to brigd.
type Client struct {
	// Use this channel to send commands to the daemon
	Send chan *proto.Command

	// Responses are sent to this channel
	Recv chan *proto.Response

	// Underlying tcp connection:
	conn net.Conn

	// Be able to tell handleMessages to stop
	quit chan bool
}

// Dial connects to a running daemon instance.
func Dial(port int) (*Client, error) {
	client := &Client{
		Send: make(chan *proto.Command),
		Recv: make(chan *proto.Response),
		quit: make(chan bool, 1),
	}

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	client.conn = conn
	tnl, err := tunnel.NewEllipticTunnel(conn)
	if err != nil {
		log.Error("Tunneling failed: ", err)
		return nil, err
	}

	go client.handleMessages(tnl)

	client.Ping()
	return client, nil
}

// handleMessages takes all messages from the Send channel
// and actually sends them over the network. It then waits
// for the response and puts it in the Recv channel.
func (c *Client) handleMessages(tnl io.ReadWriter) {
	for {
		select {
		case <-c.quit:
			return
		case msg := <-c.Send:
			if err := send(tnl, msg); err != nil {
				log.Warning("client-send: ", err)
				c.Recv <- nil
				continue
			}

			resp := &proto.Response{}
			if err := recv(tnl, resp); err != nil {
				log.Warning("client-recv: ", err)
				c.Recv <- nil
				continue
			}

			c.Recv <- resp
		}
	}
}

// Reach tries to Dial() the daemon, if not there it Launch()'es one.
func Reach(pwd, repoPath string, port int) (*Client, error) {
	// Try to Dial directly first:
	if client, err := Dial(port); err == nil {
		return client, nil
	}

	// Probably not running, find out our own binary:
	exePath, err := godaemon.GetExecutablePath()
	if err != nil {
		return nil, err
	}

	// Start a new daemon process:
	log.Info("Starting daemon: ", exePath)
	proc, err := os.StartProcess(
		exePath,
		[]string{"brig", "daemon", "-x", pwd},
		&os.ProcAttr{},
	)

	if err != nil {
		return nil, err
	}

	// Make sure it it's still referenced:
	go func() {
		log.Info("Daemon has PID: ", proc.Pid)
		if _, err := proc.Wait(); err != nil {
			log.Warning("Bad exit state: ", err)
		}
	}()

	// Wait at max 15 seconds for the daemon to start up:
	// (this means, wait till it's network interface is started)
	for i := 0; i < 15; i++ {
		client, err := Dial(port)
		fmt.Println("Try dial", client)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		return client, nil
	}

	return nil, fmt.Errorf("Daemon could not be started or took to long.")
}

// Ping returns true if the daemon is running and responds correctly.
func (c *Client) Ping() bool {
	cmd := &proto.Command{}
	cmd.CommandType = proto.MessageType_PING.Enum()

	c.Send <- cmd
	resp := <-c.Recv
	if resp != nil {
		return "PONG" == resp.GetResponse()
	}

	return false
}

// Exorcise sends a QUIT message to the daemon.
func (c *Client) Exorcise() {
	cmd := &proto.Command{}
	cmd.CommandType = proto.MessageType_QUIT.Enum()
	c.Send <- cmd
	<-c.Recv
}

// Close shuts down the daemon client
func (c *Client) Close() {
	if c != nil {
		c.quit <- true
		c.conn.Close()
	}
}

// LocalAddr returns a net.Addr with the client end of the Connection
func (c *Client) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

// RemoteAddr returns a net.Addr with the server end of the Connection
func (c *Client) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *Client) Add(absPath string) (multihash.Multihash, error) {
	cmd := &proto.Command{}
	cmd.CommandType = proto.MessageType_ADD.Enum()
	cmd.AddCommand = &proto.Command_AddCmd{
		FilePath: protobuf.String(absPath),
	}

	c.Send <- cmd
	resp := <-c.Recv
	if resp != nil && !resp.GetSuccess() {
		return nil, fmt.Errorf("client: %v", resp.GetError())
	}

	return multihash.FromB58String(resp.GetResponse())
}
