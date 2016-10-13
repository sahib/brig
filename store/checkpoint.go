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
func (c ChangeType) String() string {
	return changeTypeToString[c]
}

func (c ChangeType) Marshal() ([]byte, error) {
	dec, ok := changeTypeToString[c]
	if !ok {
		return nil, fmt.Errorf("Bad change type `%d`", c)
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

func newEmptyCheckpoint(ID uint64, hash *Hash, author id.ID) *Checkpoint {
	return &Checkpoint{
		idLink: ID,
		hash:   hash,
		index:  0,
		change: ChangeAdd,
		author: author,
	}
}

func (cp *Checkpoint) ChangeType() ChangeType { return cp.change }
func (cp *Checkpoint) Hash() *Hash            { return cp.hash }
func (cp *Checkpoint) Author() id.ID          { return cp.author }

// TODO: nice representation
func (cp *Checkpoint) String() string {
	return fmt.Sprintf(
		"%x:%x@%s(%s)",
		cp.idLink,
		cp.index,
		cp.change.String(),
		cp.hash.B58String(),
	)
}

func (cp *Checkpoint) ToProto() (*wire.Checkpoint, error) {
	return &wire.Checkpoint{
		IdLink: cp.idLink,
		Hash:   cp.hash.Bytes(),
		Change: int32(cp.change),
		Author: string(cp.author),
		Index:  cp.index,
	}, nil
}

func (cp *Checkpoint) FromProto(msg *wire.Checkpoint) error {
	cp.hash = &Hash{msg.Hash}
	cp.change = ChangeType(msg.Change)
	cp.idLink = msg.IdLink
	cp.index = msg.Index

	ID, err := id.Cast(msg.Author)
	if err != nil {
		log.Warningf("Bad author-id `%s` in proto-checkpoint: %v", msg.Author, err)
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
	cl.IDLink = pcl.IdLink
	cl.Index = pcl.Index
	return nil
}

func (cl *CheckpointLink) ToProto() (*wire.CheckpointLink, error) {
	return &wire.CheckpointLink{
		IdLink: cl.IDLink,
		Index:  cl.Index,
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

func (hy *History) Equal(hb *History) bool {
	// TODO
	return false
}

func (hy *History) MostCurrentPath(fs *FS) string {
	if len(*hy) == 0 {
		return ""
	}

	// id := (*hy)[len(*hy)-1].idLink
	// TODO: resolve id to file, return file.path?
	return ""
}

func (hy *History) AllPaths() []string {
	// TODO
	return nil
}

func (hy *History) IsPrefix(hb *History) bool {
	// TODO
	return false
}

func (hy *History) CommonRoot(hb *History) int {
	// TODO
	return -1
}

func (hy *History) ConflictingChanges(hb *History, since int) error {
	// TODO? return?
	return nil
}

//////////////////////////

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
	} else if !newHash.Equal(oldHash) {
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

	newIndex := index
	if ckp.ChangeType() != change {
		newIndex += 1
	}

	return &Checkpoint{
		idLink: idLink,
		index:  newIndex,
		hash:   hash,
		change: change,
		author: author,
	}, nil
}
