package parcello

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/blang/vfs/memfs"
)

//go:generate counterfeiter -fake-name FileSystem -o ./fake/FileSystem.go . FileSystem
//go:generate counterfeiter -fake-name FileSystemManager -o ./fake/FileSystemManager.go . FileSystemManager
//go:generate counterfeiter -fake-name File -o ./fake/File.go . File
//go:generate counterfeiter -fake-name Composer -o ./fake/Composer.go . Composer
//go:generate counterfeiter -fake-name Compressor -o ./fake/Compressor.go . Compressor

// FileSystem provides primitives to work with the underlying file system
type FileSystem interface {
	// A FileSystem implements access to a collection of named files.
	http.FileSystem
	// Walk walks the file tree rooted at root, calling walkFn for each file or
	// directory in the tree, including root.
	Walk(dir string, fn filepath.WalkFunc) error
	// OpenFile is the generalized open call; most users will use Open
	OpenFile(name string, flag int, perm os.FileMode) (File, error)
}

// FileSystemManager is a file system that can create sub-file-systems
type FileSystemManager interface {
	// FileSystem is the underlying file system
	FileSystem
	// Dir returns a sub-file-system
	Dir(name string) (FileSystemManager, error)
	// Add resource bundle to the manager
	Add(resource *Resource) error
}

// Resource represents a resource
type Resource struct {
	// Body of the resource
	Body io.ReaderAt
	// Size of the body
	Size int64
}

// BinaryResource creates a binary resource
func BinaryResource(data []byte) *Resource {
	return &Resource{
		Body: bytes.NewReader(data),
		Size: int64(len(data)),
	}
}

// ReadOnlyFile is the bundle file
type ReadOnlyFile = http.File

// File is the bundle file
type File interface {
	// Close() error
	// Read(p []byte) (n int, err error)
	// Seek(offset int64, whence int) (int64, error)
	// Readdir(count int) ([]os.FileInfo, error)
	// Stat() (os.FileInfo, error)
	// Write(p []byte) (n int, err error)
	// ReadAt(p []byte, off int64) (n int, err error)

	// A File is returned by a FileSystem's Open method and can be
	ReadOnlyFile
	// Writer is the interface that wraps the basic Write method.
	io.Writer
	// ReaderAt reads at specific position
	io.ReaderAt
}

// Composer composes the resources
type Composer interface {
	// Compose composes from an archive
	Compose(bundle *Bundle) error
}

// CompressorContext used for the compression
type CompressorContext struct {
	// FileSystem file system that contain the files which will be compressed
	FileSystem FileSystem
	// Offset that should be applied
	Offset int64
}

// Compressor compresses given resource
type Compressor interface {
	// Compress compresses given source
	Compress(ctx *CompressorContext) (*Bundle, error)
}

// Bundle represents a bundled resource
type Bundle struct {
	// Name of the resource
	Name string
	// Count returns the count of files in the bundle
	Count int
	// Body of the resource
	Body []byte
}

// Node represents a node in resource tree
type Node struct {
	// Name of the node
	Name string
	// IsDir returns true if the node is directory
	IsDir bool
	// Mutext keeps the node thread safe
	Mutex *sync.RWMutex
	// ModTime returns the last modified time
	ModTime time.Time
	// Content of the node
	Content *[]byte
	// Children of the node
	Children []*Node
}

var _ os.FileInfo = &ResourceFileInfo{}

// ResourceFileInfo represents a hierarchy node in the resource manager
type ResourceFileInfo struct {
	Node *Node
}

// Name returns the base name of the file
func (n *ResourceFileInfo) Name() string {
	return n.Node.Name
}

// Size returns the length in bytes for regular files
func (n *ResourceFileInfo) Size() int64 {
	if n.Node.IsDir {
		return 0
	}

	n.Node.Mutex.RLock()
	defer n.Node.Mutex.RUnlock()
	l := len(*(n.Node.Content))
	return int64(l)
}

// Mode returns the file mode bits
func (n *ResourceFileInfo) Mode() os.FileMode {
	return 0
}

// ModTime returns the modification time
func (n *ResourceFileInfo) ModTime() time.Time {
	return n.Node.ModTime
}

// IsDir returns true if the node is directory
func (n *ResourceFileInfo) IsDir() bool {
	return n.Node.IsDir
}

// Sys returns the underlying data source
func (n *ResourceFileInfo) Sys() interface{} {
	return nil
}

var _ File = &ResourceFile{}

// ResourceFile represents a *bytes.Buffer that can be closed
type ResourceFile struct {
	*memfs.MemFile
	node *Node
}

// NewResourceFile creates a new Buffer
func NewResourceFile(node *Node) *ResourceFile {
	return &ResourceFile{
		MemFile: memfs.NewMemFile(node.Name, node.Mutex, node.Content),
		node:    node,
	}
}

// Readdir reads the contents of the directory associated with file and
// returns a slice of up to n FileInfo values, as would be returned
func (b *ResourceFile) Readdir(n int) ([]os.FileInfo, error) {
	info := []os.FileInfo{}

	if !b.node.IsDir {
		return info, fmt.Errorf("Not supported")
	}

	for index, node := range b.node.Children {
		if index >= n && n > 0 {
			break
		}

		info = append(info, &ResourceFileInfo{Node: node})
	}

	return info, nil
}

// Stat returns the FileInfo structure describing file.
// If there is an error, it will be of type *PathError.
func (b *ResourceFile) Stat() (os.FileInfo, error) {
	return &ResourceFileInfo{Node: b.node}, nil
}

// ExecutableFunc returns the executable path
type ExecutableFunc func() (string, error)
