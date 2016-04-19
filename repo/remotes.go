package repo

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/disorganizer/brig/id"
	"gopkg.in/yaml.v2"
)

type ErrRemoteHashExists struct {
	hash, id string
}

func (e ErrRemoteHashExists) Error() string {
	return fmt.Sprintf("Hash with this id (%v) exists already: %v", e.id, e.hash)
}

// ErrNoSuchRemote is returned when an ID could not have been resolved to an ID
type ErrNoSuchRemote id.ID

func (e ErrNoSuchRemote) Error() string {
	return fmt.Sprintf("No such remote `%s` found", string(e))
}

// Remote is the metadata of a single communication partner
// It contains the id and authentication info for each partner.
type Remote interface {
	// ID returns the ID of the remote partner
	ID() id.ID
	// Hash returns the peer hash of the partner
	Hash() string
}

// NewRemote returns a struct that fulfills the Remote interface
// fille with the passed in parameters.
func NewRemote(ID id.ID, hash string) Remote {
	// Re-use the yaml remote here, but don't tell anyone.
	return &yamlRemote{
		Identity: ID,
		yamlRemoteEntry: &yamlRemoteEntry{
			PeerHash: hash,
		},
	}
}

func NewRemoteFromPeer(peer id.Peer) Remote {
	return NewRemote(peer.ID(), peer.Hash())
}

// RemoteIsEqual returns true when two remotes have the same id and hash
func RemoteIsEqual(a, b Remote) bool {
	return a.ID() == b.ID() && a.Hash() == b.Hash()
}

type FileHandle interface {
	io.Reader
	io.Writer
	io.Seeker
	io.Closer
	Truncate(size int64) error
	Sync() error
}

// RemoteStore is a store for several Remotes.
type RemoteStore interface {
	io.Closer

	// Insert stores `r` for the partner `ID`.
	// If there is already a remote with this hash but with a
	// different ID, ErrRemoteHashExists should be returned.
	// If the ID exists already, it will be overwritten.
	Insert(r Remote) error

	// Get returns the Remote info for `ID`
	Get(ID id.ID) (Remote, error)

	// Remove purges the partner with `ID` from he store.
	Remove(ID id.ID) error

	// Iter returns a channel that yields every remote in the store.
	// The elements should be sorted in the alphabetic order of the ID.
	Iter() chan Remote
}

// AsList converts a RemoteStore into a list of Remotes.
func AsList(r RemoteStore) []Remote {
	var rms []Remote

	for rm := range r.Iter() {
		rms = append(rms, rm)
	}

	return rms
}

type yamlRemote struct {
	Identity id.ID
	*yamlRemoteEntry
}

type yamlRemoteEntry struct {
	PeerHash  string
	Timestamp time.Time
}

func (ye *yamlRemote) ID() id.ID {
	return ye.Identity
}

func (ye *yamlRemote) Hash() string {
	return ye.PeerHash
}

type yamlRemotes []*yamlRemote

func (yl yamlRemotes) Len() int {
	return len(yl)
}

func (yl yamlRemotes) Less(i, j int) bool {
	return yl[i].ID() < yl[j].ID()
}

func (yl yamlRemotes) Swap(i, j int) {
	yl[i], yl[j] = yl[j], yl[i]
}

type yamlRemoteStore struct {
	mu     sync.Mutex
	path   string
	parsed map[id.ID]*yamlRemoteEntry
}

// NewYAMLRemotes returns a new remote store that stores
// its data in the open file pointed to by `fd`.
// (os.Open() returns a suitable FileHandle)
func NewYAMLRemotes(path string) (RemoteStore, error) {
	remotes := &yamlRemoteStore{path: path}
	if err := remotes.load(); err != nil {
		return nil, err
	}

	return remotes, nil
}

func (yr *yamlRemoteStore) load() error {
	fd, err := os.OpenFile(yr.path, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer fd.Close()

	data, err := ioutil.ReadAll(fd)
	if err != nil {
		return err
	}

	parsed := make(map[id.ID]*yamlRemoteEntry)
	if err := yaml.Unmarshal(data, parsed); err != nil {
		return err
	}

	yr.parsed = parsed
	return nil
}

func (yr *yamlRemoteStore) save() error {
	fd, err := os.OpenFile(yr.path, os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer fd.Close()

	data, err := yaml.Marshal(yr.parsed)
	if err != nil {
		return err
	}

	if _, err := fd.Write(data); err != nil {
		return err
	}

	if err := fd.Sync(); err != nil {
		return err
	}

	return nil
}

func (yr *yamlRemoteStore) Insert(r Remote) error {
	yr.mu.Lock()
	defer yr.mu.Unlock()

	// Sanity check:
	hash := r.Hash()
	for id, entry := range yr.parsed {
		if id != r.ID() && entry.PeerHash == hash {
			return ErrRemoteHashExists{string(id), hash}
		}
	}

	yr.parsed[r.ID()] = &yamlRemoteEntry{
		PeerHash:  r.Hash(),
		Timestamp: time.Now(),
	}

	return yr.save()
}

func (yr *yamlRemoteStore) Get(ID id.ID) (Remote, error) {
	yr.mu.Lock()
	defer yr.mu.Unlock()

	ent, ok := yr.parsed[ID]
	if !ok {
		return nil, ErrNoSuchRemote(ID)
	}

	return &yamlRemote{
		Identity:        ID,
		yamlRemoteEntry: ent,
	}, nil
}

func (yr *yamlRemoteStore) Remove(ID id.ID) error {
	yr.mu.Lock()
	defer yr.mu.Unlock()

	if _, ok := yr.parsed[ID]; !ok {
		return ErrNoSuchRemote(ID)
	}

	delete(yr.parsed, ID)
	return yr.save()
}

func (yr *yamlRemoteStore) Iter() chan Remote {
	rmCh := make(chan Remote)

	go func() {
		yr.mu.Lock()
		defer yr.mu.Unlock()

		var remotes yamlRemotes

		for ident, entry := range yr.parsed {
			remotes = append(remotes, &yamlRemote{
				Identity:        ident,
				yamlRemoteEntry: entry,
			})
		}

		sort.Sort(remotes)

		for _, rm := range remotes {
			rmCh <- rm
		}

		close(rmCh)
	}()

	return rmCh
}

func (yr *yamlRemoteStore) Close() error {
	// TODO: Needed?
	return nil
}
