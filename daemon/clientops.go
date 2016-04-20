package daemon

import (
	"bytes"
	"fmt"

	"github.com/disorganizer/brig/daemon/wire"
	"github.com/disorganizer/brig/id"
	"github.com/disorganizer/brig/store"
	storewire "github.com/disorganizer/brig/store/wire"
	"github.com/disorganizer/brig/util/protocol"
	"github.com/gogo/protobuf/proto"
)

// recvResponseBytes reads one response from the daemon and formats possible errors.
func (c *Client) recvResponse(logname string) (*wire.Response, error) {
	resp := <-c.Recv
	if resp != nil && !resp.GetSuccess() {
		return nil, fmt.Errorf("client: %v: %v", logname, resp.GetError())
	}

	return resp, nil
}

// recvResponseBytes reads one response from the daemon and formats possible errors.
func (c *Client) recvResponseBytes(logname string) ([]byte, error) {
	resp, err := c.recvResponse(logname)
	if err != nil {
		return nil, err
	}

	return resp.GetResponse(), nil
}

func (c *Client) recvResponseString(logname string) (string, error) {
	resp, err := c.recvResponseBytes(logname)
	if err != nil {
		return "", err
	}

	return string(resp), nil
}

// Add adds the data at `filePath` to brig as `repoPath`.
func (c *Client) Add(filePath, repoPath string) (string, error) {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_ADD.Enum(),
		AddCommand: &wire.Command_AddCmd{
			FilePath: proto.String(filePath),
			RepoPath: proto.String(repoPath),
		},
	}

	return c.recvResponseString("add")
}

// Cat outputs the brig file at `repoPath` to `filePath`.
func (c *Client) Cat(repoPath, filePath string) (string, error) {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_CAT.Enum(),
		CatCommand: &wire.Command_CatCmd{
			FilePath: proto.String(filePath),
			RepoPath: proto.String(repoPath),
		},
	}

	return c.recvResponseString("cat")
}

// Mount serves a fuse endpoint at the specified path.
func (c *Client) Mount(mountPath string) (string, error) {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_MOUNT.Enum(),
		MountCommand: &wire.Command_MountCmd{
			MountPoint: proto.String(mountPath),
		},
	}

	return c.recvResponseString("mount")
}

// Unmount removes a previously mounted fuse endpoint.
func (c *Client) Unmount(mountPath string) (string, error) {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_UNMOUNT.Enum(),
		UnmountCommand: &wire.Command_UnmountCmd{
			MountPoint: proto.String(mountPath),
		},
	}

	return c.recvResponseString("unmount")
}

// Remove removes the brig file at `repoPath`
func (c *Client) Remove(repoPath string, recursive bool) (string, error) {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_RM.Enum(),
		RmCommand: &wire.Command_RmCmd{
			RepoPath:  proto.String(repoPath),
			Recursive: proto.Bool(recursive),
		},
	}

	return c.recvResponseString("rm")
}

func (c *Client) Move(source, dest string) error {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_MV.Enum(),
		MvCommand: &wire.Command_MvCmd{
			Source: proto.String(source),
			Dest:   proto.String(dest),
		},
	}

	if _, err := c.recvResponseString("mv"); err != nil {
		return err
	}

	return nil
}

// Log returns a series of commits.
func (c *Client) Log() ([]*store.Commit, error) {
	c.Send <- &wire.Command{CommandType: wire.MessageType_LOG.Enum()}

	// TODO: Implement.
	return nil, nil
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

	// TODO: Sending json over protobuf is pretty hilarious/stupid.
	//       Do something else, but be consistent this time.
	protoData, err := c.recvResponseBytes("history")
	if err != nil {
		return nil, err
	}

	hist := &store.History{}
	if err := hist.Unmarshal(protoData); err != nil {
		return nil, err
	}

	return *hist, nil
}

func (c *Client) alterOnlineStatus(query wire.OnlineQuery) ([]byte, error) {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_ONLINE_STATUS.Enum(),
		OnlineStatusCommand: &wire.Command_OnlineStatusCmd{
			Query: &query,
		},
	}

	data, err := c.recvResponseBytes("online-status")
	if err != nil {
		return nil, err
	}

	return data, nil
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
	data, err := c.alterOnlineStatus(wire.OnlineQuery_IS_ONLINE)
	if err != nil {
		return false, err
	}

	return string(data) == "online", nil
}

func (c *Client) List(root string, depth int) ([]*storewire.Dirent, error) {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_LIST.Enum(),
		ListCommand: &wire.Command_ListCmd{
			Root:  proto.String(root),
			Depth: proto.Int(depth),
		},
	}

	listData, err := c.recvResponseBytes("list")
	if err != nil {
		return nil, err
	}

	dec := protocol.NewProtocolReader(bytes.NewReader(listData), nil, true)
	dirlist := &storewire.Dirlist{}

	if err := dec.Recv(dirlist); err != nil {
		return nil, err
	}

	return dirlist.Entries, nil
}

func (c *Client) Fetch(who id.ID) error {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_FETCH.Enum(),
		FetchCommand: &wire.Command_FetchCmd{
			Who: proto.String(string(who)),
		},
	}

	if _, err := c.recvResponseBytes("fetch"); err != nil {
		return err
	}

	return nil
}

func (c *Client) Mkdir(path string) error {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_MKDIR.Enum(),
		MkdirCommand: &wire.Command_MkdirCmd{
			Path: proto.String(string(path)),
		},
	}

	if _, err := c.recvResponseBytes("mkdir"); err != nil {
		return err
	}

	return nil
}

func (c *Client) AuthAdd(ident id.ID, peerHash string) error {
	c.Send <- &wire.Command{
		CommandType: wire.MessageType_AUTH_ADD.Enum(),
		AuthAddCommand: &wire.Command_AuthAddCmd{
			Who:      proto.String(string(ident)),
			PeerHash: proto.String(peerHash),
		},
	}

	if _, err := c.recvResponseBytes("auth-add"); err != nil {
		return err
	}

	return nil
}

func (c *Client) AuthPrint() (string, error) {
	c.Send <- &wire.Command{
		CommandType:      wire.MessageType_AUTH_PRINT.Enum(),
		AuthPrintCommand: &wire.Command_AuthPrintCmd{},
	}

	finger, err := c.recvResponseBytes("auth-print")
	if err != nil {
		return "", err
	}

	return string(finger), nil
}
