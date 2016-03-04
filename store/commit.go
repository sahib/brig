package store

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/disorganizer/brig/store/proto"
	protobuf "github.com/gogo/protobuf/proto"
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

var (
	// ErrNoChange means that nothing changed between two versions (of a file)
	ErrNoChange = fmt.Errorf("Nothing changed between the given versions")
)

// String formats a changetype to a human readable verb in past tense.
func (c *ChangeType) String() string {
	return changeTypeToString[*c]
}

// UnmarshalJSON reads a json string and tries to convert it to a ChangeType.
func (c *ChangeType) Unmarshal(data []byte) error {
	var ok bool
	*c, ok = stringToChangeType[string(data)]
	if !ok {
		return fmt.Errorf("Bad change type: %v", string(data))
	}

	return nil
}

// Commit groups a change set
type Commit struct {
	// Optional commit message
	Message string `json:"message"`

	// Author is the jabber id of the committer.
	Author string `json:"author"`

	// Time at this commit was conceived.
	ModTime time.Time `json:"modtime"`

	// Set of files that were changed.
	Changes map[string]*Checkpoint `json:"changes"`

	// Parent commit (only nil for initial commit)
	Parent *Commit `json:"-"`
}

// Checkpoint remembers one state of a single file.
type Checkpoint struct {
	// Hash is the hash of the file at this point.
	// It may, or may not be retrievable from ipfs.
	// For ChangeRemove the hash is the hash of the last existing file.
	Hash *Hash

	// ModTime is the time the checkpoint was made.
	ModTime time.Time

	// Size is the size of the file in bytes at this point.
	Size int64

	// Change is the detailed type of the modification.
	Change ChangeType

	// Author of the file modifications (jabber id)
	Author string
}

// TODO: nice representation
func (c *Checkpoint) String() string {
	return fmt.Sprintf("%-7s %+7s@%s", c.Change.String(), c.Hash.B58String(), c.ModTime.String())
}

func (cp *Checkpoint) toProtoMessage() (*proto.Checkpoint, error) {
	mtimeBin, err := cp.ModTime.MarshalBinary()
	if err != nil {
		return nil, err
	}

	protoCheck := &proto.Checkpoint{
		Hash:     cp.Hash.Bytes(),
		ModTime:  mtimeBin,
		FileSize: protobuf.Int64(cp.Size),
		Change:   protobuf.Int32(int32(cp.Change)),
		Author:   protobuf.String(cp.Author),
	}

	if err != nil {
		return nil, err
	}

	return protoCheck, nil
}

func (cp *Checkpoint) fromProtoMessage(msg *proto.Checkpoint) error {
	modTime := time.Time{}
	if err := modTime.UnmarshalBinary(msg.GetModTime()); err != nil {
		return err
	}

	cp.Hash = &Hash{msg.GetHash()}
	cp.ModTime = modTime
	cp.Size = msg.GetFileSize()
	cp.Change = ChangeType(msg.GetChange())
	cp.Author = msg.GetAuthor()
	return nil
}

func (cp *Checkpoint) Marshal() ([]byte, error) {
	protoCheck, err := cp.toProtoMessage()
	if err != nil {
		return nil, err
	}

	protoData, err := protobuf.Marshal(protoCheck)
	if err != nil {
		return nil, err
	}

	return protoData, nil
}

func (cp *Checkpoint) Unmarshal(data []byte) error {
	protoCheck := &proto.Checkpoint{}
	if err := protobuf.Unmarshal(data, protoCheck); err != nil {
		return err
	}

	return cp.fromProtoMessage(protoCheck)
}

// History remembers the changes made to a file.
// New changes get appended to the end.
type History []*Checkpoint

func (hy *History) Marshal() ([]byte, error) {
	protoHist := &proto.History{}

	for _, ck := range *hy {
		protoCheck, err := ck.toProtoMessage()
		if err != nil {
			return nil, err
		}

		protoHist.Hist = append(protoHist.Hist, protoCheck)
	}

	data, err := protobuf.Marshal(protoHist)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (hy *History) Unmarshal(data []byte) error {
	protoHist := &proto.History{}

	if err := protobuf.Unmarshal(data, protoHist); err != nil {
		return err
	}

	for _, protoCheck := range protoHist.Hist {
		ck := &Checkpoint{}
		if err := ck.fromProtoMessage(protoCheck); err != nil {
			return err
		}

		*hy = append(*hy, ck)
	}

	return nil
}

// MakeCheckpoint creates a new checkpoint in the version history of `curr`.
// One of old or curr may be nil (if no old version exists or new version
// does not exist anymore). It is an error to pass nil twice.
//
// If nothing changed between old and curr, ErrNoChange is returned.
func (s *Store) MakeCheckpoint(old, curr *Metadata, oldPath, currPath string) error {
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
		// TODO: Take the actual one
		Author: "alice@jabber.nullcat.de/desktop",
	}

	protoData, err := checkpoint.Marshal()
	if err != nil {
		return err
	}

	mtimeBin, err := checkpoint.ModTime.MarshalBinary()
	if err != nil {
		return err
	}

	dbErr := s.updateWithBucket("checkpoints", func(tx *bolt.Tx, bckt *bolt.Bucket) error {
		histBuck, err := bckt.CreateBucketIfNotExists([]byte(path))
		if err != nil {
			return err
		}

		return histBuck.Put(mtimeBin, protoData)
	})

	if dbErr != nil {
		return dbErr
	}

	log.Debugf("created check point: ", checkpoint)
	return nil
}

// History returns all checkpoints a file has.
// Note: even on error a empty history is returned.
func (s *Store) History(path string) (*History, error) {
	var hist History

	return &hist, s.viewWithBucket("checkpoints", func(tx *bolt.Tx, bckt *bolt.Bucket) error {
		changeBuck := bckt.Bucket([]byte(path))
		if changeBuck == nil {
			// No history yet, return empty.
			return nil
		}

		return changeBuck.ForEach(func(k, v []byte) error {
			ck := &Checkpoint{}
			if err := ck.Unmarshal(v); err != nil {
				return err
			}

			hist = append(hist, ck)
			return nil
		})
	})
}
