package bench

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"time"

	"github.com/sahib/brig/backend/httpipfs"
	"github.com/sahib/brig/catfs/mio"
	"github.com/sahib/brig/client"
	"github.com/sahib/brig/client/clienttest"
	"github.com/sahib/brig/repo/hints"
	"github.com/sahib/brig/server"
	"github.com/sahib/brig/util/testutil"
)

type Bench interface {
	SupportHints() bool
	Bench(hint hints.Hint, r io.Reader) (time.Duration, error)
	Close() error
}

var (
	dummyKey = make([]byte, 32)
)

func withTiming(fn func() error) (time.Duration, error) {
	start := time.Now()
	err := fn()
	return time.Since(start), err
}

//////////

type NullBench struct{}

func NewNullBench() NullBench {
	return NullBench{}
}

func (n NullBench) SupportHints() bool { return false }

func (n NullBench) Bench(hint hints.Hint, r io.Reader) (time.Duration, error) {
	// NOTE: Use DumbCopy, since io.Copy would use the
	// ReadFrom of ioutil.Discard. This is lightning fast.
	// We want to measure actual time to copy in memory.
	return withTiming(func() error {
		_, err := testutil.DumbCopy(ioutil.Discard, r, false, false)
		return err
	})
}

func (n NullBench) Close() error { return nil }

//////////

type serverCommon struct {
	daemon *server.Server
	client *client.Client
}

func newServerCommon(ipfsPath string) (*serverCommon, error) {
	backendName := "mock"
	if ipfsPath != "" {
		backendName = "httpipfs"
	}

	srv, err := clienttest.StartDaemon("ali", backendName, ipfsPath)
	if err != nil {
		return nil, err
	}

	ctl, err := client.Dial(context.Background(), srv.DaemonURL())
	if err != nil {
		return nil, err
	}

	return &serverCommon{
		daemon: srv,
		client: ctl,
	}, nil
}

func (sc *serverCommon) Close() error {
	sc.daemon.Close()
	sc.client.Close()
	return nil
}

type ServerStageBench struct {
	common *serverCommon
}

func NewServerStageBench(ipfsPath string) (*ServerStageBench, error) {
	common, err := newServerCommon(ipfsPath)
	if err != nil {
		return nil, err
	}

	return &ServerStageBench{common: common}, nil
}

func (s *ServerStageBench) SupportHints() bool { return true }

func (s *ServerStageBench) Bench(hint hints.Hint, r io.Reader) (time.Duration, error) {
	path := fmt.Sprintf("/path_%d", rand.Int31())

	c := string(hint.CompressionAlgo)
	e := string(hint.EncryptionAlgo)
	if err := s.common.client.HintSet(path, &c, &e); err != nil {
		return 0, err
	}

	return withTiming(func() error {
		return s.common.client.StageFromReader(path, r)
	})
}

func (s *ServerStageBench) Close() error {
	return s.common.Close()
}

type ServerCatBench struct {
	common *serverCommon
}

func NewServerCatBench(ipfsPath string) (*ServerCatBench, error) {
	common, err := newServerCommon(ipfsPath)
	if err != nil {
		return nil, err
	}

	return &ServerCatBench{common: common}, nil
}

func (s *ServerCatBench) SupportHints() bool { return true }

func (s *ServerCatBench) Bench(hint hints.Hint, r io.Reader) (time.Duration, error) {
	path := fmt.Sprintf("/path_%d", rand.Int31())

	c := string(hint.CompressionAlgo)
	e := string(hint.EncryptionAlgo)
	if err := s.common.client.HintSet(path, &c, &e); err != nil {
		return 0, err
	}

	if err := s.common.client.StageFromReader(path, r); err != nil {
		return 0, err
	}

	return withTiming(func() error {
		stream, err := s.common.client.Cat(path, true)
		if err != nil {
			return err
		}

		defer stream.Close()

		_, err = testutil.DumbCopy(ioutil.Discard, stream, false, false)
		return err
	})
}

func (s *ServerCatBench) Close() error {
	return s.common.Close()
}

//////////

type MioWriterBench struct{}

func NewMioWriterBench() *MioWriterBench {
	return &MioWriterBench{}
}

func (m *MioWriterBench) SupportHints() bool { return true }

func (m *MioWriterBench) Bench(hint hints.Hint, r io.Reader) (time.Duration, error) {
	stream, _, err := mio.NewInStream(r, "", dummyKey, hint)
	if err != nil {
		return 0, err
	}

	defer stream.Close()

	return withTiming(func() error {
		_, err := testutil.DumbCopy(ioutil.Discard, stream, false, false)
		return err
	})
}

func (m *MioWriterBench) Close() error {
	return nil
}

//////////

type MioReaderBench struct{}

func NewMioReaderBench() *MioReaderBench {
	return &MioReaderBench{}
}

func (m *MioReaderBench) SupportHints() bool { return true }

func (m *MioReaderBench) Bench(hint hints.Hint, r io.Reader) (time.Duration, error) {
	// Produce a buffer with encoded data in the right size.
	// This is not benched, only the reading of it is.
	inStream, _, err := mio.NewInStream(r, "", dummyKey, hint)
	if err != nil {
		return 0, err
	}

	defer inStream.Close()

	// Read it to memory before measuring.
	// We do not want to count the encoding in the bench time.
	streamData, err := ioutil.ReadAll(inStream)
	if err != nil {
		return 0, err
	}

	return withTiming(func() error {
		outStream, err := mio.NewOutStream(
			bytes.NewReader(streamData),
			hint.IsRaw(),
			dummyKey,
		)

		if err != nil {
			return err
		}

		defer outStream.Close()

		_, err = testutil.DumbCopy(ioutil.Discard, outStream, false, false)
		return err
	})
}

func (m *MioReaderBench) Close() error {
	return nil
}

//////////

type IpfsAddOrCatBench struct {
	ipfsPath string
	isAdd    bool
}

func NewIpfsAddBench(ipfsPath string, isAdd bool) *IpfsAddOrCatBench {
	return &IpfsAddOrCatBench{ipfsPath: ipfsPath, isAdd: isAdd}
}

func (ia *IpfsAddOrCatBench) SupportHints() bool { return false }

func (ia *IpfsAddOrCatBench) Bench(hint hints.Hint, r io.Reader) (time.Duration, error) {
	nd, err := httpipfs.NewNode(ia.ipfsPath, "")
	if err != nil {
		return 0, err
	}

	defer nd.Close()

	if ia.isAdd {
		return withTiming(func() error {
			_, err := nd.Add(r)
			return err
		})
	}

	hash, err := nd.Add(r)
	if err != nil {
		return 0, err
	}

	return withTiming(func() error {
		stream, err := nd.Cat(hash)
		if err != nil {
			return err
		}

		_, err = testutil.DumbCopy(ioutil.Discard, stream, false, false)
		return err
	})
}

func (ia *IpfsAddOrCatBench) Close() error {
	return nil
}

//////////

// TODO: fuse-{read,write}
func BenchByName(name, ipfsPath string) (Bench, error) {
	switch name {
	case "null":
		return NewNullBench(), nil
	case "brig-stage-mem":
		return NewServerStageBench("")
	case "brig-cat-mem":
		return NewServerCatBench("")
	case "brig-stage-ipfs":
		return NewServerStageBench(ipfsPath)
	case "brig-cat-ipfs":
		return NewServerCatBench(ipfsPath)
	case "mio-writer":
		return NewMioWriterBench(), nil
	case "mio-reader":
		return NewMioReaderBench(), nil
	case "ipfs-add":
		return NewIpfsAddBench(ipfsPath, true), nil
	case "ipfs-cat":
		return NewIpfsAddBench(ipfsPath, false), nil
	default:
		return nil, fmt.Errorf("no such bench: %s", name)
	}
}
