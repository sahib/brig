package store

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/disorganizer/brig/store/wire"
	"github.com/gogo/protobuf/proto"
	goipfsutil "github.com/ipfs/go-ipfs-util"
	"github.com/jbenet/go-multihash"
)

type Directory struct {
	name     string
	size     uint64
	modTime  time.Time
	parent   string
	hash     *Hash
	children map[string]*Hash
	id       uint64

	// This is not set by FromProto() and must be passed
	// on creating by FS.
	fs *FS
}

// newEmptyDirectory creates a new empty directory that is not yet present
// in the store. It should not be used directtly.
func newEmptyDirectory(fs *FS, parent *Directory, name string) (*Directory, error) {
	code := goipfsutil.DefaultIpfsHash
	length := multihash.DefaultLengths[code]

	absPath := ""
	if parent != nil {
		absPath = path.Join(NodePath(parent), name)
	}

	mh, err := multihash.Sum([]byte(absPath), code, length)
	if err != nil {
		// The programmer has likely fucked up:
		return nil, fmt.Errorf("Failed to calculate basic checksum of a string: %v", err)
	}

	id, err := fs.NextID()
	if err != nil {
		return nil, err
	}

	newDir := &Directory{
		fs:       fs,
		id:       id,
		hash:     &Hash{mh},
		name:     name,
		children: make(map[string]*Hash),
	}

	if parent != nil {
		if err := parent.Add(newDir); err != nil {
			return nil, err
		}
	}

	return newDir, nil
}

////////////// MARSHALLING ////////////////

func (d *Directory) ToProto() (*wire.Node, error) {
	binModTime, err := d.modTime.MarshalBinary()
	if err != nil {
		return nil, err
	}

	binLinks := [][]byte{}
	binNames := []string{}

	for name, link := range d.children {
		binLinks = append(binLinks, link.Bytes())
		binNames = append(binNames, name)
	}

	return &wire.Node{
		ID:       d.id,
		Type:     wire.NodeType_DIRECTORY,
		ModTime:  binModTime,
		NodeSize: d.size,
		Hash:     d.hash.Bytes(),
		Name:     d.name,
		Directory: &wire.Directory{
			Parent: d.parent,
			Links:  binLinks,
			Names:  binNames,
		},
	}, nil
}

func (d *Directory) Marshal() ([]byte, error) {
	pbd, err := d.ToProto()
	if err != nil {
		return nil, err
	}

	return proto.Marshal(pbd)
}

func (d *Directory) FromProto(pnd *wire.Node) error {
	pbd := pnd.Directory

	modTime := time.Time{}
	if err := modTime.UnmarshalBinary(pnd.ModTime); err != nil {
		return err
	}

	d.id = pnd.ID
	d.modTime = modTime
	d.parent = pbd.Parent
	d.size = pnd.NodeSize
	d.hash = &Hash{pnd.Hash}
	d.name = pnd.Name
	d.children = make(map[string]*Hash)

	// Find our place in the world:
	links := pbd.Links
	for idx, name := range pbd.Names {
		// Be cautious, input might come from everywhere:
		if idx >= 0 && idx < len(links) {
			return fmt.Errorf("Malformed input: More or less names than links in `%s`", d.name)
		}

		d.children[name] = &Hash{links[idx]}
	}

	return nil
}

func (d *Directory) Unmarshal(data []byte) error {
	pbd := &wire.Node{}
	if err := proto.Unmarshal(data, pbd); err != nil {
		return err
	}

	return d.FromProto(pbd)
}

////////////// NODE INTERFACE /////////////////

func (d *Directory) Name() string {
	return d.name
}

func (d *Directory) Hash() *Hash {
	return d.hash
}

func (d *Directory) Size() uint64 {
	return d.size
}

func (d *Directory) Path() string {
	return prefixSlash(path.Join(d.parent, d.name))
}

func (d *Directory) ModTime() time.Time {
	return d.modTime
}

func (d *Directory) NChildren() int {
	return len(d.children)
}

func (d *Directory) Child(name string) (Node, error) {
	childHash, ok := d.children[name]
	if !ok {
		return nil, nil
	}

	return d.fs.NodeByHash(childHash)
}

func (d *Directory) Parent() (Node, error) {
	if d.parent == "" {
		return nil, nil
	}

	return d.fs.LookupNode(d.parent)
}

func (d *Directory) SetParent(nd Node) error {
	if nd == nil {
		d.parent = ""
	} else {
		d.parent = nd.Path()
	}
	return nil
}

func (d *Directory) GetType() NodeType {
	return NodeTypeDirectory
}

func (d *Directory) ID() uint64 {
	return d.id
}

////////////// TREE MOVEMENT /////////////////

func (d *Directory) VisitChildren(fn func(*Directory) error) error {
	for name, hash := range d.children {
		child, err := d.fs.DirectoryByHash(hash)
		if err != nil {
			return err
		}

		if child == nil {
			return fmt.Errorf("BUG: dead link in tree: %s => %s", name, hash.B58String())
		}

		if err := fn(child); err != nil {
			return err
		}
	}

	return nil
}

func (d *Directory) Up(visit func(par *Directory) error) error {
	root, err := d.fs.Root()
	if err != nil {
		return err
	}

	elems := strings.Split(d.Path(), "/")
	dirs := []*Directory{root}
	curr := root

	for _, elem := range elems {
		if elem == "" {
			continue
		}

		childHash, ok := curr.children[elem]
		if !ok {
			return fmt.Errorf("BUG: Cannot reach self from root in Up()")
		}

		child, err := d.fs.DirectoryByHash(childHash)
		if err != nil {
			return err
		}

		dirs = append(dirs, child)
		curr = child
	}

	// Visit the nodes in reverse order, self first, root last:
	for idx := len(dirs) - 1; idx >= 0; idx-- {
		if err := visit(dirs[idx]); err != nil {
			return err
		}
	}

	return nil
}

func (d *Directory) xorHash(hash *Hash) error {
	return d.hash.Xor(hash)
}

func Walk(node Node, dfs bool, visit func(child Node) error) error {
	if node == nil {
		return nil
	}

	if node.GetType() != NodeTypeDirectory {
		return visit(node)
	}

	d, ok := node.(*Directory)
	if !ok {
		return ErrBadNode
	}

	if !dfs {
		if err := visit(node); err != nil {
			return err
		}
	}

	for _, link := range d.children {
		child, err := d.fs.NodeByHash(link)
		if err != nil {
			return err
		}
		fmt.Println("Sub Walking", link, child)

		if err := Walk(child, dfs, visit); err != nil {
			return err
		}
	}

	if dfs {
		if err := visit(node); err != nil {
			return err
		}
	}

	return nil
}

func (d *Directory) Lookup(repoPath string) (Node, error) {
	repoPath = prefixSlash(path.Clean(repoPath))
	elems := strings.Split(repoPath, "/")

	// Strip off the first empty field:
	elems = elems[1:]

	if len(elems) == 1 && elems[0] == "" {
		return d, nil
	}

	var curr Node = d
	var err error

	for _, elem := range elems {
		curr, err = curr.Child(elem)
		if err != nil {
			return nil, err
		}

		if curr == nil {
			return nil, nil
		}
	}

	return curr, nil
}

//////////// STATE ALTERING METHODS //////////////

// TODO: Grafik dafür in der Masterarbeit machen!
func (d *Directory) Add(nd Node) error {
	if _, ok := d.children[nd.Name()]; ok {
		return ErrExists
	}

	nodeSize := nd.Size()
	nodeHash := nd.Hash()

	err := d.Up(func(parent *Directory) error {
		parent.size += nodeSize
		return parent.xorHash(nodeHash)
	})

	if err != nil {
		return err
	}

	// Establish the link between parent and child:
	// (must be done last, because d's hashed changed)
	nd.SetParent(d)
	d.children[nd.Name()] = nodeHash
	return nil
}

// RemoveChild removes the child named `name` from it's children.
//
// Note that there is no general Remove() function that works on itself.
// It is therefore not possible (or a good idea) to remove the root node.
func (d *Directory) RemoveChild(nd Node) error {
	name := nd.Name()
	if _, ok := d.children[name]; !ok {
		return NoSuchFile(name)
	}

	// Unset parent from child:
	if err := nd.SetParent(nil); err != nil {
		return err
	}

	delete(d.children, name)

	nodeSize := nd.Size()
	nodeHash := nd.Hash()

	return d.Up(func(parent *Directory) error {
		parent.size -= nodeSize
		return parent.xorHash(nodeHash)
	})
}
