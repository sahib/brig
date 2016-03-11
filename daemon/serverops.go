package daemon

import (
	"bytes"
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/daemon/proto"
	"github.com/tsuibin/goxmpp2/xmpp"
	"golang.org/x/net/context"
)

type handlerFunc func(d *Server, ctx context.Context, cmd *proto.Command) ([]byte, error)

var handlerMap = map[proto.MessageType]handlerFunc{
	proto.MessageType_ADD:           handleAdd,
	proto.MessageType_CAT:           handleCat,
	proto.MessageType_PING:          handlePing,
	proto.MessageType_QUIT:          handleQuit,
	proto.MessageType_MOUNT:         handleMount,
	proto.MessageType_UNMOUNT:       handleUnmount,
	proto.MessageType_RM:            handleRm,
	proto.MessageType_MV:            handleMv,
	proto.MessageType_HISTORY:       handleHistory,
	proto.MessageType_LOG:           handleLog,
	proto.MessageType_ONLINE_STATUS: handleOnlineStatus,
	proto.MessageType_FETCH:         handleFetch,
	proto.MessageType_LIST:          handleList,
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

	err := d.Repo.OwnStore.Add(filePath, repoPath)
	if err != nil {
		return nil, err
	}

	return []byte(repoPath), nil
}

func handleCat(d *Server, ctx context.Context, cmd *proto.Command) ([]byte, error) {
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

	if err := d.Repo.OwnStore.Rm(repoPath); err != nil {
		return nil, err
	}

	return []byte(repoPath), nil
}

func handleMv(d *Server, ctx context.Context, cmd *proto.Command) ([]byte, error) {
	mvCmd := cmd.GetMvCommand()
	if err := d.Repo.OwnStore.Move(mvCmd.GetSource(), mvCmd.GetDest()); err != nil {
		return nil, err
	}

	return nil, nil
}

func handleHistory(d *Server, ctx context.Context, cmd *proto.Command) ([]byte, error) {
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

func handleLog(d *Server, ctx context.Context, cmd *proto.Command) ([]byte, error) {
	// TODO: Needs implementation.
	return nil, nil
}

func handleOnlineStatus(d *Server, ctx context.Context, cmd *proto.Command) ([]byte, error) {
	qry := cmd.GetOnlineStatusCommand().GetQuery()
	switch qry {
	case proto.OnlineQuery_IS_ONLINE:
		if d.IsOnline() {
			return []byte("online"), nil
		} else {
			return []byte("offline"), nil
		}
	case proto.OnlineQuery_GO_ONLINE:
		return nil, d.Connect(xmpp.JID(d.Repo.Jid), d.Repo.Password)
	case proto.OnlineQuery_GO_OFFLINE:
		return nil, d.Disconnect()
	}

	return nil, fmt.Errorf("handleOnlineStatus: Bad query received: %v", qry)
}

func handleFetch(d *Server, ctx context.Context, cmd *proto.Command) ([]byte, error) {
	fetchCmd := cmd.GetFetchCommand()
	who := xmpp.JID(fetchCmd.GetWho())

	if !d.XMPP.IsOnline() {
		return nil, fmt.Errorf("XMPP client is not online.")
	}

	client, err := d.XMPP.Talk(who)
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

func handleList(d *Server, ctx context.Context, cmd *proto.Command) ([]byte, error) {
	listCmd := cmd.GetListCommand()
	root, depth := listCmd.GetRoot(), listCmd.GetDepth()
	buf := &bytes.Buffer{}

	if err := d.Repo.OwnStore.ListMarshalled(buf, root, int(depth)); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
