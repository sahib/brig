package parcello

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	zipexe "github.com/daaku/go.zipexe"
	"github.com/kardianos/osext"
)

var (
	// ErrReadOnly is returned if the file is read-only and write operations are disabled.
	ErrReadOnly = errors.New("File is read-only")
	// ErrWriteOnly is returned if the file is write-only and read operations are disabled.
	ErrWriteOnly = errors.New("File is write-only")
	// ErrIsDirectory is returned if the file under operation is not a regular file but a directory.
	ErrIsDirectory = errors.New("Is directory")
)

var (
	// Manager keeps track of all resources
	Manager = DefaultManager(osext.Executable)
	// Make sure the ResourceManager implements the FileSystemManager interface
	_ FileSystemManager = &ResourceManager{}
)

// Open opens an embedded resource for read
func Open(name string) (File, error) {
	return Manager.OpenFile(name, os.O_RDONLY, 0)
}

// ManagerAt returns manager at given path
func ManagerAt(path string) FileSystemManager {
	mngr, err := Manager.Dir(path)
	if err != nil {
		panic(err)
	}
	return mngr
}

// AddResource adds resource to the default resource manager
// Note that the method may panic if the resource not exists
func AddResource(resource []byte) {
	if err := Manager.Add(BinaryResource(resource)); err != nil {
		panic(err)
	}
}

// ResourceManagerConfig represents the configuration for Resource Manager
type ResourceManagerConfig struct {
	// Path to the archive
	Path string
	// FileSystem that stores the archive
	FileSystem FileSystem
}

// ResourceManager represents a virtual in memory file system
type ResourceManager struct {
	cfg  *ResourceManagerConfig
	rw   sync.RWMutex
	root *Node
	// NewReader creates a new ZIP Reader
	NewReader func(io.ReaderAt, int64) (*zip.Reader, error)
}

// DefaultManager creates a FileSystemManager based on whether dev mode is enabled
func DefaultManager(executable ExecutableFunc) FileSystemManager {
	mode := os.Getenv("PARCELLO_DEV_ENABLED")

	if mode != "" {
		return Dir(getenv("PARCELLO_RESOURCE_DIR", "."))
	}

	path, err := executable()
	if err != nil {
		panic(err)
	}

	dir, path := filepath.Split(path)

	cfg := &ResourceManagerConfig{
		Path:       path,
		FileSystem: Dir(dir),
	}

	manager, err := NewResourceManager(cfg)
	if err != nil {
		panic(err)
	}

	return manager
}

// NewResourceManager creates a new manager
func NewResourceManager(cfg *ResourceManagerConfig) (*ResourceManager, error) {
	manager := &ResourceManager{cfg: cfg}

	file, err := cfg.FileSystem.OpenFile(cfg.Path, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	resource := &Resource{
		Body: file,
		Size: info.Size(),
	}

	_ = manager.Add(resource)

	return manager, nil
}

// Add adds resource to the manager
func (m *ResourceManager) Add(resource *Resource) error {
	m.rw.Lock()
	defer m.rw.Unlock()

	if m.root == nil {
		m.root = &Node{Name: "/", IsDir: true}
	}

	newReader := zipexe.NewReader

	if m.NewReader != nil {
		newReader = m.NewReader
	}

	reader, err := newReader(resource.Body, resource.Size)
	if err != nil {
		return err
	}

	return m.uncompress(reader)
}

func (m *ResourceManager) uncompress(reader *zip.Reader) error {
	for _, header := range reader.File {
		path := split(header.Name)
		node := add(path, m.root)

		if node == m.root || node == nil {
			return fmt.Errorf("invalid path: '%s'", header.Name)
		}

		file, err := header.Open()
		if err != nil {
			return err
		}
		defer file.Close()

		content, err := ioutil.ReadAll(file)
		if err != nil {
			return err
		}

		node.IsDir = false
		node.Content = &content
	}

	return nil
}

// Dir returns a sub-manager for given path
func (m *ResourceManager) Dir(name string) (FileSystemManager, error) {
	if _, node := find(split(name), nil, m.root); node != nil {
		if node.IsDir {
			return &ResourceManager{root: node}, nil
		}
	}

	return nil, os.ErrNotExist
}

// Open opens an embedded resource for read
func (m *ResourceManager) Open(name string) (ReadOnlyFile, error) {
	return m.OpenFile(name, os.O_RDONLY, 0)
}

// OpenFile is the generalized open call; most users will use Open
func (m *ResourceManager) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	parent, node, err := m.open(name)
	if err != nil {
		return nil, err
	}

	if isWritable(flag) && node != nil && node.IsDir {
		return nil, &os.PathError{Op: "open", Path: name, Err: ErrIsDirectory}
	}

	if hasFlag(os.O_CREATE, flag) {
		if node != nil && !hasFlag(os.O_TRUNC, flag) {
			return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrExist}
		}

		node = newNode(filepath.Base(name), parent)
	}

	if node == nil {
		return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
	}

	return newFile(node, flag)
}

func (m *ResourceManager) open(name string) (*Node, *Node, error) {
	parent, node := find(split(name), nil, m.root)
	if node != m.root && parent == nil {
		return nil, nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
	}

	return parent, node, nil
}

// Walk walks the file tree rooted at root, calling walkFn for each file or
// directory in the tree, including root.
func (m *ResourceManager) Walk(dir string, fn filepath.WalkFunc) error {
	if _, node := find(split(dir), nil, m.root); node != nil {
		return walk(dir, node, fn)
	}

	return os.ErrNotExist
}

func add(path []string, node *Node) *Node {
	if !node.IsDir || node.Content != nil {
		return nil
	}

	if len(path) == 0 {
		return node
	}

	name := path[0]

	for _, child := range node.Children {
		if child.Name == name {
			return add(path[1:], child)
		}
	}

	child := &Node{
		Mutex:   &sync.RWMutex{},
		Name:    name,
		IsDir:   true,
		ModTime: time.Now(),
	}

	node.Children = append(node.Children, child)
	return add(path[1:], child)
}

func split(path string) []string {
	parts := []string{}

	for _, part := range strings.Split(path, string(os.PathSeparator)) {
		if part != "" && part != "/" {
			parts = append(parts, part)
		}
	}

	return parts
}

func find(path []string, parent, node *Node) (*Node, *Node) {
	if len(path) == 0 || node == nil {
		return parent, node
	}

	for _, child := range node.Children {
		if path[0] == child.Name {
			if len(path) == 1 {
				return node, child
			}
			return find(path[1:], node, child)
		}
	}

	return parent, nil
}

func walk(path string, node *Node, fn filepath.WalkFunc) error {
	if err := fn(path, &ResourceFileInfo{Node: node}, nil); err != nil {
		return err
	}

	for _, child := range node.Children {
		if err := walk(filepath.Join(path, child.Name), child, fn); err != nil {
			return err
		}
	}

	return nil
}

func newNode(name string, parent *Node) *Node {
	node := &Node{
		Name:    name,
		IsDir:   false,
		ModTime: time.Now(),
	}

	parent.Children = append(parent.Children, node)
	return node
}

func newFile(node *Node, flag int) (File, error) {
	if isWritable(flag) {
		node.ModTime = time.Now()
	}

	if node.Content == nil || hasFlag(os.O_TRUNC, flag) {
		buf := make([]byte, 0)
		node.Content = &buf
		node.Mutex = &sync.RWMutex{}
	}

	f := NewResourceFile(node)

	if hasFlag(os.O_APPEND, flag) {
		_, _ = f.Seek(0, io.SeekEnd)
	}

	if hasFlag(os.O_RDWR, flag) {
		return f, nil
	}
	if hasFlag(os.O_WRONLY, flag) {
		return &woFile{f}, nil
	}

	return &roFile{f}, nil
}

func hasFlag(flag int, flags int) bool {
	return flags&flag == flag
}

func isWritable(flag int) bool {
	return hasFlag(os.O_WRONLY, flag) || hasFlag(os.O_RDWR, flag) || hasFlag(os.O_APPEND, flag)
}

type roFile struct {
	*ResourceFile
}

// Write is disabled and returns ErrorReadOnly
func (f *roFile) Write(p []byte) (n int, err error) {
	return 0, ErrReadOnly
}

// woFile wraps the given file and disables Read(..) operation.
type woFile struct {
	*ResourceFile
}

// Read is disabled and returns ErrorWroteOnly
func (f *woFile) Read(p []byte) (n int, err error) {
	return 0, ErrWriteOnly
}
