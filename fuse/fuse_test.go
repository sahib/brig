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

type brigHelp struct{}

// These helpers will be requested from test and executed on the server
// which is managing brig-fuse connection (started within test)
func (bmh *brigHelp) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.URL.Path {
	case "/mount":
		httpjson.ServePOST(bmh.makeBrigAndFuseMount).ServeHTTP(w, req)
	case "/unmount":
		httpjson.ServePOST(bmh.unmountFuseAndcloseDummyBrigFS).ServeHTTP(w, req)
	case "/fuseReMount":
		httpjson.ServePOST(bmh.makeFuseReMount).ServeHTTP(w, req)
	case "/brigStage":
		httpjson.ServePOST(bmh.brigStage).ServeHTTP(w, req)
	case "/brigGetData":
		httpjson.ServePOST(bmh.brigGetData).ServeHTTP(w, req)
	default:
		http.NotFound(w, req)
	}
}

func makeDummyBrigFS(dbPath string) (brigFuseInfo, error) {
	backend := catfs.NewMemFsBackend()
	owner := "alice"

	cfg, err := config.Open(nil, defaults.Defaults, config.StrictnessPanic)
	if err != nil {
		log.Fatalf("Could not get default brig FS config: %v", err)
		return brigFuseInfo{}, err
	}

	bfs, err := catfs.NewFilesystem(backend, dbPath, owner, false, cfg.Section("fs"))
	if err != nil {
		log.Fatalf("Failed to create brig filesystem: %v", err)
		return brigFuseInfo{}, err
	}
	bfInfo := brigFuseInfo{}
	bfInfo.bfs = bfs
	bfInfo.dbPath = dbPath

	return bfInfo, err
}

type nothing struct{} // use it to send empty request or responses to server

type brigFuseInfo struct {
	bfs       *catfs.FS
	dbPath    string
	fuseMount *Mount
}

// bfInfo will be in the global space for the server
// which manage fuse mount connection to the brig FS
var bfInfo brigFuseInfo

type mountingRequest struct {
	DbPath  string
	MntPath string
	Opts    MountOptions
}

func (bmh *brigHelp) makeBrigAndFuseMount(ctx context.Context, req mountingRequest) (*nothing, error) {
	var err error
	bfInfo, err = makeDummyBrigFS(req.DbPath)
	if err != nil {
		log.Errorf("cannot comake brig file system in %v", bfInfo.dbPath)
		return &nothing{}, err
	}

	fuseMount, err := makeFuseMount(bfInfo.bfs, req.MntPath, req.Opts)
	if err != nil {
		log.Errorf("cannot mount brig fuse file system to %v", req.MntPath)
		return &nothing{}, err
	}
	bfInfo.fuseMount = fuseMount
	return &nothing{}, err
}

func (bmh *brigHelp) makeFuseReMount(ctx context.Context, req mountingRequest) (*nothing, error) {
	fuseMount, err := makeFuseMount(bfInfo.bfs, req.MntPath, req.Opts)
	if err != nil {
		log.Errorf("cannot mount brig fuse file system to %v", req.MntPath)
		return &nothing{}, err
	}
	bfInfo.fuseMount = fuseMount
	return &nothing{}, err
}

func (bmh *brigHelp) unmountFuseAndcloseDummyBrigFS(ctx context.Context, req nothing) (*nothing, error) {
	defer os.RemoveAll(bfInfo.fuseMount.Dir)
	defer os.RemoveAll(bfInfo.dbPath)
	// first unmount fuse directory
	if err := lazyUnmount(bfInfo.fuseMount.Dir); err != nil {
		skipableErr := "exit status 1: fusermount: entry for " + bfInfo.fuseMount.Dir + " not found in /etc/mtab"
		log.Debug(skipableErr)
		if err.Error() != skipableErr {
			return &nothing{}, err
		}
	}

	// now close brig FS
	err := bfInfo.bfs.Close()
	if err != nil {
		log.Fatalf("Could not close brig filesystem: %v", err)
	}
	return &nothing{}, err
}

func makeFuseMount(bfs *catfs.FS, mntPath string, opts MountOptions) (*Mount, error) {
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

	mount, err := NewMount(bfs, mntPath, nil, opts)
	if err != nil {
		log.Fatalf("Cannot create mount: %v", err)
		return nil, err
	}
	return mount, err
}

type brigPayload struct {
	Path string
	Data []byte
}

func (bmh *brigHelp) brigStage(ctx context.Context, req brigPayload) (*nothing, error) {
	err := bfInfo.bfs.Stage(req.Path, bytes.NewReader(req.Data))
	return &nothing{}, err
}

// Get data from a file stored by brig fs
func (bmh *brigHelp) brigGetData(ctx context.Context, req brigPayload) (*brigPayload, error) {
	out := brigPayload{}
	out.Path = req.Path

	stream, err := bfInfo.bfs.Cat(req.Path)
	if err != nil {
		log.Fatalf("Could not get stream for a brig file: %v", err)
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
var brigHelper = helpers.Register("brigHelp", &brigHelp{})

type mountInfo struct { // fuse related info available to OS layer
	Dir  string
	Opts MountOptions
}

func withMount(t *testing.T, opts MountOptions, f func(ctx context.Context, control *spawntest.Control, mount *mountInfo)) {
	// set up mounts
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	control := brigHelper.Spawn(ctx, t)
	defer control.Close()

	dbPath, err := ioutil.TempDir("", "brig-fs-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir for brig file system: %v", err)
	}

	req := mountingRequest{
		DbPath:  dbPath,
		MntPath: filepath.Join(os.TempDir(), "brig-fuse-mountdir"),
		Opts:    opts,
	}

	if err := control.JSON("/mount").Call(ctx, req, &nothing{}); err != nil {
		t.Fatalf("calling helper: %v", err)
	}

	mount := mountInfo{Dir: req.MntPath, Opts: req.Opts}
	// function which required mounts
	f(ctx, control, &mount)

	// cleanup and unmount
	if err := control.JSON("/unmount").Call(ctx, nothing{}, &nothing{}); err != nil {
		t.Fatalf("calling helper: %v", err)
	}
}

func checkFuseFileMatcheToBrig(t *testing.T, ctx context.Context, control *spawntest.Control, fusePath string, brigPath string) {
	// checks if OS file content matches brig FS file content
	fuseData, err := ioutil.ReadFile(fusePath)
	require.Nil(t, err)

	// is brig seeing the same data
	req := brigPayload{Path: brigPath}
	out := brigPayload{}
	require.Nil(t, control.JSON("/brigGetData").Call(ctx, req, &out))
	require.Equal(t, len(out.Data), len(fuseData))
	if out.Data == nil {
		// this is special for the 0 length data
		out.Data = []byte{}
	}
	require.Equal(t, out.Data, fuseData)
}

// Finally we ready to do tests

// Tests for spawntest infrastructure related tests

// Just checks that our brigStage interface to brig FS does not error out
func TestBrigStage(t *testing.T) {
	withMount(t, MountOptions{}, func(ctx context.Context, control *spawntest.Control, mount *mountInfo) {
		dataIn := []byte{1, 2, 3, 4}
		filePath := "StagingTest.bin"
		req := brigPayload{Path: filePath, Data: dataIn}
		require.Nil(t, control.JSON("/brigStage").Call(ctx, req, &nothing{}))
	})
}

func TestBrigGetData(t *testing.T) {
	withMount(t, MountOptions{}, func(ctx context.Context, control *spawntest.Control, mount *mountInfo) {
		dataIn := []byte{1, 2, 3, 4}
		filePath := "StageAndReadTest.bin"
		req := brigPayload{Path: filePath, Data: dataIn}
		require.Nil(t, control.JSON("/brigStage").Call(ctx, req, &nothing{}))

		req.Data = []byte{}
		out := brigPayload{}
		require.Nil(t, control.JSON("/brigGetData").Call(ctx, req, &out))
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
				brigFilePath := fmt.Sprintf("/hello_from_brig_%d", size)
				req := brigPayload{Path: brigFilePath, Data: helloData}
				require.Nil(t, control.JSON("/brigStage").Call(ctx, req, &nothing{}))
				fuseFilePath := filepath.Join(mount.Dir, brigFilePath)
				checkFuseFileMatcheToBrig(t, ctx, control, fuseFilePath, brigFilePath)
			})
		}
	})
}

func TestWrite(t *testing.T) {
	withMount(t, MountOptions{}, func(ctx context.Context, control *spawntest.Control, mount *mountInfo) {
		for _, size := range DataSizes {
			t.Run(fmt.Sprintf("%d", size), func(t *testing.T) {
				helloData := testutil.CreateDummyBuf(size)

				brigFilePath := fmt.Sprintf("/hello_from_fuse%d", size)
				fuseFilePath := filepath.Join(mount.Dir, brigFilePath)

				// Write a simple file via the fuse layer:
				err := ioutil.WriteFile(fuseFilePath, helloData, 0644)
				if err != nil {
					t.Fatalf("Could not write simple file via fuse layer: %v", err)
				}
				checkFuseFileMatcheToBrig(t, ctx, control, fuseFilePath, brigFilePath)
			})
		}
	})
}

// Regression test for copying larger file to the mount.
func TestTouchWrite(t *testing.T) {
	withMount(t, MountOptions{}, func(ctx context.Context, control *spawntest.Control, mount *mountInfo) {
		for _, size := range DataSizes {
			t.Run(fmt.Sprintf("%d", size), func(t *testing.T) {

				brigFilePath := fmt.Sprintf("/emty_at_creation_by_brig_%d", size)
				req := brigPayload{Path: brigFilePath, Data: []byte{}}
				require.Nil(t, control.JSON("/brigStage").Call(ctx, req, &nothing{}))

				fuseFilePath := filepath.Join(mount.Dir, brigFilePath)

				// Write a simple file via the fuse layer:
				helloData := testutil.CreateDummyBuf(size)
				err := ioutil.WriteFile(fuseFilePath, helloData, 0644)
				if err != nil {
					t.Fatalf("Could not write simple file via fuse layer: %v", err)
				}
				checkFuseFileMatcheToBrig(t, ctx, control, fuseFilePath, brigFilePath)
			})
		}
	})
}

// Regression test for copying a file to a subdirectory.
func TestTouchWriteSubdir(t *testing.T) {
	withMount(t, MountOptions{}, func(ctx context.Context, control *spawntest.Control, mount *mountInfo) {
		file := "donald.png"
		subDirPath := "sub"
		brigFilePath := filepath.Join(subDirPath, file)

		fuseSubDirPath := filepath.Join(mount.Dir, subDirPath)
		fuseFilePath := filepath.Join(fuseSubDirPath, file)

		require.Nil(t, os.Mkdir(fuseSubDirPath, 0644))

		expected := []byte{1, 2, 3}
		require.Nil(t, ioutil.WriteFile(fuseFilePath, expected, 0644))

		checkFuseFileMatcheToBrig(t, ctx, control, fuseFilePath, brigFilePath)
	})
}

func TestReadOnlyFs(t *testing.T) {
	opts := MountOptions{
		ReadOnly: true,
	}
	withMount(t, opts, func(ctx context.Context, control *spawntest.Control, mount *mountInfo) {
		xData := []byte{1, 2, 3}
		req := brigPayload{Path: "/x.png", Data: xData}
		require.Nil(t, control.JSON("/brigStage").Call(ctx, req, &nothing{}))

		// Do some allowed io to check if the fs is actually working.
		// The test does not check on the kind of errors otherwise.
		xPath := filepath.Join(mount.Dir, "x.png")
		data, err := ioutil.ReadFile(xPath)
		require.Nil(t, err)
		require.Equal(t, data, xData)
		checkFuseFileMatcheToBrig(t, ctx, control, xPath, "x.png")

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
		// Populate brig FS with some files in different directories
		req := brigPayload{Path: "/u.png", Data: []byte{1, 2, 3}}
		require.Nil(t, control.JSON("/brigStage").Call(ctx, req, &nothing{}))
		checkFuseFileMatcheToBrig(t, ctx, control, filepath.Join(mount.Dir, req.Path), req.Path)

		req = brigPayload{Path: "/a/x.png", Data: []byte{2, 3, 4}}
		require.Nil(t, control.JSON("/brigStage").Call(ctx, req, &nothing{}))
		checkFuseFileMatcheToBrig(t, ctx, control, filepath.Join(mount.Dir, req.Path), req.Path)

		req = brigPayload{Path: "/a/b/y.png", Data: []byte{3, 4, 5}}
		require.Nil(t, control.JSON("/brigStage").Call(ctx, req, &nothing{}))
		checkFuseFileMatcheToBrig(t, ctx, control, filepath.Join(mount.Dir, req.Path), req.Path)

		req = brigPayload{Path: "/a/b/c/z.png", Data: []byte{4, 5, 6}}
		require.Nil(t, control.JSON("/brigStage").Call(ctx, req, &nothing{}))
		checkFuseFileMatcheToBrig(t, ctx, control, filepath.Join(mount.Dir, req.Path), req.Path)

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
		checkFuseFileMatcheToBrig(t, ctx, control, yPath, "/a/b/y.png")

		// Write to a new file
		newPath := filepath.Join(mount.Dir, "new.png")
		require.Nil(t, ioutil.WriteFile(newPath, []byte{5, 6, 7}, 0644))
		checkFuseFileMatcheToBrig(t, ctx, control, newPath, "/a/b/new.png")

		// Attempt to read file above mounted root
		inAccessiblePath := filepath.Join(mount.Dir, "u.png")
		_, err := ioutil.ReadFile(inAccessiblePath)
		require.NotNil(t, err)
	})
}
