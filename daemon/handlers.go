package daemon

import (
	"encoding/json"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/daemon/proto"
	"golang.org/x/net/context"
)

type handlerFunc func(d *Server, ctx context.Context, cmd *proto.Command) ([]byte, error)

var handlerMap = map[proto.MessageType]handlerFunc{
	proto.MessageType_ADD:     handleAdd,
	proto.MessageType_CAT:     handleCat,
	proto.MessageType_PING:    handlePing,
	proto.MessageType_QUIT:    handleQuit,
	proto.MessageType_MOUNT:   handleMount,
	proto.MessageType_UNMOUNT: handleUnmount,
	proto.MessageType_RM:      handleRm,
	proto.MessageType_HISTORY: handleHistory,
	proto.MessageType_LOG:     handleLog,
}

func handlePing(d *Server, ctx context.Context, cmd *proto.Command) ([]byte, error) {
	return []byte("PONG"), nil
}

func handleQuit(d *Server, ctx context.Context, cmd *proto.Command) ([]byte, error) {
	d.signals <- os.Interrupt
	return []byte("BYE"), nil
}

func handleAdd(d *Server, ctx context.Context, cmd *proto.Command) ([]byte, error) {
	filePath := cmd.GetAddCommand().GetFilePath()
	repoPath := cmd.GetAddCommand().GetRepoPath()

	err := d.Repo.Store.Add(filePath, repoPath)
	if err != nil {
		return nil, err
	}

	return []byte(repoPath), nil
}

func handleCat(d *Server, ctx context.Context, cmd *proto.Command) ([]byte, error) {
	filePath := cmd.GetCatCommand().GetFilePath()
	fd, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}

	srcPath := cmd.GetCatCommand().GetRepoPath()
	if err := d.Repo.Store.Cat(srcPath, fd); err != nil {
		return nil, err
	}

	return []byte(srcPath), nil
}

func handleMount(d *Server, ctx context.Context, cmd *proto.Command) ([]byte, error) {
	mountPath := cmd.GetMountCommand().GetMountPoint()

	if _, err := d.Mounts.AddMount(mountPath); err != nil {
		log.Errorf("Unable to mount `%v`: %v", mountPath, err)
		return nil, err
	}

	return []byte(mountPath), nil
}

func handleUnmount(d *Server, ctx context.Context, cmd *proto.Command) ([]byte, error) {
	mountPath := cmd.GetUnmountCommand().GetMountPoint()

	if err := d.Mounts.Unmount(mountPath); err != nil {
		log.Errorf("Unable to unmount `%v`: %v", mountPath, err)
		return nil, err
	}

	return []byte(mountPath), nil
}

func handleRm(d *Server, ctx context.Context, cmd *proto.Command) ([]byte, error) {
	repoPath := cmd.GetRmCommand().GetRepoPath()

	if err := d.Repo.Store.Rm(repoPath); err != nil {
		return nil, err
	}

	return []byte(repoPath), nil
}

func handleHistory(d *Server, ctx context.Context, cmd *proto.Command) ([]byte, error) {
	repoPath := cmd.GetHistoryCommand().GetRepoPath()

	history, err := d.Repo.Store.History(repoPath)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(history)
	if err != nil {
		return nil, err
	}

	return jsonData, err
}

func handleLog(d *Server, ctx context.Context, cmd *proto.Command) ([]byte, error) {
	// TODO: Needs implementation.
	return nil, nil
}
