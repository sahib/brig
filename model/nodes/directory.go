package nodes

import (
	"fmt"
	"path"
	"strings"
	"time"

	h "github.com/disorganizer/brig/util/hashlib"
	capnp "zombiezen.com/go/capnproto2"
)

type Directory struct {
	Base

	size     uint64
	parent   string
	children map[string]h.Hash
}

//
// // newEmptyDirectory creates a new empty directory that is not yet present
// // in the store. It should not be used directtly.
// func newEmptyDirectory(fs *FS, parent *Directory, name string) (*Directory, error) {
// 	code := goipfsutil.DefaultIpfsHash
// 	length := multihash.DefaultLengths[code]
//
// 	absPath := ""
// 	if parent != nil {
// 		absPath = path.Join(parent.Path(), name)
// 	}
//
// 	mh, err := multihash.Sum([]byte(absPath), code, length)
// 	if err != nil {
// 		// The programmer has likely fucked up:
// 		return nil, fmt.Errorf("Failed to calculate basic checksum of a string: %v", err)
// 	}
//
// 	id, err := fs.NextID()
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	newDir := &Directory{
// 		fs:       fs,
// 		id:       id,
// 		hash:     &h.Hash{mh},
// 		name:     name,
// 		children: make(map[string]*h.Hash),
// 	}
//
// 	if parent != nil {
// 		if err := parent.Add(newDir); err != nil {
// 			return nil, err
// 		}
// 	}
//
// 	return newDir, nil
// }

func (d *Directory) ToCapnp() (*capnp.Message, error) {
	// TODO: Implement.
	return nil, nil
}

func (d *Directory) FromCapnp(msg *capnp.Message) error {
	// TODO: Implement.
	return nil
}

////////////// NODE INTERFACE /////////////////

func (d *Directory) Name() string {
	return d.name
}

func (d *Directory) Size() uint64 {
	return d.size
}

func (d *Directory) Path() string {
	return prefixSlash(path.Join(d.parent, d.Base.name))
}

func (d *Directory) NChildren(lkr Linker) int {
	return len(d.children)
}

func (d *Directory) Child(lkr Linker, name string) (Node, error) {
	childHash, ok := d.children[name]
	if !ok {
		return nil, nil
	}

	return lkr.NodeByHash(childHash)
}

func (d *Directory) Parent(lkr Linker) (Node, error) {
	if d.parent == "" {
		return nil, nil
	}

	return lkr.LookupNode(d.parent)
}

func (d *Directory) SetParent(lkr Linker, nd Node) error {
	if d.Path() == "/" {
		return nil
	}

	if nd == nil {
		d.parent = ""
	} else {
		d.parent = nd.Path()
	}

	return nil
}

// ////////////// TREE MOVEMENT /////////////////

func (d *Directory) VisitChildren(lkr Linker, fn func(nd Node) error) error {
	for name, hash := range d.children {
		child, err := lkr.NodeByHash(hash)
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

func (d *Directory) Up(lkr Linker, visit func(par *Directory) error) error {
	root, err := lkr.Root()
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
			// This usually means that some link is missing.
			return fmt.Errorf("BUG: Cannot reach self from root in Up()")
		}

		childNode, err := lkr.NodeByHash(childHash)
		if err != nil {
			return err
		}

		child, ok := childNode.(*Directory)
		if !ok {
			return fmt.Errorf("BUG: Non-directory in Up(): %v", childHash)
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

func (d *Directory) IsRoot() bool {
	return d.parent == ""
}

func (d *Directory) xorHash(lkr Linker, hash h.Hash) error {
	oldHash := d.hash.Clone()
	if err := d.hash.Xor(hash); err != nil {
		return err
	}

	if d.IsRoot() {
		lkr.MemSetRoot(d)
	}

	lkr.MemIndexSwap(d, oldHash)
	return nil
}

func Walk(lkr Linker, node Node, dfs bool, visit func(child Node) error) error {
	if node == nil {
		return nil
	}

	if node.Type() != NodeTypeDirectory {
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

	for name, link := range d.children {
		child, err := lkr.NodeByHash(link)
		if err != nil {
			return err
		}

		if child == nil {
			return fmt.Errorf("Walk: could not resolve %s (%s)", name, link.B58String())
		}

		if err := Walk(lkr, child, dfs, visit); err != nil {
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

func (d *Directory) Lookup(lkr Linker, repoPath string) (Node, error) {
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
		curr, err = curr.Child(lkr, elem)
		if err != nil {
			return nil, err
		}

		if curr == nil {
			return nil, NoSuchFile(repoPath)
		}
	}

	return curr, nil
}

//////////// STATE ALTERING METHODS //////////////

func (d *Directory) SetSize(size uint64)          { d.size = size }
func (d *Directory) SetName(name string)          { d.name = name }
func (d *Directory) SetModTime(modTime time.Time) { d.Base.modTime = modTime }
func (d *Directory) SetHash(hash h.Hash)          { d.Base.hash = hash.Clone() }

func (d *Directory) Add(lkr Linker, nd Node) error {
	if nd == d {
		return fmt.Errorf("ADD-BUG: attempting to add `%s` to itself", nd.Path())
	}

	if _, ok := d.children[nd.Name()]; ok {
		return ErrExists
	}

	nodeSize := nd.Size()
	nodeHash := nd.Hash()

	err := d.Up(lkr, func(parent *Directory) error {
		parent.size += nodeSize
		return parent.xorHash(lkr, nodeHash)
	})

	if err != nil {
		return err
	}

	// Establish the link between parent and child:
	// (must be done last, because d's hash changed)
	nd.SetParent(lkr, d)
	d.children[nd.Name()] = nodeHash
	return nil
}

// RemoveChild removes the child named `name` from it's children.
//
// Note that there is no general Remove() function that works on itself.
// It is therefore not possible (or a good idea) to remove the root node.
func (d *Directory) RemoveChild(lkr Linker, nd Node) error {
	name := nd.Name()
	if _, ok := d.children[name]; !ok {
		return NoSuchFile(name)
	}

	// Unset parent from child:
	if err := nd.SetParent(lkr, nil); err != nil {
		return err
	}

	delete(d.children, name)

	nodeSize := nd.Size()
	nodeHash := nd.Hash()

	return d.Up(lkr, func(parent *Directory) error {
		parent.size -= nodeSize
		return parent.xorHash(lkr, nodeHash)
	})
}

// Assert that Directory follows the Node interface:
var _ SettableNode = &Directory{}
