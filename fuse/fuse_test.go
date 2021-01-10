// +build !windows

package fuse

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
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
	log.SetLevel(log.DebugLevel)
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
func callUnMount(t testing.TB, ctx context.Context, control *spawntest.Control) {
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
	defer callUnMount(t, ctx, control)

	// function which required mounts
	f(ctx, control, &mountInfo{
		Dir: req.MntPath,
		Opts: req.Opts,
	})
}

func checkFuseFileMatchToCatFS(t *testing.T, ctx context.Context, control *spawntest.Control, fusePath string, catfsPath string) {
	// checks if OS file content matches catFS file content
	fuseData, err := ioutil.ReadFile(fusePath)
	require.Nil(t, err)

	// is catFS seeing the same data
	checkCatfsFileContent(t, ctx, control, catfsPath, fuseData)
}

func checkCatfsFileContent(t *testing.T, ctx context.Context, control *spawntest.Control, catfsPath string, expected []byte) {
	req := catfsPayload{Path: catfsPath}
	out := catfsPayload{}
	require.Nil(t, control.JSON("/catfsGetData").Call(ctx, req, &out))
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
		require.Nil(t, control.JSON("/catfsStage").Call(ctx, req, &nothing{}))
	})
}

func TestCatfsGetData(t *testing.T) {
	withMount(t, MountOptions{}, func(ctx context.Context, control *spawntest.Control, mount *mountInfo) {
		dataIn := []byte{1, 2, 3, 4}
		filePath := "StageAndReadTest.bin"
		req := catfsPayload{Path: filePath, Data: dataIn}
		require.Nil(t, control.JSON("/catfsStage").Call(ctx, req, &nothing{}))

		req.Data = []byte{}
		out := catfsPayload{}
		require.Nil(t, control.JSON("/catfsGetData").Call(ctx, req, &out))
		require.Equal(t, out.Data, dataIn)
	})
}

// Main fuse layer tests

var (
	DataSizes = []int64{
		0, 1, 2, 4, 8, 16, 32, 64, 1024,
		2048, 4095, 4096, 4097, 147611,
		2*1024*1024+123, // in case if we have buffer size interference
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
				require.Nil(t, control.JSON("/catfsStage").Call(ctx, req, &nothing{}))
				checkCatfsFileContent(t, ctx, control, catfsFilePath, helloData)

				fuseFilePath := filepath.Join(mount.Dir, catfsFilePath)
				checkFuseFileMatchToCatFS(t, ctx, control, fuseFilePath, catfsFilePath)
			})
		}
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
				checkCatfsFileContent(t, ctx, control, catfsFilePath, helloData)
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
				require.Nil(t, control.JSON("/catfsStage").Call(ctx, req, &nothing{}))
				checkCatfsFileContent(t, ctx, control, catfsFilePath, req.Data)

				fuseFilePath := filepath.Join(mount.Dir, catfsFilePath)

				// Write a simple file via the fuse layer:
				helloData := testutil.CreateDummyBuf(size)
				err := ioutil.WriteFile(fuseFilePath, helloData, 0644)
				if err != nil {
					t.Fatalf("Could not write simple file via fuse layer: %v", err)
				}
				checkCatfsFileContent(t, ctx, control, catfsFilePath, helloData)
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

		require.Nil(t, os.Mkdir(fuseSubDirPath, 0644))

		expected := []byte{1, 2, 3}
		require.Nil(t, ioutil.WriteFile(fuseFilePath, expected, 0644))

		checkCatfsFileContent(t, ctx, control, catfsFilePath, expected)
	})
}

func TestReadOnlyFs(t *testing.T) {
	opts := MountOptions{
		ReadOnly: true,
	}
	withMount(t, opts, func(ctx context.Context, control *spawntest.Control, mount *mountInfo) {
		xData := []byte{1, 2, 3}
		req := catfsPayload{Path: "/x.png", Data: xData}
		require.Nil(t, control.JSON("/catfsStage").Call(ctx, req, &nothing{}))
		checkCatfsFileContent(t, ctx, control, "x.png", xData)

		// Do some allowed io to check if the fs is actually working.
		// The test does not check on the kind of errors otherwise.
		xPath := filepath.Join(mount.Dir, "x.png")
		checkFuseFileMatchToCatFS(t, ctx, control, xPath, "x.png")

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
		require.Nil(t, control.JSON("/catfsStage").Call(ctx, req, &nothing{}))
		checkCatfsFileContent(t, ctx, control, req.Path, data)
		checkFuseFileMatchToCatFS(t, ctx, control, filepath.Join(mount.Dir, req.Path), req.Path)

		data = []byte{2, 3, 4}
		req = catfsPayload{Path: "/a/x.png", Data: data}
		require.Nil(t, control.JSON("/catfsStage").Call(ctx, req, &nothing{}))
		checkCatfsFileContent(t, ctx, control, req.Path, data)
		checkFuseFileMatchToCatFS(t, ctx, control, filepath.Join(mount.Dir, req.Path), req.Path)

		data = []byte{3, 4, 5}
		req = catfsPayload{Path: "/a/b/y.png", Data: data}
		require.Nil(t, control.JSON("/catfsStage").Call(ctx, req, &nothing{}))
		checkCatfsFileContent(t, ctx, control, req.Path, data)
		checkFuseFileMatchToCatFS(t, ctx, control, filepath.Join(mount.Dir, req.Path), req.Path)

		data = []byte{4, 5, 6}
		req = catfsPayload{Path: "/a/b/c/z.png", Data: data}
		require.Nil(t, control.JSON("/catfsStage").Call(ctx, req, &nothing{}))
		checkCatfsFileContent(t, ctx, control, req.Path, data)
		checkFuseFileMatchToCatFS(t, ctx, control, filepath.Join(mount.Dir, req.Path), req.Path)

		// Now we need to remount fuse with different root directory
		remntReq := mountingRequest{
			MntPath: mount.Dir,
			Opts:    MountOptions{Root: "/a/b"},
		}
		require.Nil(t, control.JSON("/fuseReMount").Call(ctx, remntReq, &nothing{}))
		mount.Opts = remntReq.Opts // update with new mount options

		// See if fuse indeed provides different root
		// Read already existing file
		yPath := filepath.Join(mount.Dir, "y.png")
		checkFuseFileMatchToCatFS(t, ctx, control, yPath, "/a/b/y.png")

		// Write to a new file
		data = []byte{5, 6, 7}
		newPath := filepath.Join(mount.Dir, "new.png")
		require.Nil(t, ioutil.WriteFile(newPath, data, 0644))
		checkCatfsFileContent(t, ctx, control, "/a/b/new.png", data)

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
		1024, 2*1024, 16*1024, 64*1024, 128*1023,
		1*1024*1024, 16*1024*1024,
	}
)

func stageAndRead(b *testing.B, ctx context.Context, control *spawntest.Control, mount *mountInfo, label string, data []byte) {
	size := len(data)
	// stage data to catFS
	catfsFilePath := fmt.Sprintf("%s_file_%d", label, size)
	req := catfsPayload{Path: catfsFilePath, Data: data}
	require.Nil(b, control.JSON("/catfsStage").Call(ctx, req, &nothing{}))
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
			stageAndRead(b, ctx, control, mount,  "CompressibleContent", data)

			// Check how fast is readout of a file with random/uncompressible content
			data = testutil.CreateRandomDummyBuf(size, 1)
			stageAndRead(b, ctx, control, mount,  "RandomContent", data)
		}
	})
}

