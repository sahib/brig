package bench

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"

	"github.com/sahib/brig/client"
	"github.com/sahib/brig/client/clienttest"
	"github.com/sahib/brig/repo/hints"
	"github.com/sahib/brig/server"
	"github.com/sahib/brig/util/testutil"
)

type Bench interface {
	SupportHints() bool
	SetHint(hint hints.Hint) error
	Process(r io.Reader) error
	Close() error
}

//////////

type NullBench struct{}

func NewNullBench() NullBench {
	return NullBench{}
}

func (n NullBench) SupportHints() bool            { return false }
func (n NullBench) SetHint(hint hints.Hint) error { return nil }

func (n NullBench) Process(r io.Reader) error {
	// NOTE: Use DumbCopy, since io.Copy would use the
	// ReadFrom of ioutil.Discard. This is lightning fast.
	// We want to measure actual time to copy in memory.
	_, err := testutil.DumbCopy(ioutil.Discard, r, false, false)
	return err
}

func (n NullBench) Close() error { return nil }

//////////

// TODO: Make backend configurable we can also test with ipfs.

type ServerStageBench struct {
	daemon *server.Server
	client *client.Client
}

func NewServerStageBench() (*ServerStageBench, error) {
	srv, err := clienttest.StartDaemon("ali", "mock")
	if err != nil {
		return nil, err
	}

	ctl, err := client.Dial(context.Background(), srv.DaemonURL())
	if err != nil {
		return nil, err
	}

	return &ServerStageBench{
		daemon: srv,
		client: ctl,
	}, nil
}

func (s *ServerStageBench) SupportHints() bool { return true }

func (s *ServerStageBench) SetHint(hint hints.Hint) error {
	c := string(hint.CompressionAlgo)
	e := string(hint.EncryptionAlgo)
	return s.client.HintSet("/", &c, &e)
}

func (s *ServerStageBench) Process(r io.Reader) error {
	path := fmt.Sprintf("/path_%d", rand.Int31())
	return s.client.StageFromReader(path, r)
}

func (s *ServerStageBench) Close() error {
	s.daemon.Close()
	s.client.Close()
	return nil
}

// TODO: brig-stage-ipfs
// TODO: ipfs add
// TODO: fuse-write

func BenchByName(name string) (Bench, error) {
	switch name {
	case "null":
		return NewNullBench(), nil
	case "brig-stage":
		return NewServerStageBench()
	default:
		return nil, fmt.Errorf("no such bench: %s", name)
	}
}
