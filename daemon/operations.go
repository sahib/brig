package daemon

import (
	"fmt"

	"github.com/disorganizer/brig/daemon/proto"
	protobuf "github.com/gogo/protobuf/proto"
)

// recvResponse reads one response from the daemon and formats possible errors.
func (c *Client) recvResponse(logname string) (string, error) {
	resp := <-c.Recv
	if resp != nil && !resp.GetSuccess() {
		return "", fmt.Errorf("client: %v: %v", logname, resp.GetError())
	}

	return resp.GetResponse(), nil
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

	return c.recvResponse("add")
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

	return c.recvResponse("cat")
}

// Mount serves a fuse endpoint at the specified path.
func (c *Client) Mount(mountPath string) (string, error) {
	c.Send <- &proto.Command{
		CommandType: proto.MessageType_MOUNT.Enum(),
		MountCommand: &proto.Command_MountCmd{
			MountPoint: protobuf.String(mountPath),
		},
	}

	return c.recvResponse("mount")
}

// Unmount removes a previously mounted fuse endpoint.
func (c *Client) Unmount(mountPath string) (string, error) {
	c.Send <- &proto.Command{
		CommandType: proto.MessageType_UNMOUNT.Enum(),
		UnmountCommand: &proto.Command_UnmountCmd{
			MountPoint: protobuf.String(mountPath),
		},
	}

	return c.recvResponse("unmount")
}

// Rm removes the brig file at `repoPath`
func (c *Client) Rm(repoPath string) (string, error) {
	c.Send <- &proto.Command{
		CommandType: proto.MessageType_RM.Enum(),
		RmCommand: &proto.Command_RmCmd{
			RepoPath: protobuf.String(repoPath),
		},
	}

	return c.recvResponse("rm")
}
