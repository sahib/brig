package daemon

import (
	"github.com/disorganizer/brig/repo"
)

//////////////////////////
// COMMON CODE PORTIONS //
//////////////////////////

type Command interface {
	Stringer // Some interface for protobuf?
}

type Response interface {
	Stringer // TODO...
}

//////////////////////////
// FRONTEND DAEMON PART //
//////////////////////////

// https://github.com/docker/libchan

// Daemon is the top-level struct of the brig daemon.
type DaemonClient struct {
	// Port we're operating on
	Port int

	// Host we're talking to
	Host string

	// Use this channel to send commands to the daemon
	Send chan<- Command

	// Responses and errors are sent to this channel
	Recv <-chan Response
}

func Launch(repoPath string, host, string, port int) (*DaemonClient, error) {
	// fork, luanch daemonMain in child, return ready DaemonClient{} for father
	return nil, nil
}

func Dial(port int) (*DaemonClient, error) {
	return nil, nil
}

// Reach tries to Dial() the daemon, if not there it Launch()'es one.
func Reach(repoPath string, host string, port int) (*DaemonClient, error) {
	return nil, nil
}

func (c *DaemonClient) Close() {
	// ...
}

/////////////////////////
// BACKEND DAEMON PART //
/////////////////////////

type DaemonServer struct {
	// The repo we're working on
	Repo *repo.FsRepository

	// socket....
}

func (s *DaemonServer) daemonMain(repoPath string) {
	// Actual daemon init here.
}
