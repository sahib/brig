package daemon

import (
	"bytes"
	"fmt"

	"github.com/disorganizer/brig/daemon/proto"
	"github.com/disorganizer/brig/store"
	storeproto "github.com/disorganizer/brig/store/proto"
	"github.com/disorganizer/brig/util/protocol"
	protobuf "github.com/gogo/protobuf/proto"
	"github.com/tsuibin/goxmpp2/xmpp"
)

// recvResponseBytes reads one response from the daemon and formats possible errors.
func (c *Client) recvResponse(logname string) (*proto.Response, error) {
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
	c.Send <- &proto.Command{
		CommandType: proto.MessageType_ADD.Enum(),
		AddCommand: &proto.Command_AddCmd{
			FilePath: protobuf.String(filePath),
			RepoPath: protobuf.String(repoPath),
		},
	}

	return c.recvResponseString("add")
}

// Cat outputs the brig file at `repoPath` to `filePath`.
func (c *Client) Cat(repoPath, filePath string) (string, error) {
	c.Send <- &proto.Command{
		CommandType: proto.MessageType_CAT.Enum(),
		CatCommand: &proto.Command_CatCmd{
			FilePath: protobuf.String(filePath),
			RepoPath: protobuf.String(repoPath),
		},
	}

	return c.recvResponseString("cat")
}

// Mount serves a fuse endpoint at the specified path.
func (c *Client) Mount(mountPath string) (string, error) {
	c.Send <- &proto.Command{
		CommandType: proto.MessageType_MOUNT.Enum(),
		MountCommand: &proto.Command_MountCmd{
			MountPoint: protobuf.String(mountPath),
		},
	}

	return c.recvResponseString("mount")
}

// Unmount removes a previously mounted fuse endpoint.
func (c *Client) Unmount(mountPath string) (string, error) {
	c.Send <- &proto.Command{
		CommandType: proto.MessageType_UNMOUNT.Enum(),
		UnmountCommand: &proto.Command_UnmountCmd{
			MountPoint: protobuf.String(mountPath),
		},
	}

	return c.recvResponseString("unmount")
}

// Rm removes the brig file at `repoPath`
func (c *Client) Rm(repoPath string) (string, error) {
	c.Send <- &proto.Command{
		CommandType: proto.MessageType_RM.Enum(),
		RmCommand: &proto.Command_RmCmd{
			RepoPath: protobuf.String(repoPath),
		},
	}

	return c.recvResponseString("rm")
}

func (c *Client) Move(source, dest string) error {
	c.Send <- &proto.Command{
		CommandType: proto.MessageType_MV.Enum(),
		MvCommand: &proto.Command_MvCmd{
			Source: protobuf.String(source),
			Dest:   protobuf.String(dest),
		},
	}

	if _, err := c.recvResponseString("mv"); err != nil {
		return err
	}

	return nil
}

// Log returns a series of commits.
func (c *Client) Log() ([]*store.Commit, error) {
	c.Send <- &proto.Command{CommandType: proto.MessageType_LOG.Enum()}

	// TODO: Implement.
	return nil, nil
}

// History returns the available checkpoints for the file at repoPath.
// It might have been deleted earlier. Asking for a non-existing file
// yields an empty history, but is not an error.
func (c *Client) History(repoPath string) (store.History, error) {
	c.Send <- &proto.Command{
		CommandType: proto.MessageType_HISTORY.Enum(),
		HistoryCommand: &proto.Command_HistoryCmd{
			RepoPath: protobuf.String(repoPath),
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

func (c *Client) alterOnlineStatus(query proto.OnlineQuery) ([]byte, error) {
	c.Send <- &proto.Command{
		CommandType: proto.MessageType_ONLINE_STATUS.Enum(),
		OnlineStatusCommand: &proto.Command_OnlineStatusCmd{
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
	_, err := c.alterOnlineStatus(proto.OnlineQuery_GO_ONLINE)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Offline() error {
	_, err := c.alterOnlineStatus(proto.OnlineQuery_GO_OFFLINE)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) IsOnline() (bool, error) {
	data, err := c.alterOnlineStatus(proto.OnlineQuery_IS_ONLINE)
	if err != nil {
		return false, err
	}

	return string(data) == "online", nil
}

func (c *Client) List(root string, depth int) ([]*storeproto.Dirent, error) {
	c.Send <- &proto.Command{
		CommandType: proto.MessageType_LIST.Enum(),
		ListCommand: &proto.Command_ListCmd{
			Root:  protobuf.String(root),
			Depth: protobuf.Int(depth),
		},
	}

	listData, err := c.recvResponseBytes("list")
	if err != nil {
		return nil, err
	}

	dec := protocol.NewProtocolReader(bytes.NewReader(listData), true)
	dirlist := &storeproto.Dirlist{}

	if err := dec.Recv(dirlist); err != nil {
		return nil, err
	}

	return dirlist.Entries, nil
}

func (c *Client) Fetch(who xmpp.JID) error {
	c.Send <- &proto.Command{
		CommandType: proto.MessageType_FETCH.Enum(),
		FetchCommand: &proto.Command_FetchCmd{
			Who: protobuf.String(string(who)),
		},
	}

	_, err := c.recvResponseBytes("fetch")
	if err != nil {
		return err
	}

	return nil
}
