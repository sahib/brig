package daemon

import (
	"fmt"
	"time"

	"github.com/disorganizer/brig/daemon/wire"
	"github.com/disorganizer/brig/id"
	"github.com/disorganizer/brig/store"
	storewire "github.com/disorganizer/brig/store/wire"
)

func (c *Client) recvResponse(logname string) (*wire.Response, error) {
	resp := <-c.Recv
	if resp != nil && !resp.Success {
		return nil, fmt.Errorf("client: %v: %v", logname, resp.Error)
	}

	return resp, nil
}

// Add adds the data at `filePath` to brig as `repoPath`.
func (c *Client) Stage(filePath, repoPath string) error {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_STAGE,
		AddCommand: &wire.Command_StageCmd{
			FilePath: filePath,
			RepoPath: repoPath,
		},
	}

	if _, err := c.recvResponse("stage"); err != nil {
		return err
	}

	return nil
}

func (c *Client) Reset(repoPath, commitRef string) error {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_RESET,
		ResetCommand: &wire.Command_ResetCmd{
			RepoPath:  repoPath,
			CommitRef: commitRef,
		},
	}

	if _, err := c.recvResponse("reset"); err != nil {
		return err
	}

	return nil
}

// Cat outputs the brig file at `repoPath` to `filePath`.
func (c *Client) Cat(repoPath, filePath string) error {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_CAT,
		CatCommand: &wire.Command_CatCmd{
			FilePath: filePath,
			RepoPath: repoPath,
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
		CommandType: wire.MessageType_MOUNT,
		MountCommand: &wire.Command_MountCmd{
			MountPoint: mountPath,
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
		CommandType: wire.MessageType_UNMOUNT,
		UnmountCommand: &wire.Command_UnmountCmd{
			MountPoint: mountPath,
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
		CommandType: wire.MessageType_RM,
		RmCommand: &wire.Command_RmCmd{
			RepoPath:  repoPath,
			Recursive: recursive,
		},
	}

	if _, err := c.recvResponse("rm"); err != nil {
		return err
	}

	return nil
}

func (c *Client) Move(source, dest string) error {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_MV,
		MvCommand: &wire.Command_MvCmd{
			Source: source,
			Dest:   dest,
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
		CommandType: wire.MessageType_HISTORY,
		HistoryCommand: &wire.Command_HistoryCmd{
			RepoPath: repoPath,
		},
	}

	resp, err := c.recvResponse("history")
	if err != nil {
		return nil, err
	}

	hist := &store.History{}
	protoHist := resp.GetHistoryResp().History

	if err := hist.FromProto(protoHist); err != nil {
		return nil, err
	}

	return *hist, nil
}

func (c *Client) alterOnlineStatus(query wire.OnlineQuery) (*wire.Response, error) {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_ONLINE_STATUS,
		OnlineStatusCommand: &wire.Command_OnlineStatusCmd{
			Query: query,
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

	return resp.GetOnlineStatusResp().IsOnline, nil
}

func (c *Client) List(root string, depth int) ([]*storewire.Node, error) {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_LIST,
		ListCommand: &wire.Command_ListCmd{
			Root:  root,
			Depth: int32(depth),
		},
	}

	resp, err := c.recvResponse("list")
	if err != nil {
		return nil, err
	}

	return resp.GetListResp().Entries.GetNodes(), nil
}

func (c *Client) Sync(who id.ID) error {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_SYNC,
		SyncCommand: &wire.Command_SyncCmd{
			Who: string(who),
		},
	}

	if _, err := c.recvResponse("sync"); err != nil {
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
		CommandType: wire.MessageType_MKDIR,
		MkdirCommand: &wire.Command_MkdirCmd{
			Path:          string(path),
			CreateParents: createParents,
		},
	}

	if _, err := c.recvResponse("mkdir"); err != nil {
		return err
	}

	return nil
}

func (c *Client) RemoteAdd(ident id.ID, peerHash string) error {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_REMOTE_ADD,
		RemoteAddCommand: &wire.Command_RemoteAddCmd{
			Id:   string(ident),
			Hash: peerHash,
		},
	}

	if _, err := c.recvResponse("remote-add"); err != nil {
		return err
	}

	return nil
}

func (c *Client) RemoteRemove(ident id.ID) error {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_REMOTE_REMOVE,
		RemoteRemoveCommand: &wire.Command_RemoteRemoveCmd{
			Id: string(ident),
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
		CommandType: wire.MessageType_REMOTE_LIST,
		RemoteListCommand: &wire.Command_RemoteListCmd{
			NeedsOnline: true,
		},
	}

	resp, err := c.recvResponse("remote-list")
	if err != nil {
		return nil, err
	}

	entries := []*RemoteEntry{}
	for _, entry := range resp.GetRemoteListResp().Remotes {
		entries = append(entries, &RemoteEntry{
			Ident:    entry.Id,
			Hash:     entry.Hash,
			IsOnline: entry.IsOnline,
		})
	}

	return entries, nil
}

func (c *Client) RemoteLocate(ident id.ID, limit int, timeout time.Duration) ([]string, error) {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_REMOTE_LOCATE,
		RemoteLocateCommand: &wire.Command_RemoteLocateCmd{
			Id:        string(ident),
			PeerLimit: int32(limit),
			TimeoutMs: int32(timeout / time.Millisecond),
		},
	}

	resp, err := c.recvResponse("remote-locate")
	if err != nil {
		return nil, err
	}

	return resp.GetRemoteLocateResp().Hashes, nil
}

func (c *Client) RemoteSelf() (*RemoteEntry, error) {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_REMOTE_SELF,
	}

	resp, err := c.recvResponse("remote-self")
	if err != nil {
		return nil, err
	}

	self := resp.GetRemoteSelfResp().Self
	return &RemoteEntry{
		Ident:    self.Id,
		Hash:     self.Hash,
		IsOnline: self.IsOnline,
	}, nil
}

func (c *Client) Status() (*storewire.Node, error) {
	c.Send <- &wire.Command{CommandType: wire.MessageType_STATUS}

	resp, err := c.recvResponse("status")
	if err != nil {
		return nil, err
	}

	return resp.GetStatusResp().StageCommit, nil
}

func (c *Client) MakeCommit(msg string) error {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_COMMIT,
		CommitCommand: &wire.Command_CommitCmd{
			Message: msg,
		},
	}

	if _, err := c.recvResponse("commit"); err != nil {
		return err
	}

	return nil
}

func (c *Client) Log(from, to *store.Hash) (*storewire.Nodes, error) {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_LOG,
		LogCommand: &wire.Command_LogCmd{
			Low:  from.Bytes(),
			High: to.Bytes(),
		},
	}

	resp, err := c.recvResponse("log")
	if err != nil {
		return nil, err
	}

	return resp.GetLogResp().GetNodes(), nil
}

func (c *Client) doPin(path string, balance int) (bool, error) {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_PIN,
		PinCommand: &wire.Command_PinCmd{
			Path:    path,
			Balance: int32(balance),
		},
	}

	resp, err := c.recvResponse("pin")
	if err != nil {
		return false, err
	}

	return resp.GetPinResp().IsPinned, nil
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

func (c *Client) Export(who id.ID) ([]byte, error) {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_EXPORT,
		ExportCommand: &wire.Command_ExportCmd{
			Who: string(who),
		},
	}

	resp, err := c.recvResponse("export")
	if err != nil {
		return nil, err
	}

	return resp.GetExportResp().Data, nil
}

func (c *Client) Import(data []byte) error {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_IMPORT,
		ImportCommand: &wire.Command_ImportCmd{
			Data: data,
		},
	}

	// TODO: add a 'checkErrResponse()' func?
	_, err := c.recvResponse("import")
	return err
}
