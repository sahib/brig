package store

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
)

const (
	// ChangeInvalid indicates a bug.
	ChangeInvalid = iota

	// The file was newly added.
	ChangeAdd

	// The file was modified
	ChangeModify

	// The file was moved
	ChangeMove

	// The file was removed.
	ChangeRemove
)

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
	ErrNoChange = fmt.Errorf("Nothing changed between the given versions")
)

func (c *ChangeType) String() string {
	return changeTypeToString[*c]
}

func (c *ChangeType) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(c.String())), nil
}

func (c *ChangeType) UnmarshalJSON(data []byte) error {
	unquoted, err := strconv.Unquote(string(data))
	if err != nil {
		return err
	}

	*c = stringToChangeType[unquoted]
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
	Hash *Hash `json:"hash"`

	// ModTime is the time the checkpoint was made.
	ModTime time.Time `json:"modtime"`

	// Size is the size of the file in bytes at this point.
	Size int64 `json:"size"`

	// Change is the detailed type of the modification.
	Change *ChangeType `json:"change"`

	// Author of the file modifications (jabber id)
	Author string `json:"author"`
}

// TODO: nice representation
func (c *Checkpoint) String() string {
	return fmt.Sprintf("%-7s %+7s@%s", c.Change, c.Hash.B58String(), c.ModTime.String())
}

// History remembers the changes made to a file.
// New changes get appended to the end.
type History []*Checkpoint

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
		Change:  &change,
		// TODO: Take the actual one.
		Author: "alice@jabber.nullcat.de/desktop",
	}

	jsonPoint, err := json.Marshal(checkpoint)
	if err != nil {
		return err
	}

	mtimeJson, err := json.Marshal(checkpoint.ModTime)
	if err != nil {
		return err
	}

	dbErr := s.updateWithBucket("checkpoints", func(tx *bolt.Tx, bckt *bolt.Bucket) error {
		histBuck, err := bckt.CreateBucketIfNotExists([]byte(path))
		if err != nil {
			return err
		}

		return histBuck.Put(mtimeJson, jsonPoint)
	})

	if dbErr != nil {
		return dbErr
	}

	fmt.Println("created check point: ", checkpoint)
	return nil
}

// History returns all checkpoints a file has.
// Note: even on error a empty history is returned.
func (s *Store) History(f *File) (History, error) {
	hist := make(History, 0)

	return hist, s.viewWithBucket("checkpoints", func(tx *bolt.Tx, bckt *bolt.Bucket) error {
		changeBuck := bckt.Bucket([]byte(f.Path()))
		if changeBuck == nil {
			// No history yet, return empty.
			return nil
		}

		return changeBuck.ForEach(func(k, v []byte) error {
			ck := &Checkpoint{}
			if err := json.Unmarshal(v, &ck); err != nil {
				return err
			}

			hist = append(hist, ck)
			return nil
		})
	})
}
