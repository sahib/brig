package store

import (
	"encoding/binary"
	"fmt"
	"sort"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
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

func (c *ChangeType) Unmarshal(data []byte) error {
	var ok bool
	*c, ok = stringToChangeType[string(data)]
	if !ok {
		return fmt.Errorf("Bad change type: %v", string(data))
	}

	return nil
}

// Checkpoint remembers one state of a single file.
type Checkpoint struct {
	// Hash is the hash of the file at this point.
	// It may, or may not be retrievable from ipfs.
	// For ChangeRemove the hash is the hash of the last existing file.
	Hash *Hash

	// Index is a a unique counter on the number of checkpoints
	Index uint64

	// ModTime is the time the checkpoint was made.
	ModTime time.Time

	// Size is the size of the file in bytes at this point.
	Size int64

	// Change is the detailed type of the modification.
	Change ChangeType

	// Author of the file modifications (jabber id)
	Author id.ID

	// Path of the file:
	//   - if added/modified: the current file path.
	//   - if removed: the old file path.
	//   - if moved: The new file path.
	Path string

	// OldPath is the path of the file before moving (for ChangeMove only)
	OldPath string
}

// TODO: nice representation
func (c *Checkpoint) String() string {
	return fmt.Sprintf("%-7s %+7s@%s", c.Change.String(), c.Hash.B58String(), c.ModTime.String())
}

// Checkpoints is a list of checkpoints.
// It is used to enable sorting by path.
type Checkpoints []*Checkpoint

func (cps *Checkpoints) Len() int {
	return len(*cps)
}

func (cps *Checkpoints) Less(i, j int) bool {
	return (*cps)[i].Path < (*cps)[j].Path
}

func (cps *Checkpoints) Swap(i, j int) {
	(*cps)[i], (*cps)[j] = (*cps)[j], (*cps)[i]
}

func (cp *Checkpoint) ToProto() (*wire.Checkpoint, error) {
	mtimeBin, err := cp.ModTime.MarshalBinary()
	if err != nil {
		return nil, err
	}

	protoCheck := &wire.Checkpoint{
		Hash:     cp.Hash.Bytes(),
		ModTime:  mtimeBin,
		FileSize: proto.Int64(cp.Size),
		Change:   proto.Int32(int32(cp.Change)),
		Author:   proto.String(string(cp.Author)),
		Path:     proto.String(cp.Path),
		OldPath:  proto.String(cp.OldPath),
		Index:    proto.Uint64(cp.Index),
	}

	if err != nil {
		return nil, err
	}

	return protoCheck, nil
}

// TODO: consistent UnmarshalProto/MarshalProto functions.

func (cp *Checkpoint) FromProto(msg *wire.Checkpoint) error {
	modTime := time.Time{}
	if err := modTime.UnmarshalBinary(msg.GetModTime()); err != nil {
		return err
	}

	cp.Hash = &Hash{msg.GetHash()}
	cp.ModTime = modTime
	cp.Size = msg.GetFileSize()
	cp.Change = ChangeType(msg.GetChange())
	cp.Path = msg.GetPath()
	cp.OldPath = msg.GetOldPath()
	cp.Index = msg.GetIndex()

	ID, err := id.Cast(msg.GetAuthor())
	if err != nil {
		log.Warningf("Bad author-id `%s` in proto-checkpoint: %v", msg.GetAuthor(), err)
	} else {
		cp.Author = ID
	}

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

func (cs *Commits) UnmarshalProto(data []byte) error {
	protoCmts := &wire.Commits{}
	if err := proto.Unmarshal(data, protoCmts); err != nil {
		return err
	}

	return cs.FromProto(protoCmts)
}

func (cs *Commits) FromProto(protoCmts *wire.Commits) error {
	for _, protoCmt := range protoCmts.GetCommits() {
		cmt := NewEmptyCommit(nil, "")
		if err := cmt.FromProto(protoCmt); err != nil {
			return err
		}

		*cs = append(*cs, cmt)
	}

	return nil
}

// History remembers the changes made to a file.
// New changes get appended to the end.
type History []*Checkpoint

// Len conforming sort.Interface
func (hy *History) Len() int {
	return len(*hy)
}

func (hy *History) Less(i, j int) bool {
	return (*hy)[i].Index < (*hy)[j].Index
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

// MakeCheckpoint creates a new checkpoint in the version history of `curr`.
// One of old or curr may be nil (if no old version exists or new version
// does not exist anymore). It is an error to pass nil twice.
//
// If nothing changed between old and curr, ErrNoChange is returned.
func (st *Store) MakeCheckpoint(old, curr *Metadata, oldPath, currPath string) error {
	var change ChangeType
	var hash *Hash
	var path string
	var size int64

	if old == nil {
		change, path, hash, size = ChangeAdd, currPath, curr.hash, curr.size
	} else if curr == nil {
		change, path, hash, size = ChangeRemove, oldPath, old.hash, old.size
	} else if !curr.hash.Equal(old.hash) {
		change, path, hash, size = ChangeModify, currPath, curr.hash, curr.size
	} else if oldPath != currPath {
		change, path, hash, size = ChangeMove, currPath, curr.hash, curr.size
	} else {
		return ErrNoChange
	}

	checkpoint := &Checkpoint{
		Hash:    hash,
		ModTime: time.Now(),
		Size:    size,
		Change:  change,
		Author:  st.ID,
		Path:    path,
	}

	protoData, err := checkpoint.Marshal()
	if err != nil {
		return err
	}

	dbErr := st.updateWithBucket("checkpoints", func(tx *bolt.Tx, bckt *bolt.Bucket) error {
		histBuck, err := bckt.CreateBucketIfNotExists([]byte(path))
		if err != nil {
			return err
		}

		// On a "move" we need to move the old data to the new path.
		if change == ChangeMove {
			checkpoint.OldPath = oldPath

			if oldBuck := bckt.Bucket([]byte(oldPath)); oldBuck != nil {
				err = oldBuck.ForEach(func(k, v []byte) error {
					return histBuck.Put(k, v)
				})

				if err != nil {
					return err
				}

				if err := bckt.DeleteBucket([]byte(oldPath)); err != nil {
					return err
				}
			}
		}

		key := make([]byte, 8)
		binary.LittleEndian.PutUint64(key, checkpoint.Index)
		return histBuck.Put(key, protoData)
	})

	if dbErr != nil {
		return dbErr
	}

	dbErr = st.updateWithBucket("stage", func(tx *bolt.Tx, bckt *bolt.Bucket) error {
		return bckt.Put([]byte(path), protoData)
	})

	if dbErr != nil {
		return dbErr
	}

	log.Debugf("created check point: %v", checkpoint)
	return nil
}

// History returns all checkpoints a file has.
// Note: even on error a empty history is returned.
func (s *Store) History(path string) (*History, error) {
	var hist History

	return &hist, s.viewWithBucket("checkpoints", func(tx *bolt.Tx, bckt *bolt.Bucket) error {
		changeBuck := bckt.Bucket([]byte(path))
		if changeBuck == nil {
			return NoSuchFile(path)
		}

		err := changeBuck.ForEach(func(_, v []byte) error {
			ck := &Checkpoint{}
			if err := ck.Unmarshal(v); err != nil {
				return err
			}

			hist = append(hist, ck)
			return nil
		})

		if err != nil {
			return err
		}

		// Make sure we're in order.
		sort.Sort(&hist)
		return nil
	})
}
