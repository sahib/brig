package store

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/id"
	"github.com/disorganizer/brig/store/wire"
	"github.com/gogo/protobuf/proto"
)

var (
	// ErrNoChange means that nothing changed between two versions (of a file)
	ErrNoChange = fmt.Errorf("Nothing changed between the given versions")
)

const (
	// ChangeInvalid indicates a bug.
	ChangeInvalid = iota

	// ChangeAdd means the file was added (initially or after ChangeRemove)
	ChangeAdd

	// ChangeModify indicates a content modification.
	ChangeModify

	// ChangeMove indicates that a file's path changed.
	ChangeMove

	// ChangeRemove indicates that the file was deleted.
	// Old versions might still be accessible from the history.
	ChangeRemove
)

// ChangeType describes the nature of a change.
type ChangeType byte

var changeTypeToString = map[ChangeType]string{
	ChangeInvalid: "invalid",
	ChangeAdd:     "added",
	ChangeModify:  "modified",
	ChangeRemove:  "removed",
	ChangeMove:    "moved",
}

var stringToChangeType = map[string]ChangeType{
	"invalid":  ChangeInvalid,
	"added":    ChangeAdd,
	"modified": ChangeModify,
	"removed":  ChangeRemove,
	"moved":    ChangeMove,
}

// String formats a changetype to a human readable verb in past tense.
func (c *ChangeType) String() string {
	return changeTypeToString[*c]
}

func (c *ChangeType) Marshal() ([]byte, error) {
	dec, ok := changeTypeToString[*c]
	if !ok {
		return nil, fmt.Errorf("Bad change type `%d`", *c)
	}

	return []byte(dec), nil
}

func (c *ChangeType) Unmarshal(data []byte) error {
	var ok bool
	*c, ok = stringToChangeType[string(data)]
	if !ok {
		return fmt.Errorf("Unknown change type: %s", string(data))
	}

	return nil
}

// Checkpoint remembers one state of a single file.
type Checkpoint struct {
	// IDLink references the history of a single file
	idLink uint64

	// Hash is the hash of the file at this point.
	// It may, or may not be retrievable from ipfs.
	// For ChangeRemove the hash is the hash of the last existing file.
	hash *Hash

	// Index is a a unique counter on the number of checkpoints
	index uint64

	// Size is the size of the file in bytes at this point.
	// Change is the detailed type of the modification.
	change ChangeType

	// Author of the file modifications
	// TODO: use commit.Author?
	author id.ID
}

func (c *Checkpoint) ChangeType() *ChangeType { return &c.change }
func (c *Checkpoint) Hash() *Hash             { return c.hash }
func (c *Checkpoint) Author() id.ID           { return c.author }

// TODO: nice representation
func (c *Checkpoint) String() string {
	return fmt.Sprintf(
		"%x:%x@%s(%s)",
		c.idLink,
		c.index,
		c.change.String(),
		c.hash.B58String(),
	)
}

func newEmptyCheckpoint() *Checkpoint {
	// This is here to make sure api changes cause compile errors.
	return &Checkpoint{}
}

func (cp *Checkpoint) ToProto() (*wire.Checkpoint, error) {
	return &wire.Checkpoint{
		IdLink: proto.Uint64(cp.idLink),
		Hash:   cp.hash.Bytes(),
		Change: proto.Int32(int32(cp.change)),
		Author: proto.String(string(cp.author)),
		Index:  proto.Uint64(cp.index),
	}, nil
}

func (cp *Checkpoint) FromProto(msg *wire.Checkpoint) error {
	cp.hash = &Hash{msg.GetHash()}
	cp.change = ChangeType(msg.GetChange())
	cp.index = msg.GetIndex()

	ID, err := id.Cast(msg.GetAuthor())
	if err != nil {
		log.Warningf("Bad author-id `%s` in proto-checkpoint: %v", msg.GetAuthor(), err)
		return err
	}

	cp.author = ID
	return nil
}

func (cp *Checkpoint) Marshal() ([]byte, error) {
	protoCheck, err := cp.ToProto()
	if err != nil {
		return nil, err
	}

	protoData, err := proto.Marshal(protoCheck)
	if err != nil {
		return nil, err
	}

	return protoData, nil
}

func (cp *Checkpoint) Unmarshal(data []byte) error {
	protoCheck := &wire.Checkpoint{}
	if err := proto.Unmarshal(data, protoCheck); err != nil {
		return err
	}

	return cp.FromProto(protoCheck)
}

type CheckpointLink struct {
	IDLink uint64
	Index  uint64
}

func (cl *CheckpointLink) String() string {
	return fmt.Sprintf("%x:%x", cl.IDLink, cl.Index)
}

func (cl *CheckpointLink) FromProto(pcl *wire.CheckpointLink) error {
	cl.IDLink = pcl.GetIdLink()
	cl.Index = pcl.GetIndex()
	return nil
}

func (cl *CheckpointLink) ToProto() (*wire.CheckpointLink, error) {
	return &wire.CheckpointLink{
		IdLink: proto.Uint64(cl.IDLink),
		Index:  proto.Uint64(cl.Index),
	}, nil
}

func (cl *CheckpointLink) Resolve(fs *FS) (*Checkpoint, error) {
	// TODO: This is a bit inefficient. Just load a single checkpoint? :P
	hist, err := fs.History(cl.IDLink)
	if err != nil {
		return nil, err
	}

	ckp := hist.At(int(cl.Index))
	if ckp == nil {
		return nil, fmt.Errorf("Invalid checkpoint-link %s", cl.String())
	}

	return ckp, nil
}

////////////////////////////
// HISTORY IMPLEMENTATION //
////////////////////////////

// History remembers the changes made to a file.
// New changes get appended to the end.
type History []*Checkpoint

// Len conforming sort.Interface
func (hy *History) Len() int {
	return len(*hy)
}

func (hy *History) Less(i, j int) bool {
	return (*hy)[i].index < (*hy)[j].index
}

func (hy *History) Swap(i, j int) {
	(*hy)[i], (*hy)[j] = (*hy)[j], (*hy)[i]
}

func (hy *History) ToProto() (*wire.History, error) {
	protoHist := &wire.History{}

	for _, ck := range *hy {
		protoCheck, err := ck.ToProto()
		if err != nil {
			return nil, err
		}

		protoHist.Hist = append(protoHist.Hist, protoCheck)
	}

	return protoHist, nil
}

func (hy *History) Marshal() ([]byte, error) {
	protoHist, err := hy.ToProto()
	if err != nil {
		return nil, err
	}

	data, err := proto.Marshal(protoHist)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (hy *History) FromProto(protoHist *wire.History) error {
	for _, protoCheck := range protoHist.Hist {
		ck := &Checkpoint{}
		if err := ck.FromProto(protoCheck); err != nil {
			return err
		}

		*hy = append(*hy, ck)
	}

	return nil
}

func (hy *History) Unmarshal(data []byte) error {
	protoHist := &wire.History{}

	if err := proto.Unmarshal(data, protoHist); err != nil {
		return err
	}

	return hy.FromProto(protoHist)
}

// At is like the normal array subscription, but does not crash when
// getting passed an invalid index. If the index is invalid, nil is returned.
func (hy *History) At(index int) *Checkpoint {
	if index < 0 || index >= len(*hy) {
		return nil
	}

	return (*hy)[index]
}

func (ckp *Checkpoint) MakeLink() *CheckpointLink {
	return &CheckpointLink{
		IDLink: ckp.index,
		Index:  ckp.index,
	}
}

func (ckp *Checkpoint) Fork(author id.ID, oldHash, newHash *Hash, oldPath, newPath string) (*Checkpoint, error) {
	var change ChangeType
	var hash *Hash

	if oldHash == nil {
		change, hash = ChangeAdd, newHash
	} else if newHash == nil {
		change, hash = ChangeRemove, oldHash
	} else if newHash.Equal(oldHash) == false {
		change, hash = ChangeModify, newHash
	} else if oldPath != newPath {
		change, hash = ChangeMove, newHash
	} else {
		return nil, ErrNoChange
	}

	var idLink uint64
	var index uint64

	if ckp != nil {
		idLink = ckp.idLink
		index = ckp.index
	}

	return &Checkpoint{
		idLink: idLink,
		index:  index + 1,
		hash:   hash,
		change: change,
		author: author,
	}, nil
}
