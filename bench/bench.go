package bench

// TODO: n_allocs, compression rate?
import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"syscall"
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

// Run is a single benchmark run
type Run struct {
	Took             time.Duration
	Allocs           int64
	CompressionRatio float32
}

// Runs is a list of individual runs
type Runs []Run

// Average returns a fictional average run out of all runs
func (runs Runs) Average() Run {
	sum := Run{}
	for _, run := range runs {
		sum.Took += run.Took
		sum.Allocs += run.Allocs
		sum.CompressionRatio += run.CompressionRatio
	}

	return Run{
		Took:             sum.Took / time.Duration(len(runs)),
		Allocs:           sum.Allocs / int64(len(runs)),
		CompressionRatio: sum.CompressionRatio / float32(len(runs)),
	}
}

// Bench is the interface every benchmark needs to implement.
type Bench interface {
	// SupportHints should return true for benchmarks where
	// passing hint influences the benchmark result.
	SupportHints() bool

	// CanBeVerified should return true when the test
	// can use the verifier (i.e. is a read test)
	CanBeVerified() bool

	// Bench should read the input from `r` and apply `hint` if applicable.
	// The time needed to process all of `r` should be returned.
	Bench(hint hints.Hint, size int64, r io.Reader, w io.Writer) (*Run, error)

	// Close should clean up the benchmark.
	Close() error
}

var (
	dummyKey = make([]byte, 32)
)

func withRunStats(size int64, fn func() (int64, error)) (*Run, error) {
	start := time.Now()

	var memBefore, memAfter runtime.MemStats
	runtime.ReadMemStats(&memBefore)
	written, err := fn()
	runtime.ReadMemStats(&memAfter)
	took := time.Since(start)
	return &Run{
		Took:             took,
		CompressionRatio: float32(written) / float32(size),
		Allocs:           int64(memAfter.Mallocs) - int64(memBefore.Mallocs),
	}, err
}

//////////

type memcpyBench struct{}

func newMemcpyBench(_ string, _ bool) (Bench, error) {
	return memcpyBench{}, nil
}

func (n memcpyBench) SupportHints() bool { return false }

func (n memcpyBench) CanBeVerified() bool { return true }

func (n memcpyBench) Bench(hint hints.Hint, size int64, r io.Reader, verifier io.Writer) (*Run, error) {
	// NOTE: Use DumbCopy, since io.Copy would use the
	// ReadFrom of ioutil.Discard. This is lightning fast.
	// We want to measure actual time to copy in memory.

	return withRunStats(size, func() (int64, error) {
		return testutil.DumbCopy(verifier, r, false, false)
	})
}

func (n memcpyBench) Close() error { return nil }

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

type serverStageBench struct {
	common *serverCommon
}

func newServerStageBench(ipfsPath string, _ bool) (Bench, error) {
	common, err := newServerCommon(ipfsPath)
	if err != nil {
		return nil, err
	}

	return &serverStageBench{common: common}, nil
}

func (s *serverStageBench) SupportHints() bool { return true }

func (s *serverStageBench) CanBeVerified() bool { return false }

func (s *serverStageBench) Bench(hint hints.Hint, size int64, r io.Reader, verifier io.Writer) (*Run, error) {
	path := fmt.Sprintf("/path_%d", rand.Int31())

	c := string(hint.CompressionAlgo)
	e := string(hint.EncryptionAlgo)
	if err := s.common.client.HintSet(path, &c, &e); err != nil {
		return nil, err
	}

	return withRunStats(size, func() (int64, error) {
		return size, s.common.client.StageFromReader(path, r)
	})
}

func (s *serverStageBench) Close() error {
	return s.common.Close()
}

type serverCatBench struct {
	common *serverCommon
}

func newServerCatBench(ipfsPath string, _ bool) (Bench, error) {
	common, err := newServerCommon(ipfsPath)
	if err != nil {
		return nil, err
	}

	return &serverCatBench{common: common}, nil
}

func (s *serverCatBench) SupportHints() bool { return true }

func (s *serverCatBench) CanBeVerified() bool { return true }

func (s *serverCatBench) Bench(hint hints.Hint, size int64, r io.Reader, verifier io.Writer) (*Run, error) {
	path := fmt.Sprintf("/path_%d", rand.Int31())
	c := string(hint.CompressionAlgo)
	e := string(hint.EncryptionAlgo)

	if err := s.common.client.HintSet(path, &c, &e); err != nil {
		return nil, err
	}

	if err := s.common.client.StageFromReader(path, r); err != nil {
		return nil, err
	}

	return withRunStats(size, func() (int64, error) {
		stream, err := s.common.client.Cat(path, true)
		if err != nil {
			return 0, err
		}

		defer stream.Close()
		return testutil.DumbCopy(verifier, stream, false, false)
	})
}

func (s *serverCatBench) Close() error {
	return s.common.Close()
}

//////////

type mioWriterBench struct{}

func newMioWriterBench(_ string, _ bool) (Bench, error) {
	return &mioWriterBench{}, nil
}

func (m *mioWriterBench) SupportHints() bool { return true }

func (m *mioWriterBench) CanBeVerified() bool { return false }

func (m *mioWriterBench) Bench(hint hints.Hint, size int64, r io.Reader, verifier io.Writer) (*Run, error) {
	stream, _, err := mio.NewInStream(r, "", dummyKey, hint)
	if err != nil {
		return nil, err
	}

	return withRunStats(size, func() (int64, error) {
		defer stream.Close()
		return testutil.DumbCopy(ioutil.Discard, stream, false, false)
	})
}

func (m *mioWriterBench) Close() error {
	return nil
}

//////////

type mioReaderBench struct{}

func newMioReaderBench(_ string, _ bool) (Bench, error) {
	return &mioReaderBench{}, nil
}

func (m *mioReaderBench) SupportHints() bool { return true }

func (m *mioReaderBench) CanBeVerified() bool { return true }

func (m *mioReaderBench) Bench(hint hints.Hint, size int64, r io.Reader, verifier io.Writer) (*Run, error) {
	// Produce a buffer with encoded data in the right size.
	// This is not benched, only the reading of it is.
	inStream, _, err := mio.NewInStream(r, "", dummyKey, hint)
	if err != nil {
		return nil, err
	}

	defer inStream.Close()

	// Read it to memory before measuring.
	// We do not want to count the encoding in the bench time.
	streamData, err := ioutil.ReadAll(inStream)
	if err != nil {
		return nil, err
	}

	return withRunStats(size, func() (int64, error) {
		outStream, err := mio.NewOutStream(
			bytes.NewReader(streamData),
			hint.IsRaw(),
			dummyKey,
		)

		if err != nil {
			return -1, err
		}

		defer outStream.Close()

		return testutil.DumbCopy(verifier, outStream, false, false)
	})
}

func (m *mioReaderBench) Close() error {
	return nil
}

//////////

type ipfsAddOrCatBench struct {
	ipfsPath string
	isAdd    bool
}

func newIPFSAddBench(ipfsPath string, isAdd bool) (Bench, error) {
	return &ipfsAddOrCatBench{ipfsPath: ipfsPath, isAdd: isAdd}, nil
}

func (ia *ipfsAddOrCatBench) SupportHints() bool { return false }

func (ia *ipfsAddOrCatBench) CanBeVerified() bool { return !ia.isAdd }

func (ia *ipfsAddOrCatBench) Bench(hint hints.Hint, size int64, r io.Reader, verifier io.Writer) (*Run, error) {
	nd, err := httpipfs.NewNode(ia.ipfsPath, "")
	if err != nil {
		return nil, err
	}

	defer nd.Close()

	if ia.isAdd {
		return withRunStats(size, func() (int64, error) {
			_, err := nd.Add(r)
			return size, err
		})
	}

	hash, err := nd.Add(r)
	if err != nil {
		return nil, err
	}

	return withRunStats(size, func() (int64, error) {
		stream, err := nd.Cat(hash)
		if err != nil {
			return -1, err
		}

		return testutil.DumbCopy(verifier, stream, false, false)
	})
}

func (ia *ipfsAddOrCatBench) Close() error {
	return nil
}

//////////

type fuseWriteOrReadBench struct {
	ipfsPath string
	isWrite  bool

	tmpDir string
	ctl    *fusetest.Client
	proc   *os.Process
}

func newFuseWriteOrReadBench(ipfsPath string, isWrite bool) (Bench, error) {
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

	return &fuseWriteOrReadBench{
		ipfsPath: ipfsPath,
		isWrite:  isWrite,
		tmpDir:   tmpDir,
		proc:     proc,
		ctl:      ctl,
	}, nil
}

func (fb *fuseWriteOrReadBench) SupportHints() bool { return true }

func (fb *fuseWriteOrReadBench) CanBeVerified() bool { return !fb.isWrite }

func (fb *fuseWriteOrReadBench) Bench(hint hints.Hint, size int64, r io.Reader, verifier io.Writer) (*Run, error) {
	mountDir := filepath.Join(fb.tmpDir, "mount")
	testPath := filepath.Join(mountDir, fmt.Sprintf("/path_%d", rand.Int31()))

	const (
		xattrEnc = "user.brig.hints.encryption"
		xattrZip = "user.brig.hints.compression"
	)

	// Make sure hints are followed:
	if err := xattr.Set(mountDir, xattrEnc, []byte(hint.EncryptionAlgo)); err != nil {
		return nil, err
	}

	if err := xattr.Set(mountDir, xattrZip, []byte(hint.CompressionAlgo)); err != nil {
		return nil, err
	}

	took, err := withRunStats(size, func() (int64, error) {
		fd, err := os.OpenFile(testPath, os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return -1, err
		}

		defer fd.Close()

		return testutil.DumbCopy(fd, r, false, false)
	})

	if err != nil {
		return nil, err
	}

	if fb.isWrite {
		// test is done already, no need to read-back.
		return took, nil
	}

	took, err = withRunStats(size, func() (int64, error) {
		// NOTE: We have to use syscall.O_DIRECT here in order to
		//       bypass the kernel page cache. The write above fills it with
		//       data immediately, thus this read can yield 10x times higher
		//       results (which you still might get in practice, if lucky)
		fd, err := os.OpenFile(testPath, os.O_RDONLY|syscall.O_DIRECT, 0600)
		if err != nil {
			return -1, err
		}

		defer fd.Close()

		return testutil.DumbCopy(verifier, fd, false, false)
	})

	return took, err
}

func (fb *fuseWriteOrReadBench) Close() error {
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
		"memcpy":          newMemcpyBench,
		"brig-write-mem":  newServerStageBench,
		"brig-read-mem":   newServerCatBench,
		"brig-write-ipfs": newServerStageBench,
		"brig-read-ipfs":  newServerCatBench,
		"mio-write":       newMioWriterBench,
		"mio-read":        newMioReaderBench,
		"ipfs-write":      newIPFSAddBench,
		"ipfs-read":       newIPFSAddBench,
		"fuse-write-mem":  newFuseWriteOrReadBench,
		"fuse-write-ipfs": newFuseWriteOrReadBench,
		"fuse-read-mem":   newFuseWriteOrReadBench,
		"fuse-read-ipfs":  newFuseWriteOrReadBench,
	}
)

// ByName returns the benchmark with this name, or an error
// if none. If IPFS is used, it should be given as `ipfsPath`.
func ByName(name, ipfsPath string) (Bench, error) {
	newBench, ok := benchMap[name]
	if !ok {
		return nil, fmt.Errorf("no such bench: %s", name)
	}

	return newBench(ipfsPath, strings.Contains(name, "write"))
}

// BenchmarkNames returns all possible benchmark names
// in an defined & stable sorting.
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
			"memcpy",
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
