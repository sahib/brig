package client

import (
	"errors"
	"strings"
	"time"

	"github.com/sahib/brig/server/capnp"
	capnplib "zombiezen.com/go/capnproto2"
)

////////////////////////
// REMOTE LIST ACCESS //
////////////////////////

// RemoteFolder is a single folder shared with a remote.
type RemoteFolder struct {
	Folder           string `yaml:"Folder"`
	ReadOnly         bool   `yaml:"ReadOnly"`
	ConflictStrategy string `yaml:"ConflictStrategy"`
}

// Remote describes a single remote in the remote list.
type Remote struct {
	Name             string         `yaml:"Name"`
	Fingerprint      string         `yaml:"Fingerprint"`
	Folders          []RemoteFolder `yaml:"Folders,flow"`
	AutoUpdate       bool           `yaml:"AutoUpdate"`
	ConflictStrategy string         `yaml:"ConflictStrategy"`
	AcceptPush       bool           `yaml:"AcceptPush"`
}

func capRemoteToRemote(capRemote capnp.Remote) (*Remote, error) {
	remoteName, err := capRemote.Name()
	if err != nil {
		return nil, err
	}

	remoteFp, err := capRemote.Fingerprint()
	if err != nil {
		return nil, err
	}

	remoteFolders, err := capRemote.Folders()
	if err != nil {
		return nil, err
	}

	conflictStrategy, err := capRemote.ConflictStrategy()
	if err != nil {
		return nil, err
	}

	folders := []RemoteFolder{}
	for idx := 0; idx < remoteFolders.Len(); idx++ {
		folder := remoteFolders.At(idx)
		folderName, err := folder.Folder()
		if err != nil {
			return nil, err
		}

		cs, err := folder.ConflictStrategy()
		if err != nil {
			return nil, err
		}

		folders = append(folders, RemoteFolder{
			Folder:           folderName,
			ReadOnly:         folder.ReadOnly(),
			ConflictStrategy: cs,
		})
	}

	return &Remote{
		Name:             remoteName,
		Fingerprint:      remoteFp,
		Folders:          folders,
		AutoUpdate:       capRemote.AcceptAutoUpdates(),
		AcceptPush:       capRemote.AcceptPush(),
		ConflictStrategy: conflictStrategy,
	}, nil
}

func remoteToCapRemote(remote Remote, seg *capnplib.Segment) (*capnp.Remote, error) {
	capRemote, err := capnp.NewRemote(seg)
	if err != nil {
		return nil, err
	}

	if err := capRemote.SetName(remote.Name); err != nil {
		return nil, err
	}

	if err := capRemote.SetFingerprint(string(remote.Fingerprint)); err != nil {
		return nil, err
	}

	if err := capRemote.SetConflictStrategy(remote.ConflictStrategy); err != nil {
		return nil, err
	}

	capFolders, err := capnp.NewRemoteFolder_List(seg, int32(len(remote.Folders)))
	if err != nil {
		return nil, err
	}

	for idx, folder := range remote.Folders {
		capFolder, err := capnp.NewRemoteFolder(seg)
		if err != nil {
			return nil, err
		}

		capFolder.SetReadOnly(folder.ReadOnly)
		if err := capFolder.SetFolder(folder.Folder); err != nil {
			return nil, err
		}

		if err := capFolder.SetConflictStrategy(folder.ConflictStrategy); err != nil {
			return nil, err
		}

		if err := capFolders.Set(idx, capFolder); err != nil {
			return nil, err
		}
	}

	if err := capRemote.SetFolders(capFolders); err != nil {
		return nil, err
	}

	capRemote.SetAcceptAutoUpdates(remote.AutoUpdate)
	capRemote.SetAcceptPush(remote.AcceptPush)
	return &capRemote, nil
}

// RemoteAddOrUpdate adds a new remote described in `remote`.
// We thus authenticate this remote.
func (cl *Client) RemoteAddOrUpdate(remote Remote) error {
	call := cl.api.RemoteAddOrUpdate(cl.ctx, func(p capnp.Net_remoteAddOrUpdate_Params) error {
		capRemote, err := remoteToCapRemote(remote, p.Segment())
		if err != nil {
			return err
		}

		return p.SetRemote(*capRemote)
	})

	_, err := call.Struct()
	return err
}

// RemoteByName adds a new remote described in `remote`.
// We thus authenticate this remote.
func (cl *Client) RemoteByName(name string) (Remote, error) {
	call := cl.api.RemoteByName(cl.ctx, func(p capnp.Net_remoteByName_Params) error {
		return p.SetName(name)
	})

	res, err := call.Struct()
	if err != nil {
		return Remote{}, err
	}

	capRmt, err := res.Remote()
	if err != nil {
		return Remote{}, err
	}

	rmt, err := capRemoteToRemote(capRmt)
	if err != nil {
		return Remote{}, err
	}

	return *rmt, nil
}

// RemoteUpdate Updates the contents of `remote`.
func (cl *Client) RemoteUpdate(remote Remote) error {
	call := cl.api.RemoteUpdate(cl.ctx, func(p capnp.Net_remoteUpdate_Params) error {
		capRemote, err := remoteToCapRemote(remote, p.Segment())
		if err != nil {
			return err
		}

		return p.SetRemote(*capRemote)
	})

	_, err := call.Struct()
	return err
}

// RemoteRm removes a remote by `name` from the remote list.
func (cl *Client) RemoteRm(name string) error {
	call := cl.api.RemoteRm(cl.ctx, func(p capnp.Net_remoteRm_Params) error {
		return p.SetName(name)
	})

	_, err := call.Struct()
	return err
}

// RemoteClear clears all of the remote list.
func (cl *Client) RemoteClear() error {
	call := cl.api.RemoteClear(cl.ctx, func(p capnp.Net_remoteClear_Params) error {
		return nil
	})

	_, err := call.Struct()
	return err
}

// RemoteLs lists all remotes in the remote list.
func (cl *Client) RemoteLs() ([]Remote, error) {
	call := cl.api.RemoteLs(cl.ctx, func(p capnp.Net_remoteLs_Params) error {
		return nil
	})

	result, err := call.Struct()
	if err != nil {
		return nil, err
	}

	capRemotes, err := result.Remotes()
	if err != nil {
		return nil, err
	}

	remotes := []Remote{}
	for idx := 0; idx < capRemotes.Len(); idx++ {
		capRemote := capRemotes.At(idx)
		remote, err := capRemoteToRemote(capRemote)
		if err != nil {
			return nil, err
		}

		remotes = append(remotes, *remote)
	}

	return remotes, nil
}

// RemoteSave swaps the contents of the remote lists with the contents of `remotes`.
func (cl *Client) RemoteSave(remotes []Remote) error {
	call := cl.api.RemoteSave(cl.ctx, func(p capnp.Net_remoteSave_Params) error {
		seg := p.Segment()
		capRemotes, err := capnp.NewRemote_List(seg, int32(len(remotes)))
		if err != nil {
			return err
		}

		for idx, remote := range remotes {
			capRemote, err := remoteToCapRemote(remote, seg)
			if err != nil {
				return err
			}

			if err := capRemotes.Set(idx, *capRemote); err != nil {
				return err
			}
		}

		return p.SetRemotes(capRemotes)
	})

	_, err := call.Struct()
	return err
}

// LocateResult is a result returned by Locate()
type LocateResult struct {
	Name        string
	Addr        string
	Mask        []string
	Fingerprint string
}

func capLrToLr(capLr capnp.LocateResult) (*LocateResult, error) {
	name, err := capLr.Name()
	if err != nil {
		return nil, err
	}

	addr, err := capLr.Addr()
	if err != nil {
		return nil, err
	}

	mask, err := capLr.Mask()
	if err != nil {
		return nil, err
	}

	fingerprint, err := capLr.Fingerprint()
	if err != nil {
		return nil, err
	}

	return &LocateResult{
		Addr:        addr,
		Name:        name,
		Mask:        strings.Split(mask, ","),
		Fingerprint: fingerprint,
	}, nil
}

// NetLocate tries to find other remotes by searching of `who` described by `mask`.
// It will at max. take `timeoutSec` to search. This operation might take some time.
// The return channel will yield a LocateResult once a new result is available.
func (cl *Client) NetLocate(who, mask string, timeoutSec float64) (chan *LocateResult, error) {
	call := cl.api.NetLocate(cl.ctx, func(p capnp.Net_netLocate_Params) error {
		p.SetTimeoutSec(float64(timeoutSec))

		if err := p.SetLocateMask(mask); err != nil {
			return err
		}

		return p.SetWho(who)
	})

	result, err := call.Struct()
	if err != nil {
		return nil, err
	}

	ticket := result.Ticket()
	resultCh := make(chan *LocateResult)

	go func() {
		defer close(resultCh)

		for {
			nextCall := cl.api.NetLocateNext(cl.ctx, func(p capnp.Net_netLocateNext_Params) error {
				p.SetTicket(ticket)
				return nil
			})

			result, err := nextCall.Struct()
			if err != nil {
				continue
			}

			if !result.HasResult() {
				break
			}

			capLr, err := result.Result()
			if err != nil {
				continue
			}

			lr, err := capLrToLr(capLr)
			if err != nil {
				continue
			}

			resultCh <- lr
		}
	}()

	return resultCh, nil
}

// RemotePing pings a remote by the name `who`.
func (cl *Client) RemotePing(who string) (float64, error) {
	call := cl.api.RemotePing(cl.ctx, func(p capnp.Net_remotePing_Params) error {
		return p.SetWho(who)
	})

	result, err := call.Struct()
	if err != nil {
		return 0, err
	}

	return result.Roundtrip(), nil
}

// Whoami describes the current user state
type Whoami struct {
	CurrentUser string
	Owner       string
	Fingerprint string
	IsOnline    bool
}

// Whoami describes our own identity.
func (cl *Client) Whoami() (*Whoami, error) {
	call := cl.api.Whoami(cl.ctx, func(p capnp.Net_whoami_Params) error {
		return nil
	})

	result, err := call.Struct()
	if err != nil {
		return nil, err
	}

	capWhoami, err := result.Whoami()
	if err != nil {
		return nil, err
	}

	whoami := &Whoami{}
	whoami.CurrentUser, err = capWhoami.CurrentUser()
	if err != nil {
		return nil, err
	}

	whoami.Fingerprint, err = capWhoami.Fingerprint()
	if err != nil {
		return nil, err
	}

	whoami.Owner, err = capWhoami.Owner()
	if err != nil {
		return nil, err
	}

	whoami.IsOnline = capWhoami.IsOnline()
	return whoami, nil
}

// NetConnect connects to the ipfs network.
func (cl *Client) NetConnect() error {
	_, err := cl.api.Connect(cl.ctx, func(p capnp.Net_connect_Params) error {
		return nil
	}).Struct()
	return err
}

// NetDisconnect disconnects from the ipfs network.
func (cl *Client) NetDisconnect() error {
	_, err := cl.api.Disconnect(cl.ctx, func(p capnp.Net_disconnect_Params) error {
		return nil
	}).Struct()
	return err
}

// RemoteStatus is a entry in the remote online list.
// Fingerprint is not necessarily filled.
type RemoteStatus struct {
	Remote        Remote
	LastSeen      time.Time
	Roundtrip     time.Duration
	Err           error
	Authenticated bool
}

func capRemoteStatusToRemoteStatus(capStatus capnp.RemoteStatus) (*RemoteStatus, error) {
	capRemote, err := capStatus.Remote()
	if err != nil {
		return nil, err
	}

	remote, err := capRemoteToRemote(capRemote)
	if err != nil {
		return nil, err
	}

	msg, err := capStatus.Error()
	if err != nil {
		return nil, err
	}

	lastSeenStamp, err := capStatus.LastSeen()
	if err != nil {
		return nil, err
	}

	lastSeen := time.Now()
	if lastSeenStamp != "" {
		lastSeen, err = time.Parse(time.RFC3339, lastSeenStamp)
		if err != nil {
			return nil, err
		}
	}

	pingErr := errors.New(msg)
	if len(msg) == 0 {
		pingErr = nil
	}

	roundtripMs := time.Duration(capStatus.RoundtripMs()) * time.Millisecond
	return &RemoteStatus{
		Remote:        *remote,
		LastSeen:      lastSeen,
		Roundtrip:     roundtripMs,
		Err:           pingErr,
		Authenticated: capStatus.Authenticated(),
	}, nil
}

// RemoteOnlineList is like RemoteList but also includes IsOnline and Authenticated
// status.
func (cl *Client) RemoteOnlineList() ([]RemoteStatus, error) {
	call := cl.api.RemoteOnlineList(cl.ctx, func(p capnp.Net_remoteOnlineList_Params) error {
		return nil
	})

	result, err := call.Struct()
	if err != nil {
		return nil, err
	}

	capStatuses, err := result.Infos()
	if err != nil {
		return nil, err
	}

	statuses := []RemoteStatus{}
	for idx := 0; idx < capStatuses.Len(); idx++ {
		capStatus := capStatuses.At(idx)
		status, err := capRemoteStatusToRemoteStatus(capStatus)
		if err != nil {
			return nil, err
		}

		statuses = append(statuses, *status)
	}

	return statuses, nil
}

// Push sets a push request to `remoteName`. If `dryRun` is true,
// the push won't be send but we will still check if the push is allowed.
func (cl *Client) Push(remoteName string, dryRun bool) error {
	call := cl.api.Push(cl.ctx, func(p capnp.Net_push_Params) error {
		p.SetDryRun(dryRun)
		return p.SetRemoteName(remoteName)
	})

	_, err := call.Struct()
	return err
}
