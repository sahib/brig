package daemon

import (
	"encoding/json"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/daemon/proto"
	"github.com/disorganizer/brig/store"
	"golang.org/x/net/context"
)

type handlerFunc func(d *Server, ctx context.Context, cmd *proto.Command) (string, error)

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

func handlePing(d *Server, ctx context.Context, cmd *proto.Command) (string, error) {
	return "PONG", nil
}

func handleQuit(d *Server, ctx context.Context, cmd *proto.Command) (string, error) {
	d.signals <- os.Interrupt
	return "BYE", nil
}

func handleAdd(d *Server, ctx context.Context, cmd *proto.Command) (string, error) {
	filePath := cmd.GetAddCommand().GetFilePath()
	repoPath := cmd.GetAddCommand().GetRepoPath()

	err := d.Repo.Store.Add(filePath, repoPath)
	if err != nil {
		return "", err
	}

	return repoPath, nil
}

func handleCat(d *Server, ctx context.Context, cmd *proto.Command) (string, error) {
	filePath := cmd.GetCatCommand().GetFilePath()
	fd, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return "", err
	}

	srcPath := cmd.GetCatCommand().GetRepoPath()
	if err := d.Repo.Store.Cat(srcPath, fd); err != nil {
		return "", err
	}

	return srcPath, nil
}

func handleMount(d *Server, ctx context.Context, cmd *proto.Command) (string, error) {
	mountPath := cmd.GetMountCommand().GetMountPoint()

	if _, err := d.Mounts.AddMount(mountPath); err != nil {
		log.Errorf("Unable to mount `%v`: %v", mountPath, err)
		return "", err
	}

	return mountPath, nil
}

func handleUnmount(d *Server, ctx context.Context, cmd *proto.Command) (string, error) {
	mountPath := cmd.GetUnmountCommand().GetMountPoint()

	if err := d.Mounts.Unmount(mountPath); err != nil {
		log.Errorf("Unable to unmount `%v`: %v", mountPath, err)
		return "", err
	}

	return mountPath, nil
}

func handleRm(d *Server, ctx context.Context, cmd *proto.Command) (string, error) {
	repoPath := cmd.GetRmCommand().GetRepoPath()

	if err := d.Repo.Store.Rm(repoPath); err != nil {
		return "", err
	}

	return repoPath, nil
}

func handleHistory(d *Server, ctx context.Context, cmd *proto.Command) (string, error) {
	repoPath := cmd.GetHistoryCommand().GetRepoPath()

	file := d.Repo.Store.Root.Lookup(repoPath)
	if file == nil {
		return "", store.ErrNoSuchFile
	}

	history, err := d.Repo.Store.History(file)
	if err != nil {
		return "", err
	}

	jsonData, err := json.Marshal(history)
	if err != nil {
		return "", err
	}

	// TODO: Change handlers to return []byte?
	return string(jsonData), err
}

func handleLog(d *Server, ctx context.Context, cmd *proto.Command) (string, error) {
	return "", nil
}
