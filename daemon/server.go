package daemon

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/daemon/wire"
	"github.com/disorganizer/brig/fuse"
	"github.com/disorganizer/brig/repo"
	"github.com/disorganizer/brig/transfer"
	"github.com/disorganizer/brig/transfer/moose"
	"github.com/disorganizer/brig/util/protocol"
	"github.com/disorganizer/brig/util/tunnel"
	"github.com/gogo/protobuf/proto"
	"golang.org/x/net/context"
)

const (
	// MaxConnections is the upper limit of clients that may connect to a daemon
	// at the same time. Other client will wait in Accept().
	MaxConnections = 20
)

// This is just here to make the maxConnections prettier.
type allowOneConn struct{}

// Server is a TCP server that executed all commands
// on a single repository. Once the daemon is started, it
// attempts to open the repository, for which a password is needed.
type Server struct {
	// The repo we're working on
	Repo *repo.Repository

	// All mountpoints this daemon is serving:
	Mounts *fuse.MountTable

	// The metadata networkinglayer.
	MetaHost *transfer.Connector

	// signals (external and self triggered) arrive on this channel.
	signals chan os.Signal

	// Root context for this daemon
	ctx context.Context

	// TCP Listener for incoming connections:
	listener net.Listener

	// buffered channel with N places,
	// - every active connection holds one.
	maxConnections chan allowOneConn
}

// Summon creates a new up and running Server instance
func Summon(pwd, repoFolder string, port int) (*Server, error) {
	// Load the on-disk repository:
	log.Infof("Opening repo: %s", repoFolder)
	rep, err := repo.Open(pwd, repoFolder)
	if err != nil {
		log.Error("Could not load repository: ", err)
		return nil, err
	}

	// Listen for incoming connections.
	addr := fmt.Sprintf("localhost:%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Error("Error listening:", err.Error())
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	// Close the listener when the application closes.
	log.Info("Listening on ", addr)
	// TODO: Auth mgr and peer
	layer := moose.NewLayer(rep.IPFS)

	ctx, cancel := context.WithCancel(context.Background())
	daemon := &Server{
		Repo:           rep,
		Mounts:         fuse.NewMountTable(rep.OwnStore),
		MetaHost:       transfer.NewConnector(layer, rep),
		signals:        make(chan os.Signal, 1),
		listener:       listener,
		maxConnections: make(chan allowOneConn, MaxConnections),
		ctx:            ctx,
	}

	go daemon.loop(cancel)

	if err := daemon.Connect(); err != nil {
		return nil, err
	}

	return daemon, nil
}

// Serve waits until the Server received a quit reason.
func (d *Server) Serve() {
	<-d.ctx.Done()

	if err := d.listener.Close(); err != nil {
		log.Warningf("daemon-close: cannot close listener: %v", err)
	}

	if err := d.Disconnect(); err != nil {
		log.Warningf("Could not shut down online services: %v", err)
	}

	if err := d.Mounts.Close(); err != nil {
		log.Errorf("daemon-close: error while closing mounts: %v", err)
	}

	if err := d.Repo.Close(); err != nil {
		log.Errorf("daemon-close: unable to close repository: %v", err)
	}
}

// Handle incoming connections:
func (d *Server) loop(cancel context.CancelFunc) {
	// Forward signals to the signals channel:
	signal.Notify(d.signals, os.Interrupt, os.Kill)

	// Reserve at least cap(d.maxConnections)
	for i := 0; i < cap(d.maxConnections); i++ {
		d.maxConnections <- allowOneConn{}
	}

	for {
		select {
		case <-d.signals:
			// Break the Serve() loop
			cancel()
			return
		case <-d.maxConnections:
			// Listen for an incoming connection.
			deadline := time.Now().Add(500 * time.Millisecond)
			err := d.listener.(*net.TCPListener).SetDeadline(deadline)
			if err != nil {
				log.Errorf("BUG: SetDeadline failed: %v", err)
				return
			}

			conn, err := d.listener.Accept()
			if err != nil && err.(*net.OpError).Timeout() {
				d.maxConnections <- allowOneConn{}
				continue
			}

			if err != nil {
				log.Errorf("Error in Accept(): %v", err)
				return
			}

			// Handle connections in a new goroutine.
			go d.handleConnection(d.ctx, conn)
		default:
			log.Infof("Max number of connections hit: %d", cap(d.maxConnections))
			time.Sleep(500 * time.Millisecond)
		}
	}
}

// Handles incoming requests:
func (d *Server) handleConnection(ctx context.Context, conn net.Conn) {
	// Make sure this connection count gets released:
	defer func() {
		if err := conn.Close(); err != nil {
			log.Debugf("daemon-loop: connection drop failed: %v", err)
		}

		d.maxConnections <- allowOneConn{}
	}()

	tnl, err := tunnel.NewEllipticTunnel(conn)
	if err != nil {
		log.Error("Tunnel failed", err)
		return
	}

	p := protocol.NewProtocol(tnl, false)

	// Loop until client disconnect or dies otherwise:
	for {
		msg := &wire.Command{}
		if err := p.Recv(msg); err != nil {
			if err != io.EOF {
				log.Warning("daemon-recv: ", err)
			}
			return
		}

		log.Infof("recv: %s: %v", conn.RemoteAddr().String(), msg)
		d.handleCommand(ctx, msg, p)
	}
}

// Handles the actual incoming commands:
func (d *Server) handleCommand(ctx context.Context, cmd *wire.Command, p *protocol.Protocol) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Prepare a response template:
	resp := &wire.Response{
		ResponseType: cmd.CommandType,
		Success:      proto.Bool(false),
	}

	// Figure out which handler to call:
	handlerID := *(cmd.CommandType)
	if handler, ok := handlerMap[handlerID]; !ok {
		resp.Error = proto.String(fmt.Sprintf("No handler for Id: %v", handlerID))
	} else {
		answer, err := handler(d, ctx, cmd)

		if err != nil {
			resp.Error = proto.String(err.Error())
		} else {
			resp.Response = answer
			resp.Success = proto.Bool(true)
		}
	}

	// Send the response back to the client:
	if err := p.Send(resp); err != nil {
		log.Warning("Unable to send message back to client: ", err)
	}
}

// Connect tries to connect the client and the ipfs daemon to the outside world.
func (sv *Server) Connect() error {
	if sv.IsOnline() {
		return nil
	}

	// Check if a previous offline mode was there:
	if err := sv.Repo.IPFS.Online(); err != nil {
		return err
	}

	if err := sv.MetaHost.Connect(); err != nil {
		log.Warningf("Unable to connect metadata client: %v", err)
		return err
	}

	log.Infof("Connected to the network")

	return nil
}

// Disconnect shuts down all store services that need an connection
// to the outside.
func (sv *Server) Disconnect() (err error) {
	if !sv.IsOnline() {
		return nil
	}

	log.Debugf("Disconnecting metadata host.")

	// Try to close metadata host, even if ipfs is still running:
	if err = sv.MetaHost.Disconnect(); err != nil {
		log.Warningf("Unable to disconnect metadata host: %v", err)
	}

	log.Debugf("Disconnecting ipfs daemon.")

	if err = sv.Repo.IPFS.Close(); err != nil {
		log.Warningf("Unable to close ipfs node: %v", err)
	}

	return err
}

// IsOnline checks if both meta host and ipfs is up and running.
func (sv *Server) IsOnline() bool {
	return sv.MetaHost.IsInOnlineMode() && sv.Repo.IPFS.IsOnline()
}
