package daemon

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/daemon/proto"
	"github.com/disorganizer/brig/im"
	"github.com/disorganizer/brig/repo"
	"github.com/disorganizer/brig/util/tunnel"
	protobuf "github.com/gogo/protobuf/proto"
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

	// XMPP is the control client to the outside world.
	XMPP *im.Client

	// Handle to `ipfs daemon`
	ipfsDaemon *exec.Cmd

	// signals (external and self triggered) arrive on this channel.
	signals chan os.Signal

	// Root context for this daemon
	ctx context.Context

	// TCP Listener for incoming connections:
	listener net.Listener

	// buffered channel with N places,
	// every active connection holds one.
	maxConnections chan allowOneConn
}

// Summon creates a new up and running Server instance
func Summon(pwd, repoFolder string, port int) (*Server, error) {
	// Load the on-disk repository:
	log.Infof("Opening repo: %s", repoFolder)
	repository, err := repo.Open(pwd, repoFolder)
	if err != nil {
		log.Error("Could not load repository: ", err)
		return nil, err
	}

	// TODO: Uncomment later.
	// proc, err := ipfsutil.StartDaemon(repository.Store.IpfsCtx)
	// if err != nil {
	// 	log.Error("Unable to start ipfs daemon: ", err)
	// 	return nil, err
	// }

	// Listen for incoming connections.
	addr := fmt.Sprintf("localhost:%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Error("Error listening:", err.Error())
		return nil, err
	}

	// TODO: Uncomment later
	// xmppClient, err := im.NewClient(
	// 	&im.Config{
	// 		Jid:      xmpp.JID(repository.Jid),
	// 		Password: pwd,
	// 		TLSConfig: tls.Config{
	// 			ServerName: xmpp.JID(repository.Jid).Domain(),
	// 		},
	// 		KeyPath:              filepath.Join(repository.InternalFolder, "otr.key"),
	// 		FingerprintStorePath: filepath.Join(repository.InternalFolder, "otr.buddies"),
	// 	},
	// )
	// if err != nil {
	// 	return nil, err
	// }

	// Close the listener when the application closes.
	log.Info("Listening on ", addr)

	ctx, cancel := context.WithCancel(context.Background())
	daemon := &Server{
		Repo: repository,
		// XMPP:     xmppClient,
		signals:  make(chan os.Signal, 1),
		listener: listener,
		// ipfsDaemon:     proc,
		maxConnections: make(chan allowOneConn, MaxConnections),
		ctx:            ctx,
	}

	go daemon.loop(cancel)
	return daemon, nil
}

// Serve waits until the Server received a quit reason.
func (d *Server) Serve() {
	fmt.Println("Serving... ")
	<-d.ctx.Done()
	fmt.Println("Serving done... ")
	d.listener.Close()

	if d.ipfsDaemon != nil {
		if err := d.ipfsDaemon.Process.Kill(); err != nil {
			log.Errorf("Unable to kill off ipfs daemon: %v", err)
		}
	}

	if err := d.Repo.Close(); err != nil {
		log.Errorf("Unable to close repository: %v", err)
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
			go d.handleRequest(d.ctx, conn)
		default:
			log.Infof("Max number of connections hit: %d", cap(d.maxConnections))
			time.Sleep(500 * time.Millisecond)
		}
	}
}

// Handles incoming requests:
func (d *Server) handleRequest(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	// Make sure this connection count gets released
	defer func() {
		d.maxConnections <- allowOneConn{}
	}()

	tnl, err := tunnel.NewEllipticTunnel(conn)
	if err != nil {
		log.Error("Tunnel failed", err)
		return
	}

	// Loop until client disconnect or dies otherwise:
	for {
		msg := &proto.Command{}
		if err := recv(tnl, msg); err != nil {
			if err != io.EOF {
				log.Warning("daemon-recv: ", err)
			}
			return
		}

		log.Infof("recv: %s: %v", conn.RemoteAddr().String(), msg)
		d.handleCommand(ctx, msg, tnl)
	}
}

// Handles the actual incoming commands:
func (d *Server) handleCommand(ctx context.Context, cmd *proto.Command, conn io.ReadWriter) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Prepare a response template
	resp := &proto.Response{
		ResponseType: cmd.CommandType,
		Success:      protobuf.Bool(false),
	}

	switch *(cmd.CommandType) {
	case proto.MessageType_ADD:
		d.handleAddCommand(ctx, cmd, resp)
	case proto.MessageType_CAT:
		d.handleCatCommand(ctx, cmd, resp)
	case proto.MessageType_QUIT:
		resp.Response = protobuf.String("BYE")
		resp.Success = protobuf.Bool(true)
		d.signals <- os.Interrupt
	case proto.MessageType_PING:
		resp.Response = protobuf.String("PONG")
		resp.Success = protobuf.Bool(true)
	default:
		fmt.Println("Unknown message type.")
		return
	}

	if err := send(conn, resp); err != nil {
		log.Warning("Unable to send message back to client: ", err)
	}
}

func (d *Server) handleAddCommand(ctx context.Context, cmd *proto.Command, resp *proto.Response) {
	path := cmd.GetAddCommand().GetFilePath()
	fd, err := os.Open(path)
	if err != nil {
		resp.Error = protobuf.String(err.Error())
		return
	}

	hash, err := d.Repo.Store.Add(path, fd)
	if err != nil {
		resp.Error = protobuf.String(err.Error())
		return
	}

	resp.Success = protobuf.Bool(true)
	resp.Response = protobuf.String(hash.B58String())
}

func (d *Server) handleCatCommand(ctx context.Context, cmd *proto.Command, resp *proto.Response) {
	destPath := cmd.GetCatCommand().GetDestPath()
	fd, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		resp.Error = protobuf.String(err.Error())
		return
	}

	srcPath := cmd.GetCatCommand().GetFilePath()
	if err := d.Repo.Store.Cat(srcPath, fd); err != nil {
		resp.Error = protobuf.String(err.Error())
		return
	}

	resp.Success = protobuf.Bool(true)
	resp.Response = protobuf.String(srcPath)
}
