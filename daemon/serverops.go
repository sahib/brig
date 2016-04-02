package daemon

import (
	"bytes"
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/daemon/wire"
	"golang.org/x/net/context"
)

type handlerFunc func(d *Server, ctx context.Context, cmd *wire.Command) ([]byte, error)

var handlerMap = map[wire.MessageType]handlerFunc{
	wire.MessageType_ADD:           handleAdd,
	wire.MessageType_CAT:           handleCat,
	wire.MessageType_PING:          handlePing,
	wire.MessageType_QUIT:          handleQuit,
	wire.MessageType_MOUNT:         handleMount,
	wire.MessageType_UNMOUNT:       handleUnmount,
	wire.MessageType_RM:            handleRm,
	wire.MessageType_MV:            handleMv,
	wire.MessageType_HISTORY:       handleHistory,
	wire.MessageType_LOG:           handleLog,
	wire.MessageType_ONLINE_STATUS: handleOnlineStatus,
	wire.MessageType_FETCH:         handleFetch,
	wire.MessageType_LIST:          handleList,
	wire.MessageType_MKDIR:         handleMkdir,
	wire.MessageType_AUTH_ADD:      handleAuthAdd,
	wire.MessageType_AUTH_PRINT:    handleAuthPrint,
}

func handlePing(d *Server, ctx context.Context, cmd *wire.Command) ([]byte, error) {
	return []byte("PONG"), nil
}

func handleQuit(d *Server, ctx context.Context, cmd *wire.Command) ([]byte, error) {
	d.signals <- os.Interrupt
	return []byte("BYE"), nil
}

func handleAdd(d *Server, ctx context.Context, cmd *wire.Command) ([]byte, error) {
	filePath := cmd.GetAddCommand().GetFilePath()
	repoPath := cmd.GetAddCommand().GetRepoPath()

	err := d.Repo.OwnStore.Add(filePath, repoPath)
	if err != nil {
		return nil, err
	}

	return []byte(repoPath), nil
}

func handleCat(d *Server, ctx context.Context, cmd *wire.Command) ([]byte, error) {
	filePath := cmd.GetCatCommand().GetFilePath()
	fd, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}

	srcPath := cmd.GetCatCommand().GetRepoPath()
	if err := d.Repo.OwnStore.Cat(srcPath, fd); err != nil {
		return nil, err
	}

	return []byte(srcPath), nil
}

func handleMount(d *Server, ctx context.Context, cmd *wire.Command) ([]byte, error) {
	mountPath := cmd.GetMountCommand().GetMountPoint()

	if _, err := d.Mounts.AddMount(mountPath); err != nil {
		log.Errorf("Unable to mount `%v`: %v", mountPath, err)
		return nil, err
	}

	return []byte(mountPath), nil
}

func handleUnmount(d *Server, ctx context.Context, cmd *wire.Command) ([]byte, error) {
	mountPath := cmd.GetUnmountCommand().GetMountPoint()

	if err := d.Mounts.Unmount(mountPath); err != nil {
		log.Errorf("Unable to unmount `%v`: %v", mountPath, err)
		return nil, err
	}

	return []byte(mountPath), nil
}

func handleRm(d *Server, ctx context.Context, cmd *wire.Command) ([]byte, error) {
	rmCmd := cmd.GetRmCommand()
	repoPath := rmCmd.GetRepoPath()
	recursive := rmCmd.GetRecursive()

	if err := d.Repo.OwnStore.Remove(repoPath, recursive); err != nil {
		return nil, err
	}

	return []byte(repoPath), nil
}

func handleMv(d *Server, ctx context.Context, cmd *wire.Command) ([]byte, error) {
	mvCmd := cmd.GetMvCommand()
	if err := d.Repo.OwnStore.Move(mvCmd.GetSource(), mvCmd.GetDest()); err != nil {
		return nil, err
	}

	return nil, nil
}

func handleHistory(d *Server, ctx context.Context, cmd *wire.Command) ([]byte, error) {
	repoPath := cmd.GetHistoryCommand().GetRepoPath()

	history, err := d.Repo.OwnStore.History(repoPath)
	if err != nil {
		return nil, err
	}

	protoData, err := history.Marshal()
	if err != nil {
		return nil, err
	}

	return protoData, err
}

func handleLog(d *Server, ctx context.Context, cmd *wire.Command) ([]byte, error) {
	// TODO: Needs implementation.
	return nil, nil
}

func handleOnlineStatus(d *Server, ctx context.Context, cmd *wire.Command) ([]byte, error) {
	qry := cmd.GetOnlineStatusCommand().GetQuery()
	switch qry {
	case wire.OnlineQuery_IS_ONLINE:
		if d.IsOnline() {
			return []byte("online"), nil
		} else {
			return []byte("offline"), nil
		}
	case wire.OnlineQuery_GO_ONLINE:
		return nil, d.Connect(d.Repo.ID, d.Repo.Password)
	case wire.OnlineQuery_GO_OFFLINE:
		return nil, d.Disconnect()
	}

	return nil, fmt.Errorf("handleOnlineStatus: Bad query received: %v", qry)
}

func handleFetch(d *Server, ctx context.Context, cmd *wire.Command) ([]byte, error) {
	fetchCmd := cmd.GetFetchCommand()
	who, err := id.Cast(fetchCmd.GetWho())
	if err != nil {
		return nil, fmt.Errorf("Bad id `%s`: %v", fetchCmd.Who(), err)
	}

	if !d.MetaHost.IsOnline() {
		return nil, fmt.Errorf("Metadata Host is not online.")
	}

	client, err := d.MetaHost.Talk(who)
	if err != nil {
		return nil, err
	}

	importData, err := client.DoFetch()
	if err != nil {
		return nil, err
	}

	if err := d.Repo.OwnStore.Import(bytes.NewReader(importData)); err != nil {
		return nil, err
	}

	// TODO: what to return on success?
	return []byte("OK"), nil
}

func handleList(d *Server, ctx context.Context, cmd *wire.Command) ([]byte, error) {
	listCmd := cmd.GetListCommand()
	root, depth := listCmd.GetRoot(), listCmd.GetDepth()
	buf := &bytes.Buffer{}

	if err := d.Repo.OwnStore.ListMarshalled(buf, root, int(depth)); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func handleMkdir(d *Server, ctx context.Context, cmd *wire.Command) ([]byte, error) {
	path := cmd.GetMkdirCommand().GetPath()

	if _, err := d.Repo.OwnStore.Mkdir(path); err != nil {
		return nil, err
	}

	return []byte("OK"), nil
}

func handleAuthAdd(d *Server, ctx context.Context, cmd *wire.Command) ([]byte, error) {
	authCmd := cmd.GetAuthAddCommand()
	id, peerHash := authCmd.GetWho(), authCmd.GetPeerHash()

	if err := d.MetaHost.Auth(id, peerHash); err != nil {
		return nil, err
	}

	return []byte("OK"), nil
}

func handleAuthPrint(d *Server, ctx context.Context, cmd *wire.Command) ([]byte, error) {
	peerHash, err := d.Repo.IPFS.Identity()
	if err != nil {
		return nil, err
	}

	return []byte(peerHash), nil
}
