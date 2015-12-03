package daemon

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/daemon/proto"
	"github.com/disorganizer/brig/repo"
	"github.com/disorganizer/brig/util/ipfsutil"
	"github.com/disorganizer/brig/util/tunnel"
	protobuf "github.com/gogo/protobuf/proto"
	"golang.org/x/net/context"
)

// DaemonServer is a TCP server that executed all commands
// on a single repository.
type DaemonServer struct {
	// The repo we're working on
	Repo   *repo.Repository
	Folder string

	ipfsDaemon *exec.Cmd

	signals chan os.Signal

	// TCP Listener for incoming connections:
	listener net.Listener

	ctx context.Context
}

// Summon creates a new up and running DaemonServer instance
func Summon(pwd, repoFolder string, port int) (*DaemonServer, error) {
	// Load the on-disk repository:
	log.Infof("Opening repo: %s", repoFolder)
	repository, err := repo.Open(pwd, repoFolder)
	if err != nil {
		log.Error("Could not load repository: ", err)
		return nil, err
	}

	proc, err := ipfsutil.StartDaemon(&ipfsutil.Context{
		Path: filepath.Join(repoFolder, ".brig", "ipfs"),
	})

	if err != nil {
		log.Error("Unable to start ipfs daemon: ", err)
		return nil, err
	}

	// Listen for incoming connections.
	addr := fmt.Sprintf("localhost:%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Error("Error listening:", err.Error())
		return nil, err
	}

	// Close the listener when the application closes.
	log.Info("Listening on ", addr)

	daemon := &DaemonServer{
		signals:    make(chan os.Signal, 1),
		listener:   listener,
		Repo:       repository,
		Folder:     repoFolder,
		ipfsDaemon: proc,
	}

	ctx, cancel := context.WithCancel(context.Background())
	daemon.ctx = ctx

	// Daemon mainloop:
	go func() {
		defer cancel()

		// Forward signals to the quit channel:
		signal.Notify(daemon.signals, os.Interrupt, os.Kill)

		for {
			select {
			case <-daemon.signals:
				// Break the Serve() loop
				cancel()
				return
			default:
				// Listen for an incoming connection.
				deadline := time.Now().Add(500 * time.Millisecond)
				err := listener.(*net.TCPListener).SetDeadline(deadline)
				if err != nil {
					break
				}

				conn, err := listener.Accept()
				if err != nil && err.(*net.OpError).Timeout() {
					continue
				}

				if err != nil {
					log.Errorf("Error accepting: %v", err.Error())
					break
				}

				// Handle connections in a new goroutine.
				go daemon.handleRequest(ctx, conn)
			}
		}
	}()

	return daemon, nil
}

// Serve waits until the DaemonServer received a quit event.
func (d *DaemonServer) Serve() {
	<-d.ctx.Done()
	d.listener.Close()
	if err := d.ipfsDaemon.Process.Kill(); err != nil {
		log.Errorf("Unable to kill off ipfs daemon: %v", err)
	}

	if err := d.Repo.Close(); err != nil {
		log.Errorf("Unable to close repository: %v", err)
	}
}

// Handles incoming requests:
func (d *DaemonServer) handleRequest(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	tnl, err := tunnel.NewEllipticTunnel(conn)
	if err != nil {
		log.Error("Tunnel failed", err)
		return
	}

	for {
		msg := &proto.Command{}
		if err := recv(tnl, msg); err != nil {
			log.Warning("daemon-recv: ", err)
			return
		}

		d.handleCommand(ctx, msg, tnl)
	}
}

// Handles the actual incoming commands:
func (d *DaemonServer) handleCommand(ctx context.Context, cmd *proto.Command, conn io.ReadWriter) {
	// This might be used to enforce timeouts for operations:
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	log.Info("Processing message: ", cmd)

	// Prepare a response template
	resp := &proto.Response{}
	resp.ResponseType = cmd.CommandType

	switch *(cmd.CommandType) {
	case proto.MessageType_ADD:
	case proto.MessageType_CAT:
	case proto.MessageType_QUIT:
		resp.Response = protobuf.String("BYE")
		d.signals <- os.Interrupt
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
