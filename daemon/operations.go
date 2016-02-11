package daemon

import (
	"encoding/json"
	"fmt"

	"github.com/disorganizer/brig/daemon/proto"
	"github.com/disorganizer/brig/store"
	protobuf "github.com/gogo/protobuf/proto"
)

// recvResponse reads one response from the daemon and formats possible errors.
func (c *Client) recvResponse(logname string) ([]byte, error) {
	resp := <-c.Recv
	if resp != nil && !resp.GetSuccess() {
		return nil, fmt.Errorf("client: %v: %v", logname, resp.GetError())
	}

	return resp.GetResponse(), nil
}

func (c *Client) recvResponseString(logname string) (string, error) {
	resp, err := c.recvResponse(logname)
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

// Log returns a series of commits.
func (c *Client) Log() ([]*store.Commit, error) {
	c.Send <- &proto.Command{CommandType: proto.MessageType_LOG.Enum()}

	// TODO: Implement.
	return nil, nil
}

// Log returns a series of commits.
func (c *Client) History(repoPath string) (store.History, error) {
	c.Send <- &proto.Command{
		CommandType: proto.MessageType_HISTORY.Enum(),
		HistoryCommand: &proto.Command_HistoryCmd{
			RepoPath: protobuf.String(repoPath),
		},
	}

	jsonData, err := c.recvResponse("history")
	if err != nil {
		return nil, err
	}

	hist := make(store.History, 0)
	if err := json.Unmarshal([]byte(jsonData), &hist); err != nil {
		return nil, err
	}

	return hist, nil
}
