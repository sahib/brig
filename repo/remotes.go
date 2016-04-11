package repo

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/disorganizer/brig/id"
	"gopkg.in/yaml.v2"
)

// ErrNoSuchRemote is returned when an ID could not have been resolved to an ID
type ErrNoSuchRemote id.ID

func (e ErrNoSuchRemote) Error() string {
	return fmt.Sprintf("No such remote `%s` found", e)
}

// Remote is the metadata of a single communication partner
// It contains the id and authentication info for each partner.
type Remote interface {
	ID() id.ID
	Hash() string
}

func NewRemote(ID id.ID, hash string) Remote {
	// Re-use the yaml remote here, but don't tell anyone.
	return &yamlRemote{
		Identity: ID,
		yamlRemoteEntry: &yamlRemoteEntry{
			PeerHash: hash,
		},
	}
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
	// If it exists already, it will be overwritten.
	Insert(r Remote) error

	// Get returns the Remote info for `ID`
	Get(ID id.ID) (Remote, error)

	// Remove purges the partner with `ID` from he store.
	Remove(ID id.ID) error

	// Iter returns a channel that yields every remote in the store.
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

type yamlRemotes struct {
	mu     sync.Mutex
	fd     FileHandle
	parsed map[id.ID]*yamlRemoteEntry
}

// NewYAMLRemotes returns a new remote store that stores
// its data in the open file pointed to by `fd`.
// (os.Open() returns a suitable FileHandle)
func NewYAMLRemotes(fd FileHandle) (RemoteStore, error) {
	remotes := &yamlRemotes{fd: fd}
	if err := remotes.load(); err != nil {
		return nil, err
	}

	return remotes, nil
}

func (yr *yamlRemotes) load() error {
	if _, err := yr.fd.Seek(0, os.SEEK_SET); err != nil {
		return err
	}

	data, err := ioutil.ReadAll(yr.fd)
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

func (yr *yamlRemotes) save() error {
	data, err := yaml.Marshal(yr.parsed)
	if err != nil {
		return err
	}

	if _, err := yr.fd.Seek(0, os.SEEK_SET); err != nil {
		return err
	}

	if err := yr.fd.Truncate(0); err != nil {
		return err
	}

	if _, err := yr.fd.Write(data); err != nil {
		return err
	}

	if err := yr.fd.Sync(); err != nil {
		return err
	}

	return nil
}

func (yr *yamlRemotes) Insert(r Remote) error {
	yr.mu.Lock()
	defer yr.mu.Unlock()

	entry := &yamlRemoteEntry{
		PeerHash:  r.Hash(),
		Timestamp: time.Now(),
	}

	yr.parsed[r.ID()] = entry
	return yr.save()
}

func (yr *yamlRemotes) Get(ID id.ID) (Remote, error) {
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

func (yr *yamlRemotes) Remove(ID id.ID) error {
	yr.mu.Lock()
	defer yr.mu.Unlock()

	if _, ok := yr.parsed[ID]; !ok {
		return ErrNoSuchRemote(ID)
	}

	delete(yr.parsed, ID)
	return yr.save()
}

func (yr *yamlRemotes) Iter() chan Remote {
	rmCh := make(chan Remote)

	go func() {
		yr.mu.Lock()
		defer yr.mu.Unlock()

		for ident, entry := range yr.parsed {
			rmCh <- &yamlRemote{
				Identity:        ident,
				yamlRemoteEntry: entry,
			}
		}

		close(rmCh)
	}()

	return rmCh
}

func (yr *yamlRemotes) Close() error {
	return yr.fd.Close()
}
