package daemon

import (
	"fmt"
	"time"

	"github.com/disorganizer/brig/daemon/wire"
	"github.com/disorganizer/brig/id"
	"github.com/disorganizer/brig/store"
	storewire "github.com/disorganizer/brig/store/wire"
	"github.com/gogo/protobuf/proto"
)

func (c *Client) recvResponse(logname string) (*wire.Response, error) {
	resp := <-c.Recv
	if resp != nil && !resp.GetSuccess() {
		return nil, fmt.Errorf("client: %v: %v", logname, resp.GetError())
	}

	return resp, nil
}

// Add adds the data at `filePath` to brig as `repoPath`.
func (c *Client) Add(filePath, repoPath string) error {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_ADD.Enum(),
		AddCommand: &wire.Command_AddCmd{
			FilePath: proto.String(filePath),
			RepoPath: proto.String(repoPath),
		},
	}

	if _, err := c.recvResponse("add"); err != nil {
		return err
	}

	return nil
}

// Cat outputs the brig file at `repoPath` to `filePath`.
func (c *Client) Cat(repoPath, filePath string) error {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_CAT.Enum(),
		CatCommand: &wire.Command_CatCmd{
			FilePath: proto.String(filePath),
			RepoPath: proto.String(repoPath),
		},
	}

	if _, err := c.recvResponse("cat"); err != nil {
		return err
	}

	return nil
}

// Mount serves a fuse endpoint at the specified path.
func (c *Client) Mount(mountPath string) error {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_MOUNT.Enum(),
		MountCommand: &wire.Command_MountCmd{
			MountPoint: proto.String(mountPath),
		},
	}

	if _, err := c.recvResponse("mount"); err != nil {
		return err
	}

	return nil
}

// Unmount removes a previously mounted fuse endpoint.
func (c *Client) Unmount(mountPath string) error {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_UNMOUNT.Enum(),
		UnmountCommand: &wire.Command_UnmountCmd{
			MountPoint: proto.String(mountPath),
		},
	}

	if _, err := c.recvResponse("unmount"); err != nil {
		return err
	}

	return nil
}

// Remove removes the brig file at `repoPath`
func (c *Client) Remove(repoPath string, recursive bool) error {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_RM.Enum(),
		RmCommand: &wire.Command_RmCmd{
			RepoPath:  proto.String(repoPath),
			Recursive: proto.Bool(recursive),
		},
	}

	if _, err := c.recvResponse("rm"); err != nil {
		return err
	}

	return nil
}

func (c *Client) Move(source, dest string) error {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_MV.Enum(),
		MvCommand: &wire.Command_MvCmd{
			Source: proto.String(source),
			Dest:   proto.String(dest),
		},
	}

	if _, err := c.recvResponse("mv"); err != nil {
		return err
	}

	return nil
}

// History returns the available checkpoints for the file at repoPath.
// It might have been deleted earlier. Asking for a non-existing file
// yields an empty history, but is not an error.
func (c *Client) History(repoPath string) (store.History, error) {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_HISTORY.Enum(),
		HistoryCommand: &wire.Command_HistoryCmd{
			RepoPath: proto.String(repoPath),
		},
	}

	resp, err := c.recvResponse("history")
	if err != nil {
		return nil, err
	}

	hist := &store.History{}
	protoHist := resp.GetHistoryResp().GetHistory()

	if err := hist.FromProto(protoHist); err != nil {
		return nil, err
	}

	return *hist, nil
}

func (c *Client) alterOnlineStatus(query wire.OnlineQuery) (*wire.Response, error) {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_ONLINE_STATUS.Enum(),
		OnlineStatusCommand: &wire.Command_OnlineStatusCmd{
			Query: &query,
		},
	}

	return c.recvResponse("online-status")
}

func (c *Client) Online() error {
	_, err := c.alterOnlineStatus(wire.OnlineQuery_GO_ONLINE)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Offline() error {
	_, err := c.alterOnlineStatus(wire.OnlineQuery_GO_OFFLINE)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) IsOnline() (bool, error) {
	resp, err := c.alterOnlineStatus(wire.OnlineQuery_IS_ONLINE)
	if err != nil {
		return false, err
	}

	return resp.GetOnlineStatusResp().GetIsOnline(), nil
}

func (c *Client) List(root string, depth int) ([]*storewire.Dirent, error) {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_LIST.Enum(),
		ListCommand: &wire.Command_ListCmd{
			Root:  proto.String(root),
			Depth: proto.Int(depth),
		},
	}

	resp, err := c.recvResponse("list")
	if err != nil {
		return nil, err
	}

	dirlist := resp.GetListResp().GetDirlist()
	return dirlist.Entries, nil
}

func (c *Client) Fetch(who id.ID) error {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_FETCH.Enum(),
		FetchCommand: &wire.Command_FetchCmd{
			Who: proto.String(string(who)),
		},
	}

	if _, err := c.recvResponse("fetch"); err != nil {
		return err
	}

	return nil
}

func (c *Client) Mkdir(path string) error {
	return c.mkdir(path, false)
}

func (c *Client) MkdirAll(path string) error {
	return c.mkdir(path, true)
}

func (c *Client) mkdir(path string, createParents bool) error {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_MKDIR.Enum(),
		MkdirCommand: &wire.Command_MkdirCmd{
			Path:          proto.String(string(path)),
			CreateParents: proto.Bool(createParents),
		},
	}

	if _, err := c.recvResponse("mkdir"); err != nil {
		return err
	}

	return nil
}

func (c *Client) RemoteAdd(ident id.ID, peerHash string) error {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_REMOTE_ADD.Enum(),
		RemoteAddCommand: &wire.Command_RemoteAddCmd{
			Id:   proto.String(string(ident)),
			Hash: proto.String(peerHash),
		},
	}

	if _, err := c.recvResponse("remote-add"); err != nil {
		return err
	}

	return nil
}

func (c *Client) RemoteRemove(ident id.ID) error {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_REMOTE_REMOVE.Enum(),
		RemoteRemoveCommand: &wire.Command_RemoteRemoveCmd{
			Id: proto.String(string(ident)),
		},
	}

	if _, err := c.recvResponse("remote-remove"); err != nil {
		return err
	}

	return nil
}

type RemoteEntry struct {
	Hash     string
	Ident    string
	IsOnline bool
}

func (c *Client) RemoteList() ([]*RemoteEntry, error) {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_REMOTE_LIST.Enum(),
		RemoteListCommand: &wire.Command_RemoteListCmd{
			NeedsOnline: proto.Bool(true),
		},
	}

	resp, err := c.recvResponse("remote-list")
	if err != nil {
		return nil, err
	}

	entries := []*RemoteEntry{}
	for _, entry := range resp.GetRemoteListResp().GetRemotes() {
		entries = append(entries, &RemoteEntry{
			Ident:    entry.GetId(),
			Hash:     entry.GetHash(),
			IsOnline: entry.GetIsOnline(),
		})
	}

	return entries, nil
}

func (c *Client) RemoteLocate(ident id.ID, limit int, timeout time.Duration) ([]string, error) {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_REMOTE_LOCATE.Enum(),
		RemoteLocateCommand: &wire.Command_RemoteLocateCmd{
			Id:        proto.String(string(ident)),
			PeerLimit: proto.Int32(int32(limit)),
			TimeoutMs: proto.Int32(int32(timeout / time.Millisecond)),
		},
	}

	resp, err := c.recvResponse("remote-locate")
	if err != nil {
		return nil, err
	}

	return resp.GetRemoteLocateResp().GetHashes(), nil
}

func (c *Client) RemoteSelf() (*RemoteEntry, error) {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_REMOTE_SELF.Enum(),
	}

	resp, err := c.recvResponse("remote-self")
	if err != nil {
		return nil, err
	}

	self := resp.GetRemoteSelfResp().GetSelf()
	return &RemoteEntry{
		Ident:    self.GetId(),
		Hash:     self.GetHash(),
		IsOnline: self.GetIsOnline(),
	}, nil
}

func (c *Client) Status() (*storewire.Commit, error) {
	c.Send <- &wire.Command{CommandType: wire.MessageType_STATUS.Enum()}

	resp, err := c.recvResponse("status")
	if err != nil {
		return nil, err
	}

	return resp.GetStatusResp().GetStageCommit(), nil
}

func (c *Client) MakeCommit(msg string) error {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_COMMIT.Enum(),
		CommitCommand: &wire.Command_CommitCmd{
			Message: proto.String(msg),
		},
	}

	if _, err := c.recvResponse("commit"); err != nil {
		return err
	}

	return nil
}

func (c *Client) Log(from, to *store.Hash) (*storewire.Commits, error) {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_LOG.Enum(),
		LogCommand: &wire.Command_LogCmd{
			Low:  from.Bytes(),
			High: to.Bytes(),
		},
	}

	resp, err := c.recvResponse("log")
	if err != nil {
		return nil, err
	}

	return resp.GetLogResp().GetCommits(), nil
}

func (c *Client) doPin(path string, balance int) (bool, error) {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_PIN.Enum(),
		PinCommand: &wire.Command_PinCmd{
			Path:    proto.String(path),
			Balance: proto.Int32(int32(balance)),
		},
	}

	resp, err := c.recvResponse("pin")
	if err != nil {
		return false, err
	}

	return resp.GetPinResp().GetIsPinned(), nil
}

func (c *Client) Pin(path string) error {
	if _, err := c.doPin(path, +1); err != nil {
		return err
	}

	return nil
}

func (c *Client) Unpin(path string) error {
	if _, err := c.doPin(path, -1); err != nil {
		return err
	}

	return nil
}

func (c *Client) IsPinned(path string) (bool, error) {
	return c.doPin(path, 0)
}
