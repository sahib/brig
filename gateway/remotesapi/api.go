// Package remotesapi implements a simple API for accessing all remotes
// and their online state as well as triggering syncs and diffs.
// Normally this involves at least three different APIs.
package remotesapi

import (
	"time"

	"github.com/sahib/brig/catfs"
)

// Folder is a single folder limit for a remote.
type Folder struct {
	Folder   string `json:"folder"`
	ReadOnly bool   `json:"read_only"`
}

// Remote is a the result of List and Get.
type Remote struct {
	Name              string    `json:"name"`
	Folders           []Folder  `json:"folders"`
	Fingerprint       string    `json:"fingerprint"`
	AcceptAutoUpdates bool      `json:"accept_auto_updates"`
	IsOnline          bool      `json:"is_online"`
	IsAuthenticated   bool      `json:"is_authenticated"`
	AcceptPush        bool      `json:"accept_push"`
	ConflictStrategy  string    `json:"conflict_strategy"`
	LastSeen          time.Time `json:"last_seen"`
}

// Identity describes our own repository identity.
type Identity struct {
	Name        string `json:"name"`
	Fingerprint string `json:"fingerprint"`
}

// RemotesAPI provides a simpler interface to accessing remote information
// from repo.Repository, net.PeerServer and events.EventListener.
type RemotesAPI interface {
	List() ([]*Remote, error)
	Get(name string) (*Remote, error)
	Set(rm Remote) error
	Remove(name string) error
	Self() (Identity, error)
	OnChange(fn func())

	Sync(name string) error
	MakeDiff(name string) (*catfs.Diff, error)
}
