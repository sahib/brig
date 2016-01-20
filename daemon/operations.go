package daemon

import (
	"fmt"

	"github.com/disorganizer/brig/daemon/proto"
	protobuf "github.com/gogo/protobuf/proto"
)

func (c *Client) recvResponse(logname string) (string, error) {
	resp := <-c.Recv
	if resp != nil && !resp.GetSuccess() {
		return "", fmt.Errorf("client: %v: %v", logname, resp.GetError())
	}

	return resp.GetResponse(), nil
}

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

func (c *Client) Mount(mountPath string) (string, error) {
	c.Send <- &proto.Command{
		CommandType: proto.MessageType_MOUNT.Enum(),
		MountCommand: &proto.Command_MountCmd{
			MountPoint: protobuf.String(mountPath),
		},
	}

	return c.recvResponse("mount")
}

func (c *Client) Unmount(mountPath string) (string, error) {
	c.Send <- &proto.Command{
		CommandType: proto.MessageType_UNMOUNT.Enum(),
		UnmountCommand: &proto.Command_UnmountCmd{
			MountPoint: protobuf.String(mountPath),
		},
	}

	return c.recvResponse("unmount")
}

func (c *Client) Rm(repoPath string) (string, error) {
	c.Send <- &proto.Command{
		CommandType: proto.MessageType_RM.Enum(),
		RmCommand: &proto.Command_RmCmd{
			RepoPath: protobuf.String(repoPath),
		},
	}

	return c.recvResponse("rm")
}
