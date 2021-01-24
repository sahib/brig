// +build !windows

package fuse

import (
	"bytes"
	"context"
	"encoding/binary"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/sahib/brig/catfs"
	"github.com/sahib/brig/defaults"
	"github.com/sahib/brig/util/testutil"
	"github.com/sahib/config"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"bazil.org/fuse/fs/fstestutil/spawntest"
	"bazil.org/fuse/fs/fstestutil/spawntest/httpjson"
)

// The routines which manage fuse layer
// and OS dependent code (os.Open, and similar)  MUST BE SEPARATE OS PROCESSES.
// Note: not different go routines but processes!
// See https://github.com/bazil/fuse/issues/264#issuecomment-727269770
// This separation happens automatically during normal brig operations, but
// TESTING FUSE LAYER IN GO IS VERY TRICKY.
// See brig relevant discussion at
// https://github.com/sahib/brig/pull/77#issuecomment-754831080
// However this issue is general for any go program from version 1.9,
// as can be seen in references to the issue.
//
// bazil/fuse offers "bazil.org/fuse/fs/fstestutil/spawntest"
// infrastructure which helps run tests in different communicating via socket processes.

func init() {
	log.SetLevel(log.ErrorLevel)
}

func TestMain(m *testing.M) {
	helpers.AddFlag(flag.CommandLine)
	flag.Parse()
	helpers.RunIfNeeded()
	os.Exit(m.Run())
}

type fuseCatFSHelp struct{}

// These helpers will be requested from test and executed on the server
// which is managing catfs-fuse connection (started within test)
func (fch *fuseCatFSHelp) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.URL.Path {
	case "/mount":
		httpjson.ServePOST(fch.makeCatfsAndFuseMount).ServeHTTP(w, req)
	case "/unmount":
		httpjson.ServePOST(fch.unmountFuseAndCloseDummyCatFS).ServeHTTP(w, req)
	case "/fuseReMount":
		httpjson.ServePOST(fch.makeFuseReMount).ServeHTTP(w, req)
	case "/catfsStage":
		httpjson.ServePOST(fch.catfsStage).ServeHTTP(w, req)
	case "/catfsGetData":
		httpjson.ServePOST(fch.catfsGetData).ServeHTTP(w, req)
	default:
		http.NotFound(w, req)
	}
}

func makeDummyCatFS(dbPath string) (catfsFuseInfo, error) {
	backend := catfs.NewMemFsBackend()
	owner := "alice"

	cfg, err := config.Open(nil, defaults.Defaults, config.StrictnessPanic)
	if err != nil {
		log.Fatalf("Could not get default catFS config: %v", err)
		return catfsFuseInfo{}, err
	}

	cfs, err := catfs.NewFilesystem(backend, dbPath, owner, false, cfg.Section("fs"))
	if err != nil {
		log.Fatalf("Failed to create catfs filesystem: %v", err)
		return catfsFuseInfo{}, err
	}
	cfInfo := catfsFuseInfo{}
	cfInfo.cfs = cfs
	cfInfo.dbPath = dbPath

	return cfInfo, err
}

type nothing struct{} // use it to send empty request or responses to server

type catfsFuseInfo struct {
	cfs       *catfs.FS
	dbPath    string
	fuseMount *Mount
}

// cfInfo will be in the global space for the server
// which manage fuse mount connection to the catFS
var cfInfo catfsFuseInfo

type mountingRequest struct {
	DbPath  string
	MntPath string
	Opts    MountOptions
}

func (fch *fuseCatFSHelp) makeCatfsAndFuseMount(ctx context.Context, req mountingRequest) (*nothing, error) {
	var err error
	cfInfo, err = makeDummyCatFS(req.DbPath)
	if err != nil {
		log.Errorf("cannot make catFS in %v", cfInfo.dbPath)
		return &nothing{}, err
	}

	fuseMount, err := makeFuseMount(cfInfo.cfs, req.MntPath, req.Opts)
	if err != nil {
		log.Errorf("cannot mount catfs fuse file system to %v", req.MntPath)
		return &nothing{}, err
	}
	cfInfo.fuseMount = fuseMount
	return &nothing{}, err
}

func (fch *fuseCatFSHelp) makeFuseReMount(ctx context.Context, req mountingRequest) (*nothing, error) {
	fuseMount, err := makeFuseMount(cfInfo.cfs, req.MntPath, req.Opts)
	if err != nil {
		log.Errorf("cannot mount catfs fuse file system to %v", req.MntPath)
		return &nothing{}, err
	}
	cfInfo.fuseMount = fuseMount
	return &nothing{}, err
}

func (fch *fuseCatFSHelp) unmountFuseAndCloseDummyCatFS(ctx context.Context, req nothing) (*nothing, error) {
	defer os.RemoveAll(cfInfo.fuseMount.Dir)
	defer os.RemoveAll(cfInfo.dbPath)
	// first unmount fuse directory
	if err := lazyUnmount(cfInfo.fuseMount.Dir); err != nil {
		skipableErr := "exit status 1: fusermount: entry for " + cfInfo.fuseMount.Dir + " not found in /etc/mtab"
		log.Debug(skipableErr)
		if err.Error() != skipableErr {
			return &nothing{}, err
		}
	}

	// now close catFS
	err := cfInfo.cfs.Close()
	if err != nil {
		log.Fatalf("Could not close catfs filesystem: %v", err)
	}
	return &nothing{}, err
}

func makeFuseMount(cfs *catfs.FS, mntPath string, opts MountOptions) (*Mount, error) {
	// Make sure to unmount any mounts that are there.
	// Possibly there are some leftovers from previous failed runs.
	if err := lazyUnmount(mntPath); err != nil {
		skipableErr := "exit status 1: fusermount: entry for " + mntPath + " not found in /etc/mtab"
		log.Debug(skipableErr)
		if err.Error() != skipableErr {
			return nil, err
		}
	}

	if err := os.MkdirAll(mntPath, 0777); err != nil {
		log.Fatalf("Unable to create empty mount dir: %v", err)
		return nil, err
	}

	mount, err := NewMount(cfs, mntPath, nil, opts)
	if err != nil {
		log.Fatalf("Cannot create mount: %v", err)
		return nil, err
	}
	return mount, err
}

type catfsPayload struct {
	Path string
	Data []byte
}

func (fch *fuseCatFSHelp) catfsStage(ctx context.Context, req catfsPayload) (*nothing, error) {
	err := cfInfo.cfs.Stage(req.Path, bytes.NewReader(req.Data))
	return &nothing{}, err
}

// Get data from a file stored by catFS
func (fch *fuseCatFSHelp) catfsGetData(ctx context.Context, req catfsPayload) (*catfsPayload, error) {
	out := catfsPayload{}
	out.Path = req.Path

	stream, err := cfInfo.cfs.Cat(req.Path)
	if err != nil {
		log.Fatalf("Could not get stream for a catfs file: %v", err)
		return &out, err
	}
	result := bytes.NewBuffer(nil)
	_, err = stream.WriteTo(result)
	if err != nil {
		log.Fatalf("Streaming to a buffer failed: %v", err)
		return &out, err
	}
	out.Data = result.Bytes()

	return &out, err
}

var helpers spawntest.Registry
var fuseCatFSHelper = helpers.Register("fuseCatFSHelp", &fuseCatFSHelp{})

type mountInfo struct { // fuse related info available to OS layer
	Dir  string
	Opts MountOptions
}

// Call helper for unmount and cleanup
func callUnMount(ctx context.Context, t testing.TB, control *spawntest.Control) {
	if err := control.JSON("/unmount").Call(ctx, nothing{}, &nothing{}); err != nil {
		t.Fatalf("calling helper: %v", err)
	}
}

// Spawns helper, prepare catFS, connects it to fuse layer, and execute function f
func withMount(t testing.TB, opts MountOptions, f func(ctx context.Context, control *spawntest.Control, mount *mountInfo)) {
	// set up mounts
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	control := fuseCatFSHelper.Spawn(ctx, t)
	defer control.Close()

	dbPath, err := ioutil.TempDir("", "catfs-db-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir for catFS: %v", err)
	}

	req := mountingRequest{
		DbPath:  dbPath,
		MntPath: filepath.Join(os.TempDir(), "catfs-fuse-mountdir"),
		Opts:    opts,
	}

	if err := control.JSON("/mount").Call(ctx, req, &nothing{}); err != nil {
		t.Fatalf("calling helper: %v", err)
	}
	defer callUnMount(ctx, t, control)

	// function which required mounts
	f(ctx, control, &mountInfo{
		Dir:  req.MntPath,
		Opts: req.Opts,
	})
}

func checkFuseFileMatchToCatFS(ctx context.Context, t *testing.T, control *spawntest.Control, fusePath string, catfsPath string) {
	// checks if OS file content matches catFS file content
	fuseData, err := ioutil.ReadFile(fusePath)
	require.NoError(t, err)

	// is catFS seeing the same data
	checkCatfsFileContent(ctx, t, control, catfsPath, fuseData)
}

func checkCatfsFileContent(ctx context.Context, t *testing.T, control *spawntest.Control, catfsPath string, expected []byte) {
	req := catfsPayload{Path: catfsPath}
	out := catfsPayload{}
	require.NoError(t, control.JSON("/catfsGetData").Call(ctx, req, &out))
	require.Equal(t, len(out.Data), len(expected))
	if out.Data == nil {
		// this is special for the 0 length data
		out.Data = []byte{}
	}
	require.Equal(t, out.Data, expected)
}

// Finally we ready to do tests

// Tests for spawntest infrastructure related tests

// Just checks that our catfsStage interface to catFS does not error out
func TestCatfsStage(t *testing.T) {
	withMount(t, MountOptions{}, func(ctx context.Context, control *spawntest.Control, mount *mountInfo) {
		dataIn := []byte{1, 2, 3, 4}
		filePath := "StagingTest.bin"
		req := catfsPayload{Path: filePath, Data: dataIn}
		require.NoError(t, control.JSON("/catfsStage").Call(ctx, req, &nothing{}))
	})
}

func TestCatfsGetData(t *testing.T) {
	withMount(t, MountOptions{}, func(ctx context.Context, control *spawntest.Control, mount *mountInfo) {
		dataIn := []byte{1, 2, 3, 4}
		filePath := "StageAndReadTest.bin"
		req := catfsPayload{Path: filePath, Data: dataIn}
		require.NoError(t, control.JSON("/catfsStage").Call(ctx, req, &nothing{}))

		req.Data = []byte{}
		out := catfsPayload{}
		require.NoError(t, control.JSON("/catfsGetData").Call(ctx, req, &out))
		require.Equal(t, out.Data, dataIn)
	})
}

// Main fuse layer tests

var (
	DataSizes = []int64{
		0, 1, 2, 4, 8, 16, 32, 64, 1024,
		2048, 4095, 4096, 4097, 147611,
		2*1024*1024 + 123, // in case if we have buffer size interference
	}
)

func TestRead(t *testing.T) {
	withMount(t, MountOptions{}, func(ctx context.Context, control *spawntest.Control, mount *mountInfo) {
		for _, size := range DataSizes {
			t.Run(fmt.Sprintf("%d", size), func(t *testing.T) {
				helloData := testutil.CreateDummyBuf(size)

				// Add a simple file:
				catfsFilePath := fmt.Sprintf("/hello_from_catfs_%d", size)
				req := catfsPayload{Path: catfsFilePath, Data: helloData}
				require.NoError(t, control.JSON("/catfsStage").Call(ctx, req, &nothing{}))
				checkCatfsFileContent(ctx, t, control, catfsFilePath, helloData)

				fuseFilePath := filepath.Join(mount.Dir, catfsFilePath)
				checkFuseFileMatchToCatFS(ctx, t, control, fuseFilePath, catfsFilePath)
			})
		}
	})
}

func TestFileXattr(t *testing.T) {
	withMount(t, MountOptions{}, func(ctx context.Context, control *spawntest.Control, mount *mountInfo) {
		size := int64(4)
		helloData := testutil.CreateDummyBuf(size)

		// Add a simple file:
		catfsFilePath := fmt.Sprintf("/hello_from_catfs_%d", size)
		req := catfsPayload{Path: catfsFilePath, Data: helloData}
		require.NoError(t, control.JSON("/catfsStage").Call(ctx, req, &nothing{}))
		checkCatfsFileContent(ctx, t, control, catfsFilePath, helloData)

		fuseFilePath := filepath.Join(mount.Dir, catfsFilePath)

		// no let's see all the extended attributes list
		response := make([]byte, 1024*4) // large buffer to fit everything
		sz, err := syscall.Listxattr(fuseFilePath, response)
		require.NoError(t, err)
		response = response[:sz]
		receivedAttrs := bytes.Split(response, []byte{0})
		// every response should belong to valid attributes
		for _, attr := range receivedAttrs {
			if len(attr) == 0 {
				// protecting against empty chunk after split delimiter
				continue
			}
			_, ok := xattrMap[string(attr)]
			require.Truef(t, ok, "Invalid extended attribute '%s'", attr)
		}
		// every valid attribute should be in received Attrs list
		for attr, _ := range xattrMap {
			require.Containsf(t, receivedAttrs, []uint8(attr), "Received attributes are missing '%s'", attr)
		}
		// now let's check some attributes values
		// Note hashes are hard without direct access to catfs
		// which is accessed in different process
		response = make([]byte, 64) // large buffer to fit everything
		sz, err = syscall.Getxattr(fuseFilePath, "user.brig.pinned", response)
		require.NoError(t, err)
		response = response[:sz]
		require.Equal(t, "yes", string(response))

		response = make([]byte, 64) // large buffer to fit everything
		sz, err = syscall.Getxattr(fuseFilePath, "user.brig.explicitly_pinned", response)
		require.NoError(t, err)
		response = response[:sz]
		require.Equal(t, "no", string(response))
	})
}

func TestWrite(t *testing.T) {
	withMount(t, MountOptions{}, func(ctx context.Context, control *spawntest.Control, mount *mountInfo) {
		for _, size := range DataSizes {
			t.Run(fmt.Sprintf("%d", size), func(t *testing.T) {
				helloData := testutil.CreateDummyBuf(size)

				catfsFilePath := fmt.Sprintf("/hello_from_fuse%d", size)
				fuseFilePath := filepath.Join(mount.Dir, catfsFilePath)

				// Write a simple file via the fuse layer:
				err := ioutil.WriteFile(fuseFilePath, helloData, 0644)
				if err != nil {
					t.Fatalf("Could not write simple file via fuse layer: %v", err)
				}
				checkCatfsFileContent(ctx, t, control, catfsFilePath, helloData)
			})
		}
	})
}

// Regression test for copying larger file to the mount.
func TestTouchWrite(t *testing.T) {
	withMount(t, MountOptions{}, func(ctx context.Context, control *spawntest.Control, mount *mountInfo) {
		for _, size := range DataSizes {
			t.Run(fmt.Sprintf("%d", size), func(t *testing.T) {

				catfsFilePath := fmt.Sprintf("/empty_at_creation_by_catfs_%d", size)
				req := catfsPayload{Path: catfsFilePath, Data: []byte{}}

				require.NoError(t, control.JSON("/catfsStage").Call(ctx, req, &nothing{}))
				checkCatfsFileContent(ctx, t, control, catfsFilePath, req.Data)

				fuseFilePath := filepath.Join(mount.Dir, catfsFilePath)

				// Write a simple file via the fuse layer:
				helloData := testutil.CreateDummyBuf(size)
				err := ioutil.WriteFile(fuseFilePath, helloData, 0644)
				if err != nil {
					t.Fatalf("Could not write simple file via fuse layer: %v", err)
				}
				checkCatfsFileContent(ctx, t, control, catfsFilePath, helloData)
			})
		}
	})
}

// Regression test for copying a file to a subdirectory.
func TestTouchWriteSubdir(t *testing.T) {
	withMount(t, MountOptions{}, func(ctx context.Context, control *spawntest.Control, mount *mountInfo) {
		file := "donald.png"
		subDirPath := "sub"
		catfsFilePath := filepath.Join(subDirPath, file)

		fuseSubDirPath := filepath.Join(mount.Dir, subDirPath)
		fuseFilePath := filepath.Join(fuseSubDirPath, file)

		require.NoError(t, os.Mkdir(fuseSubDirPath, 0644))

		expected := []byte{1, 2, 3}
		require.NoError(t, ioutil.WriteFile(fuseFilePath, expected, 0644))

		checkCatfsFileContent(ctx, t, control, catfsFilePath, expected)
	})
}

func TestReadOnlyFs(t *testing.T) {
	opts := MountOptions{
		ReadOnly: true,
	}
	withMount(t, opts, func(ctx context.Context, control *spawntest.Control, mount *mountInfo) {
		xData := []byte{1, 2, 3}
		req := catfsPayload{Path: "/x.png", Data: xData}

		require.NoError(t, control.JSON("/catfsStage").Call(ctx, req, &nothing{}))
		checkCatfsFileContent(ctx, t, control, "x.png", xData)

		// Do some allowed io to check if the fs is actually working.
		// The test does not check on the kind of errors otherwise.
		xPath := filepath.Join(mount.Dir, "x.png")
		checkFuseFileMatchToCatFS(ctx, t, control, xPath, "x.png")

		// Try creating a new file:
		yPath := filepath.Join(mount.Dir, "y.png")
		require.NotNil(t, ioutil.WriteFile(yPath, []byte{4, 5, 6}, 0600))

		// Try modifying an existing file:
		require.NotNil(t, ioutil.WriteFile(xPath, []byte{4, 5, 6}, 0600))

		dirPath := filepath.Join(mount.Dir, "sub")
		require.NotNil(t, os.Mkdir(dirPath, 0644))
	})
}

func TestWithRoot(t *testing.T) {
	withMount(t, MountOptions{}, func(ctx context.Context, control *spawntest.Control, mount *mountInfo) {
		data := []byte{1, 2, 3}
		// Populate catFS with some files in different directories
		req := catfsPayload{Path: "/u.png", Data: data}

		require.NoError(t, control.JSON("/catfsStage").Call(ctx, req, &nothing{}))
		checkCatfsFileContent(ctx, t, control, req.Path, data)
		checkFuseFileMatchToCatFS(ctx, t, control, filepath.Join(mount.Dir, req.Path), req.Path)

		data = []byte{2, 3, 4}
		req = catfsPayload{Path: "/a/x.png", Data: data}
		require.NoError(t, control.JSON("/catfsStage").Call(ctx, req, &nothing{}))
		checkCatfsFileContent(ctx, t, control, req.Path, data)
		checkFuseFileMatchToCatFS(ctx, t, control, filepath.Join(mount.Dir, req.Path), req.Path)

		data = []byte{3, 4, 5}
		req = catfsPayload{Path: "/a/b/y.png", Data: data}
		require.NoError(t, control.JSON("/catfsStage").Call(ctx, req, &nothing{}))
		checkCatfsFileContent(ctx, t, control, req.Path, data)
		checkFuseFileMatchToCatFS(ctx, t, control, filepath.Join(mount.Dir, req.Path), req.Path)

		data = []byte{4, 5, 6}
		req = catfsPayload{Path: "/a/b/c/z.png", Data: data}
		require.NoError(t, control.JSON("/catfsStage").Call(ctx, req, &nothing{}))
		checkCatfsFileContent(ctx, t, control, req.Path, data)
		checkFuseFileMatchToCatFS(ctx, t, control, filepath.Join(mount.Dir, req.Path), req.Path)

		// Now we need to remount fuse with different root directory
		remntReq := mountingRequest{
			MntPath: mount.Dir,
			Opts:    MountOptions{Root: "/a/b"},
		}
		require.NoError(t, control.JSON("/fuseReMount").Call(ctx, remntReq, &nothing{}))
		mount.Opts = remntReq.Opts // update with new mount options

		// See if fuse indeed provides different root
		// Read already existing file
		yPath := filepath.Join(mount.Dir, "y.png")
		checkFuseFileMatchToCatFS(ctx, t, control, yPath, "/a/b/y.png")

		// Write to a new file
		data = []byte{5, 6, 7}
		newPath := filepath.Join(mount.Dir, "new.png")

		require.NoError(t, ioutil.WriteFile(newPath, data, 0644))
		checkCatfsFileContent(ctx, t, control, "/a/b/new.png", data)

		// Attempt to read file above mounted root
		inAccessiblePath := filepath.Join(mount.Dir, "u.png")
		_, err := ioutil.ReadFile(inAccessiblePath)
		require.NotNil(t, err)
	})
}

// Benchmarks

var (
	BenchmarkDataSizes = []int64{
		0,
		1024, 2 * 1024, 16 * 1024, 64 * 1024, 128 * 1024,
		1 * 1024 * 1024, 16 * 1024 * 1024,
	}
)

func stageAndRead(ctx context.Context, b *testing.B, control *spawntest.Control, mount *mountInfo, label string, data []byte) {
	size := len(data)
	// stage data to catFS
	catfsFilePath := fmt.Sprintf("%s_file_%d", label, size)
	req := catfsPayload{Path: catfsFilePath, Data: data}
	require.NoError(b, control.JSON("/catfsStage").Call(ctx, req, &nothing{}))
	fuseFilePath := filepath.Join(mount.Dir, catfsFilePath)

	// Read it back via fuse
	b.Run(fmt.Sprintf("%s_Size_%d", label, size), func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			ioutil.ReadFile(fuseFilePath)
		}
	})
}

func BenchmarkRead(b *testing.B) {
	withMount(b, MountOptions{}, func(ctx context.Context, control *spawntest.Control, mount *mountInfo) {
		for _, size := range BenchmarkDataSizes {
			// Check how fast is readout of a file with compressible content
			data := testutil.CreateDummyBuf(size)
			stageAndRead(ctx, b, control, mount, "CompressibleContent", data)

			// Check how fast is readout of a file with random/uncompressible content
			data = testutil.CreateRandomDummyBuf(size, 1)
			stageAndRead(ctx, b, control, mount, "RandomContent", data)
		}
	})
}

func writeDataNtimes(b *testing.B, data []byte, ntimes int) {
	// Writing could be very space demanding even for a small size,
	// Since benchmark runs many-many times, it will consume a lot of space.
	// We have to remount everything every time to start with clean catFS DB.
	// Consequently, this test takes long time, since mounting is long operation.
	require.True(b, ntimes > 0, "ntimes must be positive")
	// note ntimes =0 is bad too,
	// since execution time between StartTimer/StopTimer is too short/jittery
	// and benchmarks run forever
	label := "dummy"
	size := len(data)
	for n := 0; n < b.N; n++ {
		b.StopTimer()
		withMount(b, MountOptions{}, func(ctx context.Context, control *spawntest.Control, mount *mountInfo) {
			// Check how fast is write to a file with compressible content
			catfsFilePath := fmt.Sprintf("%s_file_%d", label, size)
			fuseFilePath := filepath.Join(mount.Dir, catfsFilePath)

			b.StartTimer()
			for i := 0; i < ntimes; i++ {
				if len(data) > 0 {
					// modification of one byte is enough
					// to generate new encrypted content for the backend
					binary.LittleEndian.PutUint64(data[0:8], uint64(i))
				}
				require.NoError(b, ioutil.WriteFile(fuseFilePath, data, 0644))
			}
			b.StopTimer()
		})
	}
}

var (
	// keep this low or you might run out of space
	NumberOfOverWrites = []int{
		1, 2, 5,
	}
)

func BenchmarkWrite(b *testing.B) {
	size := int64(10 * 1024 * 1024)

	for _, Ntimes := range NumberOfOverWrites {
		// Check how fast is write to a file with compressible content
		data := testutil.CreateDummyBuf(size)
		prefix := fmt.Sprintf("Owerwrite_%d", Ntimes)
		label := fmt.Sprintf("%s/CompressibleContent_Size_%d", prefix, size)
		b.Run(label, func(b *testing.B) {
			writeDataNtimes(b, data, Ntimes)
		})

		// Check how fast is write to a file with random/uncompressible content
		data = testutil.CreateRandomDummyBuf(size, 1)
		label = fmt.Sprintf("%s/RandomContent_Size_%d", prefix, size)
		b.Run(label, func(b *testing.B) {
			writeDataNtimes(b, data, Ntimes)
		})
	}
}
