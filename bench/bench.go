package bench

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/pkg/xattr"
	"github.com/sahib/brig/backend/httpipfs"
	"github.com/sahib/brig/catfs/mio"
	"github.com/sahib/brig/client"
	"github.com/sahib/brig/client/clienttest"
	"github.com/sahib/brig/fuse/fusetest"
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

func NewNullBench(_ string, _ bool) (Bench, error) {
	return NullBench{}, nil
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

func NewServerStageBench(ipfsPath string, _ bool) (Bench, error) {
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

func NewServerCatBench(ipfsPath string, _ bool) (Bench, error) {
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

func NewMioWriterBench(_ string, _ bool) (Bench, error) {
	return &MioWriterBench{}, nil
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

func NewMioReaderBench(_ string, _ bool) (Bench, error) {
	return &MioReaderBench{}, nil
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

func NewIpfsAddBench(ipfsPath string, isAdd bool) (Bench, error) {
	return &IpfsAddOrCatBench{ipfsPath: ipfsPath, isAdd: isAdd}, nil
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

type FuseWriteOrReadBench struct {
	ipfsPath string
	isWrite  bool

	tmpDir string
	ctl    *fusetest.Client
	proc   *os.Process
}

func NewFuseWriteOrReadBench(ipfsPath string, isWrite bool) (Bench, error) {
	tmpDir, err := ioutil.TempDir("", "brig-fuse-bench-*")
	if err != nil {
		return nil, err
	}

	unixSocket := "unix:" + filepath.Join(tmpDir, "socket")

	proc, err := fusetest.LaunchAsProcess(fusetest.Options{
		MountPath: filepath.Join(tmpDir, "mount"),
		CatfsPath: filepath.Join(tmpDir, "catfs"),
		IpfsPath:  ipfsPath,
		URL:       unixSocket,
	})

	if err != nil {
		return nil, err
	}

	// bit time to start things up:
	time.Sleep(500 * time.Millisecond)

	ctl, err := fusetest.Dial(unixSocket)
	if err != nil {
		return nil, err
	}

	return &FuseWriteOrReadBench{
		ipfsPath: ipfsPath,
		isWrite:  isWrite,
		tmpDir:   tmpDir,
		proc:     proc,
		ctl:      ctl,
	}, nil
}

func (fb *FuseWriteOrReadBench) SupportHints() bool { return true }

func (fb *FuseWriteOrReadBench) Bench(hint hints.Hint, r io.Reader) (time.Duration, error) {
	mountDir := filepath.Join(fb.tmpDir, "mount")
	testPath := filepath.Join(mountDir, fmt.Sprintf("/path_%d", rand.Int31()))

	const (
		xattrEnc = "user.brig.hints.encryption"
		xattrZip = "user.brig.hints.compression"
	)

	// Make sure hints are followed:
	if err := xattr.Set(mountDir, xattrEnc, []byte(hint.EncryptionAlgo)); err != nil {
		return 0, err
	}

	if err := xattr.Set(mountDir, xattrZip, []byte(hint.CompressionAlgo)); err != nil {
		return 0, err
	}

	took, err := withTiming(func() error {
		fd, err := os.OpenFile(testPath, os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}

		defer fd.Close()

		_, err = testutil.DumbCopy(fd, r, false, false)
		return err
	})

	if err != nil {
		return 0, err
	}

	if fb.isWrite {
		// test is done already, no need to read-back.
		return took, nil
	}

	return withTiming(func() error {
		fd, err := os.Open(testPath)
		if err != nil {
			return err
		}

		defer fd.Close()

		_, err = testutil.DumbCopy(ioutil.Discard, fd, false, false)
		return err
	})
}

func (fb *FuseWriteOrReadBench) Close() error {
	// TODO: make sure it's dead.
	fb.ctl.QuitServer()
	time.Sleep(time.Second)
	fb.proc.Kill()
	return os.RemoveAll(fb.tmpDir)
}

//////////

var (
	// Convention:
	// - If it's using ipfs, put it in the name.
	// - If it's writing things, put that in the name too as "write".
	benchMap = map[string]func(string, bool) (Bench, error){
		"null":            NewNullBench,
		"brig-write-mem":  NewServerStageBench,
		"brig-read-mem":   NewServerCatBench,
		"brig-write-ipfs": NewServerStageBench,
		"brig-read-ipfs":  NewServerCatBench,
		"mio-write":       NewMioWriterBench,
		"mio-read":        NewMioReaderBench,
		"ipfs-write":      NewIpfsAddBench,
		"ipfs-read":       NewIpfsAddBench,
		"fuse-write-mem":  NewFuseWriteOrReadBench,
		"fuse-write-ipfs": NewFuseWriteOrReadBench,
		"fuse-read-mem":   NewFuseWriteOrReadBench,
		"fuse-read-ipfs":  NewFuseWriteOrReadBench,
	}
)

func BenchByName(name, ipfsPath string) (Bench, error) {
	newBench, ok := benchMap[name]
	if !ok {
		return nil, fmt.Errorf("no such bench: %s", name)
	}

	return newBench(ipfsPath, strings.Contains(name, "write"))
}

func BenchmarkNames() []string {
	names := []string{}
	for name := range benchMap {
		names = append(names, name)
	}

	sort.Slice(names, func(i, j int) bool {
		if names[i] == names[j] {
			return false
		}

		specials := []string{
			"null",
			"mio",
		}

		for _, special := range specials {
			v := strings.HasSuffix(names[i], special)
			if v || strings.HasSuffix(names[j], special) {
				return v
			}
		}

		return names[i] < names[j]
	})

	return names
}
